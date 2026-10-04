// ABOUTME: Regression tests for portable, complete backups and API restore round trips.
// ABOUTME: Uses isolated stores/media and asserts export never silently succeeds on missing content.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"omnicollect/storage"
)

type brokenArchiveWriter struct{}

func (brokenArchiveWriter) Write([]byte) (int, error) { return 0, errors.New("injected write failure") }

func TestBackupChecksFinalizationErrors(t *testing.T) {
	snapshot := storage.Snapshot{Version: storage.SnapshotVersion, Items: []storage.Item{}, Modules: []storage.ModuleSchema{}, Settings: json.RawMessage(`{}`)}
	if err := writeBackupArchive(context.Background(), brokenArchiveWriter{}, snapshot, nil); err == nil {
		t.Fatal("archive write/finalization failure was ignored")
	}
}

func seedBackupApp(t *testing.T) (*App, storage.Item) {
	t.Helper()
	app := testApp(t)
	if err := app.store.SaveModule(storage.ModuleSchema{ID: "books", DisplayName: "Books"}); err != nil {
		t.Fatal(err)
	}
	if err := app.store.SaveSettings(`{"theme":{"mode":"dark"},"smartFolders":[{"id":"saved"}]}`); err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 16, 8))); err != nil {
		t.Fatal(err)
	}
	image := encoded.Bytes()
	for _, save := range []func(context.Context, string, []byte) error{app.mediaStore.SaveOriginal, app.mediaStore.SaveThumbnail} {
		if err := save(context.Background(), "image.jpg", image); err != nil {
			t.Fatal(err)
		}
	}
	item, err := app.store.InsertItem(storage.Item{ModuleID: "books", Title: "Preserve my history", Images: []string{"image.jpg"}, Attributes: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	return app, item
}

func TestBackupPublishesOnlyCompleteArchives(t *testing.T) {
	app, _ := seedBackupApp(t)
	output := filepath.Join(t.TempDir(), "backup.zip")
	if err := os.WriteFile(output, []byte("previous backup"), 0600); err != nil {
		t.Fatal(err)
	}
	local := app.mediaStore.(*storage.LocalMediaStore)
	if err := os.Remove(filepath.Join(local.BaseDir(), "originals", "image.jpg")); err != nil {
		t.Fatal(err)
	}
	if err := createBackupArchive(context.Background(), output, app.store, app.mediaStore); err == nil {
		t.Fatal("backup accepted missing original")
	}
	data, err := os.ReadFile(output)
	if err != nil || string(data) != "previous backup" {
		t.Fatalf("failed export replaced previous backup: %q %v", data, err)
	}
}

func TestHTTPBackupRestoreRoundTrip(t *testing.T) {
	source, original := seedBackupApp(t)
	server := httptest.NewServer(NewServer(source).buildHandler())
	defer server.Close()
	response, err := http.Get(server.URL + "/api/v1/export/backup")
	if err != nil {
		t.Fatal(err)
	}
	archiveBytes, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("backup failed: %d %s", response.StatusCode, archiveBytes)
	}
	archive, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, entry := range archive.File {
		names[entry.Name] = true
	}
	for _, name := range []string{"manifest.json", "items.json", "modules.json", "settings.json", "media/originals/image.jpg", "media/thumbnails/image.jpg"} {
		if !names[name] {
			t.Errorf("missing backup entry %s", name)
		}
	}
	target := testApp(t)
	targetServer := httptest.NewServer(NewServer(target).buildHandler())
	defer targetServer.Close()
	for _, mode := range []string{"replace", "merge", "merge", "uncertain"} {
		var upload bytes.Buffer
		form := multipart.NewWriter(&upload)
		file, err := form.CreateFormFile("backup", "backup.zip")
		if err != nil {
			t.Fatal(err)
		}
		file.Write(archiveBytes)
		form.Close()
		response, err := http.Post(targetServer.URL+"/api/v1/import/analyze", form.FormDataContentType(), &upload)
		if err != nil {
			t.Fatal(err)
		}
		var summary ImportSummary
		err = json.NewDecoder(response.Body).Decode(&summary)
		response.Body.Close()
		if err != nil || response.StatusCode != 200 {
			t.Fatalf("analyze failed: %d %v", response.StatusCode, err)
		}
		uncertain := mode == "uncertain"
		if uncertain {
			mode = "merge"
			target.store = &lostCommitAckStore{target.store}
		}
		body, _ := json.Marshal(ImportRequest{TempID: summary.TempID, Mode: mode})
		response, err = http.Post(targetServer.URL+"/api/v1/import/execute", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		result, err := io.ReadAll(response.Body)
		response.Body.Close()
		if uncertain {
			if err != nil || response.StatusCode != 400 || !bytes.Contains(result, []byte("may have changed")) || bytes.Contains(result, []byte("unchanged")) {
				t.Fatalf("dishonest outcome: %d %s %v", response.StatusCode, result, err)
			}
		} else if err != nil || response.StatusCode != 200 {
			t.Fatalf("restore failed: %d %s %v", response.StatusCode, result, err)
		}
		snapshot, err := target.store.Snapshot(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(snapshot.Items) != 1 {
			t.Fatalf("reimport duplicated items: %d", len(snapshot.Items))
		}
		got := snapshot.Items[0]
		if got.ID != original.ID || got.CreatedAt != original.CreatedAt || got.UpdatedAt != original.UpdatedAt {
			t.Fatalf("restore lost identity/history: %+v", got)
		}
		if len(snapshot.Modules) != 1 || !bytes.Contains(snapshot.Settings, []byte("smartFolders")) {
			t.Fatalf("metadata lost: %+v", snapshot)
		}
		if image, err := target.mediaStore.GetOriginal(context.Background(), got.Images[0]); err != nil || len(image) == 0 {
			t.Fatalf("media missing: %v", err)
		}
	}
}
