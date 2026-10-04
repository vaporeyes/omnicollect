// ABOUTME: Portable ZIP backups from consistent database snapshots and referenced media.
// ABOUTME: Publishes archives atomically and reports write, close, and missing-media failures.
package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"omnicollect/storage"
)

const maxArchiveEntries = 10000
const maxArchiveExpandedBytes int64 = 1 << 30
const maxArchiveCompressedBytes int64 = 256 << 20
const maxArchiveMetadataBytes int64 = 32 << 20

// backupManifest versions the portable JSON format, independent of DB backend.
type backupManifest struct {
	Format  string `json:"format"`
	Version int    `json:"version"`
}

// createBackupArchive reads a single DB snapshot and includes all referenced media.
// It never reads global user directories or copies a live SQLite database file.
func createBackupArchive(ctx context.Context, outputPath string, store storage.Store, media storage.MediaStore) error {
	snapshot, err := store.Snapshot(ctx)
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(outputPath), ".omnicollect-backup-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()
	if err := writeBackupArchive(ctx, temporary, snapshot, media); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	info, err := temporary.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxArchiveCompressedBytes {
		return fmt.Errorf("backup exceeds the 256 MB archive limit")
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), outputPath)
}

func writeBackupArchive(ctx context.Context, destination io.Writer, snapshot storage.Snapshot, media storage.MediaStore) error {
	if err := storage.ValidateSnapshot(snapshot); err != nil {
		return err
	}
	archive := zip.NewWriter(destination)
	defer archive.Close()
	count := 0
	var expanded int64
	writeEntry := func(name string, data []byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		count++
		expanded += int64(len(data))
		if count > maxArchiveEntries || expanded > maxArchiveExpandedBytes {
			return fmt.Errorf("backup exceeds archive entry or expanded-size limit")
		}
		entry, err := archive.Create(name)
		if err != nil {
			return err
		}
		_, err = entry.Write(data)
		return err
	}
	metadata := []struct {
		name  string
		value any
	}{
		{"manifest.json", backupManifest{Format: "omnicollect", Version: storage.SnapshotVersion}},
		{"items.json", snapshot.Items},
		{"modules.json", snapshot.Modules},
		{"settings.json", snapshot.Settings},
	}
	for _, entry := range metadata {
		data, err := json.Marshal(entry.value)
		if err != nil {
			return err
		}
		if int64(len(data)) > maxArchiveMetadataBytes {
			return fmt.Errorf("%s exceeds metadata limit", entry.name)
		}
		if err := writeEntry(entry.name, data); err != nil {
			return err
		}
	}
	filenames := map[string]bool{}
	for _, item := range snapshot.Items {
		for _, name := range item.Images {
			filenames[name] = true
		}
	}
	sorted := make([]string, 0, len(filenames))
	for name := range filenames {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	for _, name := range sorted {
		if err := storage.ValidateFilename(name); err != nil {
			return err
		}
		original, err := media.GetOriginal(ctx, name)
		if err != nil {
			return fmt.Errorf("backup cannot include original %q: %w", name, err)
		}
		if err := writeEntry("media/originals/"+name, original); err != nil {
			return err
		}
		thumbnail, err := media.GetThumbnail(ctx, name)
		if err != nil {
			return fmt.Errorf("backup cannot include thumbnail %q: %w", name, err)
		}
		if err := writeEntry("media/thumbnails/"+name, thumbnail); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return archive.Close()
}
