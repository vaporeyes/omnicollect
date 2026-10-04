// ABOUTME: Regression coverage for SQL filter parameters and filesystem boundaries.
// ABOUTME: Malicious names remain data or are rejected before filesystem side effects.
package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFilterFieldsAreData(t *testing.T) {
	store := newTestStore(t)
	field := "value') OR 1=1 --"
	_, err := store.InsertItem(Item{ModuleID: "test", Title: "Match", Attributes: map[string]any{field: "yes"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.InsertItem(Item{ModuleID: "test", Title: "Other", Attributes: map[string]any{field: "no"}})
	if err != nil {
		t.Fatal(err)
	}
	filters, _ := json.Marshal([]attrFilter{{Field: field, Op: "eq", Value: "yes"}})
	items, err := store.QueryItems("", "", string(filters), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "Match" {
		t.Fatalf("filter escaped: %+v", items)
	}
}

func TestModuleIDsCannotTraverse(t *testing.T) {
	store := newTestStore(t)
	for _, id := range []string{"../escaped", "a/b", `a\b`, ".", "", "/tmp/escaped"} {
		if err := store.SaveModule(ModuleSchema{ID: id, DisplayName: "Unsafe"}); err == nil {
			t.Errorf("accepted %q", id)
		}
	}
	modules, err := store.GetModules()
	if err != nil || len(modules) != 0 {
		t.Fatalf("invalid modules persisted: %v, %v", modules, err)
	}
}

func TestMediaFilesCannotEscape(t *testing.T) {
	dir := t.TempDir()
	media := NewLocalMediaStoreAt(filepath.Join(dir, "media"))
	if err := media.SaveOriginal(context.Background(), "../outside", []byte("bad")); err == nil {
		t.Fatal("accepted traversal")
	}
	if err := media.SaveOriginal(context.Background(), "test.jpg", []byte("safe")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(dir, "outside")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(media.BaseDir(), "originals", "link.jpg")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFileAt(filepath.Join(media.BaseDir(), "originals"), "link.jpg", 100); err == nil {
		t.Fatal("read escaped symlink")
	}
}
