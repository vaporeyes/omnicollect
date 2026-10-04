// ABOUTME: Shared cancellation and atomic schema/batch validation regression contracts.
// ABOUTME: Uses isolated SQLite and explicitly disposable PostgreSQL stores only.
package storage

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func mutationContract(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	module := ModuleSchema{ID: "source", DisplayName: "Source"}
	if err := s.SaveModule(module); err != nil {
		t.Fatal(err)
	}
	item, err := s.SaveItem(ctx, Item{Title: "Keep", ModuleID: module.ID})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	scoped := s.WithContext(cancelled)
	checks := map[string]func() error{
		"read":            func() error { _, e := scoped.QueryItems("", "", "", ""); return e },
		"modules":         func() error { _, e := scoped.GetModules(); return e },
		"module file":     func() error { _, e := scoped.LoadModuleFile(module.ID); return e },
		"settings":        func() error { _, e := scoped.GetSettings(); return e },
		"tags":            func() error { _, e := scoped.GetAllTags(); return e },
		"showcases":       func() error { _, e := scoped.ListShowcases(); return e },
		"showcase lookup": func() error { _, e := scoped.GetShowcaseBySlug("missing"); return e },
		"delete":          func() error { return scoped.DeleteItem(item.ID) },
		"batch delete":    func() error { _, e := scoped.DeleteItems([]string{item.ID}); return e },
		"batch update":    func() error { _, e := scoped.BulkUpdateModule([]string{item.ID}, module.ID); return e },
		"CSV":             func() error { _, e := scoped.ExportItemsCSV([]string{item.ID}, nil); return e },
		"rename tag":      func() error { _, e := scoped.RenameTag("old", "new"); return e },
		"delete tag":      func() error { _, e := scoped.DeleteTag("old"); return e },
		"save settings":   func() error { return scoped.SaveSettings(`{"test":true}`) },
		"save module":     func() error { return scoped.SaveModule(module) },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			if err := check(); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancel ignored: %v", err)
			}
		})
	}
	if err := scoped.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tags := range []string{"null", "{}", "[1]", "[", "[\"\"]"} {
		if _, err := s.QueryItems("", "", "", tags); err == nil {
			t.Fatalf("invalid tag filter accepted: %s", tags)
		}
	}
	if _, err := s.RenameTag("old", string(make([]byte, 51))); err == nil {
		t.Fatal("long tag accepted")
	}
	baseline, err := s.Snapshot(ctx)
	if err != nil || len(baseline.Items) != 1 {
		t.Fatalf("base context poisoned: %+v %v", baseline, err)
	}
	unchanged := func() {
		t.Helper()
		after, e := s.Snapshot(ctx)
		if e != nil || !reflect.DeepEqual(baseline, after) {
			t.Fatalf("failed write changed snapshot: %+v %v", after, e)
		}
	}
	if _, e := s.BulkUpdateModule([]string{item.ID}, "missing"); e == nil {
		t.Fatal("missing target accepted")
	}
	unchanged()
	incompatible := module
	incompatible.Attributes = []AttributeSchema{{Name: "required", Type: "string", Required: true}}
	if e := s.SaveModule(incompatible); e == nil {
		t.Fatal("incompatible schema accepted")
	}
	unchanged()
	for _, mode := range []string{"replace", "merge"} {
		bad := baseline
		bad.Modules = []ModuleSchema{incompatible}
		if mode == "merge" {
			bad.Items = nil
		}
		if _, e := s.Restore(ctx, bad, mode); e == nil {
			t.Fatalf("incompatible %s accepted", mode)
		}
		unchanged()
	}
	target := ModuleSchema{ID: "target", DisplayName: "Target", Attributes: incompatible.Attributes}
	if e := s.SaveModule(target); e != nil {
		t.Fatal(e)
	}
	if _, e := s.BulkUpdateModule([]string{item.ID}, target.ID); e == nil {
		t.Fatal("incompatible reassignment accepted")
	}
	target.Attributes = nil
	if e := s.SaveModule(target); e != nil {
		t.Fatal(e)
	}
	if _, e := s.BulkUpdateModule([]string{item.ID, "missing"}, target.ID); e == nil {
		t.Fatal("partial reassignment accepted")
	}
	if _, e := s.DeleteItems([]string{item.ID, item.ID}); e == nil {
		t.Fatal("duplicate IDs accepted")
	}
	if _, e := s.DeleteItems(make([]string, 501)); e == nil {
		t.Fatal("unbounded IDs accepted")
	}
	n, e := s.BulkUpdateModule([]string{item.ID}, target.ID)
	if e != nil || n != 1 {
		t.Fatalf("valid reassignment: %d %v", n, e)
	}
	patch := "{\"large\":\"" + strings.Repeat("x", 600000) + "\"}"
	if err := s.SaveSettings(patch); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(strings.Replace(patch, "large", "other", 1)); err == nil {
		t.Fatal("cumulative settings limit bypass")
	}
	settings, err := s.GetSettings()
	if err != nil || strings.Contains(settings, "other") {
		t.Fatalf("failed settings merge changed data: %v", err)
	}
	if _, e := s.SaveItem(ctx, item); !errors.Is(e, ErrEditConflict) {
		t.Fatalf("reassignment didn't invalidate edit: %v", e)
	}
}
func TestPostgresBlockedQueryCancellation(t *testing.T) {
	s, _ := postgresTestStores(t)
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec("LOCK TABLE " + s.table("items") + " IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := s.WithContext(ctx).QueryItems("", "", "", ""); err == nil {
		t.Fatal("blocked query ignored timeout")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("query cancellation was not prompt")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.QueryItems("", "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestSQLiteMutationValidation(t *testing.T) { mutationContract(t, newTestStore(t)) }
func TestPostgresMutationValidation(t *testing.T) {
	a, _ := postgresTestStores(t)
	mutationContract(t, a)
}
