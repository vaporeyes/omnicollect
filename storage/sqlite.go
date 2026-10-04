// ABOUTME: SQLiteStore implements the Store interface using local SQLite database.
// ABOUTME: Extracted from the original db.go with FTS5 search, json_extract filters, and batch ops.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using a local SQLite database with FTS5.
type SQLiteStore struct {
	db         *sql.DB
	requestCtx context.Context
}

// NewSQLiteStore opens the local database. Modules and settings live in the
// database so backups and restores can include them in a single transaction.
func NewSQLiteStore() (*SQLiteStore, error) {
	path, err := sqliteDBPath()
	if err != nil {
		return nil, err
	}
	return NewSQLiteStoreAt(path)
}

// NewSQLiteStoreAt opens a database at an explicit path, useful for isolated tests.
func NewSQLiteStoreAt(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	location := url.URL{Scheme: "file", Path: path}
	dsn := location.String() + "?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	return openSQLite(dsn)
}

// NewSQLiteStoreInMemory creates a fully isolated, filesystem-free test store.
func NewSQLiteStoreInMemory() (*SQLiteStore, error) {
	return openSQLite(":memory:")
}

func openSQLite(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite has one writer. A single connection also keeps :memory: databases
	// coherent and serializes read/modify/write settings transactions.
	db.SetMaxOpenConns(1)
	if err := createSQLiteSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

// DB returns the underlying *sql.DB for backup and other raw operations.
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

func sqliteDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "OmniCollect", "collection.db"), nil
}

func createSQLiteSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS modules (
            id TEXT PRIMARY KEY,
            display_name TEXT NOT NULL,
            description TEXT NOT NULL DEFAULT '',
            schema_json TEXT NOT NULL CHECK(json_valid(schema_json))
        )`,
		`CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL CHECK(json_valid(value))
        )`,
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			module_id TEXT NOT NULL,
			title TEXT NOT NULL,
			purchase_price REAL,
			images TEXT NOT NULL DEFAULT '[]',
			tags TEXT NOT NULL DEFAULT '[]',
			attributes TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS deletion_batches (id TEXT PRIMARY KEY, payload TEXT NOT NULL, titles TEXT NOT NULL, item_count INTEGER NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_id ON items(module_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_updated_id ON items(module_id,updated_at DESC,id ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_items_updated_id ON items(updated_at DESC, id ASC)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
			title,
			attrs_text,
			content='',
			contentless_delete=1
		)`,
		// Trigger: insert into FTS after item insert.
		// Concatenates attribute values and tag values into attrs_text for search.
		`CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
			INSERT INTO items_fts(rowid, title, attrs_text)
			VALUES (
				new.rowid,
				new.title,
				coalesce((SELECT group_concat(value, ' ')
				 FROM json_each(new.attributes)
				 WHERE type IN ('text','integer','real')), '')
				|| ' ' ||
				coalesce((SELECT group_concat(value, ' ')
				 FROM json_each(new.tags)), '')
			);
		END`,
		// Trigger: remove old FTS entry and insert new one after item update.
		// Uses DELETE FROM (not the 'delete' command) for contentless_delete=1 tables.
		`CREATE TRIGGER IF NOT EXISTS items_au AFTER UPDATE ON items BEGIN
			DELETE FROM items_fts WHERE rowid = old.rowid;
			INSERT INTO items_fts(rowid, title, attrs_text)
			VALUES (
				new.rowid,
				new.title,
				coalesce((SELECT group_concat(value, ' ')
				 FROM json_each(new.attributes)
				 WHERE type IN ('text','integer','real')), '')
				|| ' ' ||
				coalesce((SELECT group_concat(value, ' ')
				 FROM json_each(new.tags)), '')
			);
		END`,
		// Trigger: remove FTS entry after item delete.
		// Uses DELETE FROM (not the 'delete' command) for contentless_delete=1 tables.
		`CREATE TRIGGER IF NOT EXISTS items_ad AFTER DELETE ON items BEGIN
			DELETE FROM items_fts WHERE rowid = old.rowid;
		END`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing DDL: %w\nStatement: %s", err, stmt)
		}
	}

	// Showcases table for public gallery links
	showcaseStatements := []string{
		`CREATE TABLE IF NOT EXISTS showcases (
			id TEXT PRIMARY KEY,
			slug TEXT NOT NULL UNIQUE,
			tenant_id TEXT NOT NULL,
			module_id TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_showcases_slug ON showcases(slug)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_showcases_tenant_module ON showcases(tenant_id, module_id)`,
	}
	for _, stmt := range showcaseStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing showcase DDL: %w\nStatement: %s", err, stmt)
		}
	}

	// Migration: add tags column to existing databases that lack it
	migrateSQLiteAddTagsColumn(db)

	return nil
}

