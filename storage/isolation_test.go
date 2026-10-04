// ABOUTME: Regression tests for filesystem isolation of in-memory storage fixtures.
// ABOUTME: Ensures settings and modules never leak between stores or outlive Close.
package storage

import (
	"testing"
)

func TestInMemoryStoresIsolateFilesystem(t *testing.T) {
	a, err := NewSQLiteStoreInMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := NewSQLiteStoreInMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := a.SaveSettings(`{"theme":{"mode":"dark"}}`); err != nil {
		t.Fatal(err)
	}
	if err := a.SaveModule(ModuleSchema{ID: "isolated", DisplayName: "Isolated"}); err != nil {
		t.Fatal(err)
	}
	settings, err := b.GetSettings()
	if err != nil || settings != "{}" {
		t.Fatalf("settings leaked: %s, %v", settings, err)
	}
	modules, err := b.GetModules()
	if err != nil || len(modules) != 0 {
		t.Fatalf("modules leaked: %v, %v", modules, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetSettings(); err == nil {
		t.Fatal("closed store remained accessible")
	}
}
