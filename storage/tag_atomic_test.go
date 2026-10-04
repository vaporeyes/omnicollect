// ABOUTME: Set-based tag writes span multiple pages and roll back completely on database failure.
// ABOUTME: Uses only isolated stores and an injected SQLite trigger to reject a mutation.
package storage

import (
	"fmt"
	"reflect"
	"testing"
)

func largeTagMutation(t *testing.T, s Store) {
	for i := 0; i < 600; i++ {
		if _, err := s.InsertItem(Item{ModuleID: "bulk-tags", Title: fmt.Sprintf("item%d", i), Tags: []string{"old", "new"}, Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	if n, err := s.RenameTag("old", "new"); err != nil || n != 600 {
		t.Fatalf("rename %d %v", n, err)
	}
	tags, err := s.GetAllTags()
	if err != nil || len(tags) != 1 || tags[0].Count != 600 {
		t.Fatalf("merged counts %v %v", tags, err)
	}
	if n, err := s.DeleteTag("new"); err != nil || n != 600 {
		t.Fatalf("delete %d %v", n, err)
	}
	tags, err = s.GetAllTags()
	if err != nil || len(tags) != 0 {
		t.Fatalf("delete incomplete %v %v", tags, err)
	}
}
func TestSQLiteLargeTagMutation(t *testing.T) { largeTagMutation(t, newTestStore(t)) }
func TestPostgresLargeTagMutation(t *testing.T) {
	s, _ := postgresTestStores(t)
	largeTagMutation(t, s)
}
func TestSQLiteTagMutationRollback(t *testing.T) {
	s := newTestStore(t)
	for _, title := range []string{"first", "reject"} {
		if _, err := s.InsertItem(Item{ModuleID: "tags", Title: title, Tags: []string{"old", "keep"}, Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.QueryItemPage("", "", "", "", 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`CREATE TRIGGER reject_tag_change BEFORE UPDATE OF tags ON items WHEN OLD.title='reject' BEGIN SELECT RAISE(ABORT,'injected'); END`); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func() (int64, error){func() (int64, error) { return s.RenameTag("old", "new") }, func() (int64, error) { return s.DeleteTag("old") }} {
		if _, err := mutate(); err == nil {
			t.Fatal("injected failure ignored")
		}
		after, err := s.QueryItemPage("", "", "", "", 200, 0)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatalf("partial mutation: %v %v", after, err)
		}
	}
}