// migrateSQLiteAddTagsColumn adds the tags column if it doesn't exist.
// Silently ignores errors (column already exists).
func migrateSQLiteAddTagsColumn(db *sql.DB) {
	_, _ = db.Exec(`ALTER TABLE items ADD COLUMN tags TEXT NOT NULL DEFAULT '[]'`)
	// Recreate FTS triggers to include tags in indexed text.
	// DROP + CREATE is idempotent since the main DDL uses IF NOT EXISTS
	// but triggers need to be updated to include the tags column.
	db.Exec(`DROP TRIGGER IF EXISTS items_ai`)
	db.Exec(`DROP TRIGGER IF EXISTS items_au`)
	db.Exec(`CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
		INSERT INTO items_fts(rowid, title, attrs_text)
		VALUES (
			new.rowid,
			new.title,
			coalesce((SELECT group_concat(value, ' ')
			 FROM json_each(new.attributes)
			 WHERE type IN ('text','integer','real')), '')
			|| ' ' ||
			coalesce((SELECT group_concat(value, ' ')
			 FROM json_each(new.tags)), '')
		);
	END`)
	db.Exec(`CREATE TRIGGER IF NOT EXISTS items_au AFTER UPDATE ON items BEGIN
		DELETE FROM items_fts WHERE rowid = old.rowid;
		INSERT INTO items_fts(rowid, title, attrs_text)
		VALUES (
			new.rowid,
			new.title,
			coalesce((SELECT group_concat(value, ' ')
			 FROM json_each(new.attributes)
			 WHERE type IN ('text','integer','real')), '')
			|| ' ' ||
			coalesce((SELECT group_concat(value, ' ')
			 FROM json_each(new.tags)), '')
		);
	END`)
}

// QueryItems retrieves items with optional FTS search, module filter,
// attribute filters, and tag filters.
func (s *SQLiteStore) QueryItems(query string, moduleID string, filtersJSON string, tagsJSON string) ([]Item, error) {
	return readLegacyItems(s.queryItems, query, moduleID, filtersJSON, tagsJSON)
}

