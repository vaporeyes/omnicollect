// ABOUTME: Adversarial ZIP, legacy compatibility, and retryable media staging regressions.
// ABOUTME: Ensures failed imports cannot commit metadata or overwrite differently named live images.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"omnicollect/storage"
)

type archiveTestEntry struct {
	name string
	data []byte
	mode fs.FileMode
}

func testZIP(t *testing.T, entries []archiveTestEntry) *zip.Reader {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = file.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return zr
}
func emptyPortableEntries() []archiveTestEntry {
	return []archiveTestEntry{{"manifest.json", []byte(`{"format":"omnicollect","version":1}`), 0}, {"items.json", []byte(`[]`), 0}, {"modules.json", []byte(`[]`), 0}, {"settings.json", []byte(`{}`), 0}}
}
func TestArchiveRejectsAmbiguousUnsafeAndMalformedEntries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry archiveTestEntry
	}{
		{"duplicate", archiveTestEntry{"items.json", []byte(`[]`), 0}},
		{"traversal", archiveTestEntry{"media/originals/../outside.jpg", nil, 0}},
		{"absolute", archiveTestEntry{"/items.json", nil, 0}},
		{"nested", archiveTestEntry{"media/originals/sub/image.jpg", nil, 0}},
		{"backslash", archiveTestEntry{"media\\originals\\image.jpg", nil, 0}},
		{"symlink", archiveTestEntry{"media/originals/image.jpg", []byte("/etc/passwd"), fs.ModeSymlink | 0600}},
		{"unknown", archiveTestEntry{"executable", nil, 0}},
		{"mixed formats", archiveTestEntry{"collection.db", nil, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := loadBackup(context.Background(), testZIP(t, append(emptyPortableEntries(), tc.entry)))
			if err == nil {
				t.Fatal("malformed backup accepted")
			}
		})
	}
	for _, tc := range []struct {
		name  string
		index int
		value string
	}{
		{"future version", 0, `{"format":"omnicollect","version":900}`},
		{"wrong format", 0, `{"format":"other","version":1}`},
		{"invalid JSON", 1, `[{`}, {"trailing JSON", 1, `[] {}`}, {"null items", 1, `null`},
		{"invalid metadata", 1, `[{"id":"broken"}]`}, {"null modules", 2, `null`},
		{"invalid module", 2, `[{"id":"../bad"}]`}, {"invalid settings", 3, `[]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries := emptyPortableEntries()
			entries[tc.index].data = []byte(tc.value)
			if _, _, _, err := loadBackup(context.Background(), testZIP(t, entries)); err == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
}
func TestArchiveDirectoryBudgetBeforeZIPAllocation(t *testing.T) {
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	for i := 0; i <= maxArchiveEntries; i++ {
		if _, err := writer.Create(fmt.Sprintf("entry-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "oversized-directory.zip")
	for _, lieAboutCount := range []bool{false, true} {
		if lieAboutCount {
			end := data.Bytes()[data.Len()-22:]
			binary.LittleEndian.PutUint16(end[8:], 1)
			binary.LittleEndian.PutUint16(end[10:], 1)
		}
		if err := os.WriteFile(filename, data.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		if zr, err := openBackupZip(filename); err == nil {
			zr.Close()
			t.Fatal("excessive central directory accepted")
		}
	}
}

func TestArchiveEnforcesBudgetsAndIntegrity(t *testing.T) {
	for _, mutation := range []func(*zip.Reader){
		func(zr *zip.Reader) { zr.File[0].UncompressedSize64 = 1 << 63 },
		func(zr *zip.Reader) { zr.File[0].CompressedSize64 = 1 << 63 },
		func(zr *zip.Reader) { zr.File[0].UncompressedSize64 = 1 },
		func(zr *zip.Reader) { zr.File[0].CRC32 ^= 1 },
		func(zr *zip.Reader) { zr.File = make([]*zip.File, maxArchiveEntries+1) },
	} {
		zr := testZIP(t, emptyPortableEntries())
		mutation(zr)
		if err := validateArchive(context.Background(), zr); err == nil {
			t.Fatal("invalid size/checksum accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := validateArchive(ctx, testZIP(t, emptyPortableEntries())); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
func TestLegacySQLiteBackupWithoutTags(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE items(id TEXT,module_id TEXT,title TEXT,purchase_price REAL,images TEXT,attributes TEXT,created_at TEXT,updated_at TEXT);
 INSERT INTO items VALUES('original-id','books','Old book',NULL,'[]','{}','2020-01-02 03:04:05','2021-01-02 03:04:05');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	zr := testZIP(t, []archiveTestEntry{{"collection.db", data, 0}, {"modules/books.json", []byte(`{"id":"books","displayName":"Books"}`), 0}})
	snapshot, format, warnings, err := loadBackup(context.Background(), zr)
	if err != nil {
		t.Fatal(err)
	}
	if format != "local" || len(warnings) == 0 || len(snapshot.Items) != 1 || len(snapshot.Modules) != 1 {
		t.Fatalf("legacy backup lost: %+v", snapshot)
	}
	if snapshot.Items[0].ID != "original-id" || snapshot.Items[0].CreatedAt != "2020-01-02T03:04:05Z" {
		t.Fatal("legacy history lost")
	}
	if len(snapshot.Items[0].Tags) != 0 {
		t.Fatal("legacy tags not defaulted")
	}
}

type failingThumbnailMedia struct {
	storage.MediaStore
	fail bool
}

func (m *failingThumbnailMedia) SaveThumbnail(ctx context.Context, name string, data []byte) error {
	if m.fail {
		return errors.New("injected thumbnail failure")
	}
	return m.MediaStore.SaveThumbnail(ctx, name, data)
}
func TestMediaStagingFailureKeepsMetadataAndHandleForRetry(t *testing.T) {
	source, _ := seedBackupApp(t)
	target, _ := seedBackupApp(t)
	before, err := target.store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	oldBytes, err := target.mediaStore.GetOriginal(context.Background(), "image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "backup.zip")
	if err := createBackupArchive(context.Background(), filename, source.store, source.mediaStore); err != nil {
		t.Fatal(err)
	}
	media := &failingThumbnailMedia{MediaStore: target.mediaStore, fail: true}
	target.mediaStore = media
	server := NewServer(target)
	id, err := server.imports.register(filename, "local")
	if err != nil {
		t.Fatal(err)
	}
	execute := func() *httptest.ResponseRecorder {
		body, _ := json.Marshal(ImportRequest{TempID: id, Mode: "replace"})
		response := httptest.NewRecorder()
		server.mux.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/import/execute", bytes.NewReader(body)))
		return response
	}
	response := execute()
	if response.Code != 400 || !strings.Contains(response.Body.String(), "collection unchanged") {
		t.Fatalf("failure misreported: %d %s", response.Code, response.Body.String())
	}
	after, err := target.store.Snapshot(context.Background())
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("metadata changed on staging failure", err)
	}
	got, err := target.mediaStore.GetOriginal(context.Background(), "image.jpg")
	if err != nil || !bytes.Equal(oldBytes, got) {
		t.Fatal("live media overwritten")
	}
	media.fail = false
	if response = execute(); response.Code != 200 {
		t.Fatalf("retry failed: %d %s", response.Code, response.Body.String())
	}
	if response = execute(); response.Code != 404 {
		t.Fatalf("successful handle reused: %d", response.Code)
	}
}
func TestMissingOriginalFailsBeforeRestore(t *testing.T) {
	app, item := seedBackupApp(t)
	snapshot, err := app.store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	target := testApp(t)
	_, err = stageBackupMedia(context.Background(), testZIP(t, emptyPortableEntries()), &snapshot, target.mediaStore)
	if err == nil {
		t.Fatal("missing original accepted")
	}
	if snapshot.Items[0].Images[0] != item.Images[0] {
		t.Fatal("failed staging changed references")
	}
}
