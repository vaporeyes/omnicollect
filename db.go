// ABOUTME: SQLite database initialization, schema DDL, and FTS5 triggers.
// ABOUTME: Provides item persistence with full-text search via modernc.org/sqlite.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// initDB opens the SQLite database and creates tables if needed.
// The database file lives in the user's config directory.
func initDB() (*sql.DB, error) {
	dbPath, err := dbFilePath()
	if err != nil {
		return nil, fmt.Errorf("resolving db path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return db, nil
}

// dbFilePath returns the path to the SQLite database file.
func dbFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "OmniCollect", "collection.db"), nil
}

// createSchema runs DDL statements for the items table, FTS5 index, and triggers.
func createSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			module_id TEXT NOT NULL,
			title TEXT NOT NULL,
			purchase_price REAL,
			images TEXT NOT NULL DEFAULT '[]',
			attributes TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_id ON items(module_id)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
			title,
			attrs_text,
			content='',
			contentless_delete=1
		)`,
		// Trigger: insert into FTS after item insert
		`CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
			INSERT INTO items_fts(rowid, title, attrs_text)
			VALUES (
				new.rowid,
				new.title,
				(SELECT group_concat(value, ' ')
				 FROM json_each(new.attributes)
				 WHERE type IN ('text','integer','real'))
			);
		END`,
		// Trigger: remove old FTS entry and insert new one after item update
		`CREATE TRIGGER IF NOT EXISTS items_au AFTER UPDATE ON items BEGIN
			INSERT INTO items_fts(items_fts, rowid, title, attrs_text)
			VALUES (
				'delete',
				old.rowid,
				old.title,
				(SELECT group_concat(value, ' ')
				 FROM json_each(old.attributes)
				 WHERE type IN ('text','integer','real'))
			);
			INSERT INTO items_fts(rowid, title, attrs_text)
			VALUES (
				new.rowid,
				new.title,
				(SELECT group_concat(value, ' ')
				 FROM json_each(new.attributes)
				 WHERE type IN ('text','integer','real'))
			);
		END`,
		// Trigger: remove FTS entry after item delete
		`CREATE TRIGGER IF NOT EXISTS items_ad AFTER DELETE ON items BEGIN
			INSERT INTO items_fts(items_fts, rowid, title, attrs_text)
			VALUES (
				'delete',
				old.rowid,
				old.title,
				(SELECT group_concat(value, ' ')
				 FROM json_each(old.attributes)
				 WHERE type IN ('text','integer','real'))
			);
		END`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing DDL: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

// insertItem creates a new item in the database.
func insertItem(db *sql.DB, item Item) (Item, error) {
	item.ID = uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	item.CreatedAt = now
	item.UpdatedAt = now

	imagesJSON, err := json.Marshal(item.Images)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling images: %w", err)
	}

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling attributes: %w", err)
	}

	_, err = db.Exec(
		`INSERT INTO items (id, module_id, title, purchase_price, images, attributes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ModuleID, item.Title, item.PurchasePrice,
		string(imagesJSON), string(attrsJSON), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("inserting item: %w", err)
	}

	return item, nil
}

// updateItem updates an existing item in the database.
func updateItem(db *sql.DB, item Item) (Item, error) {
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	imagesJSON, err := json.Marshal(item.Images)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling images: %w", err)
	}

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling attributes: %w", err)
	}

	result, err := db.Exec(
		`UPDATE items SET module_id=?, title=?, purchase_price=?, images=?, attributes=?, updated_at=?
		 WHERE id=?`,
		item.ModuleID, item.Title, item.PurchasePrice,
		string(imagesJSON), string(attrsJSON), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return Item{}, fmt.Errorf("updating item: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return Item{}, fmt.Errorf("item not found: %s", item.ID)
	}

	return item, nil
}

// queryItems retrieves items with optional FTS search and module filter.
func queryItems(db *sql.DB, query string, moduleID string) ([]Item, error) {
	var rows *sql.Rows
	var err error

	if query != "" {
		// Sanitize for FTS5: wrap in quotes, escape internal quotes, append wildcard
		// for partial matching. Prevents syntax panics from unclosed quotes or
		// reserved keywords (e.g., AND, OR, NOT) in user input.
		safeQuery := "\"" + strings.ReplaceAll(query, "\"", "\"\"") + "\"*"

		// FTS5 search path
		baseSQL := `SELECT i.id, i.module_id, i.title, i.purchase_price, i.images, i.attributes, i.created_at, i.updated_at
			FROM items i
			JOIN items_fts ON items_fts.rowid = i.rowid
			WHERE items_fts MATCH ?`
		if moduleID != "" {
			rows, err = db.Query(baseSQL+" AND i.module_id = ? ORDER BY rank", safeQuery, moduleID)
		} else {
			rows, err = db.Query(baseSQL+" ORDER BY rank", safeQuery)
		}
	} else if moduleID != "" {
		rows, err = db.Query(
			`SELECT id, module_id, title, purchase_price, images, attributes, created_at, updated_at
			 FROM items WHERE module_id = ? ORDER BY updated_at DESC`, moduleID)
	} else {
		rows, err = db.Query(
			`SELECT id, module_id, title, purchase_price, images, attributes, created_at, updated_at
			 FROM items ORDER BY updated_at DESC`)
	}

	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

// scanItems reads item rows from a query result set.
func scanItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var item Item
		var imagesJSON, attrsJSON string
		err := rows.Scan(
			&item.ID, &item.ModuleID, &item.Title, &item.PurchasePrice,
			&imagesJSON, &attrsJSON, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}

		if err := json.Unmarshal([]byte(imagesJSON), &item.Images); err != nil {
			item.Images = []string{}
		}
		if err := json.Unmarshal([]byte(attrsJSON), &item.Attributes); err != nil {
			item.Attributes = map[string]any{}
		}

		items = append(items, item)
	}

	if items == nil {
		items = []Item{}
	}

	return items, rows.Err()
}