func (s *SQLiteStore) queryItems(query string, moduleID string, filtersJSON string, tagsJSON string, limit, offset int) ([]Item, error) {
	if err := ValidateQuery(query, moduleID, filtersJSON, tagsJSON); err != nil {
		return nil, err
	}
	filters, err := parseFilters(filtersJSON)
	if err != nil {
		return nil, err
	}

	tagClause, tagArgs := buildSQLiteTagClause(tagsJSON, "i")
	tagClauseNoAlias, tagArgsNoAlias := buildSQLiteTagClause(tagsJSON, "")

	var rows *sql.Rows

	if query != "" {
		// Sanitize for FTS5: wrap in quotes, escape internal quotes, append wildcard
		// for partial matching. Prevents syntax panics from unclosed quotes or
		// reserved keywords (e.g., AND, OR, NOT) in user input.
		safeQuery := "\"" + strings.ReplaceAll(query, "\"", "\"\"") + "\"*"

		// FTS5 search path
		baseSQL := `SELECT i.id, i.module_id, i.title, i.purchase_price, i.images, i.tags, i.attributes, i.created_at, i.updated_at
			FROM items i
			JOIN items_fts ON items_fts.rowid = i.rowid
			WHERE items_fts MATCH ?`
		queryArgs := []any{safeQuery}

		if moduleID != "" {
			baseSQL += " AND i.module_id = ?"
			queryArgs = append(queryArgs, moduleID)
		}

		filterClauses, filterArgs := buildSQLiteFilterClauses(filters, "i")
		for _, c := range filterClauses {
			baseSQL += " AND " + c
		}
		queryArgs = append(queryArgs, filterArgs...)

		if tagClause != "" {
			baseSQL += " AND " + tagClause
			queryArgs = append(queryArgs, tagArgs...)
		}

		baseSQL += " ORDER BY rank, i.id ASC"
		baseSQL, queryArgs = appendItemPage(baseSQL, queryArgs, limit, offset, false)

		rows, err = s.db.QueryContext(s.operationContext(), baseSQL, queryArgs...)
	} else {
		baseSQL := `SELECT id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at
			FROM items`
		var queryArgs []any
		var whereParts []string

		if moduleID != "" {
			whereParts = append(whereParts, "module_id = ?")
			queryArgs = append(queryArgs, moduleID)
		}

		filterClauses, filterArgs := buildSQLiteFilterClauses(filters, "")
		whereParts = append(whereParts, filterClauses...)
		queryArgs = append(queryArgs, filterArgs...)

		if tagClauseNoAlias != "" {
			whereParts = append(whereParts, tagClauseNoAlias)
			queryArgs = append(queryArgs, tagArgsNoAlias...)
		}

		if len(whereParts) > 0 {
			baseSQL += " WHERE " + strings.Join(whereParts, " AND ")
		}
		baseSQL += " ORDER BY updated_at DESC, id ASC"
		baseSQL, queryArgs = appendItemPage(baseSQL, queryArgs, limit, offset, false)

		rows, err = s.db.QueryContext(s.operationContext(), baseSQL, queryArgs...)
	}

	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

// InsertItem creates a new item in the database.
func (s *SQLiteStore) InsertItem(item Item) (Item, error) {
	item.ID = uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	item.CreatedAt = now
	item.UpdatedAt = now

	item.Tags = normalizeTags(item.Tags)

	imagesJSON, err := json.Marshal(item.Images)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling images: %w", err)
	}

	tagsJSON, err := json.Marshal(item.Tags)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling tags: %w", err)
	}

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling attributes: %w", err)
	}

	_, err = s.db.ExecContext(s.operationContext(),
		`INSERT INTO items (id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ModuleID, item.Title, item.PurchasePrice,
		string(imagesJSON), string(tagsJSON), string(attrsJSON), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("inserting item: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item in the database.
func (s *SQLiteStore) UpdateItem(item Item) (Item, error) {
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	item.Tags = normalizeTags(item.Tags)

	imagesJSON, err := json.Marshal(item.Images)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling images: %w", err)
	}

	tagsJSON, err := json.Marshal(item.Tags)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling tags: %w", err)
	}

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling attributes: %w", err)
	}

	result, err := s.db.ExecContext(s.operationContext(),
		`UPDATE items SET module_id=?, title=?, purchase_price=?, images=?, tags=?, attributes=?, updated_at=?
		 WHERE id=?`,
		item.ModuleID, item.Title, item.PurchasePrice,
		string(imagesJSON), string(tagsJSON), string(attrsJSON), item.UpdatedAt, item.ID,
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

// DeleteItem removes an item from the database by ID.
func (s *SQLiteStore) DeleteItem(id string) error {
	if err := validateIDs([]string{id}); err != nil {
		return err
	}
	result, err := s.db.ExecContext(s.operationContext(), `DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found: %s", id)
	}
	return nil
}

