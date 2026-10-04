// ABOUTME: Integration tests against an explicitly configured disposable PostgreSQL database.
// ABOUTME: Exercises concurrent tenant isolation using the real connection pool and schema DDL.
package storage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func postgresTestStores(t *testing.T) (*PostgresStore, *PostgresStore) {
	t.Helper()
	dsn := os.Getenv("OMNICOLLECT_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("OMNICOLLECT_TEST_POSTGRES_URL is not configured")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("integration database name must end in _test")
	}
	root, err := NewPostgresStoreNoTenant(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	root.db.SetMaxOpenConns(8)
	prefix := "tenant_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	a, b := root.WithTenantSchema(prefix+"_a"), root.WithTenantSchema(prefix+"_b")
	for _, s := range []*PostgresStore{a, b} {
		schema := s.TenantSchema()
		if err := root.ProvisionTenant(schema); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := root.db.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
				t.Error(err)
			}
		})
	}
	return a, b
}

func TestPostgresConcurrentTenantIsolation(t *testing.T) {
	a, b := postgresTestStores(t)
	var wg sync.WaitGroup
	for n, s := range []*PostgresStore{a, b} {
		for worker := 0; worker < 4; worker++ {
			wg.Add(1)
			go func(n int, s *PostgresStore, worker int) {
				defer wg.Done()
				title := fmt.Sprintf("tenant-%d", n)
				for i := 0; i < 15; i++ {
					if _, err := s.InsertItem(Item{ModuleID: "test", Title: title, Attributes: map[string]any{}, Images: []string{}}); err != nil {
						t.Error(err)
						return
					}
					items, err := s.QueryItems("", "", "", "")
					if err != nil {
						t.Error(err)
						return
					}
					for _, item := range items {
						if item.Title != title {
							t.Errorf("tenant leak: got %q, want %q", item.Title, title)
							return
						}
					}
				}
			}(n, s, worker)
		}
	}
	wg.Wait()
	for _, s := range []*PostgresStore{a, b} {
		items, err := s.QueryItems("", "", "", "")
		if err != nil || len(items) != 60 {
			t.Fatalf("wrong tenant count: %d, %v", len(items), err)
		}
	}
}

func TestPostgresFilterFieldsAreData(t *testing.T) {
	s, _ := postgresTestStores(t)
	field := "value') OR 1=1 --"
	for _, v := range []string{"yes", "no"} {
		if _, err := s.InsertItem(Item{ModuleID: "test", Title: v, Attributes: map[string]any{field: v}, Images: []string{}}); err != nil {
			t.Fatal(err)
		}
	}
	filters, _ := json.Marshal([]attrFilter{{Field: field, Op: "eq", Value: "yes"}})
	items, err := s.QueryItems("", "", string(filters), "")
	if err != nil || len(items) != 1 || items[0].Title != "yes" {
		t.Fatalf("unsafe filter: %+v %v", items, err)
	}
}
