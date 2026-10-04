// ABOUTME: Cross-backend contracts for bounded page reads and independent summaries.
// ABOUTME: Covers stable ties, combined filters, tenant isolation, cancellation and overflow.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math"
	"reflect"
	"sort"
	"testing"
)

func pageContract(t *testing.T, s Store) {
	t.Helper()
	empty, err := s.QueryItemPage("", "", "", "", 3, 0)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 || empty.HasMore {
		t.Fatalf("empty: %+v %v", empty, err)
	}
	ids := []string{}
	price := 2.5
	for i := 0; i < 7; i++ {
		item, err := s.InsertItem(Item{ModuleID: "a", Title: "needle", PurchasePrice: &price, Tags: []string{"red"}, Attributes: map[string]any{"tier": 1}})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.ID)
	}
	// Force equal update times so the ID tie-breaker, rather than insertion timing, is exercised.
	switch db := s.(type) {
	case *SQLiteStore:
		_, err = db.db.Exec("UPDATE items SET updated_at = '2026-01-01T00:00:00Z'")
	case *PostgresStore:
		_, err = db.db.Exec("UPDATE " + db.table("items") + " SET updated_at = '2026-01-01T00:00:00Z'")
	}
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	for _, query := range []string{"", "needle"} {
		got := []string{}
		for offset := 0; offset < 7; offset += 3 {
			page, err := s.QueryItemPage(query, "a", `[{"field":"tier","op":"eq","value":1}]`, `["red"]`, 3, offset)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.Items) > 3 || page.HasMore != (offset < 6) || page.Offset != offset || page.Limit != 3 {
				t.Fatalf("page %+v", page)
			}
			for _, item := range page.Items {
				got = append(got, item.ID)
			}
		}
		if !reflect.DeepEqual(got, ids) {
			t.Fatalf("order/gaps: %v want %v", got, ids)
		}
	}
	filtered, err := s.QueryItemPage("", "other", "", "", 3, 0)
	if err != nil || len(filtered.Items) != 0 {
		t.Fatalf("scope %+v %v", filtered, err)
	}
	summary, err := s.CollectionSummary()
	if err != nil || summary.Items != 7 || summary.PricedItems != 7 || summary.PurchaseTotal == nil || *summary.PurchaseTotal != 17.5 || len(summary.Modules) != 1 {
		t.Fatalf("summary %+v %v", summary, err)
	}
	for _, bounds := range [][2]int{{0, 0}, {201, 0}, {1, -1}, {1, 1000001}} {
		if _, err := s.QueryItemPage("", "", "", "", bounds[0], bounds[1]); err == nil {
			t.Fatal("accepted invalid bounds")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.WithContext(ctx).QueryItemPage("", "", "", "", 1, 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("page cancel: %v", err)
	}
	if _, err := s.WithContext(ctx).CollectionSummary(); !errors.Is(err, context.Canceled) {
		t.Fatalf("summary cancel: %v", err)
	}
	max, negative := math.MaxFloat64, -1.0
	for _, p := range []*float64{nil, &negative, &max} {
		if _, err := s.InsertItem(Item{ModuleID: "b", Title: "prices", PurchasePrice: p, Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	summary, err = s.CollectionSummary()
	if err != nil || !summary.ValueAvailable || summary.PurchaseTotal == nil || *summary.PurchaseTotal != math.MaxFloat64 {
		t.Fatalf("finite maximum lost: %+v %v", summary, err)
	}
	if _, err := s.InsertItem(Item{ModuleID: "b", Title: "overflow", PurchasePrice: &max, Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	summary, err = s.CollectionSummary()
	if err != nil || summary.ValueAvailable || summary.PurchaseTotal != nil || summary.InvalidPrices != 1 || summary.PricedItems != 9 {
		t.Fatalf("overflow %+v %v", summary, err)
	}
	if _, err = json.Marshal(summary); err != nil {
		t.Fatal(err)
	}
}

func TestSQLiteReadContracts(t *testing.T) { pageContract(t, newTestStore(t)) }
func TestPostgresReadContracts(t *testing.T) {
	a, b := postgresTestStores(t)
	pageContract(t, a)
	if _, err := b.InsertItem(Item{Title: "private B", ModuleID: "b", Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	page, err := b.QueryItemPage("", "", "", "", 10, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Title != "private B" {
		t.Fatalf("tenant page %+v %v", page, err)
	}
	summary, err := b.CollectionSummary()
	if err != nil || summary.Items != 1 {
		t.Fatalf("tenant summary %+v %v", summary, err)
	}
}
func TestSummarySnapshotAndGroupBound(t *testing.T) {
	for _, backend := range []string{"sqlite", "postgres"} {
		t.Run(backend, func(t *testing.T) {
			var s Store
			if backend == "sqlite" {
				s = newTestStore(t)
			} else {
				s, _ = postgresTestStores(t)
			}
			var group errgroup.Group
			group.Go(func() error {
				for i := 0; i < 80; i++ {
					if _, err := s.InsertItem(Item{Title: "concurrent", ModuleID: fmt.Sprintf("m%d", i), Attributes: map[string]any{}}); err != nil {
						return err
					}
				}
				return nil
			})
			group.Go(func() error {
				for i := 0; i < 40; i++ {
					summary, err := s.CollectionSummary()
					if err != nil {
						return err
					}
					var count int64
					for _, m := range summary.Modules {
						count += m.Items
					}
					if count != summary.Items {
						return fmt.Errorf("mixed summary snapshot: %d != %d", count, summary.Items)
					}
				}
				return nil
			})
			if err := group.Wait(); err != nil {
				t.Fatal(err)
			}
			for i := 80; i < MaxSummaryModules+1; i++ {
				if _, err := s.InsertItem(Item{Title: "bounded", ModuleID: fmt.Sprintf("m%d", i), Attributes: map[string]any{}}); err != nil {
					t.Fatal(err)
				}
			}
			summary, err := s.CollectionSummary()
			if err != nil || !summary.ModulesTruncated || len(summary.Modules) != MaxSummaryModules || summary.Items != MaxSummaryModules+1 {
				t.Fatalf("group bound %+v %v", summary, err)
			}
		})
	}
}
