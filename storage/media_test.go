// ABOUTME: Shared filesystem and S3 tenant namespace and cancellation contracts.
// ABOUTME: Uses a local S3 HTTP stub to verify actual object keys and prevent global-key regressions.
package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

func TestMediaTenantNamespaceContract(t *testing.T) {
	var mu sync.Mutex
	objects := map[string][]byte{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
				w.WriteHeader(500)
				return
			}
			mu.Lock()
			objects[r.URL.Path] = data
			mu.Unlock()
			w.Header().Set("ETag", `"test"`)
			w.WriteHeader(200)
		case "GET", "HEAD":
			mu.Lock()
			data, ok := objects[r.URL.Path]
			mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			if r.Method != "HEAD" {
				w.Write(data)
			}
		default:
			w.WriteHeader(200)
		}
	}))
	defer server.Close()
	s3, err := NewS3MediaStore(server.URL, "bucket", "test-key", "test-secret", "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	for name, base := range map[string]MediaStore{"local": NewLocalMediaStoreAt(t.TempDir()), "s3": s3} {
		t.Run(name, func(t *testing.T) {
			a, err := base.ForTenant("tenant_alice")
			if err != nil {
				t.Fatal(err)
			}
			b, err := base.ForTenant("tenant_bob")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := base.ForTenant("../tenant_bob"); err == nil {
				t.Fatal("unsafe namespace accepted")
			}
			if err := a.SaveOriginal(context.Background(), "same.png", []byte("alice")); err != nil {
				t.Fatal(err)
			}
			if err := a.CheckOriginal(context.Background(), "same.png"); err != nil {
				t.Fatal(err)
			}
			if err := b.CheckOriginal(context.Background(), "same.png"); err == nil {
				t.Fatal("cross-tenant HEAD")
			}
			if _, err := b.GetOriginal(context.Background(), "same.png"); err == nil {
				t.Fatal("cross-tenant read")
			}
			if err := b.SaveOriginal(context.Background(), "same.png", []byte("bob")); err != nil {
				t.Fatal(err)
			}
			if err := a.SaveThumbnail(context.Background(), "same.png", []byte("alice-thumb")); err != nil {
				t.Fatal(err)
			}
			if _, err := b.GetThumbnail(context.Background(), "same.png"); err == nil {
				t.Fatal("cross-tenant thumbnail")
			}
			data, err := a.GetOriginal(context.Background(), "same.png")
			if err != nil || !bytes.Equal(data, []byte("alice")) {
				t.Fatalf("original overwritten: %q %v", data, err)
			}
			if _, err := base.GetOriginal(context.Background(), "same.png"); err == nil {
				t.Fatal("scoped file leaked to root")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := a.SaveOriginal(ctx, "cancelled.png", []byte("no")); err == nil {
				t.Fatal("cancelled write succeeded")
			}
			if _, err := a.GetOriginal(ctx, "same.png"); err == nil {
				t.Fatal("cancelled read succeeded")
			}
		})
	}
	mu.Lock()
	defer mu.Unlock()
	for _, key := range []string{"/bucket/tenants/tenant_alice/originals/same.png", "/bucket/tenants/tenant_bob/originals/same.png", "/bucket/tenants/tenant_alice/thumbnails/same.png"} {
		if _, ok := objects[key]; !ok {
			t.Errorf("missing tenant object key %s", key)
		}
	}
}
