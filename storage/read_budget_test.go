// ABOUTME: Cross-backend encoded-byte/count limits fail before bulk metadata decoding.
// ABOUTME: Uses tiny helper limits and isolated oversized legacy fixtures, never production data.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func readBudgetContract(t *testing.T, s Store, db *sql.DB, table func(string) string, pg bool) {
	t.Helper()
	ctx := context.Background()
	if err := s.SaveModule(ModuleSchema{ID: "budget", DisplayName: "Budget"}); err != nil {
		t.Fatal(err)
	}
	item, err := s.InsertItem(Item{ModuleID: "budget", Title: "éé", Attributes: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Four encoded UTF-8 bytes must not pass a three-byte budget.
	if err = checkReadBudget(ctx, tx, table("items"), encodedBytes("title", pg), 1, 3, 0); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("byte budget %v", err)
	}
	if err = checkReadBudget(ctx, tx, table("items"), encodedBytes("title", pg), 1, 4, 4); err != nil {
		t.Fatal(err)
	}
	tx.Rollback()
	if _, err = s.InsertItem(Item{ModuleID: "budget", Title: "a", Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = checkReadBudget(ctx, tx, table("items"), encodedBytes("title", pg), 1, 0, 0); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("count budget %v", err)
	}
	if err = checkReadBudget(ctx, tx, table("items"), encodedBytes("title", pg), 2, 4, 4); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("aggregate budget %v", err)
	}
	if err = checkReadBudget(ctx, tx, table("items"), encodedBytes("title", pg), 2, 5, 4); err != nil {
		t.Fatal(err)
	}
	tx.Rollback()
	// Row preflight rejects data before unmarshalling oversized JSON metadata.
	raw, _ := json.Marshal(map[string]any{"large": strings.Repeat("x", MaxMetadataRowBytes)})
	if _, err = db.Exec("UPDATE "+table("items")+" SET attributes="+parameter(pg, 1)+" WHERE id="+parameter(pg, 2), string(raw), item.ID); err != nil {
		t.Fatal(err)
	}
	if snap, err := s.Snapshot(ctx); !errors.Is(err, ErrMetadataBudget) || snap.Items != nil {
		t.Fatalf("snapshot partial/budget %+v %v", snap, err)
	}
	if err = s.SaveModule(ModuleSchema{ID: "budget", DisplayName: "Changed"}); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("schema row budget %v", err)
	}
	if out, err := s.ExportItemsCSV([]string{item.ID}, nil); !errors.Is(err, ErrMetadataBudget) || out != "" {
		t.Fatalf("CSV budget %v", err)
	}
	if _, err = db.Exec("UPDATE " + table("items") + " SET attributes='{}'"); err != nil {
		t.Fatal(err)
	}
	oversized, _ := json.Marshal(ModuleSchema{ID: "budget", DisplayName: "Budget", Description: strings.Repeat("x", 256<<10)})
	if _, err = db.Exec("UPDATE "+table("modules")+" SET schema_json="+parameter(pg, 1), string(oversized)); err != nil {
		t.Fatal(err)
	}
	if modules, err := s.GetModules(); !errors.Is(err, ErrMetadataBudget) || modules != nil {
		t.Fatalf("modules partial/budget %v", err)
	}
	if _, err = s.Snapshot(ctx); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("snapshot module budget %v", err)
	}
	if err = s.SaveModule(ModuleSchema{ID: "budget", DisplayName: "Budget"}); err != nil {
		t.Fatal(err)
	}
	// Distinct tag output is bounded with a lookahead, never a silent subset.
	tags := make([]string, MaxTagResults+1)
	for i := range tags {
		tags[i] = fmt.Sprintf("tag%05d", i)
	}
	if _, err = s.InsertItem(Item{ModuleID: "budget", Title: "Many tags", Tags: tags, Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if result, err := s.GetAllTags(); !errors.Is(err, ErrMetadataBudget) || result != nil {
		t.Fatalf("tag truncation %d %v", len(result), err)
	}
	// Module count rejection is separate from total byte or per-module budgets.
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxSnapshotModules-1; i++ {
		id := fmt.Sprintf("mod%d", i)
		raw, _ := json.Marshal(ModuleSchema{ID: id, DisplayName: id})
		if _, err = tx.Exec("INSERT INTO "+table("modules")+" (id,display_name,description,schema_json) VALUES ("+parameter(pg, 1)+","+parameter(pg, 2)+",'',"+parameter(pg, 3)+")", id, id, string(raw)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveModule(ModuleSchema{ID: "overflow", DisplayName: "Overflow"}); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("module write budget %v", err)
	}
	if result, err := s.GetModules(); err != nil || len(result) != maxSnapshotModules {
		t.Fatalf("module write was not rolled back: %d %v", len(result), err)
	}
	incoming := Snapshot{Version: SnapshotVersion, Modules: []ModuleSchema{{ID: "overflow", DisplayName: "Overflow"}}, Settings: json.RawMessage(`{"overflow":true}`)}
	if _, err = s.Restore(ctx, incoming, "merge"); !errors.Is(err, ErrMetadataBudget) {
		t.Fatalf("merge catalog budget %v", err)
	}
	if result, err := s.GetModules(); err != nil || len(result) != maxSnapshotModules {
		t.Fatalf("merge catalog was not rolled back: %d %v", len(result), err)
	}
	if settings, err := s.GetSettings(); err != nil || strings.Contains(settings, `"overflow"`) {
		t.Fatalf("merge changed settings: %v %v", settings, err)
	}
	if _, err = db.Exec("INSERT INTO " + table("modules") + " (id,display_name,description,schema_json) VALUES ('overflow','Overflow','','{\"id\":\"overflow\",\"displayName\":\"Overflow\"}')"); err != nil {
		t.Fatal(err)
	}
	if result, err := s.GetModules(); !errors.Is(err, ErrMetadataBudget) || result != nil {
		t.Fatalf("module count truncation %v", err)
	}
}
func TestSQLiteReadBudgets(t *testing.T) {
	s := newTestStore(t)
	readBudgetContract(t, s, s.db, func(n string) string { return n }, false)
}
func TestPostgresReadBudgets(t *testing.T) {
	s, _ := postgresTestStores(t)
	readBudgetContract(t, s, s.db, s.table, true)
}
func TestCSVOutputBudgets(t *testing.T) {
	attributes := map[string]any{}
	for i := 0; i < MaxCSVColumns; i++ {
		attributes[fmt.Sprintf("key%d", i)] = "value"
	}
	if out, err := buildCSV([]Item{{Attributes: attributes}}, nil); !errors.Is(err, ErrMetadataBudget) || out != "" {
		t.Fatalf("columns %v", err)
	}
	if out, err := buildCSV([]Item{{Title: strings.Repeat("x", MaxCSVBytes)}}, nil); !errors.Is(err, ErrMetadataBudget) || out != "" {
		t.Fatalf("bytes %v", err)
	}
}
