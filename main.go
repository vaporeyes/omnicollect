// ABOUTME: Entry point for OmniCollect. Supports three modes:
// ABOUTME: --serve for standalone HTTP, --migrate for SQLite-to-Postgres, or default Wails desktop.
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

	"omnicollect/storage"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	serve := flag.Bool("serve", false, "Start as standalone HTTP server (no desktop window)")
	port := flag.Int("port", 8080, "HTTP server port (used with --serve)")
	migrate := flag.Bool("migrate", false, "Migrate SQLite database to PostgreSQL and exit")
	source := flag.String("source", "", "Path to SQLite database file (used with --migrate)")
	tenant := flag.String("tenant", "default", "Tenant ID for migration (used with --migrate)")
	modulesPath := flag.String("modules", "", "Path to modules directory (used with --migrate, defaults to ~/.omnicollect/modules)")
	flag.Parse()

	if *migrate {
		runMigration(*source, *tenant, *modulesPath)
		return
	}

	app := NewApp()

	if *serve {
		cfg := LoadConfig()
		cfg.Port = *port
		app.InitWithConfig(cfg)

		srv := NewServer(app)

		// Serve frontend static files at root
		distDir := filepath.Join("frontend", "dist")
		if _, err := os.Stat(distDir); err == nil {
			srv.mux.Handle("/", ServeFrontend(distDir))
		} else {
			log.Printf("warning: frontend/dist not found; run 'cd frontend && npm run build' first")
		}

		log.Fatal(srv.ListenAndServe(cfg.Port))
		return
	}

	// Desktop mode: start embedded HTTP server on random port, then Wails
	app.Init()
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
			Handler: newLocalFileHandler(app),
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
// from the appropriate MediaStore for thumbnail and original image display.
func newLocalFileHandler(app *App) http.Handler {
	mux := http.NewServeMux()

	if localStore, ok := app.mediaStore.(*storage.LocalMediaStore); ok {
		mediaBase := localStore.BaseDir()
		mux.Handle("/thumbnails/", http.StripPrefix("/thumbnails/",
			http.FileServer(http.Dir(filepath.Join(mediaBase, "thumbnails")))))
		mux.Handle("/originals/", http.StripPrefix("/originals/",
			http.FileServer(http.Dir(filepath.Join(mediaBase, "originals")))))
	}

	return mux
}

// runMigration handles the --migrate CLI mode.
func runMigration(sourcePath, tenantID, modulesDir string) {
	if sourcePath == "" {
		// Default SQLite path
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("could not resolve config dir: %v", err)
		}
		sourcePath = filepath.Join(configDir, "OmniCollect", "collection.db")
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Fatalf("source database not found: %s", sourcePath)
	}

	cfg := LoadConfig()
	if !cfg.IsCloudDB() {
		log.Fatal("DATABASE_URL environment variable is required for migration")
	}

	if tenantID != "" {
		cfg.TenantID = tenantID
	}

	pgStore, err := storage.NewPostgresStore(cfg.DatabaseURL, cfg.TenantID)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	log.Printf("migrating from %s to PostgreSQL (tenant: %s)", sourcePath, cfg.TenantID)

	result, err := storage.MigrateToPostgres(sourcePath, pgStore, modulesDir)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("migration complete: %d items, %d modules migrated", result.ItemsMigrated, result.ModulesMigrated)
	if len(result.Errors) > 0 {
		log.Printf("migration had %d errors:", len(result.Errors))
		for _, e := range result.Errors {
			log.Printf("  - %s", e)
		}
	}
}
