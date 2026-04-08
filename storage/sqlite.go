// ABOUTME: SQLiteStore implements the Store interface using local SQLite database.
// ABOUTME: Extracted from the original db.go with FTS5 search, json_extract filters, and batch ops.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using a local SQLite database with FTS5.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) the local SQLite database and runs DDL.
func NewSQLiteStore() (*SQLiteStore, error) {
	dbPath, err := sqliteDBPath()
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

	if err := createSQLiteSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
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

// QueryItems retrieves items with optional FTS search, module filter, and attribute filters.
func (s *SQLiteStore) QueryItems(query string, moduleID string, filtersJSON string) ([]Item, error) {
	filters, err := parseFilters(filtersJSON)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

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
		baseSQL += " ORDER BY rank"

		rows, err = s.db.Query(baseSQL, queryArgs...)
	} else {
		baseSQL := `SELECT id, module_id, title, purchase_price, images, attributes, created_at, updated_at
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

		if len(whereParts) > 0 {
			baseSQL += " WHERE " + strings.Join(whereParts, " AND ")
		}
		baseSQL += " ORDER BY updated_at DESC"

		rows, err = s.db.Query(baseSQL, queryArgs...)
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

	_, err = s.db.Exec(
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

// UpdateItem updates an existing item in the database.
func (s *SQLiteStore) UpdateItem(item Item) (Item, error) {
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	imagesJSON, err := json.Marshal(item.Images)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling images: %w", err)
	}

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return Item{}, fmt.Errorf("marshaling attributes: %w", err)
	}

	result, err := s.db.Exec(
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

// DeleteItem removes an item from the database by ID.
func (s *SQLiteStore) DeleteItem(id string) error {
	result, err := s.db.Exec(`DELETE FROM items WHERE id = ?`, id)
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
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := s.db.Begin()
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

	result, err := tx.Exec(
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
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	placeholders := make([]string, len(ids))
	args := []any{newModuleID, now}
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	result, err := tx.Exec(
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
	if len(ids) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := s.db.Query(
		"SELECT id, module_id, title, purchase_price, images, attributes, created_at, updated_at FROM items WHERE id IN ("+strings.Join(placeholders, ",")+") ORDER BY updated_at DESC",
		args...,
	)
	if err != nil {
		return "", fmt.Errorf("querying items for CSV: %w", err)
	}
	defer rows.Close()

	items, err := scanItems(rows)
	if err != nil {
		return "", err
	}

	return buildCSV(items, modules), nil
}

// GetModules returns module schemas from the filesystem (local mode).
func (s *SQLiteStore) GetModules() ([]ModuleSchema, error) {
	return loadModuleSchemasFromDisk()
}

// SaveModule writes a module schema JSON file to disk.
func (s *SQLiteStore) SaveModule(schema ModuleSchema) error {
	return saveModuleFileToDisk(&schema)
}

// LoadModuleFile returns the raw JSON content of a module schema file.
func (s *SQLiteStore) LoadModuleFile(id string) (string, error) {
	path, err := findModuleFileOnDisk(id)
	if err != nil {
		return "", fmt.Errorf("finding module file: %w", err)
	}
	if path == "" {
		return "", fmt.Errorf("module not found: %s", id)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading module file: %w", err)
	}

	return string(data), nil
}

// GetSettings reads settings from the local JSON file.
func (s *SQLiteStore) GetSettings() (string, error) {
	path, err := settingsPath()
	if err != nil {
		return "{}", fmt.Errorf("resolving settings path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "{}", nil
		}
		return "{}", fmt.Errorf("reading settings: %w", err)
	}

	return string(data), nil
}

// SaveSettings writes settings JSON to the local file.
func (s *SQLiteStore) SaveSettings(settingsJSON string) error {
	var check json.RawMessage
	if err := json.Unmarshal([]byte(settingsJSON), &check); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	path, err := settingsPath()
	if err != nil {
		return fmt.Errorf("resolving settings path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	// Pretty-print for readability
	var formatted json.RawMessage
	if err := json.Unmarshal([]byte(settingsJSON), &formatted); err == nil {
		pretty, err := json.MarshalIndent(formatted, "", "  ")
		if err == nil {
			settingsJSON = string(pretty)
		}
	}

	if err := os.WriteFile(path, []byte(settingsJSON), 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Helper functions ---

func settingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "OmniCollect", "settings.json"), nil
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
		return fmt.Sprintf("json_extract(%sattributes, '$.%s')", prefix, field)
	}

	for _, f := range filters {
		expr := col(f.Field)
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

// buildCSV generates CSV content from items and module schemas.
func buildCSV(items []Item, modules []ModuleSchema) string {
	moduleNames := make(map[string]string)
	for _, m := range modules {
		moduleNames[m.ID] = m.DisplayName
	}

	// Collect all unique attribute keys
	attrKeys := make(map[string]bool)
	for _, item := range items {
		for k := range item.Attributes {
			attrKeys[k] = true
		}
	}
	sortedKeys := make([]string, 0, len(attrKeys))
	for k := range attrKeys {
		sortedKeys = append(sortedKeys, k)
	}
	for i := 0; i < len(sortedKeys); i++ {
		for j := i + 1; j < len(sortedKeys); j++ {
			if sortedKeys[j] < sortedKeys[i] {
				sortedKeys[i], sortedKeys[j] = sortedKeys[j], sortedKeys[i]
			}
		}
	}

	var b strings.Builder
	header := []string{"id", "title", "module", "purchasePrice", "createdAt", "updatedAt"}
	header = append(header, sortedKeys...)
	b.WriteString(csvRow(header))

	for _, item := range items {
		modName := moduleNames[item.ModuleID]
		if modName == "" {
			modName = item.ModuleID
		}
		price := ""
		if item.PurchasePrice != nil {
			price = fmt.Sprintf("%.2f", *item.PurchasePrice)
		}
		row := []string{item.ID, item.Title, modName, price, item.CreatedAt, item.UpdatedAt}
		for _, k := range sortedKeys {
			v := item.Attributes[k]
			if v == nil {
				row = append(row, "")
			} else {
				row = append(row, fmt.Sprintf("%v", v))
			}
		}
		b.WriteString(csvRow(row))
	}

	return b.String()
}

func csvRow(fields []string) string {
	escaped := make([]string, len(fields))
	for i, f := range fields {
		if strings.ContainsAny(f, ",\"\n\r") {
			escaped[i] = "\"" + strings.ReplaceAll(f, "\"", "\"\"") + "\""
		} else {
			escaped[i] = f
		}
	}
	return strings.Join(escaped, ",") + "\n"
}

// --- Module file operations (local filesystem) ---

func localModulesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".omnicollect", "modules"), nil
}

func loadModuleSchemasFromDisk() ([]ModuleSchema, error) {
	dir, err := localModulesDir()
	if err != nil {
		return nil, fmt.Errorf("resolving modules dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating modules dir: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading modules dir: %w", err)
	}

	var schemas []ModuleSchema
	seen := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		schema, err := parseModuleSchemaFile(path)
		if err != nil {
			log.Printf("skipping malformed schema %s: %v", entry.Name(), err)
			continue
		}

		if existing, ok := seen[schema.ID]; ok {
			log.Printf("skipping duplicate module ID %q in %s (already defined in %s)", schema.ID, entry.Name(), existing)
			continue
		}

		seen[schema.ID] = entry.Name()
		schemas = append(schemas, *schema)
	}

	if schemas == nil {
		schemas = []ModuleSchema{}
	}

	return schemas, nil
}

func parseModuleSchemaFile(path string) (*ModuleSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var schema ModuleSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if err := ValidateModuleSchema(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

func findModuleFileOnDisk(moduleID string) (string, error) {
	dir, err := localModulesDir()
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		schema, err := parseModuleSchemaFile(path)
		if err != nil {
			continue
		}
		if schema.ID == moduleID {
			return path, nil
		}
	}
	return "", nil
}

func saveModuleFileToDisk(schema *ModuleSchema) error {
	dir, err := localModulesDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating modules dir: %w", err)
	}

	existingPath, err := findModuleFileOnDisk(schema.ID)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(dir, schema.ID+".json")
	if existingPath != "" && existingPath != targetPath {
		targetPath = existingPath
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("writing schema file: %w", err)
	}

	return nil
}

// ValidateModuleSchema checks required fields and attribute uniqueness.
func ValidateModuleSchema(schema *ModuleSchema) error {
	if schema.ID == "" {
		return fmt.Errorf("missing required field: id")
	}
	if schema.DisplayName == "" {
		return fmt.Errorf("missing required field: displayName")
	}

	validTypes := map[string]bool{
		"string": true, "number": true, "boolean": true, "date": true, "enum": true,
	}

	attrNames := make(map[string]bool)
	for _, attr := range schema.Attributes {
		if attr.Name == "" {
			return fmt.Errorf("attribute missing name")
		}
		if !validTypes[attr.Type] {
			return fmt.Errorf("attribute %q has unrecognized type %q", attr.Name, attr.Type)
		}
		if attrNames[attr.Name] {
			return fmt.Errorf("duplicate attribute name: %q", attr.Name)
		}
		attrNames[attr.Name] = true

		if attr.Type == "enum" && len(attr.Options) == 0 {
			return fmt.Errorf("enum attribute %q must have at least one option", attr.Name)
		}
	}

	return nil
}
