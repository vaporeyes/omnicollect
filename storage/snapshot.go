// ABOUTME: Consistent database snapshots and atomic restore for both storage backends.
// ABOUTME: Preserves item identity/history and commits items, modules, and settings together.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// Snapshot is a portable logical backup of one collection database or tenant.
// Media bytes are archived separately; image references retain their filenames.
type Snapshot struct {
	Version  int             `json:"version"`
	Items    []Item          `json:"items"`
	Modules  []ModuleSchema  `json:"modules"`
	Settings json.RawMessage `json:"settings"`
}

// RestoreResult counts records committed by a restore, never attempted writes.
type RestoreResult struct {
	Items   int
	Modules int
}

const SnapshotVersion = 1
const maxSnapshotItems = 100000
const maxSnapshotModules = 1000

// ValidateSnapshot rejects malformed or ambiguous data before any destructive work.
func ValidateSnapshot(snapshot Snapshot) error {
	if snapshot.Version != 0 && snapshot.Version != SnapshotVersion {
		return fmt.Errorf("unsupported backup version %d", snapshot.Version)
	}
	if len(snapshot.Items) > maxSnapshotItems || len(snapshot.Modules) > maxSnapshotModules {
		return fmt.Errorf("backup exceeds record limit")
	}
	if len(snapshot.Settings) > 1<<20 {
		return fmt.Errorf("backup settings exceed size limit")
	}
	if len(snapshot.Settings) > 0 {
		if _, err := settingsObject(string(snapshot.Settings)); err != nil {
			return err
		}
	}
	modules := map[string]bool{}
	for _, module := range snapshot.Modules {
		if err := ValidateModuleSchema(&module); err != nil {
			return err
		}
		if modules[module.ID] {
			return fmt.Errorf("duplicate module id %q", module.ID)
		}
		modules[module.ID] = true
	}
	items := map[string]bool{}
	for _, item := range snapshot.Items {
		if strings.TrimSpace(item.ID) == "" || len(item.ID) > 128 {
			return fmt.Errorf("invalid restored item id")
		}
		if items[item.ID] {
			return fmt.Errorf("duplicate item id %q", item.ID)
		}
		items[item.ID] = true
		if strings.TrimSpace(item.Title) == "" || len(item.Title) > 4096 || !safeModuleID.MatchString(item.ModuleID) {
			return fmt.Errorf("invalid item %q", item.ID)
		}
		if item.PurchasePrice != nil && (math.IsNaN(*item.PurchasePrice) || math.IsInf(*item.PurchasePrice, 0)) {
			return fmt.Errorf("invalid price for %q", item.ID)
		}
		for _, timestamp := range []string{item.CreatedAt, item.UpdatedAt} {
			if _, err := time.Parse(time.RFC3339Nano, timestamp); err != nil {
				return fmt.Errorf("invalid timestamp for %q: %w", item.ID, err)
			}
		}
		for _, filename := range item.Images {
			if err := ValidateFilename(filename); err != nil {
				return fmt.Errorf("invalid image on %q: %w", item.ID, err)
			}
		}
		if _, err := json.Marshal(item.Attributes); err != nil {
			return fmt.Errorf("invalid attributes on %q: %w", item.ID, err)
		}
	}
	return nil
}

func (s *SQLiteStore) Snapshot(ctx context.Context) (Snapshot, error) {
	return readSnapshot(ctx, s.db, false, func(name string) string { return name })
}

func (s *PostgresStore) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := s.validateTenant(); err != nil {
		return Snapshot{}, err
	}
	return readSnapshot(ctx, s.db, true, s.table)
}

