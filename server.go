// ABOUTME: HTTP server setup with routing, CORS middleware, and static file serving.
// ABOUTME: Wraps the App struct methods as REST endpoints under /api/v1/.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"omnicollect/auth"
	"omnicollect/mediahttp"
	"omnicollect/showcase"
	"omnicollect/storage"
)

// storeCtxKey is the context key for per-request tenant-scoped stores.
type storeCtxKey struct{}
type mediaCtxKey struct{}

func (s *Server) requestMediaStore(r *http.Request) storage.MediaStore {
	if media, ok := r.Context().Value(mediaCtxKey{}).(storage.MediaStore); ok {
		return media
	}
	return s.app.mediaStore
}

// requestStore returns the tenant-scoped store for the current request.
// Falls back to s.app.store if no per-request store is set (local mode).
func (s *Server) requestStore(r *http.Request) storage.Store {
	if store, ok := r.Context().Value(storeCtxKey{}).(storage.Store); ok {
		return store.WithContext(r.Context())
	}
	return s.app.store.WithContext(r.Context())
}

// Server wraps the App and provides HTTP routing.
type Server struct {
	app     *App
	mux     *http.ServeMux
	imports importSessions
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
	s.mux.HandleFunc("GET /api/v1/items/page", s.handleItemPage)
	s.mux.HandleFunc("GET /api/v1/items/summary", s.handleCollectionSummary)
	s.mux.HandleFunc("POST /api/v1/items", s.handleSaveItem)
	s.mux.HandleFunc("DELETE /api/v1/items/{id}", s.handleDeleteItem)
	s.mux.HandleFunc("POST /api/v1/items/batch-delete", s.handleDeleteItems)
	s.mux.HandleFunc("GET /api/v1/recovery", s.handleListDeletions)
	s.mux.HandleFunc("POST /api/v1/recovery/{id}", s.handleRecoverDeletion)
	s.mux.HandleFunc("DELETE /api/v1/recovery/{id}", s.handleDiscardDeletion)
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

	s.mux.HandleFunc("GET /originals/{filename}", s.handleMedia("originals"))
	s.mux.HandleFunc("GET /thumbnails/{filename}", s.handleMedia("thumbnails"))
}

// corsMiddleware permits same-origin requests and explicitly configured frontends.
func corsMiddleware(next http.Handler) http.Handler {
	allowed := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := 15 * time.Second
		switch r.URL.Path {
		case "/api/v1/images/upload", "/api/v1/import/analyze", "/api/v1/import/execute", "/api/v1/export/backup":
			timeout = 3 * time.Minute
		case "/api/v1/ai/analyze":
			timeout = 90 * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		origin := r.Header.Get("Origin")
		if origin != "" {
			u, err := url.Parse(origin)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			ok := err == nil && u.Host == r.Host && u.Scheme == scheme
			// Wails serves assets and API routes from its own app origin.
			if err == nil && u.Host == r.Host && (u.Scheme == "wails" || u.Scheme == "wails-asset-server") {
				ok = true
			}
			for _, value := range allowed {
				if origin == strings.TrimSpace(value) {
					ok = true
				}
			}
			if !ok {
				writeError(w, http.StatusForbidden, "origin is not allowed")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, If-Match")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, ETag")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path != "/api/v1/images/upload" && r.URL.Path != "/api/v1/import/analyze" {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
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
			_, cloud := s.app.store.(*storage.PostgresStore)
			if cloud || s.app.config.IsAuthEnabled() {
				media, err := s.app.mediaStore.ForTenant(tenantID)
				if err != nil {
					writeError(w, http.StatusServiceUnavailable, "media namespace unavailable")
					return
				}
				r = r.WithContext(context.WithValue(r.Context(), mediaCtxKey{}, media))
			}
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
		provisioner = func(ctx context.Context, tenantID string) error {
			return pgStore.WithContext(ctx).(*storage.PostgresStore).ProvisionTenant(tenantID)
		}
	}

	// Inner handler: tenant scoping + mux
	inner := s.tenantScopeMiddleware(s.mux)

	var authedHandler http.Handler
	if cfg.IsAuthEnabled() {
		log.Printf("Auth enabled: issuer=%s audience=%s", cfg.AuthIssuer, cfg.AuthAudience)
		jwtMiddleware := auth.NewJWTMiddleware(cfg.AuthIssuer, cfg.AuthAudience, provisioner)
		protected := jwtMiddleware(inner)

		// Only health is public; private media requires the same auth as items.
		authedHandler = auth.ExemptPaths(
			[]string{"/api/v1/health"},
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
	topMux.HandleFunc("GET /showcase/{slug}", showcase.HandleShowcase(s.app.store))
	topMux.HandleFunc("GET /showcase/{slug}/media/{kind}/{filename}", showcase.HandleMedia(s.app.store, s.app.mediaStore))
	topMux.Handle("/api/", authedHandler)
	topMux.Handle("/thumbnails/", authedHandler)
	topMux.Handle("/originals/", authedHandler)
	// The SPA and its login callback must load before a token is available.
	topMux.Handle("/", s.mux)

	return corsMiddleware(topMux)
}

// Start begins listening on the given port. Port 0 picks a random available port.
// Returns the listener (to get the actual port) and starts serving in a goroutine.
func (s *Server) Start(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
	if err != nil {
		return nil, fmt.Errorf("starting server: %w", err)
	}
	log.Printf("HTTP server listening on %s", ln.Addr().String())
	server := s.httpServer("")
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server: %v", err)
		}
	}()
	return ln, nil
}

// ListenAndServe blocks on the given port (for standalone mode).
func (s *Server) ListenAndServe(port int) error {
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, fmt.Sprint(port))
	log.Printf("HTTP server listening on %s", addr)
	return s.httpServer(addr).ListenAndServe()
}

func (s *Server) handleMedia(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediahttp.Serve(w, r, s.requestMediaStore(r), kind, r.PathValue("filename"))
	}
}

func (s *Server) httpServer(addr string) *http.Server {
	return &http.Server{Addr: addr, Handler: s.buildHandler(), ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout: 5 * time.Minute, WriteTimeout: 5 * time.Minute, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
}
