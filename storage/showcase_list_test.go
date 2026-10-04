// ABOUTME: Showcase listings fail explicitly at their row budget without leaking other tenants.
// ABOUTME: Complete-at-limit and overflow contracts use isolated SQLite and disposable PostgreSQL.
package storage

import (
	"database/sql"
	"errors"
	"testing"
)

func showcaseListContract(t *testing.T, s Store, db *sql.DB, table, tenant string, pg bool) {
	t.Helper()
	query := "WITH RECURSIVE sequence(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM sequence WHERE n<" + parameter(pg, 1) + ") INSERT INTO " + table + " (id,slug,tenant_id,module_id,enabled,created_at,updated_at) SELECT " + parameter(pg, 2) + "||CAST(n AS TEXT)," + parameter(pg, 3) + "||CAST(n AS TEXT)," + parameter(pg, 4) + ",'module-'||CAST(n AS TEXT),FALSE,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z' FROM sequence"
	if _, err := db.Exec(query, MaxShowcaseResults+1, tenant+"-id-", tenant+"-slug-", tenant); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM "+table+" WHERE tenant_id="+parameter(pg, 1), tenant) })
	if result, err := s.ListShowcases(); !errors.Is(err, ErrMetadataBudget) || result != nil {
		t.Fatalf("overflow returned partial data: %d %v", len(result), err)
	}
	if _, err := db.Exec("DELETE FROM "+table+" WHERE id="+parameter(pg, 1), tenant+"-id-10001"); err != nil {
		t.Fatal(err)
	}
	if result, err := s.ListShowcases(); err != nil || len(result) != MaxShowcaseResults {
		t.Fatalf("complete at limit: %d %v", len(result), err)
	}
}
func TestSQLiteShowcaseListBudget(t *testing.T) {
	s := newTestStore(t)
	showcaseListContract(t, s, s.db, "showcases", "fixture", false)
}
func TestPostgresShowcaseListBudget(t *testing.T) {
	s, other := postgresTestStores(t)
	showcaseListContract(t, s, s.db, "public.showcases", s.TenantSchema(), true)
	if result, err := other.ListShowcases(); err != nil || len(result) != 0 {
		t.Fatalf("tenant leak: %d %v", len(result), err)
	}
}
