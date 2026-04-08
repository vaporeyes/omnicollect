// ABOUTME: PostgresStore implements the Store interface using PostgreSQL with schema-per-tenant.
// ABOUTME: Uses tsvector/tsquery for FTS, JSONB for attributes, GIN indexes for performance.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresStore implements Store using PostgreSQL with schema-per-tenant isolation.
type PostgresStore struct {
	db           *sql.DB
	tenantSchema string
	mu           sync.RWMutex
}

// NewPostgresStore connects to PostgreSQL and initializes the tenant schema.
func NewPostgresStore(databaseURL, tenantID string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	schema := "tenant_" + sanitizeTenantID(tenantID)
	store := &PostgresStore{db: db, tenantSchema: schema}

	if err := store.initTenantSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing tenant schema: %w", err)
	}

	return store, nil
}

// NewPostgresStoreNoTenant connects to PostgreSQL without initializing a tenant schema.
// Used in auth-enabled mode where tenants are provisioned dynamically per request.
func NewPostgresStoreNoTenant(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &PostgresStore{db: db, tenantSchema: "public"}, nil
}

// Ping checks database connectivity.
func (s *PostgresStore) Ping(ctx interface{ Deadline() (time.Time, bool) }) error {
	return s.db.Ping()
}

// DB returns the underlying *sql.DB for migration operations.
func (s *PostgresStore) DB() *sql.DB {
	return s.db
}

// TenantSchema returns the schema name for this tenant.
func (s *PostgresStore) TenantSchema() string {
	return s.tenantSchema
}

// SetTenantSchema changes the active tenant schema for subsequent operations.
// Used by auth middleware to scope requests to the authenticated user's tenant.
func (s *PostgresStore) SetTenantSchema(schema string) {
	s.mu.Lock()
	s.tenantSchema = schema
	s.mu.Unlock()
}

// ProvisionTenant creates the schema and DDL tables for a tenant if they
// do not already exist. Idempotent -- safe to call multiple times.
func (s *PostgresStore) ProvisionTenant(tenantID string) error {
	saved := s.tenantSchema
	s.mu.Lock()
	s.tenantSchema = tenantID
	s.mu.Unlock()
	err := s.initTenantSchema()
	s.mu.Lock()
	s.tenantSchema = saved
	s.mu.Unlock()
	return err
}

func sanitizeTenantID(id string) string {
	// Only allow alphanumeric and underscores
	var sb strings.Builder
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			sb.WriteRune(c)
		}
	}
	if sb.Len() == 0 {
		return "default"
	}
	return sb.String()
}

