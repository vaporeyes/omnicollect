// ABOUTME: PostgresStore implements the Store interface using PostgreSQL with schema-per-tenant.
// ABOUTME: Uses tsvector/tsquery for FTS, JSONB for attributes, GIN indexes for performance.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"omnicollect/tenantid"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresStore implements Store using PostgreSQL with schema-per-tenant isolation.
type PostgresStore struct {
	db           *sql.DB
	requestCtx   context.Context
	tenantSchema string
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

	schema := tenantid.Local(tenantID)
	store := &PostgresStore{db: db, tenantSchema: schema}

	if err := store.initTenantSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing tenant schema: %w", err)
	}

	if err := initPublicShowcasesTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing public showcases table: %w", err)
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

	if err := initPublicShowcasesTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing public showcases table: %w", err)
	}

	return &PostgresStore{db: db, tenantSchema: "public"}, nil
}

// Ping checks database connectivity.
func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// DB returns the underlying *sql.DB for migration operations.
func (s *PostgresStore) DB() *sql.DB {
	return s.db
}

// TenantSchema returns the schema name for this tenant.
func (s *PostgresStore) TenantSchema() string {
	return s.tenantSchema
}

// table qualifies every tenant table independently of pooled session state.
func (s *PostgresStore) table(name string) string {
	return pq.QuoteIdentifier(s.TenantSchema()) + "." + pq.QuoteIdentifier(name)
}

// WithTenantSchema returns a shallow copy of the store scoped to the given tenant.
// The copy shares the same *sql.DB connection pool but has its own tenantSchema,
// so concurrent requests for different tenants do not race on shared state.
func (s *PostgresStore) WithTenantSchema(schema string) *PostgresStore {
	return &PostgresStore{db: s.db, tenantSchema: schema, requestCtx: s.requestCtx}
}

// ProvisionTenant creates the schema and DDL tables for a tenant if they
// do not already exist. Idempotent -- safe to call multiple times.
func (s *PostgresStore) ProvisionTenant(tenantID string) error {
	return s.WithTenantSchema(tenantID).initTenantSchema()
}

