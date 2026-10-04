// ABOUTME: Cross-backend atomic deletion recovery, schema/conflict and capacity contracts.
// ABOUTME: Uses only isolated stores; validates versions, tenant ownership and import replacement.
package storage

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func recoveryContract(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	module := ModuleSchema{ID: "recover", DisplayName: "Recover"}
	if err := s.SaveModule(module); err != nil {
		t.Fatal(err)
	}
	first, err := s.SaveItem(ctx, Item{Title: "First", ModuleID: module.ID, Tags: []string{"tag"}, Attributes: map[string]any{"unknown": "preserve"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.SaveItem(ctx, Item{Title: "Second", ModuleID: module.ID})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DeleteWithRecovery([]string{first.ID, "missing"}); err == nil {
		t.Fatal("partial delete accepted")
	}
	batches, _ := s.ListDeletions()
	if len(batches) != 0 {
		t.Fatal("failed delete retained a batch")
	}
	batch, err := s.DeleteWithRecovery([]string{first.ID, second.ID})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Deleted != 2 || batch.RecoveryID == "" {
		t.Fatalf("batch %+v", batch)
	}
	page, _ := s.QueryItemPage("", "", "", "", 10, 0)
	if len(page.Items) != 0 {
		t.Fatal("delete not committed")
	}
	module.Attributes = []AttributeSchema{{Name: "newRequired", Type: "string", Required: true}}
	if err = s.SaveModule(module); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RecoverDeletion(batch.RecoveryID); err == nil {
		t.Fatal("obsolete schema recovered")
	}
	module.Attributes = nil
	if err = s.SaveModule(module); err != nil {
		t.Fatal(err)
	}
	// Reintroduce one ID through merge: recovery must not overwrite or partially restore.
	conflict := first
	if second.ID > first.ID {
		conflict = second
	}
	// The final INSERT conflicts, proving earlier inserts roll back as well.
	snapshot.Items = []Item{conflict}
	if _, err = s.Restore(ctx, snapshot, "merge"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RecoverDeletion(batch.RecoveryID); err == nil {
		t.Fatal("overwrote current ID")
	}
	page, _ = s.QueryItemPage("", "", "", "", 10, 0)
	if len(page.Items) != 1 {
		t.Fatal("partial recovery committed")
	}
	if err = s.DeleteItem(conflict.ID); err != nil {
		t.Fatal(err)
	}
	restored, err := s.RecoverDeletion(batch.RecoveryID)
	if err != nil || restored != 2 {
		t.Fatalf("restore %d %v", restored, err)
	}
	page, err = s.QueryItemPage("", "", "", "", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range page.Items {
		if item.ID == first.ID {
			if item.CreatedAt != first.CreatedAt || item.UpdatedAt == first.UpdatedAt || item.Attributes["unknown"] != "preserve" || len(item.Tags) != 1 {
				t.Fatalf("lost history/data %+v", item)
			}
		}
	}
	if _, err = s.SaveItem(ctx, first); !errors.Is(err, ErrEditConflict) {
		t.Fatalf("old editor survived recovery: %v", err)
	}
	if _, err = s.RecoverDeletion(batch.RecoveryID); err == nil {
		t.Fatal("batch restored twice")
	}
	batch, err = s.DeleteWithRecovery([]string{first.ID, second.ID})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.WithContext(cancelled).RecoverDeletion(batch.RecoveryID); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel %v", err)
	}
	if _, err = s.WithContext(cancelled).ListDeletions(); !errors.Is(err, context.Canceled) {
		t.Fatalf("list cancel %v", err)
	}
	if err = s.WithContext(cancelled).DiscardDeletion(batch.RecoveryID); !errors.Is(err, context.Canceled) {
		t.Fatalf("discard cancel %v", err)
	}
	if _, err = s.Restore(ctx, Snapshot{}, "replace"); err != nil {
		t.Fatal(err)
	}
	batches, err = s.ListDeletions()
	if err != nil || len(batches) != 0 {
		t.Fatalf("replace retained recovery %+v %v", batches, err)
	}
}
func TestSQLiteRecovery(t *testing.T) { recoveryContract(t, newTestStore(t)) }
func TestPostgresRecovery(t *testing.T) {
	a, b := postgresTestStores(t)
	recoveryContract(t, a)
	if err := a.SaveModule(ModuleSchema{ID: "a", DisplayName: "A"}); err != nil {
		t.Fatal(err)
	}
	item, err := a.SaveItem(context.Background(), Item{ModuleID: "a", Title: "private"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := a.DeleteWithRecovery([]string{item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.RecoverDeletion(batch.RecoveryID); err == nil {
		t.Fatal("cross-tenant recovery")
	}
	if err = b.DiscardDeletion(batch.RecoveryID); err == nil {
		t.Fatal("cross-tenant discard")
	}
	list, err := b.ListDeletions()
	if err != nil || len(list) != 0 {
		t.Fatalf("cross-tenant list %+v %v", list, err)
	}
	// Concurrent attempts can consume the batch only once.
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); n, _ := a.RecoverDeletion(batch.RecoveryID); results <- n }()
	}
	wg.Wait()
	close(results)
	total := 0
	for n := range results {
		total += n
	}
	if total != 1 {
		t.Fatalf("restored %d times", total)
	}
	batch, err = a.DeleteWithRecovery([]string{item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.DiscardDeletion(batch.RecoveryID); err != nil {
		t.Fatal(err)
	}
	list, err = a.ListDeletions()
	if err != nil || len(list) != 0 {
		t.Fatalf("discard %+v %v", list, err)
	}
}
func TestRecoveryCapacityAndAtomicCapture(t *testing.T) {
	s := newTestStore(t)
	item, err := s.InsertItem(Item{Title: "kept", ModuleID: "a", Attributes: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec("CREATE TRIGGER fail_delete BEFORE DELETE ON items BEGIN SELECT RAISE(ABORT,'injected'); END")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DeleteWithRecovery([]string{item.ID}); err == nil {
		t.Fatal("injected deletion succeeded")
	}
	batches, _ := s.ListDeletions()
	if len(batches) != 0 {
		t.Fatal("orphan recovery batch")
	}
	page, _ := s.QueryItemPage("", "", "", "", 1, 0)
	if len(page.Items) != 1 {
		t.Fatal("failed deletion lost item")
	}
	if _, err = s.db.Exec("DROP TRIGGER fail_delete"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < MaxRecoveryBatches; i++ {
		row, err := s.InsertItem(Item{Title: "bounded", ModuleID: "a", Attributes: map[string]any{}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.DeleteWithRecovery([]string{row.ID}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.DeleteWithRecovery([]string{item.ID}); err == nil {
		t.Fatal("silently evicted recovery data")
	}
	page, _ = s.QueryItemPage("", "", "", "", 1, 0)
	if len(page.Items) != 1 {
		t.Fatal("capacity error deleted record")
	}
	batches, err = s.ListDeletions()
	if err != nil {
		t.Fatal(err)
	}
	if err = s.DiscardDeletion(batches[0].RecoveryID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RecoverDeletion(batches[0].RecoveryID); err == nil {
		t.Fatal("discarded batch recovered")
	}
	if _, err = s.DeleteWithRecovery([]string{item.ID}); err != nil {
		t.Fatal(err)
	}
}
func TestRecoveryPersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.db")
	s, err := NewSQLiteStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveModule(ModuleSchema{ID: "m", DisplayName: "M"}); err != nil {
		t.Fatal(err)
	}
	item, err := s.SaveItem(context.Background(), Item{ModuleID: "m", Title: "Retained"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := s.DeleteWithRecovery([]string{item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := NewSQLiteStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	list, err := reopened.ListDeletions()
	if err != nil || len(list) != 1 || list[0].RecoveryID != batch.RecoveryID || list[0].Titles[0] != "Retained" {
		t.Fatalf("lost batch %+v %v", list, err)
	}
	if _, err = reopened.RecoverDeletion(batch.RecoveryID); err != nil {
		t.Fatal(err)
	}
}
func TestRecoveryRejectsOversizedBatch(t *testing.T) {
	s := newTestStore(t)
	item, err := s.InsertItem(Item{ModuleID: "a", Title: "Large legacy item", Attributes: map[string]any{"large": strings.Repeat("é", MaxRecoveryBatchBytes/2)}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DeleteWithRecovery([]string{item.ID}); !errors.Is(err, ErrInvalidDeletion) {
		t.Fatalf("size limit %v", err)
	}
	summary, err := s.CollectionSummary()
	if err != nil || summary.Items != 1 {
		t.Fatalf("oversize deletion changed collection %+v %v", summary, err)
	}
}
