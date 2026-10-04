// ABOUTME: Cross-backend public cover-image reference authorization contracts.
// ABOUTME: Checks module boundaries, additional-image privacy, cancellation, and tenant isolation.
package storage

import (
	"context"
	"testing"
)

func showcaseMediaContract(t *testing.T, store Store) {
	t.Helper()
	if err := store.SaveModule(ModuleSchema{ID: "books", DisplayName: "Books"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.InsertItem(Item{ModuleID: "books", Title: "Public cover", Images: []string{"cover.png", "receipt.png"}, Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		module, name string
		allowed      bool
	}{{"books", "cover.png", true}, {"books", "receipt.png", false}, {"other", "cover.png", false}, {"books", "unknown.png", false}} {
		allowed, err := store.HasShowcaseImage(context.Background(), tc.module, tc.name)
		if err != nil || allowed != tc.allowed {
			t.Fatalf("reference %+v: %v %v", tc, allowed, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.HasShowcaseImage(ctx, "books", "cover.png"); err == nil {
		t.Fatal("cancelled authorization query succeeded")
	}
}
func TestSQLiteShowcaseMediaContract(t *testing.T) { showcaseMediaContract(t, newTestStore(t)) }
func TestPostgresShowcaseMediaContract(t *testing.T) {
	a, b := postgresTestStores(t)
	showcaseMediaContract(t, a)
	allowed, err := b.HasShowcaseImage(context.Background(), "books", "cover.png")
	if err != nil || allowed {
		t.Fatalf("cross-tenant public media reference: %v %v", allowed, err)
	}
}
