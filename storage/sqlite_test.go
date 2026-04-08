// ABOUTME: Unit tests for SQLiteStore covering all Store interface methods.
// ABOUTME: Uses in-memory SQLite databases for fast, isolated test execution.
package storage

import (
	"encoding/json"
	"strings"
	"testing"
)

// newTestStore creates an in-memory SQLiteStore for testing.
// Registers t.Cleanup to close the database when the test completes.
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	store, err := NewSQLiteStoreInMemory()
	if err != nil {
		t.Fatalf("creating test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// --- Item CRUD ---

func TestInsertItem(t *testing.T) {
	store := newTestStore(t)

	price := 29.99
	item := Item{
		ModuleID:      "comics",
		Title:         "Amazing Spider-Man #1",
		PurchasePrice: &price,
		Images:        []string{"img1.jpg"},
		Attributes:    map[string]any{"condition": "Mint", "year": float64(1963)},
	}

	saved, err := store.InsertItem(item)
	if err != nil {
		t.Fatalf("InsertItem: %v", err)
	}
	if saved.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if saved.Title != item.Title {
		t.Errorf("title: got %q, want %q", saved.Title, item.Title)
	}
	if saved.CreatedAt == "" || saved.UpdatedAt == "" {
		t.Error("expected timestamps to be set")
	}

	// Query back and verify
	items, err := store.QueryItems("", "comics", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	got := items[0]
	if got.ID != saved.ID {
		t.Errorf("ID: got %q, want %q", got.ID, saved.ID)
	}
	if *got.PurchasePrice != price {
		t.Errorf("price: got %v, want %v", *got.PurchasePrice, price)
	}
	if len(got.Images) != 1 || got.Images[0] != "img1.jpg" {
		t.Errorf("images: got %v, want [img1.jpg]", got.Images)
	}
	if got.Attributes["condition"] != "Mint" {
		t.Errorf("condition: got %v, want Mint", got.Attributes["condition"])
	}
}

func TestUpdateItem(t *testing.T) {
	store := newTestStore(t)

	price := 10.0
	item := Item{
		ModuleID:      "comics",
		Title:         "Original Title",
		PurchasePrice: &price,
		Images:        []string{},
		Attributes:    map[string]any{"condition": "Poor"},
	}
	saved, err := store.InsertItem(item)
	if err != nil {
		t.Fatalf("InsertItem: %v", err)
	}

	// Update title and attributes
	saved.Title = "Updated Title"
	saved.Attributes = map[string]any{"condition": "Mint", "year": float64(2020)}
	updated, err := store.UpdateItem(saved)
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("title: got %q, want %q", updated.Title, "Updated Title")
	}

	// Query back and verify
	items, err := store.QueryItems("", "comics", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Updated Title" {
		t.Errorf("persisted title: got %q", items[0].Title)
	}
	if items[0].Attributes["year"] != float64(2020) {
		t.Errorf("persisted year: got %v", items[0].Attributes["year"])
	}
}

func TestDeleteItem(t *testing.T) {
	store := newTestStore(t)

	item := Item{
		ModuleID:   "comics",
		Title:      "To Be Deleted",
		Images:     []string{},
		Attributes: map[string]any{},
	}
	saved, err := store.InsertItem(item)
	if err != nil {
		t.Fatalf("InsertItem: %v", err)
	}

	if err := store.DeleteItem(saved.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}

	items, err := store.QueryItems("", "", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestDeleteItem_NotFound(t *testing.T) {
	store := newTestStore(t)

	err := store.DeleteItem("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}

// --- Search / Query ---

func TestQueryItems_TextSearch(t *testing.T) {
	store := newTestStore(t)

	titles := []string{"Amazing Spider-Man", "Batman Returns", "Spider-Man 2099"}
	for _, title := range titles {
		_, err := store.InsertItem(Item{
			ModuleID: "comics", Title: title, Images: []string{}, Attributes: map[string]any{},
		})
		if err != nil {
			t.Fatalf("InsertItem %q: %v", title, err)
		}
	}

	results, err := store.QueryItems("Spider", "", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for 'Spider', got %d", len(results))
	}
	for _, r := range results {
		if !strings.Contains(r.Title, "Spider") {
			t.Errorf("unexpected result: %q", r.Title)
		}
	}
}

func TestQueryItems_ModuleFilter(t *testing.T) {
	store := newTestStore(t)

	_, _ = store.InsertItem(Item{ModuleID: "comics", Title: "Comic A", Images: []string{}, Attributes: map[string]any{}})
	_, _ = store.InsertItem(Item{ModuleID: "coins", Title: "Coin A", Images: []string{}, Attributes: map[string]any{}})
	_, _ = store.InsertItem(Item{ModuleID: "comics", Title: "Comic B", Images: []string{}, Attributes: map[string]any{}})

	results, err := store.QueryItems("", "coins", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for module 'coins', got %d", len(results))
	}
	if results[0].Title != "Coin A" {
		t.Errorf("title: got %q, want %q", results[0].Title, "Coin A")
	}
}

func TestQueryItems_NoResults(t *testing.T) {
	store := newTestStore(t)

	_, _ = store.InsertItem(Item{ModuleID: "comics", Title: "Something", Images: []string{}, Attributes: map[string]any{}})

	results, err := store.QueryItems("zzzznotfound", "", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// --- Attribute Filters ---

func TestQueryItems_EnumFilter(t *testing.T) {
	store := newTestStore(t)

	conditions := []string{"Mint", "Fine", "Poor", "Mint"}
	for i, cond := range conditions {
		_, _ = store.InsertItem(Item{
			ModuleID: "comics", Title: "Item " + cond,
			Images: []string{}, Attributes: map[string]any{"condition": cond, "idx": float64(i)},
		})
	}

	filters := `[{"field":"condition","op":"in","values":["Mint","Fine"]}]`
	results, err := store.QueryItems("", "", filters, "")
	if err != nil {
		t.Fatalf("QueryItems with enum filter: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results (2 Mint + 1 Fine), got %d", len(results))
	}
}

func TestQueryItems_BooleanFilter(t *testing.T) {
	store := newTestStore(t)

	_, _ = store.InsertItem(Item{ModuleID: "comics", Title: "Graded", Images: []string{}, Attributes: map[string]any{"isGraded": true}})
	_, _ = store.InsertItem(Item{ModuleID: "comics", Title: "Ungraded", Images: []string{}, Attributes: map[string]any{"isGraded": false}})

	filters := `[{"field":"isGraded","op":"eq","value":true}]`
	results, err := store.QueryItems("", "", filters, "")
	if err != nil {
		t.Fatalf("QueryItems with boolean filter: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Graded" {
		t.Errorf("title: got %q, want %q", results[0].Title, "Graded")
	}
}

func TestQueryItems_NumberRangeFilter(t *testing.T) {
	store := newTestStore(t)

	years := []float64{1960, 1975, 1990, 2020}
	for i, yr := range years {
		_, _ = store.InsertItem(Item{
			ModuleID: "comics", Title: "Item " + string(rune('A'+i)),
			Images: []string{}, Attributes: map[string]any{"year": yr},
		})
	}

	filters := `[{"field":"year","op":"gte","value":1970},{"field":"year","op":"lte","value":2000}]`
	results, err := store.QueryItems("", "", filters, "")
	if err != nil {
		t.Fatalf("QueryItems with range filter: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (1975, 1990), got %d", len(results))
	}
}

func TestQueryItems_CombinedFilters(t *testing.T) {
	store := newTestStore(t)

	// Insert items with different combos of condition and year
	data := []struct {
		title     string
		condition string
		year      float64
	}{
		{"Mint 1960", "Mint", 1960},
		{"Mint 1990", "Mint", 1990},
		{"Fine 1990", "Fine", 1990},
		{"Poor 1990", "Poor", 1990},
	}
	for _, d := range data {
		_, _ = store.InsertItem(Item{
			ModuleID: "comics", Title: d.title,
			Images: []string{}, Attributes: map[string]any{"condition": d.condition, "year": d.year},
		})
	}

	// Filter: condition IN (Mint, Fine) AND year >= 1980
	filters := `[{"field":"condition","op":"in","values":["Mint","Fine"]},{"field":"year","op":"gte","value":1980}]`
	results, err := store.QueryItems("", "", filters, "")
	if err != nil {
		t.Fatalf("QueryItems combined: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (Mint 1990, Fine 1990), got %d", len(results))
	}
}

// --- Batch Operations ---

func TestDeleteItems_Batch(t *testing.T) {
	store := newTestStore(t)

	var ids []string
	for i := 0; i < 3; i++ {
		saved, err := store.InsertItem(Item{
			ModuleID: "comics", Title: "Batch " + string(rune('A'+i)),
			Images: []string{}, Attributes: map[string]any{},
		})
		if err != nil {
			t.Fatalf("InsertItem: %v", err)
		}
		ids = append(ids, saved.ID)
	}

	deleted, err := store.DeleteItems(ids[:2])
	if err != nil {
		t.Fatalf("DeleteItems: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted count: got %d, want 2", deleted)
	}

	remaining, err := store.QueryItems("", "", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining item, got %d", len(remaining))
	}
}

func TestBulkUpdateModule(t *testing.T) {
	store := newTestStore(t)

	var ids []string
	for i := 0; i < 2; i++ {
		saved, err := store.InsertItem(Item{
			ModuleID: "old-module", Title: "Item " + string(rune('A'+i)),
			Images: []string{}, Attributes: map[string]any{},
		})
		if err != nil {
			t.Fatalf("InsertItem: %v", err)
		}
		ids = append(ids, saved.ID)
	}

	updated, err := store.BulkUpdateModule(ids, "new-module")
	if err != nil {
		t.Fatalf("BulkUpdateModule: %v", err)
	}
	if updated != 2 {
		t.Errorf("updated count: got %d, want 2", updated)
	}

	items, err := store.QueryItems("", "new-module", "", "")
	if err != nil {
		t.Fatalf("QueryItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items in new-module, got %d", len(items))
	}
}

// --- CSV Export ---

func TestExportItemsCSV(t *testing.T) {
	store := newTestStore(t)

	price1 := 10.0
	saved1, _ := store.InsertItem(Item{
		ModuleID: "comics", Title: "Comic One", PurchasePrice: &price1,
		Images: []string{}, Attributes: map[string]any{"condition": "Mint"},
	})
	saved2, _ := store.InsertItem(Item{
		ModuleID: "coins", Title: "Coin One",
		Images: []string{}, Attributes: map[string]any{"year": float64(1990)},
	})

	modules := []ModuleSchema{
		{ID: "comics", DisplayName: "Comics"},
		{ID: "coins", DisplayName: "Coins"},
	}

	csv, err := store.ExportItemsCSV([]string{saved1.ID, saved2.ID}, modules)
	if err != nil {
		t.Fatalf("ExportItemsCSV: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header + 2 data), got %d", len(lines))
	}

	header := lines[0]
	if !strings.Contains(header, "id") || !strings.Contains(header, "title") {
		t.Errorf("header missing expected columns: %s", header)
	}
	if !strings.Contains(header, "condition") || !strings.Contains(header, "year") {
		t.Errorf("header missing attribute columns: %s", header)
	}
}

func TestExportItemsCSV_Empty(t *testing.T) {
	store := newTestStore(t)

	csv, err := store.ExportItemsCSV([]string{}, nil)
	if err != nil {
		t.Fatalf("ExportItemsCSV: %v", err)
	}
	if csv != "" {
		t.Errorf("expected empty string for empty IDs, got %q", csv)
	}
}

// --- Modules ---

func TestGetModules_Empty(t *testing.T) {
	store := newTestStore(t)

	// GetModules reads from disk, which in a test context may or may not have
	// modules. We just verify it returns without error and gives a slice.
	modules, err := store.GetModules()
	if err != nil {
		t.Fatalf("GetModules: %v", err)
	}
	if modules == nil {
		t.Error("expected non-nil slice")
	}
}

func TestSaveAndGetModules(t *testing.T) {
	store := newTestStore(t)

	schema := ModuleSchema{
		ID:          "test-mod",
		DisplayName: "Test Module",
		Attributes: []AttributeSchema{
			{Name: "color", Type: "string"},
		},
	}

	if err := store.SaveModule(schema); err != nil {
		t.Fatalf("SaveModule: %v", err)
	}

	// Verify we can load it back
	modules, err := store.GetModules()
	if err != nil {
		t.Fatalf("GetModules: %v", err)
	}

	found := false
	for _, m := range modules {
		if m.ID == "test-mod" {
			found = true
			if m.DisplayName != "Test Module" {
				t.Errorf("displayName: got %q, want %q", m.DisplayName, "Test Module")
			}
		}
	}
	if !found {
		t.Error("saved module not found in GetModules result")
	}
}

// --- Settings ---

func TestGetSettings_Empty(t *testing.T) {
	store := newTestStore(t)

	settings, err := store.GetSettings()
	// Either returns empty JSON or an error; both are acceptable
	if err != nil && settings != "{}" {
		t.Fatalf("GetSettings: unexpected error: %v", err)
	}
	if settings == "" {
		t.Error("expected non-empty settings string (at least '{}')")
	}
}

func TestSaveAndGetSettings(t *testing.T) {
	store := newTestStore(t)

	input := `{"theme":"dark","language":"en"}`
	if err := store.SaveSettings(input); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	got, err := store.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	// The store may pretty-print the JSON, so compare parsed values
	var expected, actual map[string]any
	json.Unmarshal([]byte(input), &expected)
	json.Unmarshal([]byte(got), &actual)

	if actual["theme"] != expected["theme"] || actual["language"] != expected["language"] {
		t.Errorf("settings mismatch: got %v, want %v", actual, expected)
	}
}
