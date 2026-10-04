// ABOUTME: Shared authoritative-save, stale-edit, and cancellation regression contracts.
// ABOUTME: Runs identical assertions against isolated SQLite and disposable PostgreSQL.
package storage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func itemSaveContract(t *testing.T, store Store) {
	t.Helper()
	ctx := context.Background()
	schema := ModuleSchema{ID: "books", DisplayName: "Books", Attributes: []AttributeSchema{{Name: "condition", Type: "enum", Required: true, Options: []string{"good", "poor"}}, {Name: "count", Type: "number"}, {Name: "date", Type: "date"}}}
	if err := store.SaveModule(schema); err != nil {
		t.Fatal(err)
	}
	draft := Item{ModuleID: "books", Title: "  Example  ", Attributes: map[string]any{"condition": "good", "legacy": "keep"}}
	saved, err := store.SaveItem(ctx, draft)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Title != "Example" || saved.ID == "" || saved.UpdatedAt == "" || saved.Attributes["legacy"] != "keep" {
		t.Fatalf("invalid saved item: %+v", saved)
	}
	noVersion := saved
	noVersion.UpdatedAt = ""
	if _, err := store.SaveItem(ctx, noVersion); !errors.Is(err, ErrEditVersionRequired) {
		t.Fatalf("missing edit version accepted: %v", err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, title := range []string{"first", "second"} {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			edit := saved
			edit.Title = title
			_, err := store.SaveItem(ctx, edit)
			results <- err
		}(title)
	}
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ErrEditConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("lost-update guard failed: wins=%d conflicts=%d", wins, conflicts)
	}
	items, err := store.QueryItems("", "books", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].CreatedAt != saved.CreatedAt || items[0].UpdatedAt == saved.UpdatedAt {
		t.Fatal("history/version lost")
	}
	current := items[0]
	current.Title = "next edit"
	current, err = store.SaveItem(ctx, current)
	if err != nil {
		t.Fatalf("round-trip version rejected: %v", err)
	}
	current.Tags = []string{"old"}
	current, err = store.SaveItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RenameTag("old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveItem(ctx, current); !errors.Is(err, ErrEditConflict) {
		t.Fatalf("tag mutation did not invalidate edit: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := store.SaveItem(cancelled, draft); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel ignored: %v", err)
	}
	for _, attributes := range []map[string]any{{}, {"condition": "invalid"}, {"condition": "good", "count": "three"}, {"condition": "good", "date": "2024-02-31"}, {"condition": "good", "__proto__": "bad"}} {
		invalid := draft
		invalid.Attributes = attributes
		if _, err := store.SaveItem(ctx, invalid); err == nil {
			t.Fatalf("invalid attributes accepted: %+v", attributes)
		}
	}
	invalid := draft
	invalid.Tags = []string{strings.Repeat("Ⱥ", 20)} // Lowercase expands UTF-8 from 40 to 60 bytes.
	if _, err := store.SaveItem(ctx, invalid); err == nil {
		t.Fatal("expanded Unicode tag must be rejected, not byte-truncated")
	}
	invalid.Tags = nil
	invalid.ModuleID = "missing"
	if _, err := store.SaveItem(ctx, invalid); err == nil {
		t.Fatal("missing module accepted")
	}
}
func TestSQLiteAuthoritativeSave(t *testing.T) { itemSaveContract(t, newTestStore(t)) }
func TestPostgresAuthoritativeSave(t *testing.T) {
	a, _ := postgresTestStores(t)
	itemSaveContract(t, a)
}