func readSnapshot(ctx context.Context, db *sql.DB, postgres bool, table func(string) string) (Snapshot, error) {
	options := &sql.TxOptions{ReadOnly: true}
	if postgres {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return Snapshot{}, err
	}
	defer tx.Rollback()
	if err = checkReadBudget(ctx, tx, table("items"), itemReadBytes(postgres), maxSnapshotItems, MaxSnapshotReadBytes, MaxMetadataRowBytes); err != nil {
		return Snapshot{}, err
	}
	if err = checkReadBudget(ctx, tx, table("modules"), encodedBytes("schema_json", postgres), maxSnapshotModules, MaxModuleReadBytes, 256<<10); err != nil {
		return Snapshot{}, err
	}
	if err = checkReadBudget(ctx, tx, table("settings")+" WHERE key='settings'", encodedBytes("value", postgres), 1, 1<<20, 1<<20); err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{Version: SnapshotVersion, Items: []Item{}, Modules: []ModuleSchema{}, Settings: json.RawMessage(`{}`)}
	rows, err := tx.QueryContext(ctx, "SELECT id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at FROM "+table("items")+" ORDER BY id")
	if err != nil {
		return Snapshot{}, err
	}
	// database/sql converts TIMESTAMPTZ to RFC3339Nano when scanning into string.
	snapshot.Items, err = scanSnapshotItems(rows)
	rows.Close()
	if err != nil {
		return Snapshot{}, err
	}
	rows, err = tx.QueryContext(ctx, "SELECT schema_json FROM "+table("modules")+" ORDER BY id")
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Modules, err = scanModules(rows)
	rows.Close()
	if err != nil {
		return Snapshot{}, err
	}
	var settings string
	err = tx.QueryRowContext(ctx, "SELECT value FROM "+table("settings")+" WHERE key='settings'").Scan(&settings)
	if err != nil && err != sql.ErrNoRows {
		return Snapshot{}, err
	}
	if err == nil {
		snapshot.Settings = json.RawMessage(settings)
	}
	if err := ValidateSnapshot(snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func scanSnapshotItem(rows *sql.Rows) (Item, error) {
	var item Item
	var images, tags, attributes string
	if err := rows.Scan(&item.ID, &item.ModuleID, &item.Title, &item.PurchasePrice, &images, &tags, &attributes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	if err := json.Unmarshal([]byte(images), &item.Images); err != nil {
		return item, fmt.Errorf("corrupt image references on %q: %w", item.ID, err)
	}
	if err := json.Unmarshal([]byte(tags), &item.Tags); err != nil {
		return item, fmt.Errorf("corrupt tags on %q: %w", item.ID, err)
	}
	if err := json.Unmarshal([]byte(attributes), &item.Attributes); err != nil {
		return item, fmt.Errorf("corrupt attributes on %q: %w", item.ID, err)
	}
	return item, nil
}
func scanSnapshotItems(rows *sql.Rows) ([]Item, error) {
	items := []Item{}
	for rows.Next() {
		item, err := scanSnapshotItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) Restore(ctx context.Context, snapshot Snapshot, mode string) (RestoreResult, error) {
	return restoreSnapshot(ctx, s.db, snapshot, mode, false, func(name string) string { return name }, "")
}

func (s *PostgresStore) Restore(ctx context.Context, snapshot Snapshot, mode string) (RestoreResult, error) {
	if err := s.validateTenant(); err != nil {
		return RestoreResult{}, err
	}
	return restoreSnapshot(ctx, s.db, snapshot, mode, true, s.table, s.TenantSchema())
}

func restoreSnapshot(ctx context.Context, db *sql.DB, snapshot Snapshot, mode string, postgres bool, table func(string) string, tenant string) (RestoreResult, error) {
	zero := RestoreResult{}
	if mode != "replace" && mode != "merge" {
		return zero, fmt.Errorf("invalid restore mode")
	}
	if err := ValidateSnapshot(snapshot); err != nil {
		return zero, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	if postgres {
		// Exclude concurrent mutation during replace/upsert without relying on session state.
		if _, err := tx.ExecContext(ctx, "LOCK TABLE "+table("modules")+", "+table("items")+", "+table("settings")+", "+table("deletion_batches")+" IN EXCLUSIVE MODE"); err != nil {
			return zero, err
		}
	}
	parameter := func(n int) string {
		if postgres {
			return fmt.Sprintf("$%d", n)
		}
		return "?"
	}
	modules := map[string]ModuleSchema{}
	if mode == "merge" {
		if err := checkReadBudget(ctx, tx, table("modules"), encodedBytes("schema_json", postgres), maxSnapshotModules, MaxModuleReadBytes, 256<<10); err != nil {
			return zero, err
		}
		rows, err := tx.QueryContext(ctx, "SELECT schema_json FROM "+table("modules"))
		if err != nil {
			return zero, err
		}
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return zero, err
			}
			var module ModuleSchema
			if err := json.Unmarshal([]byte(raw), &module); err != nil {
				rows.Close()
				return zero, err
			}
			modules[module.ID] = module
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return zero, err
		}
	}
	for _, module := range snapshot.Modules {
		modules[module.ID] = module
	}
	for _, item := range snapshot.Items {
		if _, exists := modules[item.ModuleID]; !exists {
			return zero, fmt.Errorf("item %q references missing module %q", item.ID, item.ModuleID)
		}
	}
	for _, item := range snapshot.Items {
		if err := ValidateItem(item, modules[item.ModuleID]); err != nil {
			return zero, fmt.Errorf("invalid restored item %q: %w", item.ID, err)
		}
	}
	if mode == "merge" {
		replaced := map[string]bool{}
		for _, item := range snapshot.Items {
			replaced[item.ID] = true
		}
		for _, module := range snapshot.Modules {
			if err := validateSchemaItems(ctx, tx, table, postgres, module, replaced); err != nil {
				return zero, err
			}
		}
	}
	if mode == "replace" {
		for _, name := range []string{"items", "modules", "settings", "deletion_batches"} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table(name)); err != nil {
				return zero, err
			}
		}
		// A restored collection must not accidentally publish replacement data.
		if postgres {
			if _, err := tx.ExecContext(ctx, "UPDATE public.showcases SET enabled=false WHERE tenant_id=$1", tenant); err != nil {
				return zero, err
			}
		} else if _, err := tx.ExecContext(ctx, "UPDATE showcases SET enabled=0"); err != nil {
			return zero, err
		}
	}
	moduleSQL := "INSERT INTO " + table("modules") + " (id,display_name,description,schema_json) VALUES (" + parameter(1) + "," + parameter(2) + "," + parameter(3) + "," + parameter(4) + ") ON CONFLICT(id) DO UPDATE SET display_name=excluded.display_name, description=excluded.description, schema_json=excluded.schema_json"
	for _, module := range snapshot.Modules {
		data, err := json.Marshal(module)
		if err != nil {
			return zero, err
		}
		if _, err := tx.ExecContext(ctx, moduleSQL, module.ID, module.DisplayName, module.Description, string(data)); err != nil {
			return zero, fmt.Errorf("restoring module %q: %w", module.ID, err)
		}
	}
	if err := checkReadBudget(ctx, tx, table("modules"), encodedBytes("schema_json", postgres), maxSnapshotModules, MaxModuleReadBytes, 256<<10); err != nil {
		return zero, err
	}
	placeholders := make([]string, 9)
	for i := range placeholders {
		placeholders[i] = parameter(i + 1)
	}
	itemSQL := "INSERT INTO " + table("items") + " (id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at) VALUES (" + strings.Join(placeholders, ",") + ") ON CONFLICT(id) DO UPDATE SET module_id=excluded.module_id,title=excluded.title,purchase_price=excluded.purchase_price,images=excluded.images,tags=excluded.tags,attributes=excluded.attributes,created_at=excluded.created_at,updated_at=excluded.updated_at"
	for _, item := range snapshot.Items {
		if item.Images == nil {
			item.Images = []string{}
		}
		if item.Tags == nil {
			item.Tags = []string{}
		}
		if item.Attributes == nil {
			item.Attributes = map[string]any{}
		}
		images, _ := json.Marshal(item.Images)
		tags, _ := json.Marshal(item.Tags)
		attrs, _ := json.Marshal(item.Attributes)
		if _, err := tx.ExecContext(ctx, itemSQL, item.ID, item.ModuleID, item.Title, item.PurchasePrice, string(images), string(tags), string(attrs), item.CreatedAt, item.UpdatedAt); err != nil {
			return zero, fmt.Errorf("restoring item %q: %w", item.ID, err)
		}
	}
	settings := "{}"
	if len(snapshot.Settings) > 0 {
		settings = string(snapshot.Settings)
	}
	if mode == "merge" {
		var current string
		err := tx.QueryRowContext(ctx, "SELECT value FROM "+table("settings")+" WHERE key='settings'").Scan(&current)
		if err == sql.ErrNoRows {
			current = "{}"
		} else if err != nil {
			return zero, err
		}
		settings, err = mergeSettings(current, settings)
		if err != nil {
			return zero, err
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO "+table("settings")+" (key,value) VALUES ('settings',"+parameter(1)+") ON CONFLICT(key) DO UPDATE SET value=excluded.value", settings); err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	return RestoreResult{Items: len(snapshot.Items), Modules: len(snapshot.Modules)}, nil
}
