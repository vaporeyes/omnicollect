// ABOUTME: Tests consistent local tenant mapping and fail-closed provisioning.
// ABOUTME: Ensures unsuccessful provisioning cannot expose a fallback tenant.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnicollect/tenantid"
)

func TestProvisioningReceivesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	h := NewLocalTenantMiddleware("cancel", func(ctx context.Context, id string) error {
		calls++
		return ctx.Err()
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, input := range []context.Context{ctx, context.Background()} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest("GET", "/", nil).WithContext(input))
		expected := 200
		if input.Err() != nil {
			expected = 503
		}
		if response.Code != expected {
			t.Fatalf("status %d", response.Code)
		}
	}
	if calls != 2 {
		t.Fatal("cancelled provisioning was cached")
	}
}

func TestLocalTenantProvisioningFailureAndRetry(t *testing.T) {
	calls := 0
	middleware := NewLocalTenantMiddleware("Local|ID", func(ctx context.Context, id string) error {
		if id != tenantid.Local("Local|ID") {
			t.Fatalf("wrong identity: %s", id)
		}
		calls++
		if calls == 1 {
			return fmt.Errorf("temporary failure")
		}
		return nil
	})
	handled := 0
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if TenantIDFromContext(r.Context()) != tenantid.Local("Local|ID") {
			t.Fatal("wrong request tenant")
		}
		handled++
		w.WriteHeader(200)
	}))
	for _, status := range []int{503, 200, 200} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest("GET", "/", nil))
		if res.Code != status {
			t.Fatalf("status: %d, want %d", res.Code, status)
		}
	}
	if handled != 2 || calls != 2 {
		t.Fatalf("handled=%d provision calls=%d", handled, calls)
	}
}
