// ABOUTME: Entry point for OmniCollect. Supports two modes:
// ABOUTME: --serve for standalone HTTP server, or default Wails desktop shell with embedded server.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	serve := flag.Bool("serve", false, "Start as standalone HTTP server (no desktop window)")
	port := flag.Int("port", 8080, "HTTP server port (used with --serve)")
	flag.Parse()

	app := NewApp()
	app.Init()

	if *serve {
		// Standalone HTTP server mode
		srv := NewServer(app)

		// Serve frontend static files at root
		distDir := filepath.Join("frontend", "dist")
		if _, err := os.Stat(distDir); err == nil {
			srv.mux.Handle("/", ServeFrontend(distDir))
		} else {
			log.Printf("warning: frontend/dist not found; run 'cd frontend && npm run build' first")
		}

		log.Fatal(srv.ListenAndServe(*port))
	}

	// Desktop mode: start embedded HTTP server on random port, then Wails
	srv := NewServer(app)
	ln, err := srv.Start(0) // port 0 = random available
	if err != nil {
		log.Fatalf("failed to start embedded server: %v", err)
	}
	serverURL := fmt.Sprintf("http://localhost:%d", ln.Addr().(*net.TCPAddr).Port)
	log.Printf("Embedded server at %s", serverURL)

	err = wails.Run(&options.App{
		Title:  "OmniCollect",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: newLocalFileHandler(),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// newLocalFileHandler returns an http.Handler that serves media files
// from ~/.omnicollect/media/ for thumbnail and original image display.
func newLocalFileHandler() http.Handler {
	home, _ := os.UserHomeDir()
	mediaBase := filepath.Join(home, ".omnicollect", "media")

	mux := http.NewServeMux()
	mux.Handle("/thumbnails/", http.StripPrefix("/thumbnails/",
		http.FileServer(http.Dir(filepath.Join(mediaBase, "thumbnails")))))
	mux.Handle("/originals/", http.StripPrefix("/originals/",
		http.FileServer(http.Dir(filepath.Join(mediaBase, "originals")))))
	return mux
}
