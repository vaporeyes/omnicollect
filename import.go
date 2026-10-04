// ABOUTME: Strict, bounded portable and legacy ZIP backup loading and media staging.
// ABOUTME: Stages immutable validated images before the atomic metadata restore commit.
package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"omnicollect/storage"
)

type ImportSummary struct {
	Format      string   `json:"format"`
	ItemCount   int      `json:"itemCount"`
	ImageCount  int      `json:"imageCount"`
	ModuleCount int      `json:"moduleCount"`
	Warnings    []string `json:"warnings"`
	TempID      string   `json:"tempId"`
}
type ImportResult struct {
	ItemsImported   int      `json:"itemsImported"`
	ImagesRestored  int      `json:"imagesRestored"`
	ModulesImported int      `json:"modulesImported"`
	Warnings        []string `json:"warnings"`
}
type ImportRequest struct {
	TempID string `json:"tempId"`
	Mode   string `json:"mode"`
}

// Bound simultaneous archive expansion/parsing independently from image decode.
var archiveWorkers = make(chan struct{}, 1)

func acquireArchiveWorker(ctx context.Context) (func(), error) {
	select {
	case archiveWorkers <- struct{}{}:
		return func() { <-archiveWorkers }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func archiveEntryLimit(name string) (int64, error) {
	switch name {
	case "collection.db":
		return 128 << 20, nil
	case "items.json", "modules.json":
		return maxArchiveMetadataBytes, nil
	case "settings.json":
		return 1 << 20, nil
	case "manifest.json":
		return 4096, nil
	}
	for _, prefix := range []string{"modules/", "media/originals/", "media/thumbnails/"} {
		if strings.HasPrefix(name, prefix) {
			filename := strings.TrimPrefix(name, prefix)
			if err := storage.ValidateFilename(filename); err != nil {
				return 0, err
			}
			switch prefix {
			case "modules/":
				if !strings.HasSuffix(filename, ".json") {
					return 0, fmt.Errorf("invalid module entry")
				}
				return 1 << 20, nil
			case "media/originals/":
				return maxImageFileSize, nil
			default:
				return 2 << 20, nil
			}
		}
	}
	return 0, fmt.Errorf("unexpected backup entry %q", name)
}

// validateArchive checks the directory AND streams every entry through its CRC
// verifier. Directory-declared sizes alone are not trusted for decompression.
func validateArchive(ctx context.Context, zr *zip.Reader) error {
	if len(zr.File) > maxArchiveEntries {
		return fmt.Errorf("backup exceeds entry limit")
	}
	seen := map[string]bool{}
	var expanded, compressed uint64
	for _, file := range zr.File {
		name := file.Name
		if seen[name] {
			return fmt.Errorf("duplicate backup entry %q", name)
		}
		seen[name] = true
		if strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || strings.TrimSuffix(name, "/") != path.Clean(name) {
			return fmt.Errorf("unsafe backup path %q", name)
		}
		if file.FileInfo().IsDir() {
			if name != "modules/" && name != "media/" && name != "media/originals/" && name != "media/thumbnails/" {
				return fmt.Errorf("unexpected backup directory %q", name)
			}
			if file.UncompressedSize64 != 0 || file.CompressedSize64 != 0 {
				return fmt.Errorf("backup directory has content")
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return fmt.Errorf("non-regular backup entry %q", name)
		}
		limit, err := archiveEntryLimit(name)
		if err != nil {
			return err
		}
		if file.UncompressedSize64 > uint64(limit) || file.UncompressedSize64 > uint64(maxArchiveExpandedBytes)-expanded || file.CompressedSize64 > uint64(maxArchiveCompressedBytes)-compressed {
			return fmt.Errorf("backup exceeds expanded, compressed, or per-entry size limit")
		}
		expanded += file.UncompressedSize64
		compressed += file.CompressedSize64
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		limit, _ := archiveEntryLimit(file.Name)
		reader, err := file.Open()
		if err != nil {
			return err
		}
		count, err := io.Copy(io.Discard, io.LimitReader(contextReader{ctx, reader}, limit+1))
		closeErr := reader.Close()
		if err != nil {
			return fmt.Errorf("invalid entry %q: %w", file.Name, err)
		}
		if closeErr != nil {
			return closeErr
		}
		if count > limit || uint64(count) != file.UncompressedSize64 {
			return fmt.Errorf("entry size mismatch for %q", file.Name)
		}
	}
	return nil
}

func findEntry(zr *zip.Reader, name string) *zip.File {
	for _, file := range zr.File {
		if file.Name == name {
			return file
		}
	}
	return nil
}
func readEntry(ctx context.Context, file *zip.File) ([]byte, error) {
	if file == nil {
		return nil, fmt.Errorf("required backup entry is missing")
	}
	limit, err := archiveEntryLimit(file.Name)
	if err != nil {
		return nil, err
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(contextReader{ctx, reader}, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("entry %q exceeds size limit", file.Name)
	}
	return data, nil
}
func decodeEntry(ctx context.Context, zr *zip.Reader, name string, target any) error {
	data, err := readEntry(ctx, findEntry(zr, name))
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func detectBackupFormat(zr *zip.Reader) (string, error) {
	local, cloud := findEntry(zr, "collection.db") != nil, findEntry(zr, "items.json") != nil
	if local == cloud {
		return "", fmt.Errorf("backup must contain exactly one of collection.db or items.json")
	}
	if local {
		return "local", nil
	}
	return "cloud", nil
}

func loadBackup(ctx context.Context, zr *zip.Reader) (storage.Snapshot, string, []string, error) {
	snapshot := storage.Snapshot{Version: storage.SnapshotVersion, Items: []storage.Item{}, Modules: []storage.ModuleSchema{}, Settings: json.RawMessage(`{}`)}
	warnings := []string{}
	if err := validateArchive(ctx, zr); err != nil {
		return snapshot, "", nil, err
	}
	format, err := detectBackupFormat(zr)
	if err != nil {
		return snapshot, "", nil, err
	}
	if findEntry(zr, "manifest.json") != nil {
		var manifest backupManifest
		if err := decodeEntry(ctx, zr, "manifest.json", &manifest); err != nil {
			return snapshot, format, nil, err
		}
		if manifest.Format != "omnicollect" || manifest.Version != storage.SnapshotVersion || format != "cloud" {
			return snapshot, format, nil, fmt.Errorf("unsupported backup format or version")
		}
		if findEntry(zr, "modules.json") == nil || findEntry(zr, "settings.json") == nil {
			return snapshot, format, nil, fmt.Errorf("portable backup is missing required metadata")
		}
	} else {
		warnings = append(warnings, "Legacy backup: settings or media may be absent.")
	}
	if format == "local" {
		if err := readLegacyDatabase(ctx, zr, &snapshot); err != nil {
			return snapshot, format, nil, err
		}
		for _, file := range zr.File {
			if strings.HasPrefix(file.Name, "modules/") && !file.FileInfo().IsDir() {
				var module storage.ModuleSchema
				if err := decodeEntry(ctx, zr, file.Name, &module); err != nil {
					return snapshot, format, nil, err
				}
				snapshot.Modules = append(snapshot.Modules, module)
			}
		}
	} else {
		if err := decodeEntry(ctx, zr, "items.json", &snapshot.Items); err != nil {
			return snapshot, format, nil, err
		}
		if snapshot.Items == nil {
			return snapshot, format, nil, fmt.Errorf("items.json must be an array")
		}
		if findEntry(zr, "modules.json") != nil {
			if err := decodeEntry(ctx, zr, "modules.json", &snapshot.Modules); err != nil {
				return snapshot, format, nil, err
			}
			if snapshot.Modules == nil {
				return snapshot, format, nil, fmt.Errorf("modules.json must be an array")
			}
		}
	}
	if file := findEntry(zr, "settings.json"); file != nil {
		data, err := readEntry(ctx, file)
		if err != nil {
			return snapshot, format, nil, err
		}
		snapshot.Settings = data
	}
	if err := storage.ValidateSnapshot(snapshot); err != nil {
		return snapshot, format, nil, err
	}
	missing := 0
	for name := range snapshotImages(snapshot) {
		if findEntry(zr, "media/originals/"+name) == nil {
			missing++
		}
	}
	if missing > 0 {
		warnings = append(warnings, fmt.Sprintf("%d referenced originals are absent; they must already exist in the destination to restore.", missing))
	}
	return snapshot, format, warnings, nil
}

func openBackupZip(filename string) (*zip.ReadCloser, error) {
	if err := preflightZIP(filename); err != nil {
		return nil, err
	}
	return zip.OpenReader(filename)
}
func analyzeBackupZip(ctx context.Context, filename string) (ImportSummary, error) {
	release, err := acquireArchiveWorker(ctx)
	if err != nil {
		return ImportSummary{}, err
	}
	defer release()
	zr, err := openBackupZip(filename)
	if err != nil {
		return ImportSummary{}, err
	}
	defer zr.Close()
	snapshot, format, warnings, err := loadBackup(ctx, &zr.Reader)
	if err != nil {
		return ImportSummary{}, err
	}
	return ImportSummary{Format: format, ItemCount: len(snapshot.Items), ModuleCount: len(snapshot.Modules), ImageCount: len(snapshotImages(snapshot)), Warnings: warnings}, nil
}
func snapshotImages(snapshot storage.Snapshot) map[string]bool {
	names := map[string]bool{}
	for _, item := range snapshot.Items {
		for _, name := range item.Images {
			names[name] = true
		}
	}
	return names
}

// stageBackupMedia remaps references to validated content-addressed keys, without
// overwriting any differently named live content. Thumbnails are regenerated.
// Failed stages may leave unreferenced immutable files; never delete them here:
// a concurrent upload/restore could already reference the same content.
func stageBackupMedia(ctx context.Context, zr *zip.Reader, snapshot *storage.Snapshot, media storage.MediaStore) (int, error) {
	remap := map[string]string{}
	for old := range snapshotImages(*snapshot) {
		var data []byte
		var err error
		if entry := findEntry(zr, "media/originals/"+old); entry != nil {
			data, err = readEntry(ctx, entry)
		} else {
			data, err = media.GetOriginal(ctx, old)
		}
		if err != nil {
			return 0, fmt.Errorf("missing or unreadable original %q: %w", old, err)
		}
		processed, err := processImageBytes(ctx, data)
		if err != nil {
			return 0, fmt.Errorf("invalid original %q: %w", old, err)
		}
		if _, err := persistProcessedImage(ctx, media, processed); err != nil {
			return 0, err
		}
		remap[old] = processed.Filename
	}
	for i := range snapshot.Items {
		for j, old := range snapshot.Items[i].Images {
			snapshot.Items[i].Images[j] = remap[old]
		}
	}
	return len(remap), nil
}

// readLegacyDatabase opens only the uploaded database, read-only with trusted
// schema disabled. It handles backups from before the tags column existed.
func readLegacyDatabase(ctx context.Context, zr *zip.Reader, snapshot *storage.Snapshot) error {
	data, err := readEntry(ctx, findEntry(zr, "collection.db"))
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "omnicollect-legacy-*.db")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	u := url.URL{Scheme: "file", Path: file.Name(), RawQuery: "mode=ro&immutable=1"}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err = db.ExecContext(ctx, "PRAGMA trusted_schema=OFF"); err != nil {
		return err
	}
	columns, err := db.QueryContext(ctx, "PRAGMA table_info(items)")
	if err != nil {
		return err
	}
	tags := "'[]'"
	for columns.Next() {
		var cid, notnull, pk int
		var name, kind string
		var defaultValue any
		if err := columns.Scan(&cid, &name, &kind, &notnull, &defaultValue, &pk); err != nil {
			columns.Close()
			return err
		}
		if name == "tags" {
			tags = "tags"
		}
	}
	err = columns.Err()
	columns.Close()
	if err != nil {
		return err
	}
	rows, err := db.QueryContext(ctx, "SELECT id,module_id,title,purchase_price,images,"+tags+",attributes,created_at,updated_at FROM items LIMIT 100001")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item storage.Item
		var images, tagData, attributes string
		var price sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.ModuleID, &item.Title, &price, &images, &tagData, &attributes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		if price.Valid {
			item.PurchasePrice = &price.Float64
		}
		for _, pair := range []struct {
			data  string
			value any
		}{{images, &item.Images}, {tagData, &item.Tags}, {attributes, &item.Attributes}} {
			if err := json.Unmarshal([]byte(pair.data), pair.value); err != nil {
				return fmt.Errorf("invalid legacy item JSON: %w", err)
			}
		}
		for _, timestamp := range []*string{&item.CreatedAt, &item.UpdatedAt} {
			if parsed, err := time.Parse("2006-01-02 15:04:05", *timestamp); err == nil {
				*timestamp = parsed.UTC().Format(time.RFC3339Nano)
			}
		}
		snapshot.Items = append(snapshot.Items, item)
	}
	return rows.Err()
}