func (s *PostgresStore) initTenantSchema() error {
	if err := s.validateTenant(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(s.operationContext(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize provisioning of the same tenant across processes and requests.
	if _, err := tx.ExecContext(s.operationContext(), `SELECT pg_advisory_xact_lock(hashtext($1))`, s.TenantSchema()); err != nil {
		return err
	}
	schema := pq.QuoteIdentifier(s.TenantSchema())
	statements := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema),
		fmt.Sprintf(`SET LOCAL search_path TO %s`, schema),
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
		`CREATE TABLE IF NOT EXISTS deletion_batches (id TEXT PRIMARY KEY, payload TEXT NOT NULL, titles TEXT NOT NULL, item_count INTEGER NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_id ON items(module_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_module_updated_id ON items(module_id,updated_at DESC,id ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_items_updated_id ON items(updated_at DESC, id ASC)`,
		// Migration: add tags column if table existed before tags feature
		`ALTER TABLE items ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'`,
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
		if _, err := tx.ExecContext(s.operationContext(), stmt); err != nil {
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
		$$ LANGUAGE plpgsql`, schema)
	if _, err := tx.ExecContext(s.operationContext(), triggerFn); err != nil {
		return fmt.Errorf("creating trigger function: %w", err)
	}

	// Create trigger if not exists (drop and recreate to ensure correctness)
	dropTrigger := fmt.Sprintf(`DROP TRIGGER IF EXISTS items_search_update ON %s.items`, schema)
	if _, err := tx.ExecContext(s.operationContext(), dropTrigger); err != nil {
		return fmt.Errorf("dropping trigger: %w", err)
	}

	createTrigger := fmt.Sprintf(`CREATE TRIGGER items_search_update
		BEFORE INSERT OR UPDATE ON %s.items
		FOR EACH ROW EXECUTE FUNCTION %s.items_search_update()`, schema, schema)
	if _, err := tx.ExecContext(s.operationContext(), createTrigger); err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}

	return tx.Commit()
}

var tenantSchemaPattern = regexp.MustCompile(`^tenant_[a-z0-9_]{1,56}$`)

// validateTenant rejects identifiers PostgreSQL would truncate or treat as public.
func (s *PostgresStore) validateTenant() error {
	schema := s.TenantSchema()
	if !tenantSchemaPattern.MatchString(schema) {
		return fmt.Errorf("invalid tenant schema")
	}
	return nil
}

// QueryItems retrieves items with optional tsvector search, module filter,
// JSONB attribute filters, and tag filters.
func (s *PostgresStore) QueryItems(query string, moduleID string, filtersJSON string, tagsJSON string) ([]Item, error) {
	return readLegacyItems(s.queryItems, query, moduleID, filtersJSON, tagsJSON)
}

func (s *PostgresStore) queryItems(query string, moduleID string, filtersJSON string, tagsJSON string, limit, offset int) ([]Item, error) {
	if err := ValidateQuery(query, moduleID, filtersJSON, tagsJSON); err != nil {
		return nil, err
	}
	if err := s.validateTenant(); err != nil {
		return nil, err
	}

	filters, err := parseFilters(filtersJSON)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if query != "" {
		baseSQL := `SELECT i.id, i.module_id, i.title, i.purchase_price, i.images, i.tags, i.attributes, i.created_at, i.updated_at
			FROM ` + s.table("items") + ` i, plainto_tsquery('english', $1) q
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

		baseSQL += " ORDER BY ts_rank(i.search_vector, q) DESC, i.id ASC"
		baseSQL, queryArgs = appendItemPage(baseSQL, queryArgs, limit, offset, true)

		rows, err = s.db.QueryContext(s.operationContext(), baseSQL, queryArgs...)
	} else {
		baseSQL := `SELECT id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at
			FROM ` + s.table("items") + ``
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
		baseSQL += " ORDER BY updated_at DESC, id ASC"
		baseSQL, queryArgs = appendItemPage(baseSQL, queryArgs, limit, offset, true)

		rows, err = s.db.QueryContext(s.operationContext(), baseSQL, queryArgs...)
	}

	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	return scanPgItems(rows)
}

// InsertItem creates a new item with a generated UUID.
func (s *PostgresStore) InsertItem(item Item) (Item, error) {
	if err := s.validateTenant(); err != nil {
		return Item{}, err
	}

	item.ID = uuid.New().String()
	now := time.Now().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
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
		`INSERT INTO `+s.table("items")+` (id, module_id, title, purchase_price, images, tags, attributes, created_at, updated_at)
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
	if err := s.validateTenant(); err != nil {
		return Item{}, err
	}

	item.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)

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
		`UPDATE `+s.table("items")+` SET module_id=$1, title=$2, purchase_price=$3, images=$4, tags=$5, attributes=$6, updated_at=$7
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
	if err := validateIDs([]string{id}); err != nil {
		return err
	}
	if err := s.validateTenant(); err != nil {
		return err
	}

	result, err := s.db.ExecContext(s.operationContext(), `DELETE FROM `+s.table("items")+` WHERE id = $1`, id)
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
	if err := validateIDs(ids); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := s.validateTenant(); err != nil {
		return 0, err
	}

	tx, err := s.db.BeginTx(s.operationContext(), nil)
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

	result, err := tx.ExecContext(s.operationContext(),
		"DELETE FROM "+s.table("items")+" WHERE id IN ("+strings.Join(placeholders, ",")+")",
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
	if err := validateIDs(ids); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := s.validateTenant(); err != nil {
		return 0, err
	}

	tx, err := s.db.BeginTx(s.operationContext(), nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReassignment(s.operationContext(), tx, s.table, true, ids, newModuleID); err != nil {
		return 0, err
	}
	now := time.Now().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
	placeholders := make([]string, len(ids))
	args := []any{newModuleID, now}
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, id)
	}

	result, err := tx.ExecContext(s.operationContext(),
		"UPDATE "+s.table("items")+" SET module_id = $1, updated_at = $2 WHERE id IN ("+strings.Join(placeholders, ",")+")",
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
	if err := s.validateTenant(); err != nil {
		return "", err
	}
	return exportItemsCSV(s.operationContext(), s.db, s.table("items"), true, ids, modules)
}

// GetModules returns all module schemas from the modules table.
func (s *PostgresStore) GetModules() ([]ModuleSchema, error) {
	if err := s.validateTenant(); err != nil {
		return nil, err
	}

	return readModules(s.operationContext(), s.db, s.table("modules"), true)
}

// SaveModule upserts a module schema into the modules table.
func (s *PostgresStore) SaveModule(schema ModuleSchema) error {
	if err := s.validateTenant(); err != nil {
		return err
	}
	return saveModule(s.operationContext(), s.db, s.table, true, schema)
}

// LoadModuleFile returns the schema JSON for a given module ID.
func (s *PostgresStore) LoadModuleFile(id string) (string, error) {
	if err := s.validateTenant(); err != nil {
		return "", err
	}

	var schemaJSON string
	err := s.db.QueryRowContext(s.operationContext(), `SELECT schema_json FROM `+s.table("modules")+` WHERE id = $1`, id).Scan(&schemaJSON)
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
	if err := s.validateTenant(); err != nil {
		return "{}", err
	}

	rows, err := s.db.QueryContext(s.operationContext(), `SELECT key, value FROM `+s.table("settings")+``)
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
		if key == "settings" {
			object, err := settingsObject(value)
			if err != nil {
				return "{}", err
			}
			for name, setting := range object {
				result[name] = setting
			}
		} else {
			result[key] = json.RawMessage(value)
		}
	}
	if err := rows.Err(); err != nil {
		return "{}", err
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
	if _, err := settingsObject(settingsJSON); err != nil {
		return err
	}

	if err := s.validateTenant(); err != nil {
		return err
	}

	result, err := s.db.ExecContext(s.operationContext(),
		`INSERT INTO `+s.table("settings")+` AS existing (key, value)
         SELECT 'settings', $1::jsonb WHERE octet_length(($1::jsonb)::text)<=1048576
		 ON CONFLICT (key) DO UPDATE SET value = existing.value || EXCLUDED.value
         WHERE octet_length((existing.value || EXCLUDED.value)::text)<=1048576`,
		settingsJSON,
	)
	if err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return fmt.Errorf("merged settings exceed 1 MB limit")
	}
	return nil
}

// initPublicShowcasesTable creates the showcases table in the public schema.
// Public schema is used so slugs can be looked up without tenant context.
func initPublicShowcasesTable(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS public.showcases (
			id TEXT PRIMARY KEY,
			slug TEXT NOT NULL UNIQUE,
			tenant_id TEXT NOT NULL,
			module_id TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_showcases_slug ON public.showcases(slug)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_showcases_tenant_module ON public.showcases(tenant_id, module_id)`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing showcase DDL: %w\nStatement: %s", err, stmt)
		}
	}
	return nil
}

// GetShowcaseBySlug looks up a showcase by its URL slug across all tenants.
func (s *PostgresStore) GetShowcaseBySlug(slug string) (*Showcase, error) {
	var sc Showcase
	err := s.db.QueryRowContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM public.showcases WHERE slug = $1`,
		slug,
	).Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &sc.Enabled, &sc.CreatedAt, &sc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying showcase by slug: %w", err)
	}
	return &sc, nil
}

// GetShowcaseForModule returns the showcase for a module within the current tenant.
func (s *PostgresStore) GetShowcaseForModule(moduleID string) (*Showcase, error) {
	var sc Showcase
	err := s.db.QueryRowContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM public.showcases WHERE tenant_id = $1 AND module_id = $2`,
		s.TenantSchema(), moduleID,
	).Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &sc.Enabled, &sc.CreatedAt, &sc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying showcase for module: %w", err)
	}
	return &sc, nil
}

// UpsertShowcase creates or updates a showcase record. Slug is never overwritten on update.
func (s *PostgresStore) UpsertShowcase(showcase Showcase) error {
	_, err := s.db.ExecContext(s.operationContext(),
		`INSERT INTO public.showcases (id, slug, tenant_id, module_id, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT(tenant_id, module_id) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = EXCLUDED.updated_at`,
		showcase.ID, showcase.Slug, showcase.TenantID, showcase.ModuleID, showcase.Enabled, showcase.CreatedAt, showcase.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting showcase: %w", err)
	}
	return nil
}

// ListShowcases returns all showcases for the current tenant.
func (s *PostgresStore) ListShowcases() ([]Showcase, error) {
	rows, err := s.db.QueryContext(s.operationContext(),
		`SELECT id, slug, tenant_id, module_id, enabled, created_at, updated_at FROM public.showcases WHERE tenant_id = $1 ORDER BY created_at DESC,id LIMIT $2`,
		s.TenantSchema(), MaxShowcaseResults+1,
	)
	if err != nil {
		return nil, fmt.Errorf("querying showcases: %w", err)
	}
	defer rows.Close()

	var showcases []Showcase
	for rows.Next() {
		var sc Showcase
		if err := rows.Scan(&sc.ID, &sc.Slug, &sc.TenantID, &sc.ModuleID, &sc.Enabled, &sc.CreatedAt, &sc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning showcase: %w", err)
		}
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
func (s *PostgresStore) Close() error {
	if s.requestCtx != nil {
		return nil
	} // Borrowed request handles do not own the pool.
	return s.db.Close()
}

// GetAllTags returns all distinct tags with item counts.
func (s *PostgresStore) GetAllTags() ([]TagCount, error) {
	if err := s.validateTenant(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(s.operationContext(),
		`SELECT tag, COUNT(*) FROM `+s.table("items")+`, jsonb_array_elements_text(tags) AS tag GROUP BY tag ORDER BY tag LIMIT 10001`,
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

// RenameTag renames a tag across all items using JSONB operators.
func (s *PostgresStore) RenameTag(oldName, newName string) (int64, error) {
	newName = strings.ToLower(strings.TrimSpace(newName))
	if newName == "" {
		return 0, fmt.Errorf("new tag name cannot be empty")
	}
	if len(newName) > 50 || strings.TrimSpace(oldName) == "" || len(oldName) > 50 {
		return 0, fmt.Errorf("tag names must contain 1-50 bytes")
	}

	if err := s.validateTenant(); err != nil {
		return 0, err
	}

	result, err := s.db.ExecContext(s.operationContext(),
		`UPDATE `+s.table("items")+` SET tags = (
 SELECT COALESCE(jsonb_agg(tag ORDER BY position),'[]'::jsonb) FROM (
 SELECT CASE WHEN value=$1 THEN $2 ELSE value END AS tag,MIN(ordinal) AS position
 FROM jsonb_array_elements_text(tags) WITH ORDINALITY AS source(value,ordinal)
 GROUP BY 1) AS renamed), updated_at=GREATEST(clock_timestamp(),updated_at+interval '1 microsecond') WHERE tags ? $1`,
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
	if strings.TrimSpace(name) == "" || len(name) > 50 {
		return 0, fmt.Errorf("tag names must contain 1-50 bytes")
	}
	if err := s.validateTenant(); err != nil {
		return 0, err
	}

	result, err := s.db.ExecContext(s.operationContext(),
		`UPDATE `+s.table("items")+` SET tags = tags - $1, updated_at=GREATEST(clock_timestamp(),updated_at+interval '1 microsecond') WHERE tags ? $1`,
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
		expr := fmt.Sprintf("%sattributes->>$%d", prefix, *paramIdx)
		args = append(args, field)
		*paramIdx++
		return expr
	}

	for _, f := range filters {
		if f.Op == "in" && len(f.Values) == 0 {
			continue
		}
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

		item.CreatedAt = createdAt.Format(time.RFC3339Nano)
		item.UpdatedAt = updatedAt.Format(time.RFC3339Nano)

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
