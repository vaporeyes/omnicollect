// ABOUTME: Verifies snapshots observe one committed state across items and settings.
// ABOUTME: Exercises concurrent writers against both real PostgreSQL and file-backed SQLite.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func testSnapshotConsistency(t *testing.T, s Store, db *sql.DB, table func(string) string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	snapshot := backupFixture()
	snapshot.Items = snapshot.Items[:1]
	snapshot.Items[0].Title = "0"
	snapshot.Settings = json.RawMessage(`{"generation":0}`)
	if _, err := s.Restore(ctx, snapshot, "replace"); err != nil {
		t.Fatal(err)
	}
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		for n := 1; n <= 60; n++ {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, "UPDATE "+table("items")+" SET title=$1", strconv.Itoa(n))
			if err == nil {
				_, err = tx.ExecContext(ctx, "UPDATE "+table("settings")+" SET value=$1 WHERE key='settings'", fmt.Sprintf(`{"generation":%d}`, n))
			}
			if err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}
		}
		return nil
	})
	group.Go(func() error {
		for n := 0; n < 60; n++ {
			got, err := s.Snapshot(ctx)
			if err != nil {
				return err
			}
			var settings struct {
				Generation int `json:"generation"`
			}
			if err := json.Unmarshal(got.Settings, &settings); err != nil {
				return err
			}
			if len(got.Items) != 1 || got.Items[0].Title != strconv.Itoa(settings.Generation) {
				return fmt.Errorf("mixed snapshot: items=%+v settings=%s", got.Items, got.Settings)
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLiteSnapshotConsistency(t *testing.T) {
	s, err := NewSQLiteStoreAt(filepath.Join(t.TempDir(), "collection.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	testSnapshotConsistency(t, s, s.DB(), func(name string) string { return name })
}

func TestPostgresSnapshotConsistency(t *testing.T) {
	s, _ := postgresTestStores(t)
	testSnapshotConsistency(t, s, s.DB(), s.table)
}
