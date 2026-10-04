// ABOUTME: Legacy array reads return complete small results or an explicit paging requirement.
// ABOUTME: Both backends reject oversized results without returning a truncated array.
package storage

import (
	"errors"
	"testing"
)

func legacyReadContract(t *testing.T, s Store) {
	for i := 0; i < MaxItemPageSize+1; i++ {
		if _, err := s.InsertItem(Item{ModuleID: "legacy", Title: "Legacy", Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	items, err := s.QueryItems("", "legacy", "", "")
	if !errors.Is(err, ErrPaginationRequired) || items != nil {
		t.Fatalf("silent truncation %d %v", len(items), err)
	}
	page, err := s.QueryItemPage("", "legacy", "", "", MaxItemPageSize, 0)
	if err != nil || len(page.Items) != MaxItemPageSize || !page.HasMore {
		t.Fatalf("paging unavailable %+v %v", page, err)
	}
}
func TestSQLiteLegacyReadBound(t *testing.T) { legacyReadContract(t, newTestStore(t)) }
func TestPostgresLegacyReadBound(t *testing.T) {
	a, _ := postgresTestStores(t)
	legacyReadContract(t, a)
}
