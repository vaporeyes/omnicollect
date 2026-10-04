// ABOUTME: Shared bounded batch and schema compatibility validation inside write transactions.
// ABOUTME: Rejects missing targets and incompatible schema changes before any records are changed.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateQuery rejects malformed tag filters instead of silently broadening searches.
func ValidateQuery(query, moduleID, filters, tags string) error {
	if len(query) > 4096 || len(filters) > 65536 || len(tags) > 8192 {
		return fmt.Errorf("query exceeds size limits")
	}
	if moduleID != "" && !safeModuleID.MatchString(moduleID) {
		return fmt.Errorf("invalid module id")
	}
	if _, err := parseFilters(filters); err != nil {
		return err
	}
	if tags != "" {
		var names []string
		if err := json.Unmarshal([]byte(tags), &names); err != nil || names == nil || len(names) > 100 {
			return fmt.Errorf("tags must be an array of at most 100 names")
		}
		for _, name := range names {
			if strings.TrimSpace(name) == "" || len(name) > 50 {
				return fmt.Errorf("invalid tag filter")
			}
		}
	}
	return nil
}
func validateIDs(ids []string) error {
	if len(ids) > 500 {
		return fmt.Errorf("select at most 500 items per operation")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" || len(id) > 128 || seen[id] {
			return fmt.Errorf("invalid or duplicate item id")
		}
		seen[id] = true
	}
	return nil
}
func parameter(postgres bool, n int) string {
	if postgres {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

func validateReassignment(ctx context.Context, tx *sql.Tx, table func(string) string, postgres bool, ids []string, moduleID string) error {
	var raw string
	query := "SELECT schema_json FROM " + table("modules") + " WHERE id=" + parameter(postgres, 1)
	if postgres {
		query += " FOR SHARE"
	}
	if err := tx.QueryRowContext(ctx, query, moduleID).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("target module does not exist")
		}
		return err
	}
	var schema ModuleSchema
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return err
	}
	args := make([]any, len(ids))
	holders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		holders[i] = parameter(postgres, i+1)
	}
	if err := checkReadBudget(ctx, tx, table("items")+" WHERE id IN ("+strings.Join(holders, ",")+")", itemReadBytes(postgres), 500, 8<<20, MaxMetadataRowBytes, args...); err != nil {
		return err
	}
	query = "SELECT id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at FROM " + table("items") + " WHERE id IN (" + strings.Join(holders, ",") + ") ORDER BY id"
	if postgres {
		query += " FOR UPDATE"
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	items, err := scanSnapshotItems(rows)
	rows.Close()
	if err != nil {
		return err
	}
	if len(items) != len(ids) {
		return fmt.Errorf("one or more selected items no longer exist")
	}
	for _, item := range items {
		item.ModuleID = moduleID
		if err := ValidateItem(item, schema); err != nil {
			return fmt.Errorf("item %q is incompatible: %w", item.ID, err)
		}
	}
	return nil
}
func validateSchemaItems(ctx context.Context, tx *sql.Tx, table func(string) string, postgres bool, schema ModuleSchema, skip map[string]bool) error {
	if err := checkReadBudget(ctx, tx, table("items")+" WHERE module_id="+parameter(postgres, 1), itemReadBytes(postgres), 0, 0, MaxMetadataRowBytes, schema.ID); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at FROM "+table("items")+" WHERE module_id="+parameter(postgres, 1), schema.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanSnapshotItem(rows)
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if skip[item.ID] {
			continue
		}
		if err := ValidateItem(item, schema); err != nil {
			return fmt.Errorf("schema would invalidate item %q: %w", item.ID, err)
		}
	}
	return rows.Err()
}
func saveModule(ctx context.Context, db *sql.DB, table func(string) string, postgres bool, schema ModuleSchema) error {
	if err := ValidateModuleSchema(&schema); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if postgres {
		// Block new item saves before taking the item lock, preserving lock order.
		if _, err := tx.ExecContext(ctx, "LOCK TABLE "+table("modules")+" IN EXCLUSIVE MODE"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "LOCK TABLE "+table("items")+" IN SHARE MODE"); err != nil {
			return err
		}
	}
	if err := validateSchemaItems(ctx, tx, table, postgres, schema, nil); err != nil {
		return err
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	query := "INSERT INTO " + table("modules") + " (id,display_name,description,schema_json) VALUES (" + parameter(postgres, 1) + "," + parameter(postgres, 2) + "," + parameter(postgres, 3) + "," + parameter(postgres, 4) + ") ON CONFLICT(id) DO UPDATE SET display_name=excluded.display_name,description=excluded.description,schema_json=excluded.schema_json"
	if postgres {
		query += ",updated_at=NOW()"
	}
	if _, err := tx.ExecContext(ctx, query, schema.ID, schema.DisplayName, schema.Description, string(raw)); err != nil {
		return err
	}
	if err := checkReadBudget(ctx, tx, table("modules"), encodedBytes("schema_json", postgres), maxSnapshotModules, MaxModuleReadBytes, 256<<10); err != nil {
		return err
	}
	return tx.Commit()
}
