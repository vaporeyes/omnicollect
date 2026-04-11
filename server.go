// ABOUTME: HTTP server setup with routing, CORS middleware, and static file serving.
// ABOUTME: Wraps the App struct methods as REST endpoints under /api/v1/.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"omnicollect/auth"
	"omnicollect/showcase"
	"omnicollect/storage"
)

// storeCtxKey is the context key for per-request tenant-scoped stores.
type storeCtxKey struct{}

// requestStore returns the tenant-scoped store for the current request.
// Falls back to s.app.store if no per-request store is set (local mode).
func (s *Server) requestStore(r *http.Request) storage.Store {
	if store, ok := r.Context().Value(storeCtxKey{}).(storage.Store); ok {
		return store
	}
	return s.app.store
}

// Server wraps the App and provides HTTP routing.
type Server struct {
	app *App
	mux *http.ServeMux
}

// NewServer creates a server with all routes registered.
func NewServer(app *App) *Server {
	s := &Server{app: app, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Items
	s.mux.HandleFunc("GET /api/v1/items", s.handleGetItems)
	s.mux.HandleFunc("POST /api/v1/items", s.handleSaveItem)
	s.mux.HandleFunc("DELETE /api/v1/items/{id}", s.handleDeleteItem)
	s.mux.HandleFunc("POST /api/v1/items/batch-delete", s.handleDeleteItems)
	s.mux.HandleFunc("POST /api/v1/items/batch-update-module", s.handleBulkUpdateModule)

	// Tags
	s.mux.HandleFunc("GET /api/v1/tags", s.handleGetAllTags)
	s.mux.HandleFunc("POST /api/v1/tags/rename", s.handleRenameTag)
	s.mux.HandleFunc("DELETE /api/v1/tags/{name}", s.handleDeleteTag)

	// Modules
	s.mux.HandleFunc("GET /api/v1/modules", s.handleGetModules)
	s.mux.HandleFunc("POST /api/v1/modules", s.handleSaveModule)
	s.mux.HandleFunc("GET /api/v1/modules/{id}/file", s.handleLoadModuleFile)

	// Images
	s.mux.HandleFunc("POST /api/v1/images/upload", s.handleUploadImage)

	// Export
	s.mux.HandleFunc("GET /api/v1/export/backup", s.handleExportBackup)
	s.mux.HandleFunc("POST /api/v1/export/csv", s.handleExportCSV)

	// Import
	s.mux.HandleFunc("POST /api/v1/import/analyze", s.handleAnalyzeBackup)
	s.mux.HandleFunc("POST /api/v1/import/execute", s.handleExecuteImport)

	// AI
	s.mux.HandleFunc("POST /api/v1/ai/analyze", s.handleAnalyzeItem)
	s.mux.HandleFunc("GET /api/v1/ai/status", s.handleAIStatus)

	// Settings
	s.mux.HandleFunc("GET /api/v1/settings", s.handleGetSettings)
	s.mux.HandleFunc("PUT /api/v1/settings", s.handleSaveSettings)

	// Showcases (authenticated)
	s.mux.HandleFunc("POST /api/v1/showcases/toggle", s.handleToggleShowcase)
	s.mux.HandleFunc("GET /api/v1/showcases", s.handleListShowcases)

	// Health
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Media file serving: local filesystem or S3 proxy depending on MediaStore type
	if localStore, ok := s.app.mediaStore.(*storage.LocalMediaStore); ok {
		// Local mode: serve directly from filesystem
		mediaBase := localStore.BaseDir()
		s.mux.Handle("/thumbnails/", http.StripPrefix("/thumbnails/",
			http.FileServer(http.Dir(filepath.Join(mediaBase, "thumbnails")))))
		s.mux.Handle("/originals/", http.StripPrefix("/originals/",
			http.FileServer(http.Dir(filepath.Join(mediaBase, "originals")))))
	} else {
		// Cloud mode: proxy from S3
		s.mux.HandleFunc("/thumbnails/", s.handleMediaProxy("/thumbnails/"))
		s.mux.HandleFunc("/originals/", s.handleMediaProxy("/originals/"))
	}
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// tenantScopeMiddleware creates a per-request tenant-scoped store copy
// from the tenant ID in the request context (set by auth middleware).
// This avoids a race condition where concurrent requests for different
// tenants would overwrite each other's search_path on the shared store.
func (s *Server) tenantScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := auth.TenantIDFromContext(r.Context())
		if tenantID != "" {
			if pgStore, ok := s.app.store.(*storage.PostgresStore); ok {
				scoped := pgStore.WithTenantSchema(tenantID)
				ctx := context.WithValue(r.Context(), storeCtxKey{}, storage.Store(scoped))
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// buildHandler constructs the full middleware chain around the mux.
// Auth middleware is applied conditionally based on config.
// The /showcase/ route is public (no auth) and registered outside the auth chain.
func (s *Server) buildHandler() http.Handler {
	cfg := s.app.config

	// Build the provisioner for PostgresStore
	var provisioner auth.TenantProvisioner
	if pgStore, ok := s.app.store.(*storage.PostgresStore); ok {
		provisioner = func(tenantID string) error {
			return pgStore.ProvisionTenant(tenantID)
		}
	}

	// Inner handler: tenant scoping + mux
	inner := s.tenantScopeMiddleware(s.mux)

	var authedHandler http.Handler
	if cfg.IsAuthEnabled() {
		log.Printf("Auth enabled: issuer=%s audience=%s", cfg.AuthIssuer, cfg.AuthAudience)
		jwtMiddleware := auth.NewJWTMiddleware(cfg.AuthIssuer, cfg.AuthAudience, provisioner)
		protected := jwtMiddleware(inner)

		// Exempt health check, media serving, and OPTIONS from auth
		authedHandler = auth.ExemptPaths(
			[]string{"/api/v1/health", "/thumbnails/", "/originals/"},
			protected,
			inner,
		)
	} else {
		log.Printf("Auth disabled: using local tenant %q", cfg.TenantID)
		localMiddleware := auth.NewLocalTenantMiddleware(cfg.TenantID, provisioner)
		authedHandler = localMiddleware(inner)
	}

	// Top-level mux: public showcase route outside auth, everything else through auth
	topMux := http.NewServeMux()
	topMux.HandleFunc("GET /showcase/{slug...}", showcase.HandleShowcase(s.app.store))
	topMux.Handle("/", authedHandler)

	return corsMiddleware(topMux)
}

// Start begins listening on the given port. Port 0 picks a random available port.
// Returns the listener (to get the actual port) and starts serving in a goroutine.
func (s *Server) Start(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("starting server: %w", err)
	}
	log.Printf("HTTP server listening on %s", ln.Addr().String())
	go http.Serve(ln, s.buildHandler())
	return ln, nil
}

// ListenAndServe blocks on the given port (for standalone mode).
func (s *Server) ListenAndServe(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP server listening on %s", addr)
	return http.ListenAndServe(addr, s.buildHandler())
}
