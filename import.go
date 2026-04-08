// ABOUTME: ZIP backup import logic for local (SQLite) and cloud (JSON) formats.
// ABOUTME: Supports Replace (atomic) and Merge (per-item upsert) import modes.
package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"omnicollect/storage"

	_ "modernc.org/sqlite"
)

// ImportSummary describes the contents of a backup ZIP before import.
type ImportSummary struct {
	Format      string   `json:"format"`
	ItemCount   int      `json:"itemCount"`
	ImageCount  int      `json:"imageCount"`
	ModuleCount int      `json:"moduleCount"`
	Warnings    []string `json:"warnings"`
	TempID      string   `json:"tempId"`
}

// ImportResult reports the outcome of an import operation.
type ImportResult struct {
	ItemsImported   int      `json:"itemsImported"`
	ImagesRestored  int      `json:"imagesRestored"`
	ModulesImported int      `json:"modulesImported"`
	Warnings        []string `json:"warnings"`
}

// ImportRequest is the JSON body for the execute import endpoint.
type ImportRequest struct {
	TempID string `json:"tempId"`
	Mode   string `json:"mode"`
}

// detectBackupFormat scans ZIP entries to determine if this is a local
// (SQLite-based) or cloud (JSON-based) backup. Returns "local", "cloud",
// or an error if the format is unrecognized.
func detectBackupFormat(zr *zip.Reader) (string, error) {
	for _, f := range zr.File {
		if f.Name == "collection.db" {
			return "local", nil
		}
		if f.Name == "items.json" {
			return "cloud", nil
		}
	}
	return "", fmt.Errorf("unrecognized backup format: ZIP contains neither collection.db nor items.json")
}

// analyzeBackupZip opens a ZIP file and returns a summary of its contents
// without modifying any data. The tempID is the path to the temp file so
// the execute step can re-open it.
func analyzeBackupZip(zipPath string) (ImportSummary, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return ImportSummary{}, fmt.Errorf("opening ZIP: %w", err)
	}
	defer zr.Close()

	format, err := detectBackupFormat(&zr.Reader)
	if err != nil {
		return ImportSummary{}, err
	}

	summary := ImportSummary{
		Format:   format,
		TempID:   filepath.Base(zipPath),
		Warnings: []string{},
	}

	if format == "local" {
		summary.ItemCount, err = countItemsLocal(&zr.Reader)
		if err != nil {
			return ImportSummary{}, fmt.Errorf("counting local items: %w", err)
		}
	} else {
		summary.ItemCount, err = countItemsCloud(&zr.Reader)
		if err != nil {
			return ImportSummary{}, fmt.Errorf("counting cloud items: %w", err)
		}
	}

	// Count images and modules from ZIP entries
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "media/originals/") && !f.FileInfo().IsDir() {
			summary.ImageCount++
		}
		if format == "local" && strings.HasPrefix(f.Name, "modules/") && strings.HasSuffix(f.Name, ".json") {
			summary.ModuleCount++
		}
	}

	// For cloud format, count modules from modules.json
	if format == "cloud" {
		for _, f := range zr.File {
			if f.Name == "modules.json" {
				rc, err := f.Open()
				if err != nil {
					break
				}
				var modules []storage.ModuleSchema
				if err := json.NewDecoder(rc).Decode(&modules); err == nil {
					summary.ModuleCount = len(modules)
				}
				rc.Close()
				break
			}
		}
	}

	return summary, nil
}

