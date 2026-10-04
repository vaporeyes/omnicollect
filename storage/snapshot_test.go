// ABOUTME: Shared backup/restore contract tests for SQLite and PostgreSQL.
// ABOUTME: Exercises identity preservation, idempotence, cancellation, and rollback after actual database failures.
package storage

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func backupFixture() Snapshot {
	return Snapshot{Version: SnapshotVersion, Modules: []ModuleSchema{{ID: "books", DisplayName: "Books", Attributes: []AttributeSchema{{Name: "author", Type: "string"}}}},
		Settings: json.RawMessage(`{"theme":{"mode":"dark"},"smartFolders":[{"id":"saved"}]}`),
		Items: []Item{
			{ID: "original-one", ModuleID: "books", Title: "Restored first", Images: []string{"image.jpg"}, Tags: []string{"rare"}, Attributes: map[string]any{"author": "Ada"}, CreatedAt: "2020-01-02T03:04:05.123456Z", UpdatedAt: "2021-02-03T04:05:06.234567Z"},
			{ID: "original-two", ModuleID: "books", Title: "Restored second", Images: []string{}, Tags: []string{}, Attributes: map[string]any{}, CreatedAt: "2019-01-01T00:00:00Z", UpdatedAt: "2022-01-01T00:00:00Z"},
		}}
}

func semanticSnapshot(t *testing.T, snapshot Snapshot) any {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func testSnapshotRoundTrip(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	if err := s.SaveModule(ModuleSchema{ID: "obsolete", DisplayName: "Obsolete"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertItem(Item{ModuleID: "obsolete", Title: "Old", Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(`{"old":true}`); err != nil {
		t.Fatal(err)
	}
	want := backupFixture()
	result, err := s.Restore(ctx, want, "replace")
	if err != nil {
		t.Fatal(err)
	}
	if result.Items != 2 || result.Modules != 1 {
		t.Fatalf("wrong committed counts: %+v", result)
	}
	got, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, got), semanticSnapshot(t, want)) {
		t.Fatalf("round trip mismatch\ngot: %+v\nwant: %+v", got, want)
	}
	for i := 0; i < 2; i++ {
		if _, err := s.Restore(ctx, want, "merge"); err != nil {
			t.Fatal(err)
		}
	}
	got, err = s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, got), semanticSnapshot(t, want)) {
		t.Fatalf("merge changed identities/history or duplicated data: %+v", got)
	}
	items, err := s.QueryItems("Ada", "", "", "")
	if err != nil || len(items) != 1 || items[0].ID != "original-one" {
		t.Fatalf("restored search index broken: %+v %v", items, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.Restore(cancelled, Snapshot{}, "replace"); err == nil {
		t.Fatal("cancelled restore succeeded")
	}
	got, err = s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, got), semanticSnapshot(t, want)) {
		t.Fatal("cancellation lost data")
	}
}

func TestSQLiteSnapshotRoundTrip(t *testing.T) { testSnapshotRoundTrip(t, newTestStore(t)) }
func TestPostgresSnapshotRoundTrip(t *testing.T) {
	s, _ := postgresTestStores(t)
	testSnapshotRoundTrip(t, s)
}

func testRestoreRollback(t *testing.T, s Store) {
	ctx := context.Background()
	if err := s.SaveModule(ModuleSchema{ID: "original", DisplayName: "Original"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertItem(Item{ModuleID: "original", Title: "Keep me", Attributes: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(`{"keep":true}`); err != nil {
		t.Fatal(err)
	}
	before, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	incoming := backupFixture()
	incoming.Items[1].Title = "Fail"
	// The test database rejects the SECOND inserted item after deletes and a first insert.
	if result, err := s.Restore(ctx, incoming, "replace"); err == nil || result.Items != 0 {
		t.Fatalf("failed replace reported success: %+v %v", result, err)
	}
	after, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, after), semanticSnapshot(t, before)) {
		t.Fatalf("replace failure lost original data: %+v", after)
	}
	if result, err := s.Restore(ctx, incoming, "merge"); err == nil || result.Items != 0 {
		t.Fatalf("failed merge reported success: %+v %v", result, err)
	}
	after, err = s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(semanticSnapshot(t, after), semanticSnapshot(t, before)) {
		t.Fatal("merge partially committed")
	}
}

func TestSQLiteRestoreRollback(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.db.Exec(`CREATE TRIGGER fail_restore BEFORE INSERT ON items WHEN new.title='Fail' BEGIN SELECT RAISE(ABORT, 'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	testRestoreRollback(t, s)
}

func TestPostgresRestoreRollback(t *testing.T) {
	s, _ := postgresTestStores(t)
	if _, err := s.db.Exec("ALTER TABLE " + s.table("items") + " ADD CONSTRAINT fail_restore CHECK(title <> 'Fail')"); err != nil {
		t.Fatal(err)
	}
	testRestoreRollback(t, s)
}

func TestSnapshotRejectsAmbiguity(t *testing.T) {
	s := newTestStore(t)
	for _, kind := range []string{"duplicate", "module", "timestamp", "filename", "version", "settings"} {
		snapshot := backupFixture()
		switch kind {
		case "duplicate":
			snapshot.Items = append(snapshot.Items, snapshot.Items[0])
		case "module":
			snapshot.Modules[0].ID = "../escape"
		case "timestamp":
			snapshot.Items[0].CreatedAt = "yesterday"
		case "filename":
			snapshot.Items[0].Images = []string{"../secret"}
		case "version":
			snapshot.Version = 999
		case "settings":
			snapshot.Settings = json.RawMessage(`null`)
		}
		if _, err := s.Restore(context.Background(), snapshot, "replace"); err == nil {
			t.Errorf("accepted %s", kind)
		}
	}
}
