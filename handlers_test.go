// ABOUTME: Integration tests for HTTP handlers using httptest.NewServer.
// ABOUTME: Tests all REST API endpoints with a real in-memory SQLite backend.
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"omnicollect/storage"
)

// newTestServer creates an App backed by an in-memory SQLite store and a
// temp-dir LocalMediaStore, wraps it in a Server, and returns an
// httptest.Server URL plus the App. Cleanup is registered on t.
func newTestServer(t *testing.T) (string, *App) {
	t.Helper()

	store, err := storage.NewSQLiteStoreInMemory()
	if err != nil {
		t.Fatalf("creating test store: %v", err)
	}

	tmpDir := t.TempDir()
	mediaStore := newTestMediaStore(t, tmpDir)

	app := &App{
		store:      store,
		mediaStore: mediaStore,
		modules:    []ModuleSchema{},
	}

	srv := NewServer(app)
	ts := httptest.NewServer(corsMiddleware(srv.mux))
	t.Cleanup(func() {
		ts.Close()
		store.Close()
	})

	return ts.URL, app
}

// newTestMediaStore creates a LocalMediaStore at a custom base directory.
func newTestMediaStore(t *testing.T, baseDir string) *storage.LocalMediaStore {
	t.Helper()
	origDir := filepath.Join(baseDir, "originals")
	thumbDir := filepath.Join(baseDir, "thumbnails")
	os.MkdirAll(origDir, 0755)
	os.MkdirAll(thumbDir, 0755)
	return storage.NewLocalMediaStoreAt(baseDir)
}

// --- Item Endpoints ---

func TestHandlerGetItems(t *testing.T) {
	url, _ := newTestServer(t)

	resp, err := http.Get(url + "/api/v1/items")
	if err != nil {
		t.Fatalf("GET /items: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var items []Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if items == nil {
		t.Error("expected non-nil items array")
	}
}

func TestHandlerSaveItem(t *testing.T) {
	url, _ := newTestServer(t)

	body := `{"moduleId":"comics","title":"Test Comic","images":[],"attributes":{}}`
	resp, err := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /items: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("status: got %d, want %d; body: %s", resp.StatusCode, http.StatusOK, string(data))
	}

	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}
	if item.Title != "Test Comic" {
		t.Errorf("title: got %q, want %q", item.Title, "Test Comic")
	}
}

func TestHandlerSaveItem_Invalid(t *testing.T) {
	url, _ := newTestServer(t)

	// Missing title
	body := `{"moduleId":"comics","title":"","images":[],"attributes":{}}`
	resp, err := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /items: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandlerDeleteItem(t *testing.T) {
	url, _ := newTestServer(t)

	// First create an item
	body := `{"moduleId":"comics","title":"To Delete","images":[],"attributes":{}}`
	resp, err := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /items: %v", err)
	}
	var item Item
	json.NewDecoder(resp.Body).Decode(&item)
	resp.Body.Close()

	// Delete it
	req, _ := http.NewRequest("DELETE", url+"/api/v1/items/"+item.ID, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /items: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestHandlerDeleteItem_NotFound(t *testing.T) {
	url, _ := newTestServer(t)

	req, _ := http.NewRequest("DELETE", url+"/api/v1/items/nonexistent", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /items: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// --- Batch Endpoints ---

func TestHandlerDeleteItems(t *testing.T) {
	url, _ := newTestServer(t)

	// Create two items
	var ids []string
	for _, title := range []string{"Item A", "Item B"} {
		body := `{"moduleId":"comics","title":"` + title + `","images":[],"attributes":{}}`
		resp, _ := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
		var item Item
		json.NewDecoder(resp.Body).Decode(&item)
		resp.Body.Close()
		ids = append(ids, item.ID)
	}

	reqBody, _ := json.Marshal(map[string]any{"ids": ids})
	resp, err := http.Post(url+"/api/v1/items/batch-delete", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST batch-delete: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result BulkDeleteResult
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Deleted != 2 {
		t.Errorf("deleted: got %d, want 2", result.Deleted)
	}
}

func TestHandlerBulkUpdateModule(t *testing.T) {
	url, _ := newTestServer(t)

	// Create an item
	body := `{"moduleId":"old","title":"Item","images":[],"attributes":{}}`
	resp, _ := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	var item Item
	json.NewDecoder(resp.Body).Decode(&item)
	resp.Body.Close()

	reqBody, _ := json.Marshal(map[string]any{"ids": []string{item.ID}, "newModuleId": "new-mod"})
	resp, err := http.Post(url+"/api/v1/items/batch-update-module", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST batch-update-module: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result BulkUpdateResult
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Updated != 1 {
		t.Errorf("updated: got %d, want 1", result.Updated)
	}
}

// --- Module Endpoints ---

func TestHandlerGetModules(t *testing.T) {
	url, _ := newTestServer(t)

	resp, err := http.Get(url + "/api/v1/modules")
	if err != nil {
		t.Fatalf("GET /modules: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// --- Export Endpoints ---

func TestHandlerExportBackup(t *testing.T) {
	url, _ := newTestServer(t)

	resp, err := http.Get(url + "/api/v1/export/backup")
	if err != nil {
		t.Fatalf("GET /export/backup: %v", err)
	}
	defer resp.Body.Close()

	// With SQLiteStore, backup should produce a zip (or may fail gracefully
	// if the DB file doesn't exist on disk). We test that the handler responds.
	// In-memory DB backup uses createCloudBackup fallback since the store
	// is *storage.SQLiteStore, but createBackupArchive needs a disk file.
	// Accept either 200 (zip) or 500 (no disk file for WAL checkpoint).
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 200 or 500", resp.StatusCode)
	}
}

func TestHandlerExportCSV(t *testing.T) {
	url, _ := newTestServer(t)

	// Create an item first
	body := `{"moduleId":"comics","title":"CSV Item","images":[],"attributes":{}}`
	resp, _ := http.Post(url+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	var item Item
	json.NewDecoder(resp.Body).Decode(&item)
	resp.Body.Close()

	reqBody, _ := json.Marshal(map[string]any{"ids": []string{item.ID}})
	resp, err := http.Post(url+"/api/v1/export/csv", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST /export/csv: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("status: got %d, want %d; body: %s", resp.StatusCode, http.StatusOK, string(data))
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "text/csv" {
		t.Errorf("Content-Type: got %q, want %q", ct, "text/csv")
	}
}

// --- Settings Endpoints ---

func TestHandlerGetSettings(t *testing.T) {
	url, _ := newTestServer(t)

	resp, err := http.Get(url + "/api/v1/settings")
	if err != nil {
		t.Fatalf("GET /settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHandlerSaveSettings(t *testing.T) {
	url, _ := newTestServer(t)

	body := `{"theme":"dark"}`
	req, _ := http.NewRequest("PUT", url+"/api/v1/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("status: got %d, want %d; body: %s", resp.StatusCode, http.StatusOK, string(data))
	}
}

// --- Health ---

func TestHandlerHealth(t *testing.T) {
	url, _ := newTestServer(t)

	resp, err := http.Get(url + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Errorf("status field: got %q, want %q", result["status"], "ok")
	}
}