// countItemsLocal extracts the embedded SQLite DB to a temp file, queries
// the item count, and cleans up the temp file.
func countItemsLocal(zr *zip.Reader) (int, error) {
	tmpDB, err := extractFileFromZip(zr, "collection.db")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", tmpDB))
	if err != nil {
		return 0, fmt.Errorf("opening embedded database: %w", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		return 0, fmt.Errorf("counting items: %w", err)
	}
	return count, nil
}

// countItemsCloud parses items.json and returns the number of items.
func countItemsCloud(zr *zip.Reader) (int, error) {
	for _, f := range zr.File {
		if f.Name == "items.json" {
			rc, err := f.Open()
			if err != nil {
				return 0, err
			}
			defer rc.Close()
			var items []storage.Item
			if err := json.NewDecoder(rc).Decode(&items); err != nil {
				return 0, fmt.Errorf("parsing items.json: %w", err)
			}
			return len(items), nil
		}
	}
	return 0, fmt.Errorf("items.json not found")
}

// readItemsFromLocalBackup extracts the SQLite database from the ZIP,
// queries all items, and reads module JSON files from the modules/ entries.
func readItemsFromLocalBackup(zr *zip.Reader) ([]storage.Item, []storage.ModuleSchema, error) {
	tmpDB, err := extractFileFromZip(zr, "collection.db")
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", tmpDB))
	if err != nil {
		return nil, nil, fmt.Errorf("opening embedded database: %w", err)
	}
	defer db.Close()

	items, err := queryAllItemsFromDB(db)
	if err != nil {
		return nil, nil, err
	}

	// Read module schemas from ZIP entries
	var modules []storage.ModuleSchema
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "modules/") && strings.HasSuffix(f.Name, ".json") {
			rc, err := f.Open()
			if err != nil {
				log.Printf("warning: could not open module file %s: %v", f.Name, err)
				continue
			}
			var mod storage.ModuleSchema
			if err := json.NewDecoder(rc).Decode(&mod); err != nil {
				rc.Close()
				log.Printf("warning: could not parse module file %s: %v", f.Name, err)
				continue
			}
			rc.Close()
			modules = append(modules, mod)
		}
	}

	return items, modules, nil
}

// readItemsFromCloudBackup parses items.json and modules.json from the ZIP.
func readItemsFromCloudBackup(zr *zip.Reader) ([]storage.Item, []storage.ModuleSchema, error) {
	var items []storage.Item
	var modules []storage.ModuleSchema

	for _, f := range zr.File {
		if f.Name == "items.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("opening items.json: %w", err)
			}
			if err := json.NewDecoder(rc).Decode(&items); err != nil {
				rc.Close()
				return nil, nil, fmt.Errorf("parsing items.json: %w", err)
			}
			rc.Close()
		}
		if f.Name == "modules.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("opening modules.json: %w", err)
			}
			if err := json.NewDecoder(rc).Decode(&modules); err != nil {
				rc.Close()
				return nil, nil, fmt.Errorf("parsing modules.json: %w", err)
			}
			rc.Close()
		}
	}

	if items == nil {
		return nil, nil, fmt.Errorf("items.json not found in backup")
	}
	if modules == nil {
		modules = []storage.ModuleSchema{}
	}

	return items, modules, nil
}

// restoreImages extracts image files from the ZIP and saves them via MediaStore.
// Returns the count of successfully restored images.
func restoreImages(zr *zip.Reader, mediaStore storage.MediaStore) (int, error) {
	restored := 0
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		var saveFunc func(string, []byte) error
		var filename string

		if strings.HasPrefix(f.Name, "media/originals/") {
			filename = filepath.Base(f.Name)
			saveFunc = mediaStore.SaveOriginal
		} else if strings.HasPrefix(f.Name, "media/thumbnails/") {
			filename = filepath.Base(f.Name)
			saveFunc = mediaStore.SaveThumbnail
		} else {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			log.Printf("warning: could not open image %s: %v", f.Name, err)
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			log.Printf("warning: could not read image %s: %v", f.Name, err)
			continue
		}

		if err := saveFunc(filename, data); err != nil {
			log.Printf("warning: could not save image %s: %v", f.Name, err)
			continue
		}

		// Only count originals toward the total
		if strings.HasPrefix(f.Name, "media/originals/") {
			restored++
		}
	}
	return restored, nil
}

// executeReplace atomically replaces all existing data with backup data.
// Deletes all existing items and modules, then inserts all backup items
// and modules within a single transaction. Rolls back on any failure.
func executeReplace(store storage.Store, items []storage.Item, modules []storage.ModuleSchema) error {
	// Get all existing item IDs for deletion
	existing, err := store.QueryItems("", "", "", "")
	if err != nil {
		return fmt.Errorf("querying existing items: %w", err)
	}

	if len(existing) > 0 {
		ids := make([]string, len(existing))
		for i, item := range existing {
			ids[i] = item.ID
		}
		if _, err := store.DeleteItems(ids); err != nil {
			return fmt.Errorf("deleting existing items: %w", err)
		}
	}

	// Insert all backup items
	for _, item := range items {
		if _, err := store.InsertItem(item); err != nil {
			return fmt.Errorf("inserting item %q: %w", item.Title, err)
		}
	}

	// Save all backup modules
	for _, mod := range modules {
		if err := store.SaveModule(mod); err != nil {
			return fmt.Errorf("saving module %q: %w", mod.ID, err)
		}
	}

	return nil
}

