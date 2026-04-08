// ABOUTME: SQLite to PostgreSQL data migration tool.
// ABOUTME: Copies items, modules (from disk), and settings into a PostgresStore tenant schema.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// MigrationResult reports counts after a migration run.
type MigrationResult struct {
	ItemsMigrated   int
	ModulesMigrated int
	Errors          []string
}

// MigrateToPostgres reads from a SQLite database file and module JSON files,
// then writes all data into the given PostgresStore's tenant schema.
func MigrateToPostgres(sqlitePath string, pgStore *PostgresStore, modulesDir string) (MigrationResult, error) {
	result := MigrationResult{}

	// Open source SQLite database
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)", sqlitePath)
	srcDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return result, fmt.Errorf("opening source SQLite: %w", err)
	}
	defer srcDB.Close()

	// Migrate items
	itemCount, errs := migrateItems(srcDB, pgStore)
	result.ItemsMigrated = itemCount
	result.Errors = append(result.Errors, errs...)

	// Migrate modules from filesystem
	modCount, errs := migrateModules(pgStore, modulesDir)
	result.ModulesMigrated = modCount
	result.Errors = append(result.Errors, errs...)

	// Migrate settings
	if err := migrateSettings(sqlitePath, pgStore); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("settings: %v", err))
	}

	return result, nil
}

func migrateItems(srcDB *sql.DB, pgStore *PostgresStore) (int, []string) {
	rows, err := srcDB.Query(`SELECT id, module_id, title, purchase_price, images, coalesce(tags, '[]'), attributes, created_at, updated_at FROM items`)
	if err != nil {
		return 0, []string{fmt.Sprintf("querying items: %v", err)}
	}
	defer rows.Close()

	var errors []string
	count := 0

	for rows.Next() {
		var item Item
		var imagesJSON, tagsJSON, attrsJSON string
		err := rows.Scan(
			&item.ID, &item.ModuleID, &item.Title, &item.PurchasePrice,
			&imagesJSON, &tagsJSON, &attrsJSON, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			errors = append(errors, fmt.Sprintf("scanning item: %v", err))
			continue
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

		// Insert directly into PG with existing ID and timestamps
		if err := pgStore.setSearchPath(); err != nil {
			errors = append(errors, fmt.Sprintf("setting search path: %v", err))
			continue
		}

		imgJSON, _ := json.Marshal(item.Images)
		tagJSON, _ := json.Marshal(item.Tags)
		attrJSON, _ := json.Marshal(item.Attributes)

		_, err = pgStore.db.Exec(
			`INSERT INTO items (id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (id) DO NOTHING`,
			item.ID, item.ModuleID, item.Title, item.PurchasePrice,
			string(imgJSON), string(tagJSON), string(attrJSON), item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			errors = append(errors, fmt.Sprintf("inserting item %s: %v", item.ID, err))
			continue
		}
		count++
	}

	return count, errors
}

func migrateModules(pgStore *PostgresStore, modulesDir string) (int, []string) {
	if modulesDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return 0, []string{fmt.Sprintf("resolving home dir: %v", err)}
		}
		modulesDir = filepath.Join(home, ".omnicollect", "modules")
	}

	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, []string{fmt.Sprintf("reading modules dir: %v", err)}
	}

	var errors []string
	count := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(modulesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("reading %s: %v", entry.Name(), err))
			continue
		}

		var schema ModuleSchema
		if err := json.Unmarshal(data, &schema); err != nil {
			errors = append(errors, fmt.Sprintf("parsing %s: %v", entry.Name(), err))
			continue
		}

		if err := pgStore.SaveModule(schema); err != nil {
			errors = append(errors, fmt.Sprintf("saving module %s: %v", schema.ID, err))
			continue
		}

		count++
		log.Printf("migrated module: %s (%s)", schema.ID, schema.DisplayName)
	}

	return count, errors
}

func migrateSettings(sqlitePath string, pgStore *PostgresStore) error {
	// Settings are stored in a JSON file, not in SQLite.
	// Try to read the settings file from the config directory.
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil // No settings to migrate
	}

	settingsPath := filepath.Join(configDir, "OmniCollect", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // No settings file
	}

	return pgStore.SaveSettings(string(data))
}
