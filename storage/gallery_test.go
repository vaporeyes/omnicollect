// ABOUTME: Shared gallery contracts enforce bounded projections, ordering and isolation.
// ABOUTME: Uses isolated SQLite and disposable PostgreSQL, never user databases.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
)

func galleryContract(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	if err := s.SaveModule(ModuleSchema{ID: "gallery", DisplayName: "Public collection"}); err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for i := 0; i < 55; i++ {
		item, err := s.InsertItem(Item{ModuleID: "gallery", Title: "Title", Images: []string{"cover.png", "private.png"}, Tags: []string{"secret-tag"}, Attributes: map[string]any{"secret": "private-attribute"}})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.ID)
	}
	snapshot, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range snapshot.Items {
		snapshot.Items[i].UpdatedAt = "2026-01-01T00:00:00Z"
	}
	if _, err = s.Restore(ctx, snapshot, "replace"); err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	first, err := s.ReadGalleryPage("gallery", 1)
	if err != nil {
		t.Fatal(err)
	}
	if first.TotalItems != 55 || len(first.Items) != 24 || first.TotalPages != 3 || first.CollectionName != "Public collection" {
		t.Fatalf("first %+v", first)
	}
	second, err := s.ReadGalleryPage("gallery", 2)
	if err != nil {
		t.Fatal(err)
	}
	last, err := s.ReadGalleryPage("gallery", 99)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 24 || len(last.Items) != 7 || last.Page != 3 {
		t.Fatalf("pages %+v %+v", second, last)
	}
	all := append(append(first.Items, second.Items...), last.Items...)
	for i, item := range all {
		if item.ID != ids[i] || item.PrimaryImage != "cover.png" {
			t.Fatalf("order/projection %+v", item)
		}
	}
	raw, _ := json.Marshal(first)
	for _, secret := range []string{"private.png", "secret-tag", "private-attribute"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("private data %s", raw)
		}
	}
	for _, page := range []int{0, -1, MaxGalleryPages + 1} {
		if _, err = s.ReadGalleryPage("gallery", page); err == nil {
			t.Fatal("unbounded page")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.WithContext(cancelled).ReadGalleryPage("gallery", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel %v", err)
	}
	if _, err = s.ReadGalleryPage("missing", 1); err == nil {
		t.Fatal("missing module accepted")
	}
	if err = s.SaveModule(ModuleSchema{ID: "empty", DisplayName: "Empty"}); err != nil {
		t.Fatal(err)
	}
	empty, err := s.ReadGalleryPage("empty", 99)
	if err != nil || empty.Page != 1 || empty.TotalPages != 1 || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty %+v %v", empty, err)
	}
}
func galleryConsistency(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	large, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	small := large
	small.Items = large.Items[:1]
	small.Modules = append([]ModuleSchema{}, large.Modules...)
	for i := range large.Modules {
		if large.Modules[i].ID == "gallery" {
			large.Modules[i].DisplayName = "Large"
			small.Modules[i].DisplayName = "Small"
		}
	}
	if _, err = s.Restore(ctx, large, "replace"); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 20; i++ {
			snapshot := large
			if i%2 == 0 {
				snapshot = small
			}
			if _, err := s.Restore(ctx, snapshot, "replace"); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	defer func() {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	for i := 0; i < 40; i++ {
		page, err := s.ReadGalleryPage("gallery", 1)
		if err != nil {
			t.Fatal(err)
		}
		if page.CollectionName == "Large" {
			if page.TotalItems != 55 || len(page.Items) != 24 {
				t.Fatalf("mixed large snapshot %+v", page)
			}
		} else if page.CollectionName == "Small" {
			if page.TotalItems != 1 || len(page.Items) != 1 {
				t.Fatalf("mixed small snapshot %+v", page)
			}
		} else {
			t.Fatalf("unexpected name %+v", page)
		}
	}
}
func TestSQLiteGallery(t *testing.T) {
	s := newTestStore(t)
	galleryContract(t, s)
	galleryConsistency(t, s)
}
func TestPostgresGallery(t *testing.T) {
	a, b := postgresTestStores(t)
	galleryContract(t, a)
	galleryConsistency(t, a)
	if err := b.SaveModule(ModuleSchema{ID: "gallery", DisplayName: "Other tenant"}); err != nil {
		t.Fatal(err)
	}
	page, err := b.ReadGalleryPage("gallery", 1)
	if err != nil || page.TotalItems != 0 || page.CollectionName != "Other tenant" {
		t.Fatalf("tenant leak %+v %v", page, err)
	}
}
