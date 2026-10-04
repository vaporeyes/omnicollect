// ABOUTME: HTTP boundary checks for bounded queries, strict JSON, and request deadlines.
// ABOUTME: Uses isolated in-memory databases and disposable test servers.
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCRUDRejectsMalformedInput(t *testing.T) {
	base, _ := newTestServer(t)
	for _, path := range []string{"/api/v1/items?tags=null", "/api/v1/items?tags=%7B%7D", "/api/v1/items?filters=not-json"} {
		response, err := http.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Fatalf("%s: %d", path, response.StatusCode)
		}
	}
	for _, path := range []string{"/api/v1/items/batch-delete", "/api/v1/items/batch-update-module", "/api/v1/modules", "/api/v1/tags/rename", "/api/v1/import/execute", "/api/v1/showcases/toggle"} {
		response, err := http.Post(base+path, "application/json", strings.NewReader(`{} {}`))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Fatalf("%s: %d", path, response.StatusCode)
		}
	}
}
func TestRequestDeadlineAndCancellation(t *testing.T) {
	for path, maximum := range map[string]time.Duration{"/api/v1/items": 15 * time.Second, "/api/v1/import/execute": 3 * time.Minute, "/api/v1/ai/analyze": 90 * time.Second} {
		h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deadline, ok := r.Context().Deadline()
			if !ok || time.Until(deadline) > maximum || time.Until(deadline) < maximum-time.Second {
				t.Fatalf("wrong deadline for %s", path)
			}
		}))
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Err() != context.Canceled {
			t.Fatal("parent cancellation lost")
		}
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/v1/items", nil).WithContext(ctx))
}
