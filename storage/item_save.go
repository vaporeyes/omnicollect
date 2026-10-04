// ABOUTME: Authoritative schema-aware item saves with cancellable transactions.
// ABOUTME: Uses updatedAt compare-and-swap to reject stale or unversioned edits.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrEditConflict = errors.New("item changed or was deleted; reload before saving your draft")
var ErrEditVersionRequired = errors.New("updatedAt is required when editing an item")

// ValidateItem preserves unknown attributes for evolving schemas, but validates
// every declared field and bounds the entire record before database work.
func ValidateItem(item Item, schema ModuleSchema) error {
	if item.ModuleID != schema.ID {
		return fmt.Errorf("item module does not match schema")
	}
	if strings.TrimSpace(item.Title) == "" || len(item.Title) > 4096 {
		return fmt.Errorf("title must contain 1-4096 bytes")
	}
	if len(item.ID) > 128 || len(item.Images) > 50 || len(item.Tags) > 100 || len(item.Attributes) > 256 {
		return fmt.Errorf("item exceeds field limits")
	}
	if item.PurchasePrice != nil && (math.IsNaN(*item.PurchasePrice) || math.IsInf(*item.PurchasePrice, 0) || *item.PurchasePrice < 0) {
		return fmt.Errorf("purchase price must be a finite nonnegative number")
	}
	for _, name := range item.Images {
		if err := ValidateFilename(name); err != nil {
			return err
		}
	}
	for _, tag := range item.Tags {
		if strings.TrimSpace(tag) == "" || len(tag) > 50 {
			return fmt.Errorf("tags must contain 1-50 bytes")
		}
	}
	for key := range item.Attributes {
		if unsafeAttributeName(key) {
			return fmt.Errorf("invalid attribute name %q", key)
		}
	}
	raw, err := json.Marshal(item.Attributes)
	if err != nil || len(raw) > 256<<10 {
		return fmt.Errorf("attributes must be valid JSON within 256 KB")
	}
	for _, field := range schema.Attributes {
		value := item.Attributes[field.Name]
		empty := value == nil
		if str, ok := value.(string); ok {
			empty = strings.TrimSpace(str) == ""
		}
		if empty {
			if field.Required {
				return fmt.Errorf("%s is required", field.Name)
			}
			continue
		}
		valid := false
		switch field.Type {
		case "string":
			_, valid = value.(string)
		case "boolean":
			_, valid = value.(bool)
		case "number":
			// Normalize trusted Go callers through JSON just as HTTP callers are decoded.
			encoded, e := json.Marshal(value)
			var number float64
			valid = e == nil && json.Unmarshal(encoded, &number) == nil && !math.IsNaN(number) && !math.IsInf(number, 0)
		case "date":
			if str, ok := value.(string); ok {
				_, err := time.Parse("2006-01-02", str)
				valid = err == nil
			}
		case "enum":
			if str, ok := value.(string); ok {
				for _, option := range field.Options {
					if str == option {
						valid = true
						break
					}
				}
			}
		}
		if !valid {
			return fmt.Errorf("%s must be a valid %s", field.Name, field.Type)
		}
	}
	return nil
}
func unsafeAttributeName(name string) bool {
	return name == "" || len(name) > 128 || name == "__proto__" || name == "constructor" || name == "prototype"
}

func (s *SQLiteStore) SaveItem(ctx context.Context, item Item) (Item, error) {
	return saveItem(ctx, s.db, false, func(name string) string { return name }, item)
}
func (s *PostgresStore) SaveItem(ctx context.Context, item Item) (Item, error) {
	if err := s.validateTenant(); err != nil {
		return Item{}, err
	}
	return saveItem(ctx, s.db, true, s.table, item)
}
func saveItem(ctx context.Context, db *sql.DB, postgres bool, table func(string) string, item Item) (Item, error) {
	expected := item.UpdatedAt
	if item.ID != "" {
		if expected == "" {
			return Item{}, ErrEditVersionRequired
		}
		if _, err := time.Parse(time.RFC3339Nano, expected); err != nil {
			return Item{}, fmt.Errorf("invalid edit version")
		}
	}
	p := func(n int) string {
		if postgres {
			return fmt.Sprintf("$%d", n)
		}
		return "?"
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Item{}, err
	}
	defer tx.Rollback()
	query := "SELECT schema_json FROM " + table("modules") + " WHERE id=" + p(1)
	if postgres {
		query += " FOR SHARE"
	}
	var schemaJSON string
	if err := tx.QueryRowContext(ctx, query, item.ModuleID).Scan(&schemaJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Item{}, fmt.Errorf("module does not exist")
		}
		return Item{}, err
	}
	var schema ModuleSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return Item{}, err
	}
	item.Tags = normalizeTags(item.Tags)
	if err := ValidateItem(item, schema); err != nil {
		return Item{}, err
	}
	item.Title = strings.TrimSpace(item.Title)
	if item.Images == nil {
		item.Images = []string{}
	}
	if item.Attributes == nil {
		item.Attributes = map[string]any{}
	}
	images, _ := json.Marshal(item.Images)
	tags, _ := json.Marshal(item.Tags)
	attrs, _ := json.Marshal(item.Attributes)
	now := time.Now().UTC().Truncate(time.Microsecond)
	if previous, err := time.Parse(time.RFC3339Nano, expected); err == nil && !now.After(previous) {
		now = previous.Add(time.Microsecond)
	}
	stamp := now.Format(time.RFC3339Nano)
	args := []any{item.ModuleID, item.Title, item.PurchasePrice, string(images), string(tags), string(attrs), stamp}
	if item.ID == "" {
		item.ID = uuid.NewString()
		args = append(args, item.ID, stamp)
		query = "INSERT INTO " + table("items") + " (module_id,title,purchase_price,images,tags,attributes,updated_at,id,created_at) VALUES ("
		for i := 1; i <= 9; i++ {
			if i > 1 {
				query += ","
			}
			query += p(i)
		}
		query += ")"
	} else {
		args = append(args, item.ID, expected)
		query = "UPDATE " + table("items") + " SET module_id=" + p(1) + ",title=" + p(2) + ",purchase_price=" + p(3) + ",images=" + p(4) + ",tags=" + p(5) + ",attributes=" + p(6) + ",updated_at=" + p(7) + " WHERE id=" + p(8) + " AND updated_at=" + p(9)
	}
	err = tx.QueryRowContext(ctx, query+" RETURNING created_at,updated_at", args...).Scan(&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, ErrEditConflict
	}
	if err != nil {
		return Item{}, err
	}
	if err := tx.Commit(); err != nil {
		return Item{}, err
	}
	return item, nil
}
