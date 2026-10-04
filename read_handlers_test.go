// ABOUTME: HTTP contracts for explicit bounded pages and filter-independent summaries.
// ABOUTME: Uses isolated SQLite with enough records to exceed the default page.
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"omnicollect/storage"
	"testing"
)

func TestPagedReadHTTPContracts(t *testing.T) {
	base, app := newTestServer(t)
	for i := 0; i < 105; i++ {
		if _, err := app.store.InsertItem(storage.Item{ModuleID: "comics", Title: "needle", Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	get := func(path string, target any) int {
		t.Helper()
		response, err := http.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if target != nil && response.StatusCode == 200 {
			if err := json.NewDecoder(response.Body).Decode(target); err != nil {
				t.Fatal(err)
			}
		} else {
			io.Copy(io.Discard, response.Body)
		}
		return response.StatusCode
	}
	var page storage.ItemPage
	if status := get("/api/v1/items/page", &page); status != 200 || len(page.Items) != 100 || !page.HasMore || page.Limit != 100 {
		t.Fatalf("default: %d %+v", status, page)
	}
	if status := get("/api/v1/items/page?limit=100&offset=100", &page); status != 200 || len(page.Items) != 5 || page.HasMore {
		t.Fatalf("last: %d %+v", status, page)
	}
	if status := get("/api/v1/items/page?moduleId=absent", &page); status != 200 || len(page.Items) != 0 || page.Items == nil {
		t.Fatalf("empty: %d %+v", status, page)
	}
	var summary storage.CollectionSummary
	if status := get("/api/v1/items/summary", &summary); status != 200 || summary.Items != 105 || len(summary.Modules) != 1 || summary.PurchaseTotal != nil {
		t.Fatalf("summary: %d %+v", status, summary)
	}
	for _, suffix := range []string{"limit=0", "limit=201", "limit=", "offset=-1", "offset=1000001", "limit=abc", "limit=1&limit=2", "sort=id", "tags=null", "filters=%5B", "limit=1;offset=2"} {
		if status := get("/api/v1/items/page?"+suffix, nil); status != 400 {
			t.Errorf("%s returned %d", suffix, status)
		}
	}
	if status := get("/api/v1/items/summary?moduleId=comics", nil); status != 400 {
		t.Fatalf("summary accepted filters: %d", status)
	}
	// Small legacy results remain complete arrays, never silent partial pages.
	var legacy []Item
	if status := get("/api/v1/items", &legacy); status != 200 || len(legacy) != 105 {
		t.Fatalf("legacy changed: %d %d", status, len(legacy))
	}
	for i := 105; i < 201; i++ {
		if _, err := app.store.InsertItem(storage.Item{ModuleID: "comics", Title: "needle", Attributes: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	if status := get("/api/v1/items", nil); status != 422 {
		t.Fatalf("oversized legacy result returned %d", status)
	}
	if status := get("/api/v1/items/page?limit=200", &page); status != 200 || len(page.Items) != 200 || !page.HasMore {
		t.Fatalf("bounded alternative %d %+v", status, page)
	}
}
