// ABOUTME: HTTP server setup with routing, CORS middleware, and static file serving.
// ABOUTME: Wraps the App struct methods as REST endpoints under /api/v1/.
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"omnicollect/storage"
)

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

	// Modules
	s.mux.HandleFunc("GET /api/v1/modules", s.handleGetModules)
	s.mux.HandleFunc("POST /api/v1/modules", s.handleSaveModule)
	s.mux.HandleFunc("GET /api/v1/modules/{id}/file", s.handleLoadModuleFile)

	// Images
	s.mux.HandleFunc("POST /api/v1/images/upload", s.handleUploadImage)

	// Export
	s.mux.HandleFunc("GET /api/v1/export/backup", s.handleExportBackup)
	s.mux.HandleFunc("POST /api/v1/export/csv", s.handleExportCSV)

	// Settings
	s.mux.HandleFunc("GET /api/v1/settings", s.handleGetSettings)
	s.mux.HandleFunc("PUT /api/v1/settings", s.handleSaveSettings)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start begins listening on the given port. Port 0 picks a random available port.
// Returns the listener (to get the actual port) and starts serving in a goroutine.
func (s *Server) Start(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("starting server: %w", err)
	}
	log.Printf("HTTP server listening on %s", ln.Addr().String())
	go http.Serve(ln, corsMiddleware(s.mux))
	return ln, nil
}

// ListenAndServe blocks on the given port (for standalone mode).
func (s *Server) ListenAndServe(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP server listening on %s", addr)
	return http.ListenAndServe(addr, corsMiddleware(s.mux))
}
