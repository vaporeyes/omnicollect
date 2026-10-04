// ABOUTME: Tag merge/delete contracts preserve order, deduplicate destinations and invalidate drafts.
// ABOUTME: Both database implementations must apply complete atomic mutations, not buffered partial batches.
package storage

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func tagMutationContract(t *testing.T, s Store) {
	t.Helper()
	if err := s.SaveModule(ModuleSchema{ID: "tags", DisplayName: "Tags"}); err != nil {
		t.Fatal(err)
	}
	for i, tags := range [][]string{{"old", "keep", "new"}, {"new", "keep", "old"}, {"keep", "old"}, {"old"}} {
		item, err := s.InsertItem(Item{Title: fmt.Sprintf("item%d", i), ModuleID: "tags", Tags: tags, Attributes: map[string]any{}})
		if err != nil {
			t.Fatal(err)
		}
		if n, err := s.RenameTag("old", " NEW "); err != nil || n != 1 {
			t.Fatalf("rename: %d %v", n, err)
		}
		page, err := s.QueryItemPage("", "tags", "", "", 200, 0)
		if err != nil {
			t.Fatal(err)
		}
		var changed Item
		for _, v := range page.Items {
			if v.ID == item.ID {
				changed = v
			}
		}
		want := [][]string{{"new", "keep"}, {"new", "keep"}, {"keep", "new"}, {"new"}}[i]
		if !reflect.DeepEqual(changed.Tags, want) {
			t.Fatalf("merge tags got %v want %v", changed.Tags, want)
		}
		if changed.UpdatedAt == item.UpdatedAt {
			t.Fatal("tag mutation did not invalidate edit version")
		}
	}
	counts, err := s.GetAllTags()
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range counts {
		if tag.Name == "new" && tag.Count != 4 {
			t.Fatalf("merged count %+v", tag)
		}
	}
	if n, err := s.RenameTag("missing", "new"); err != nil || n != 0 {
		t.Fatalf("missing %d %v", n, err)
	}
	if n, err := s.RenameTag("new", "new"); err != nil || n != 4 {
		t.Fatalf("self rename %d %v", n, err)
	}
	if n, err := s.DeleteTag("new"); err != nil || n != 4 {
		t.Fatalf("delete %d %v", n, err)
	}
	page, err := s.QueryItemPage("", "tags", "", "", 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range page.Items {
		for _, tag := range item.Tags {
			if tag == "new" {
				t.Fatal("deleted tag survived")
			}
		}
		if item.Tags == nil {
			t.Fatal("empty tags must be array")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.WithContext(ctx).RenameTag("keep", "changed"); err == nil {
		t.Fatal("cancelled mutation succeeded")
	}
}
func TestSQLiteTagMutationContract(t *testing.T) { tagMutationContract(t, newTestStore(t)) }
func TestPostgresTagMutationContract(t *testing.T) {
	s, _ := postgresTestStores(t)
	tagMutationContract(t, s)
}
