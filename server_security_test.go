// ABOUTME: Exercises production HTTP routing and origin validation rather than a bare mux.
// ABOUTME: Covers authenticated SPA bootstrapping, protected API routes, and opaque import paths.
package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"omnicollect/storage"
)

func testApp(t *testing.T) *App {
	t.Helper()
	store, err := storage.NewSQLiteStoreInMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return &App{store: store, mediaStore: newTestMediaStore(t, t.TempDir())}
}

func TestAuthenticatedFrontendIsPublic(t *testing.T) {
	app := testApp(t)
	app.config = Config{AuthIssuer: "https://issuer.example/", AuthAudience: "test"}
	s := NewServer(app)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("SPA")) })
	h := s.buildHandler()
	for _, path := range []string{"/", "/callback?code=test", "/assets/app.js"} {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest("GET", path, nil))
		if res.Code != 200 || res.Body.String() != "SPA" {
			t.Fatalf("frontend blocked: %s %d", path, res.Code)
		}
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest("GET", "/api/v1/items", nil))
	if res.Code != 401 {
		t.Fatalf("unprotected API: %d", res.Code)
	}
}

func TestCORSRejectsUntrustedOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, tc := range []struct {
		origin string
		status int
	}{{"https://evil.example", 403}, {"http://example.com", 200}, {"http://localhost:5173", 200}} {
		req := httptest.NewRequest("POST", "http://example.com/api/v1/items", nil)
		req.Header.Set("Origin", tc.origin)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != tc.status {
			t.Errorf("origin %s: %d", tc.origin, res.Code)
		}
	}
}

func TestImportPathCannotDeleteFiles(t *testing.T) {
	victim := filepath.Join(t.TempDir(), "keep.txt")
	if err := os.WriteFile(victim, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewServer(testApp(t))
	req := httptest.NewRequest("POST", "/api/v1/import/execute", bytes.NewBufferString(`{"tempId":"../../`+victim+`","mode":"replace"}`))
	res := httptest.NewRecorder()
	s.buildHandler().ServeHTTP(res, req)
	if res.Code != 404 {
		t.Fatalf("unexpected status: %d", res.Code)
	}
	if _, err := os.Stat(victim); err != nil {
		t.Fatal("victim removed", err)
	}
}