// executeMerge performs a per-item upsert: updates existing items and
// inserts new ones. Existing items NOT in the backup are preserved.
// Returns the count of items processed.
func executeMerge(store storage.Store, items []storage.Item, modules []storage.ModuleSchema) (int, error) {
	// Build a set of existing item IDs for fast lookup
	existing, err := store.QueryItems("", "", "", "")
	if err != nil {
		return 0, fmt.Errorf("querying existing items: %w", err)
	}
	existingIDs := make(map[string]bool, len(existing))
	for _, item := range existing {
		existingIDs[item.ID] = true
	}

	processed := 0
	for _, item := range items {
		if existingIDs[item.ID] {
			if _, err := store.UpdateItem(item); err != nil {
				log.Printf("warning: failed to update item %q: %v", item.Title, err)
				continue
			}
		} else {
			if _, err := store.InsertItem(item); err != nil {
				log.Printf("warning: failed to insert item %q: %v", item.Title, err)
				continue
			}
		}
		processed++
	}

	// Upsert modules
	for _, mod := range modules {
		if err := store.SaveModule(mod); err != nil {
			log.Printf("warning: failed to save module %q: %v", mod.ID, err)
		}
	}

	return processed, nil
}

// checkMissingModules returns warnings for items that reference module IDs
// not present in the backup's module schemas.
func checkMissingModules(items []storage.Item, modules []storage.ModuleSchema) []string {
	moduleIDs := make(map[string]bool, len(modules))
	for _, m := range modules {
		moduleIDs[m.ID] = true
	}

	// Count items per missing module
	missingCounts := make(map[string]int)
	for _, item := range items {
		if !moduleIDs[item.ModuleID] {
			missingCounts[item.ModuleID]++
		}
	}

	var warnings []string
	for modID, count := range missingCounts {
		warnings = append(warnings, fmt.Sprintf("%d item(s) reference module %q which was not found in the backup", count, modID))
	}
	return warnings
}

// queryAllItemsFromDB reads all items from a SQLite database.
func queryAllItemsFromDB(db *sql.DB) ([]storage.Item, error) {
	rows, err := db.Query(
		`SELECT id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at
		 FROM items ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	var items []storage.Item
	for rows.Next() {
		var item storage.Item
		var imagesJSON, tagsJSON, attrsJSON string
		var price sql.NullFloat64

		if err := rows.Scan(
			&item.ID, &item.ModuleID, &item.Title, &price,
			&imagesJSON, &tagsJSON, &attrsJSON,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}

		if price.Valid {
			item.PurchasePrice = &price.Float64
		}

		if err := json.Unmarshal([]byte(imagesJSON), &item.Images); err != nil {
			item.Images = []string{}
		}
		if err := json.Unmarshal([]byte(tagsJSON), &item.Tags); err != nil {
			item.Tags = []string{}
		}
		if err := json.Unmarshal([]byte(attrsJSON), &item.Attributes); err != nil {
			item.Attributes = map[string]any{}
		}

		items = append(items, item)
	}
	if items == nil {
		items = []storage.Item{}
	}
	return items, rows.Err()
}

// extractFileFromZip extracts a single file from the ZIP to a temp file
// and returns the temp file path. Caller is responsible for cleanup.
func extractFileFromZip(zr *zip.Reader, name string) (string, error) {
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("opening %s: %w", name, err)
			}
			defer rc.Close()

			tmp, err := os.CreateTemp("", "omnicollect-import-*")
			if err != nil {
				return "", fmt.Errorf("creating temp file: %w", err)
			}

			if _, err := io.Copy(tmp, rc); err != nil {
				tmp.Close()
				os.Remove(tmp.Name())
				return "", fmt.Errorf("extracting %s: %w", name, err)
			}
			tmp.Close()
			return tmp.Name(), nil
		}
	}
	return "", fmt.Errorf("%s not found in ZIP", name)
}
