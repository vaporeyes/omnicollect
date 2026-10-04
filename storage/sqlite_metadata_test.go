// ABOUTME: Tests database-backed settings and modules, including file-backed persistence.
// ABOUTME: Guarantees the metadata needed for atomic backup/restore lives inside SQLite.
package storage

import (
	"path/filepath"
	"testing"
)

func TestSQLiteMetadataPersistsInDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection with spaces.db")
	s, err := NewSQLiteStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveModule(ModuleSchema{ID: "books", DisplayName: "Books"}); err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := s.SaveSettings(`{"theme":{"mode":"dark"}}`); err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = NewSQLiteStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	modules, err := s.GetModules()
	if err != nil || len(modules) != 1 || modules[0].ID != "books" {
		t.Fatalf("metadata not persisted: %+v %v", modules, err)
	}
	settings, err := s.GetSettings()
	if err != nil || settings != `{"theme":{"mode":"dark"}}` {
		t.Fatalf("settings not persisted: %s %v", settings, err)
	}
	for _, table := range []string{"modules", "settings"} {
		var count int
		if err := s.DB().QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("missing DB metadata: %s %d %v", table, count, err)
		}
	}
}
