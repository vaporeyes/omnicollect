// ABOUTME: HTTP handler functions wrapping existing App methods for REST endpoints.
// ABOUTME: Each handler parses the request, calls the App method, and writes a JSON response.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"omnicollect/ai"
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

func decodeJSONBody(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("request must contain exactly one JSON value")
	}
	return nil
}

// Items

func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	moduleID := r.URL.Query().Get("moduleId")
	filtersJSON := r.URL.Query().Get("filters")
	tagsJSON := r.URL.Query().Get("tags")
	if err := storage.ValidateQuery(query, moduleID, filtersJSON, tagsJSON); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := s.requestStore(r).QueryItems(query, moduleID, filtersJSON, tagsJSON)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrPaginationRequired) {
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fromStorageItems(items))
}

func (s *Server) handleSaveItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	// Filenames refer only to media readable in this tenant namespace.
	if len(item.Images) > 50 {
		writeError(w, http.StatusBadRequest, "too many images")
		return
	}
	for _, name := range item.Images {
		if err := s.requestMediaStore(r).CheckOriginal(ctx, name); err != nil {
			writeError(w, http.StatusBadRequest, "image is not available in this collection")
			return
		}
	}
	result, err := s.requestStore(r).SaveItem(ctx, toStorageItem(item))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, storage.ErrEditConflict) {
			status = http.StatusConflict
		}
		if errors.Is(err, storage.ErrEditVersionRequired) {
			status = http.StatusPreconditionRequired
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			status = http.StatusRequestTimeout
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fromStorageItem(result))
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "item ID is required")
		return
	}
	batch, err := s.requestStore(r).DeleteWithRecovery([]string{id})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrInvalidDeletion) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, storage.ErrDeleteSelectionMissing) {
			status = http.StatusNotFound
		}
		writeError(w, status, "Deletion was not confirmed; inspect the recovery list before retrying: "+err.Error())
		return
	}
	w.Header().Set("X-Recovery-ID", batch.RecoveryID)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteItems(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(body.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "no item IDs provided")
		return
	}
	batch, err := s.requestStore(r).DeleteWithRecovery(body.IDs)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrInvalidDeletion) || errors.Is(err, storage.ErrDeleteSelectionMissing) {
			status = http.StatusBadRequest
		}
		writeError(w, status, "Deletion was not confirmed; inspect the recovery list before retrying: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Server) handleBulkUpdateModule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs         []string `json:"ids"`
		NewModuleID string   `json:"newModuleId"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(body.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "no item IDs provided")
		return
	}
	if body.NewModuleID == "" {
		writeError(w, http.StatusBadRequest, "new module ID is required")
		return
	}
	updated, err := s.requestStore(r).BulkUpdateModule(body.IDs, body.NewModuleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, BulkUpdateResult{Updated: updated})
}

// Tags

func (s *Server) handleGetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.requestStore(r).GetAllTags()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (s *Server) handleRenameTag(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.OldName == "" || body.NewName == "" {
		writeError(w, http.StatusBadRequest, "oldName and newName are required")
		return
	}
	count, err := s.requestStore(r).RenameTag(body.OldName, body.NewName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"updated": count})
}

func (s *Server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "tag name is required")
		return
	}
	count, err := s.requestStore(r).DeleteTag(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"updated": count})
}

// Modules

func (s *Server) handleGetModules(w http.ResponseWriter, r *http.Request) {
	modules, err := s.requestStore(r).GetModules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toMainModules(modules))
}

func (s *Server) handleSaveModule(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading body: "+err.Error())
		return
	}
	var schema storage.ModuleSchema
	if err := json.Unmarshal(body, &schema); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := storage.ValidateModuleSchema(&schema); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	store := s.requestStore(r)
	if err := store.SaveModule(schema); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fromStorageModule(schema))
}

func (s *Server) handleLoadModuleFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "module ID is required")
		return
	}
	content, err := s.requestStore(r).LoadModuleFile(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

// Images
var imageUploads = make(chan struct{}, 4)
var archiveUploads = make(chan struct{}, 2)

func limitUploads(w http.ResponseWriter, workers chan struct{}) (func(), bool) {
	select {
	case workers <- struct{}{}:
		return func() { <-workers }, true
	default:
		w.Header().Set("Retry-After", "5")
		writeError(w, http.StatusTooManyRequests, "too many uploads; retry shortly")
		return nil, false
	}
}

func (s *Server) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	release, ok := limitUploads(w, imageUploads)
	if !ok {
		return
	}
	defer release()
	r.Body = http.MaxBytesReader(w, r.Body, (30<<20)+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "invalid multipart upload or image exceeds 30 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
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

	data, err := processImageFile(r.Context(), tmp.Name())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := persistProcessedImage(r.Context(), s.requestMediaStore(r), data)
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

	if err := createBackupArchive(r.Context(), tmpPath, s.requestStore(r), s.requestMediaStore(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "creating backup: "+err.Error())
		return
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
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	store := s.requestStore(r)
	modules, err := store.GetModules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading modules: "+err.Error())
		return
	}
	csv, err := store.ExportItemsCSV(body.IDs, modules)
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
	content, err := s.requestStore(r).GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load settings")
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
	if err := s.requestStore(r).SaveSettings(string(body)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Import

func (s *Server) handleAnalyzeBackup(w http.ResponseWriter, r *http.Request) {
	release, ok := limitUploads(w, archiveUploads)
	if !ok {
		return
	}
	defer release()
	r.Body = http.MaxBytesReader(w, r.Body, (256<<20)+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "invalid multipart upload or backup exceeds 256 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, _, err := r.FormFile("backup")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no backup file: "+err.Error())
		return
	}
	defer file.Close()

	// Save to temp file
	tmp, err := os.CreateTemp("", "omnicollect-import-*.zip")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating temp file: "+err.Error())
		return
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, file); err != nil {
		os.Remove(tmp.Name())
		writeError(w, http.StatusInternalServerError, "saving upload: "+err.Error())
		return
	}
	tmp.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	summary, err := analyzeBackupZip(ctx, tmp.Name())
	if err != nil {
		os.Remove(tmp.Name())
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// The client receives an opaque, tenant-owned handle, never a filename.
	summary.TempID, err = s.imports.register(tmp.Name(), importOwner(r))
	if err != nil {
		os.Remove(tmp.Name())
		writeError(w, http.StatusTooManyRequests, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleExecuteImport(w http.ResponseWriter, r *http.Request) {
	var req ImportRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.TempID == "" || (req.Mode != "replace" && req.Mode != "merge") {
		writeError(w, http.StatusBadRequest, "tempId and mode (replace|merge) are required")
		return
	}

	tmpPath, err := s.imports.claim(req.TempID, importOwner(r))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	succeeded := false
	defer func() { s.imports.release(req.TempID, succeeded) }()

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	release, err := acquireArchiveWorker(ctx)
	if err != nil {
		writeError(w, http.StatusRequestTimeout, err.Error())
		return
	}
	defer release()
	zr, err := openBackupZip(tmpPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not open backup")
		return
	}
	defer zr.Close()
	snapshot, _, warnings, err := loadBackup(ctx, &zr.Reader)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	restoredImages, err := stageBackupMedia(ctx, &zr.Reader, &snapshot, s.requestMediaStore(r))
	if err != nil {
		writeError(w, http.StatusBadRequest, "collection unchanged: media staging failed: "+err.Error())
		return
	}
	result, err := s.requestStore(r).Restore(ctx, snapshot, req.Mode)
	if err != nil {
		writeError(w, http.StatusBadRequest, "restore was not confirmed; the collection may have changed. Refresh and inspect it before retrying: "+err.Error())
		return
	}
	response := ImportResult{ItemsImported: result.Items, ModulesImported: result.Modules, ImagesRestored: restoredImages, Warnings: warnings}
	// No mutable App module cache: requests read their own tenant-scoped store.

	succeeded = true
	writeJSON(w, http.StatusOK, response)
}

// AI

func (s *Server) handleAnalyzeItem(w http.ResponseWriter, r *http.Request) {
	if s.app.aiProvider == nil {
		writeError(w, http.StatusNotImplemented, "AI is not configured")
		return
	}

	var body struct {
		ImageFilename string `json:"imageFilename"`
		ModuleID      string `json:"moduleId"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.ImageFilename == "" || body.ModuleID == "" {
		writeError(w, http.StatusBadRequest, "imageFilename and moduleId are required")
		return
	}

	// Read image bytes from MediaStore
	imageBytes, err := s.requestMediaStore(r).GetOriginal(r.Context(), body.ImageFilename)
	if err != nil {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}

	// Look up module schema
	modules, err := s.requestStore(r).GetModules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading modules: "+err.Error())
		return
	}
	var schema *storage.ModuleSchema
	for _, m := range modules {
		if m.ID == body.ModuleID {
			schema = &m
			break
		}
	}
	if schema == nil {
		writeError(w, http.StatusNotFound, "module not found: "+body.ModuleID)
		return
	}

	// Build prompt and call AI
	prompt := ai.BuildPrompt(*schema)
	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)

	rawResponse, err := s.app.aiProvider.AnalyzeImage(r.Context(), imageBase64, http.DetectContentType(imageBytes), prompt)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "AI analysis failed: "+err.Error())
		return
	}

	// Parse and validate AI response against schema
	attributes, title, warnings := ai.ParseAndValidateResponse(rawResponse, *schema)

	result := map[string]any{
		"title":      title,
		"attributes": attributes,
		"warnings":   warnings,
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAIStatus(w http.ResponseWriter, r *http.Request) {
	cfg := s.app.config
	result := map[string]any{
		"enabled":   cfg.IsAIEnabled(),
		"cloudMode": cfg.IsCloudDB(),
	}
	if cfg.IsAIEnabled() {
		result["provider"] = cfg.AIProvider
		result["model"] = cfg.AIModel
	}
	writeJSON(w, http.StatusOK, result)
}

// Showcases

func (s *Server) handleToggleShowcase(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ModuleID string `json:"moduleId"`
		Enabled  bool   `json:"enabled"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.ModuleID == "" {
		writeError(w, http.StatusBadRequest, "moduleId is required")
		return
	}

	// Determine tenant ID for the showcase record
	store := s.requestStore(r)
	tenantID := "default"
	if pgStore, ok := store.(*storage.PostgresStore); ok {
		tenantID = pgStore.TenantSchema()
	}

	// Check if showcase already exists for this module (slug stability)
	existing, err := store.GetShowcaseForModule(body.ModuleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	var sc storage.Showcase
	if existing != nil {
		// Update existing: preserve slug, just toggle enabled
		sc = *existing
		sc.Enabled = body.Enabled
		sc.UpdatedAt = now
	} else {
		// First time: generate slug from module display name
		moduleName := body.ModuleID
		modules, err := store.GetModules()
		if err == nil {
			for _, m := range modules {
				if m.ID == body.ModuleID {
					moduleName = m.DisplayName
					break
				}
			}
		}

		sc = storage.Showcase{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Slug:      storage.GenerateShowcaseSlug(moduleName),
			TenantID:  tenantID,
			ModuleID:  body.ModuleID,
			Enabled:   body.Enabled,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if err := store.UpsertShowcase(sc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return showcase with URL
	result := map[string]any{
		"id":        sc.ID,
		"slug":      sc.Slug,
		"moduleId":  sc.ModuleID,
		"enabled":   sc.Enabled,
		"url":       "/showcase/" + sc.Slug,
		"createdAt": sc.CreatedAt,
		"updatedAt": sc.UpdatedAt,
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleListShowcases(w http.ResponseWriter, r *http.Request) {
	showcases, err := s.requestStore(r).ListShowcases()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, showcases)
}

// Health

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	result := map[string]string{
		"status":   "ok",
		"database": "connected",
		"storage":  "connected",
	}

	// Check database connectivity
	if pgStore, ok := s.requestStore(r).(*storage.PostgresStore); ok {
		if err := pgStore.Ping(r.Context()); err != nil {
			result["status"] = "error"
			result["database"] = "disconnected"
		}
	}

	// Check S3 connectivity
	if s3Store, ok := s.app.mediaStore.(*storage.S3MediaStore); ok {
		if err := s3Store.Ping(r.Context()); err != nil {
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
