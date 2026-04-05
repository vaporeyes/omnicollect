// ABOUTME: Entry point for the OmniCollect Wails application.
// ABOUTME: Configures window options, binds the App struct, and embeds frontend assets.
package main

import (
	"embed"
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
	app := NewApp()

	err := wails.Run(&options.App{
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