// DeleteItems removes multiple items in a single atomic transaction.
func (s *SQLiteStore) DeleteItems(ids []string) (int64, error) {
	if err := validateIDs(ids); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := s.db.BeginTx(s.operationContext(), nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	result, err := tx.ExecContext(s.operationContext(),
		"DELETE FROM items WHERE id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}

	deleted, _ := result.RowsAffected()
	return deleted, nil
}

// BulkUpdateModule changes the module_id of multiple items in one transaction.
func (s *SQLiteStore) BulkUpdateModule(ids []string, newModuleID string) (int64, error) {
	if err := validateIDs(ids); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := s.db.BeginTx(s.operationContext(), nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReassignment(s.operationContext(), tx, func(name string) string { return name }, false, ids, newModuleID); err != nil {
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	placeholders := make([]string, len(ids))
	args := []any{newModuleID, now}
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	result, err := tx.ExecContext(s.operationContext(),
		"UPDATE items SET module_id = ?, updated_at = ? WHERE id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("updating items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}

	updated, _ := result.RowsAffected()
	return updated, nil
}

// ExportItemsCSV queries items by ID and generates a CSV string.
func (s *SQLiteStore) ExportItemsCSV(ids []string, modules []ModuleSchema) (string, error) {
	return exportItemsCSV(s.operationContext(), s.db, "items", false, ids, modules)
}

// GetModules returns all module schemas from SQLite.
func (s *SQLiteStore) GetModules() ([]ModuleSchema, error) {
	return readModules(s.operationContext(), s.db, "modules", false)
}

// SaveModule validates and upserts a module in the same database as its items.
func (s *SQLiteStore) SaveModule(schema ModuleSchema) error {

	return saveModule(s.operationContext(), s.db, func(name string) string { return name }, false, schema)
}

// LoadModuleFile returns JSON for editing or exporting a module.
func (s *SQLiteStore) LoadModuleFile(id string) (string, error) {
	var data string
	err := s.db.QueryRowContext(s.operationContext(), "SELECT schema_json FROM modules WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("module not found: %s", id)
	}
	return data, err
}

// GetSettings returns the top-level settings object.
func (s *SQLiteStore) GetSettings() (string, error) {
	var data string
	err := s.db.QueryRowContext(s.operationContext(), "SELECT value FROM settings WHERE key = 'settings'").Scan(&data)
	if err == sql.ErrNoRows {
		return "{}", nil
	}
	return data, err
}

// SaveSettings atomically replaces supplied top-level keys and preserves others.
func (s *SQLiteStore) SaveSettings(settingsJSON string) error {
	// json_patch would delete keys whose value is null; preserve explicit nulls
	// consistently with PostgreSQL by merging within a database transaction.
	if _, err := settingsObject(settingsJSON); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(s.operationContext(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current string
	err = tx.QueryRowContext(s.operationContext(), "SELECT value FROM settings WHERE key = 'settings'").Scan(&current)
	if err == sql.ErrNoRows {
		current = "{}"
	} else if err != nil {
		return err
	}
	merged, err := mergeSettings(current, settingsJSON)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(s.operationContext(), `INSERT INTO settings (key, value) VALUES ('settings', ?)
 ON CONFLICT(key) DO UPDATE SET value=excluded.value`, merged); err != nil {
		return err
	}
	return tx.Commit()
}

// GetShowcaseBySlug looks up a showcase by its URL slug (cross-tenant).
func (s *SQLiteStore) GetShowcaseBySlug(slug string) (*Showcase, error) {
	var sc Showcase
	var enabled int
	err := s.db.QueryRowContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM showcases WHERE slug = ?`,
		slug,
	).Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &enabled, &sc.CreatedAt, &sc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying showcase by slug: %w", err)
	}
	sc.Enabled = enabled != 0
	return &sc, nil
}

// GetShowcaseForModule returns the showcase for a module within the default tenant.
func (s *SQLiteStore) GetShowcaseForModule(moduleID string) (*Showcase, error) {
	var sc Showcase
	var enabled int
	err := s.db.QueryRowContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM showcases WHERE module_id = ?`,
		moduleID,
	).Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &enabled, &sc.CreatedAt, &sc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying showcase for module: %w", err)
	}
	sc.Enabled = enabled != 0
	return &sc, nil
}

// UpsertShowcase creates or updates a showcase record. Slug is never overwritten on update.
func (s *SQLiteStore) UpsertShowcase(showcase Showcase) error {
	enabledInt := 0
	if showcase.Enabled {
		enabledInt = 1
	}
	_, err := s.db.ExecContext(s.operationContext(),
		`INSERT INTO showcases (id, slug, tenant_id, module_id, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(tenant_id, module_id) DO UPDATE SET enabled = excluded.enabled, updated_at = excluded.updated_at`,
		showcase.ID, showcase.Slug, showcase.TenantID, showcase.ModuleID, enabledInt, showcase.CreatedAt, showcase.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting showcase: %w", err)
	}
	return nil
}

// ListShowcases returns all showcases (single-tenant in SQLite mode).
func (s *SQLiteStore) ListShowcases() ([]Showcase, error) {
	rows, err := s.db.QueryContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM showcases ORDER BY created_at DESC,id LIMIT ?`, MaxShowcaseResults+1,
	)
	if err != nil {
		return nil, fmt.Errorf("querying showcases: %w", err)
	}
	defer rows.Close()

	var showcases []Showcase
	for rows.Next() {
		var sc Showcase
		var enabled int
		if err := rows.Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &enabled, &sc.CreatedAt, &sc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning showcase: %w", err)
		}
		sc.Enabled = enabled != 0
		if len(showcases) >= MaxShowcaseResults {
			return nil, ErrMetadataBudget
		}
		showcases = append(showcases, sc)
	}
	if showcases == nil {
		showcases = []Showcase{}
	}
	return showcases, rows.Err()
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	if s.requestCtx != nil {
		return nil
	} // Borrowed request handles do not own the pool.
	return s.db.Close()
}

// GetAllTags returns all distinct tags with item counts.
func (s *SQLiteStore) GetAllTags() ([]TagCount, error) {
	rows, err := s.db.QueryContext(s.operationContext(),
		`SELECT value, COUNT(*) FROM items, json_each(items.tags) GROUP BY value ORDER BY value LIMIT 10001`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()

	var tags []TagCount
	for rows.Next() {
		var tc TagCount
		if err := rows.Scan(&tc.Name, &tc.Count); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		if len(tags) >= MaxTagResults {
			return nil, ErrMetadataBudget
		}
		tags = append(tags, tc)
	}
	if tags == nil {
		tags = []TagCount{}
	}
	return tags, rows.Err()
}

// RenameTag renames a tag across all items that contain it.
func (s *SQLiteStore) RenameTag(oldName, newName string) (int64, error) {
	newName = strings.ToLower(strings.TrimSpace(newName))
	if newName == "" {
		return 0, fmt.Errorf("new tag name cannot be empty")
	}
	if len(newName) > 50 || strings.TrimSpace(oldName) == "" || len(oldName) > 50 {
		return 0, fmt.Errorf("tag names must contain 1-50 bytes")
	}

	result, err := s.db.ExecContext(s.operationContext(), `UPDATE items SET tags=(
 SELECT json_group_array(tag) FROM (
  SELECT CASE WHEN value=?1 THEN ?2 ELSE value END AS tag,MIN(key) AS position
  FROM json_each(items.tags) GROUP BY tag ORDER BY position
 )),updated_at=?3 WHERE EXISTS(SELECT 1 FROM json_each(items.tags) WHERE value=?1)`,
		oldName, newName, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, fmt.Errorf("renaming tag: %w", err)
	}
	return result.RowsAffected()
}

// DeleteTag removes a tag from all items that contain it.
func (s *SQLiteStore) DeleteTag(name string) (int64, error) {
	if strings.TrimSpace(name) == "" || len(name) > 50 {
		return 0, fmt.Errorf("tag names must contain 1-50 bytes")
	}
	result, err := s.db.ExecContext(s.operationContext(), `UPDATE items SET tags=(
 SELECT json_group_array(value) FROM (SELECT value FROM json_each(items.tags) WHERE value<>?1 ORDER BY key)
 ),updated_at=?2 WHERE EXISTS(SELECT 1 FROM json_each(items.tags) WHERE value=?1)`,
		name, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, fmt.Errorf("deleting tag: %w", err)
	}
	return result.RowsAffected()
}

// --- Helper functions ---

// normalizeTags lowercases, trims and deduplicates without byte-truncating UTF-8.
// Authoritative callers validate the normalized tags before writing.
func normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || seen[t] {
			continue
		}

		seen[t] = true
		result = append(result, t)
	}
	return result
}

// buildSQLiteTagClause creates a WHERE clause for tag filtering.
// tagsJSON is a JSON array of tag names. Returns empty string if no tags.
func buildSQLiteTagClause(tagsJSON string, tableAlias string) (string, []any) {
	if tagsJSON == "" {
		return "", nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil || len(tags) == 0 {
		return "", nil
	}

	prefix := ""
	if tableAlias != "" {
		prefix = tableAlias + "."
	}

	placeholders := make([]string, len(tags))
	args := make([]any, len(tags))
	for i, t := range tags {
		placeholders[i] = "?"
		args[i] = t
	}

	clause := fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%stags) WHERE value IN (%s))",
		prefix, strings.Join(placeholders, ","))
	return clause, args
}

// attrFilter represents a single attribute filter from the frontend.
type attrFilter struct {
	Field  string   `json:"field"`
	Op     string   `json:"op"`
	Value  any      `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

func parseFilters(filtersJSON string) ([]attrFilter, error) {
	if filtersJSON == "" {
		return nil, nil
	}
	var filters []attrFilter
	if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
		return nil, fmt.Errorf("parsing filters: %w", err)
	}
	if len(filters) > 50 {
		return nil, fmt.Errorf("too many filters")
	}
	for _, f := range filters {
		if f.Field == "" || len(f.Field) > 128 {
			return nil, fmt.Errorf("invalid filter field")
		}
		switch f.Op {
		case "in":
			if len(f.Values) > 100 {
				return nil, fmt.Errorf("too many filter values")
			}
		case "eq", "gte", "lte":
			if f.Value == nil {
				return nil, fmt.Errorf("filter value is required")
			}
		default:
			return nil, fmt.Errorf("invalid filter operator %q", f.Op)
		}
	}
	return filters, nil
}

func buildSQLiteFilterClauses(filters []attrFilter, tableAlias string) ([]string, []any) {
	var clauses []string
	var args []any
	col := func(field string) string {
		prefix := ""
		if tableAlias != "" {
			prefix = tableAlias + "."
		}
		if field == "purchasePrice" {
			return prefix + "purchase_price"
		}
		return fmt.Sprintf("json_extract(%sattributes, ?)", prefix)
	}

	for _, f := range filters {
		if f.Op == "in" && len(f.Values) == 0 {
			continue
		}
		expr := col(f.Field)
		if f.Field != "purchasePrice" {
			key, _ := json.Marshal(f.Field)
			path := "$." + string(key)
			// The expression is used twice in each clause.
			args = append(args, path, path)
		}
		switch f.Op {
		case "in":
			if len(f.Values) == 0 {
				continue
			}
			placeholders := make([]string, len(f.Values))
			for i, v := range f.Values {
				placeholders[i] = "?"
				args = append(args, v)
			}
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s IN (%s))", expr, expr, strings.Join(placeholders, ",")))
		case "eq":
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s = ?)", expr, expr))
			args = append(args, f.Value)
		case "gte":
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s >= ?)", expr, expr))
			args = append(args, f.Value)
		case "lte":
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s <= ?)", expr, expr))
			args = append(args, f.Value)
		}
	}
	return clauses, args
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var item Item
		var imagesJSON, tagsJSON, attrsJSON string
		err := rows.Scan(
			&item.ID, &item.ModuleID, &item.Title, &item.PurchasePrice,
			&imagesJSON, &tagsJSON, &attrsJSON, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
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
		items = []Item{}
	}

	return items, rows.Err()
}

