// ABOUTME: Integration regression for migrating current SQLite metadata to PostgreSQL.
// ABOUTME: Verifies IDs, timestamps, modules, and settings come from the selected database.
package storage

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMigrationUsesDatabaseMetadata(t *testing.T) {
	destination, _ := postgresTestStores(t)
	path := filepath.Join(t.TempDir(), "source.db")
	source, err := NewSQLiteStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	want := backupFixture()
	if _, err := source.Restore(context.Background(), want, "replace"); err != nil {
		source.Close()
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	result, err := MigrateToPostgres(path, destination, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.ItemsMigrated != 2 || result.ModulesMigrated != 1 || len(result.Errors) != 0 {
		t.Fatalf("wrong migration result: %+v", result)
	}
	got, err := destination.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, got), semanticSnapshot(t, want)) {
		t.Fatalf("migration changed data: %+v", got)
	}
}
