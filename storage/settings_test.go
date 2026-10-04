// ABOUTME: Shared regression tests for top-level settings and concurrent partial updates.
// ABOUTME: Both backends preserve unrelated keys and reject non-object JSON.
package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func testSettingsContract(t *testing.T, s Store) {
	t.Helper()
	if err := s.SaveSettings(`{"smartFolders":[{"id":"saved"}]}`); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(`{"theme":{"mode":"dark"}}`); err != nil {
		t.Fatal(err)
	}
	raw, err := s.GetSettings()
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}
	if got["theme"] == nil || got["smartFolders"] == nil || got["settings"] != nil {
		t.Fatalf("wrong shape: %s", raw)
	}
	for _, bad := range []string{"null", "[]", `"string"`, "true"} {
		if err := s.SaveSettings(bad); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.SaveSettings(fmt.Sprintf(`{"key%d":%d}`, i, i)); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	raw, err = s.GetSettings()
	if err != nil {
		t.Fatal(err)
	}
	json.Unmarshal([]byte(raw), &got)
	if len(got) != 14 {
		t.Fatalf("lost concurrent update: %s", raw)
	}
}

func TestSQLiteSettingsContract(t *testing.T) { testSettingsContract(t, newTestStore(t)) }
func TestPostgresSettingsContract(t *testing.T) {
	a, _ := postgresTestStores(t)
	testSettingsContract(t, a)
}