// buildCSV generates CSV content from items and module schemas.
func buildCSV(items []Item, modules []ModuleSchema) (string, error) {
	moduleNames := make(map[string]string)
	for _, m := range modules {
		moduleNames[m.ID] = m.DisplayName
	}

	// Collect all unique attribute keys
	attrKeys := make(map[string]bool)
	for _, item := range items {
		for k := range item.Attributes {
			attrKeys[k] = true
			if len(attrKeys) > MaxCSVColumns-7 {
				return "", ErrMetadataBudget
			}
		}
	}
	sortedKeys := make([]string, 0, len(attrKeys))
	for k := range attrKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var b strings.Builder
	header := []string{"id", "title", "module", "purchasePrice", "tags", "createdAt", "updatedAt"}
	header = append(header, sortedKeys...)
	line := csvRow(header)
	if len(line) > MaxCSVBytes {
		return "", ErrMetadataBudget
	}
	b.WriteString(line)

	for _, item := range items {
		modName := moduleNames[item.ModuleID]
		if modName == "" {
			modName = item.ModuleID
		}
		price := ""
		if item.PurchasePrice != nil {
			price = fmt.Sprintf("%.2f", *item.PurchasePrice)
		}
		tagsStr := strings.Join(item.Tags, ", ")
		row := []string{item.ID, item.Title, modName, price, tagsStr, item.CreatedAt, item.UpdatedAt}
		for _, k := range sortedKeys {
			v := item.Attributes[k]
			if v == nil {
				row = append(row, "")
			} else {
				row = append(row, fmt.Sprintf("%v", v))
			}
		}
		line = csvRow(row)
		if b.Len()+len(line) > MaxCSVBytes {
			return "", ErrMetadataBudget
		}
		b.WriteString(line)
	}

	return b.String(), nil
}