func (s *PostgresStore) initTenantSchema() error {
	statements := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, s.tenantSchema),
		fmt.Sprintf(`SET search_path TO %s`, s.tenantSchema),
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			module_id TEXT NOT NULL,
			title TEXT NOT NULL,
			purchase_price DOUBLE PRECISION,
			images JSONB NOT NULL DEFAULT '[]',
			tags JSONB NOT NULL DEFAULT '[]',
			attributes JSONB NOT NULL DEFAULT '{}',
			search_vector tsvector,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_id ON items(module_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_tags ON items USING GIN(tags)`,
		`CREATE INDEX IF NOT EXISTS idx_items_search_vector ON items USING GIN(search_vector)`,
		`CREATE TABLE IF NOT EXISTS modules (
			id TEXT PRIMARY KEY,
			display_name TEXT NOT NULL,
			description TEXT,
			schema_json JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL
		)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("executing DDL: %w\nStatement: %s", err, stmt)
		}
	}

	// Create or replace the search vector trigger function.
	// Includes attribute values and tag values in the search index.
	triggerFn := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s.items_search_update() RETURNS trigger AS $$
		DECLARE
			attr_text TEXT;
			tags_text TEXT;
		BEGIN
			SELECT string_agg(value::text, ' ')
			INTO attr_text
			FROM jsonb_each_text(NEW.attributes);

			SELECT string_agg(t::text, ' ')
			INTO tags_text
			FROM jsonb_array_elements_text(NEW.tags) AS t;

			NEW.search_vector := to_tsvector('english', coalesce(NEW.title, '') || ' ' || coalesce(attr_text, '') || ' ' || coalesce(tags_text, ''));
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`, s.tenantSchema)
	if _, err := s.db.Exec(triggerFn); err != nil {
		return fmt.Errorf("creating trigger function: %w", err)
	}

	// Create trigger if not exists (drop and recreate to ensure correctness)
	dropTrigger := fmt.Sprintf(`DROP TRIGGER IF EXISTS items_search_update ON %s.items`, s.tenantSchema)
	if _, err := s.db.Exec(dropTrigger); err != nil {
		return fmt.Errorf("dropping trigger: %w", err)
	}

	createTrigger := fmt.Sprintf(`CREATE TRIGGER items_search_update
		BEFORE INSERT OR UPDATE ON %s.items
		FOR EACH ROW EXECUTE FUNCTION %s.items_search_update()`, s.tenantSchema, s.tenantSchema)
	if _, err := s.db.Exec(createTrigger); err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}

	// Migration: add tags column to existing databases that lack it
	s.db.Exec(fmt.Sprintf(`ALTER TABLE %s.items ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'`, s.tenantSchema))
	s.db.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_items_tags ON %s.items USING GIN(tags)`, s.tenantSchema))

	return nil
}

func (s *PostgresStore) setSearchPath() error {
	_, err := s.db.Exec(fmt.Sprintf("SET search_path TO %s", s.tenantSchema))
	return err
}

// QueryItems retrieves items with optional tsvector search, module filter,
// JSONB attribute filters, and tag filters.
func (s *PostgresStore) QueryItems(query string, moduleID string, filtersJSON string, tagsJSON string) ([]Item, error) {
	if err := s.setSearchPath(); err != nil {
		return nil, err
	}

	filters, err := parseFilters(filtersJSON)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if query != "" {
		baseSQL := `SELECT i.id, i.module_id, i.title, i.purchase_price, i.images, i.tags, i.attributes, i.created_at, i.updated_at
			FROM items i, plainto_tsquery('english', $1) q
			WHERE i.search_vector @@ q`
		queryArgs := []any{query}
		paramIdx := 2

		if moduleID != "" {
			baseSQL += fmt.Sprintf(" AND i.module_id = $%d", paramIdx)
			queryArgs = append(queryArgs, moduleID)
			paramIdx++
		}

		filterClauses, filterArgs := buildPgFilterClauses(filters, "i", &paramIdx)
		for _, c := range filterClauses {
			baseSQL += " AND " + c
		}
		queryArgs = append(queryArgs, filterArgs...)

		tagClause, tagArgs := buildPgTagClause(tagsJSON, "i", &paramIdx)
		if tagClause != "" {
			baseSQL += " AND " + tagClause
			queryArgs = append(queryArgs, tagArgs...)
		}

		baseSQL += " ORDER BY ts_rank(i.search_vector, q) DESC"

		rows, err = s.db.Query(baseSQL, queryArgs...)
	} else {
		baseSQL := `SELECT id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at
			FROM items`
		var queryArgs []any
		var whereParts []string
		paramIdx := 1

		if moduleID != "" {
			whereParts = append(whereParts, fmt.Sprintf("module_id = $%d", paramIdx))
			queryArgs = append(queryArgs, moduleID)
			paramIdx++
		}

		filterClauses, filterArgs := buildPgFilterClauses(filters, "", &paramIdx)
		whereParts = append(whereParts, filterClauses...)
		queryArgs = append(queryArgs, filterArgs...)

		tagClause, tagArgs := buildPgTagClause(tagsJSON, "", &paramIdx)
		if tagClause != "" {
			whereParts = append(whereParts, tagClause)
			queryArgs = append(queryArgs, tagArgs...)
		}

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

	return scanPgItems(rows)
}

// InsertItem creates a new item with a generated UUID.
func (s *PostgresStore) InsertItem(item Item) (Item, error) {
	if err := s.setSearchPath(); err != nil {
		return Item{}, err
	}

	item.ID = uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
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

	_, err = s.db.Exec(
		`INSERT INTO items (id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		item.ID, item.ModuleID, item.Title, item.PurchasePrice,
		string(imagesJSON), string(tagsJSON), string(attrsJSON), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("inserting item: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item.
func (s *PostgresStore) UpdateItem(item Item) (Item, error) {
	if err := s.setSearchPath(); err != nil {
		return Item{}, err
	}

	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

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

	result, err := s.db.Exec(
		`UPDATE items SET module_id=$1, title=$2, purchase_price=$3, images=$4, tags=$5, attributes=$6, updated_at=$7
		 WHERE id=$8`,
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

// DeleteItem removes an item by ID.
func (s *PostgresStore) DeleteItem(id string) error {
	if err := s.setSearchPath(); err != nil {
		return err
	}

	result, err := s.db.Exec(`DELETE FROM items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found: %s", id)
	}
	return nil
}

// DeleteItems removes multiple items in a transaction.
func (s *PostgresStore) DeleteItems(ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if err := s.setSearchPath(); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
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

// BulkUpdateModule changes the module_id of multiple items.
func (s *PostgresStore) BulkUpdateModule(ids []string, newModuleID string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if err := s.setSearchPath(); err != nil {
		return 0, err
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
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, id)
	}

	result, err := tx.Exec(
		"UPDATE items SET module_id = $1, updated_at = $2 WHERE id IN ("+strings.Join(placeholders, ",")+")",
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

// ExportItemsCSV queries items by ID and generates CSV.
func (s *PostgresStore) ExportItemsCSV(ids []string, modules []ModuleSchema) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}
	if err := s.setSearchPath(); err != nil {
		return "", err
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := s.db.Query(
		"SELECT id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at FROM items WHERE id IN ("+strings.Join(placeholders, ",")+") ORDER BY updated_at DESC",
		args...,
	)
	if err != nil {
		return "", fmt.Errorf("querying items for CSV: %w", err)
	}
	defer rows.Close()

	items, err := scanPgItems(rows)
	if err != nil {
		return "", err
	}

	return buildCSV(items, modules), nil
}

// GetModules returns all module schemas from the modules table.
func (s *PostgresStore) GetModules() ([]ModuleSchema, error) {
	if err := s.setSearchPath(); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT id, display_name, description, schema_json FROM modules ORDER BY display_name`)
	if err != nil {
		return nil, fmt.Errorf("querying modules: %w", err)
	}
	defer rows.Close()

	var modules []ModuleSchema
	for rows.Next() {
		var id, displayName string
		var description sql.NullString
		var schemaJSON string

		if err := rows.Scan(&id, &displayName, &description, &schemaJSON); err != nil {
			return nil, fmt.Errorf("scanning module: %w", err)
		}

		var schema ModuleSchema
		if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
			// Fallback: use the row data directly
			schema = ModuleSchema{
				ID:          id,
				DisplayName: displayName,
				Description: description.String,
			}
		}
		// Ensure ID/DisplayName/Description match the row
		schema.ID = id
		schema.DisplayName = displayName
		schema.Description = description.String

		modules = append(modules, schema)
	}

	if modules == nil {
		modules = []ModuleSchema{}
	}

	return modules, rows.Err()
}

// SaveModule upserts a module schema into the modules table.
func (s *PostgresStore) SaveModule(schema ModuleSchema) error {
	if err := s.setSearchPath(); err != nil {
		return err
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO modules (id, display_name, description, schema_json, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (id) DO UPDATE SET display_name=$2, description=$3, schema_json=$4, updated_at=NOW()`,
		schema.ID, schema.DisplayName, schema.Description, string(schemaJSON),
	)
	if err != nil {
		return fmt.Errorf("saving module: %w", err)
	}

	return nil
}

// LoadModuleFile returns the schema JSON for a given module ID.
func (s *PostgresStore) LoadModuleFile(id string) (string, error) {
	if err := s.setSearchPath(); err != nil {
		return "", err
	}

	var schemaJSON string
	err := s.db.QueryRow(`SELECT schema_json FROM modules WHERE id = $1`, id).Scan(&schemaJSON)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("module not found: %s", id)
	}
	if err != nil {
		return "", fmt.Errorf("loading module: %w", err)
	}

	return schemaJSON, nil
}

// GetSettings returns the combined settings JSON from the settings table.
func (s *PostgresStore) GetSettings() (string, error) {
	if err := s.setSearchPath(); err != nil {
		return "{}", err
	}

	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return "{}", fmt.Errorf("querying settings: %w", err)
	}
	defer rows.Close()

	result := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return "{}", fmt.Errorf("scanning setting: %w", err)
		}
		result[key] = json.RawMessage(value)
	}

	if len(result) == 0 {
		return "{}", nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "{}", fmt.Errorf("marshaling settings: %w", err)
	}

	return string(data), nil
}

// SaveSettings writes settings JSON to the settings table.
// The JSON is stored as a single row with key "settings".
func (s *PostgresStore) SaveSettings(settingsJSON string) error {
	var check json.RawMessage
	if err := json.Unmarshal([]byte(settingsJSON), &check); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := s.setSearchPath(); err != nil {
		return err
	}

	_, err := s.db.Exec(
		`INSERT INTO settings (key, value) VALUES ('settings', $1)
		 ON CONFLICT (key) DO UPDATE SET value = $1`,
		settingsJSON,
	)
	if err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// GetAllTags returns all distinct tags with item counts.
func (s *PostgresStore) GetAllTags() ([]TagCount, error) {
	if err := s.setSearchPath(); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`SELECT tag, COUNT(*) FROM items, jsonb_array_elements_text(tags) AS tag GROUP BY tag ORDER BY tag`,
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
		tags = append(tags, tc)
	}
	if tags == nil {
		tags = []TagCount{}
	}
	return tags, rows.Err()
}

// RenameTag renames a tag across all items using JSONB operators.
func (s *PostgresStore) RenameTag(oldName, newName string) (int64, error) {
	newName = strings.ToLower(strings.TrimSpace(newName))
	if newName == "" {
		return 0, fmt.Errorf("new tag name cannot be empty")
	}
	if len(newName) > 50 {
		newName = newName[:50]
	}

	if err := s.setSearchPath(); err != nil {
		return 0, err
	}

	result, err := s.db.Exec(
		`UPDATE items SET tags = (tags - $1) || to_jsonb($2::text) WHERE tags ? $1`,
		oldName, newName,
	)
	if err != nil {
		return 0, fmt.Errorf("renaming tag: %w", err)
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// DeleteTag removes a tag from all items using JSONB operators.
func (s *PostgresStore) DeleteTag(name string) (int64, error) {
	if err := s.setSearchPath(); err != nil {
		return 0, err
	}

	result, err := s.db.Exec(
		`UPDATE items SET tags = tags - $1 WHERE tags ? $1`,
		name,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting tag: %w", err)
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// --- PostgreSQL-specific helpers ---

// buildPgTagClause creates a WHERE clause for tag filtering using JSONB ?| operator.
func buildPgTagClause(tagsJSON string, tableAlias string, paramIdx *int) (string, []any) {
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
		placeholders[i] = fmt.Sprintf("$%d", *paramIdx)
		args[i] = t
		*paramIdx++
	}

	clause := fmt.Sprintf("%stags ?| array[%s]", prefix, strings.Join(placeholders, ","))
	return clause, args
}

func buildPgFilterClauses(filters []attrFilter, tableAlias string, paramIdx *int) ([]string, []any) {
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
		return fmt.Sprintf("%sattributes->>'%s'", prefix, field)
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
				placeholders[i] = fmt.Sprintf("$%d", *paramIdx)
				args = append(args, v)
				*paramIdx++
			}
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s IN (%s))", expr, expr, strings.Join(placeholders, ",")))
		case "eq":
			clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s = $%d)", expr, expr, *paramIdx))
			args = append(args, f.Value)
			*paramIdx++
		case "gte":
			if f.Field == "purchasePrice" {
				clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s >= $%d)", expr, expr, *paramIdx))
			} else {
				clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND (%s)::numeric >= $%d)", expr, expr, *paramIdx))
			}
			args = append(args, f.Value)
			*paramIdx++
		case "lte":
			if f.Field == "purchasePrice" {
				clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND %s <= $%d)", expr, expr, *paramIdx))
			} else {
				clauses = append(clauses, fmt.Sprintf("(%s IS NOT NULL AND (%s)::numeric <= $%d)", expr, expr, *paramIdx))
			}
			args = append(args, f.Value)
			*paramIdx++
		}
	}
	return clauses, args
}

func scanPgItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var item Item
		var imagesJSON, tagsJSON, attrsJSON string
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&item.ID, &item.ModuleID, &item.Title, &item.PurchasePrice,
			&imagesJSON, &tagsJSON, &attrsJSON, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}

		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)

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
