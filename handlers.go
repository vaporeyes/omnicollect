// ABOUTME: HTTP handler functions wrapping existing App methods for REST endpoints.
// ABOUTME: Each handler parses the request, calls the App method, and writes a JSON response.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"omnicollect/storage"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Items

func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	moduleID := r.URL.Query().Get("moduleId")
	filtersJSON := r.URL.Query().Get("filters")
	items, err := s.app.GetItems(query, moduleID, filtersJSON)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleSaveItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	saved, err := s.app.SaveItem(item)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.app.DeleteItem(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteItems(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.app.DeleteItems(body.IDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBulkUpdateModule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs         []string `json:"ids"`
		NewModuleID string   `json:"newModuleId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result, err := s.app.BulkUpdateModule(body.IDs, body.NewModuleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Modules

func (s *Server) handleGetModules(w http.ResponseWriter, r *http.Request) {
	modules, err := s.app.GetActiveModules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, modules)
}

func (s *Server) handleSaveModule(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading body: "+err.Error())
		return
	}
	mod, err := s.app.SaveCustomModule(string(body))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mod)
}

func (s *Server) handleLoadModuleFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	content, err := s.app.LoadModuleFile(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

// Images

func (s *Server) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max
	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no image file: "+err.Error())
		return
	}
	defer file.Close()

	// Write to temp file for processing
	tmp, err := os.CreateTemp("", "omnicollect-upload-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating temp file: "+err.Error())
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, file); err != nil {
		writeError(w, http.StatusInternalServerError, "saving upload: "+err.Error())
		return
	}
	tmp.Close()

	result, err := s.app.ProcessImage(tmp.Name())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Export

func (s *Server) handleExportBackup(w http.ResponseWriter, r *http.Request) {
	tmp, err := os.CreateTemp("", "omnicollect-backup-*.zip")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating temp file: "+err.Error())
		return
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	// Use SQLite-native backup if available, otherwise export via Store
	if sqliteStore, ok := s.app.store.(*storage.SQLiteStore); ok {
		if err := createBackupArchive(tmpPath, sqliteStore.DB()); err != nil {
			writeError(w, http.StatusInternalServerError, "creating backup: "+err.Error())
			return
		}
	} else {
		if err := createCloudBackup(tmpPath, s.app.store); err != nil {
			writeError(w, http.StatusInternalServerError, "creating backup: "+err.Error())
			return
		}
	}

	filename := fmt.Sprintf("omnicollect-backup-%s.zip", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	http.ServeFile(w, r, tmpPath)
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	csv, err := s.app.store.ExportItemsCSV(body.IDs, toStorageModules(s.app.modules))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("omnicollect-export-%d-items.csv", len(body.IDs))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(csv))
}

// Settings

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	content, err := s.app.LoadSettings()
	if err != nil {
		// Return empty object if no settings
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading body: "+err.Error())
		return
	}
	if err := s.app.SaveSettings(string(body)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Health

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	result := map[string]string{
		"status":   "ok",
		"database": "connected",
		"storage":  "connected",
	}

	// Check database connectivity
	if pgStore, ok := s.app.store.(*storage.PostgresStore); ok {
		if err := pgStore.Ping(context.Background()); err != nil {
			result["status"] = "error"
			result["database"] = "disconnected"
		}
	}

	// Check S3 connectivity
	if s3Store, ok := s.app.mediaStore.(*storage.S3MediaStore); ok {
		if err := s3Store.Ping(context.Background()); err != nil {
			result["status"] = "error"
			result["storage"] = "disconnected"
		}
	}

	status := http.StatusOK
	if result["status"] == "error" {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, result)
}

// ServeFrontend serves the built Vue frontend for non-API routes.
func ServeFrontend(distDir string) http.Handler {
	fs := http.FileServer(http.Dir(distDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try the exact file first
		path := filepath.Join(distDir, r.URL.Path)
		if _, err := os.Stat(path); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html for client-side routing
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
	})
}

// handleMediaProxy proxies media requests to S3 via the S3MediaStore.
func (s *Server) handleMediaProxy(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path[len(prefix):]
		if filename == "" {
			http.NotFound(w, r)
			return
		}

		s3Store, ok := s.app.mediaStore.(*storage.S3MediaStore)
		if !ok {
			http.NotFound(w, r)
			return
		}

		var data []byte
		var err error
		if prefix == "/originals/" {
			data, err = s3Store.GetOriginal(context.Background(), filename)
		} else {
			data, err = s3Store.GetThumbnail(context.Background(), filename)
		}
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Detect content type from extension
		ct := "application/octet-stream"
		ext := filepath.Ext(filename)
		switch ext {
		case ".jpg", ".jpeg":
			ct = "image/jpeg"
		case ".png":
			ct = "image/png"
		case ".gif":
			ct = "image/gif"
		case ".webp":
			ct = "image/webp"
		}

		w.Header().Set("Content-Type", ct)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(data)
	}
}