func csvRow(fields []string) string {
	escaped := make([]string, len(fields))
	for i, f := range fields {
		trimmed := strings.TrimLeftFunc(f, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) || r == '\uFEFF' })
		risky := strings.HasPrefix(f, "\t") || strings.HasPrefix(f, "\r") || strings.HasPrefix(f, "\n")
		for _, r := range trimmed {
			risky = risky || strings.ContainsRune("=+-@＝＋－＠", r)
			break
		}
		if risky {
			f = "'" + f
		}
		if strings.ContainsAny(f, ",\"\n\r") {
			escaped[i] = "\"" + strings.ReplaceAll(f, "\"", "\"\"") + "\""
		} else {
			escaped[i] = f
		}
	}
	return strings.Join(escaped, ",") + "\n"
}

// ValidateModuleSchema checks required fields and attribute uniqueness.
func ValidateModuleSchema(schema *ModuleSchema) error {
	if raw, err := json.Marshal(schema); err != nil || len(raw) > 256<<10 {
		return fmt.Errorf("module schema exceeds 256 KB limit")
	}
	if !safeModuleID.MatchString(schema.ID) {
		return fmt.Errorf("module id must contain 1-128 letters, digits, underscores or hyphens")
	}
	if strings.TrimSpace(schema.DisplayName) == "" || len(schema.DisplayName) > 256 || len(schema.Description) > 4096 || len(schema.Attributes) > 128 {
		return fmt.Errorf("missing required field: displayName")
	}

	validTypes := map[string]bool{
		"string": true, "number": true, "boolean": true, "date": true, "enum": true,
	}

	attrNames := make(map[string]bool)
	for _, attr := range schema.Attributes {
		if unsafeAttributeName(attr.Name) {
			return fmt.Errorf("attribute missing name")
		}
		if !validTypes[attr.Type] {
			return fmt.Errorf("attribute %q has unrecognized type %q", attr.Name, attr.Type)
		}
		if attrNames[attr.Name] {
			return fmt.Errorf("duplicate attribute name: %q", attr.Name)
		}
		attrNames[attr.Name] = true

		if len(attr.Options) > 256 {
			return fmt.Errorf("too many enum options")
		}
		seenOptions := map[string]bool{}
		for _, option := range attr.Options {
			if strings.TrimSpace(option) == "" || len(option) > 1024 || seenOptions[option] {
				return fmt.Errorf("invalid or duplicate enum option")
			}
			seenOptions[option] = true
		}
		if attr.Type == "enum" && len(attr.Options) == 0 {
			return fmt.Errorf("enum attribute %q must have at least one option", attr.Name)
		}
	}

	return nil
}
