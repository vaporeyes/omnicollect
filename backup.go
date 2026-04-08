// ABOUTME: ZIP archive export for backup of database, media, and module schemas.
// ABOUTME: Uses streaming compression via Go standard library archive/zip.
package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"omnicollect/storage"
	"os"
	"path/filepath"
	"strings"
)

// createBackupArchive generates a ZIP file containing the database,
// media files, and module schemas. Uses streaming writes to handle
// large media directories without buffering in memory.
func createBackupArchive(outputPath string, db *sql.DB) error {
	// Checkpoint WAL to ensure database file is consistent
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("WAL checkpoint: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	// Add database file
	dbPath, err := dbFilePath()
	if err != nil {
		return fmt.Errorf("resolving db path: %w", err)
	}
	if err := addFileToZip(zw, dbPath, "collection.db"); err != nil {
		return fmt.Errorf("adding database: %w", err)
	}

	// Add media directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home dir: %w", err)
	}

	mediaBase := filepath.Join(home, ".omnicollect", "media")
	if err := addDirToZip(zw, mediaBase, "media"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("adding media: %w", err)
	}

	// Add module schemas
	modulesBase, err := modulesDir()
	if err != nil {
		return fmt.Errorf("resolving modules dir: %w", err)
	}
	if err := addDirToZip(zw, modulesBase, "modules"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("adding modules: %w", err)
	}

	return nil
}

// createCloudBackup exports items and modules from a Store into a JSON-based ZIP.
// Used when the backend is PostgreSQL (no local SQLite file to copy).
func createCloudBackup(outputPath string, store storage.Store) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	// Export items as JSON
	items, err := store.QueryItems("", "", "", "")
	if err != nil {
		return fmt.Errorf("querying items: %w", err)
	}
	itemsJSON, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling items: %w", err)
	}
	w, err := zw.Create("items.json")
	if err != nil {
		return err
	}
	w.Write(itemsJSON)

	// Export modules as JSON
	modules, err := store.GetModules()
	if err != nil {
		return fmt.Errorf("querying modules: %w", err)
	}
	modulesJSON, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling modules: %w", err)
	}
	mw, err := zw.Create("modules.json")
	if err != nil {
		return err
	}
	mw.Write(modulesJSON)

	// Export settings
	settings, err := store.GetSettings()
	if err == nil && settings != "" {
		sw, err := zw.Create("settings.json")
		if err != nil {
			return err
		}
		sw.Write([]byte(settings))
	}

	return nil
}

// addFileToZip adds a single file to the ZIP archive at the given archive path.
func addFileToZip(zw *zip.Writer, srcPath string, archivePath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = archivePath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// addDirToZip recursively adds all files in a directory to the ZIP archive.
func addDirToZip(zw *zip.Writer, srcDir string, archivePrefix string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Compute relative path within the archive
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		archivePath := archivePrefix + "/" + strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		return addFileToZip(zw, path, archivePath)
	})
}
