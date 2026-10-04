// ABOUTME: Local-mode middleware that injects a fixed tenant ID without JWT validation.
// ABOUTME: Used when AUTH_ISSUER_URL is empty to preserve backward compatibility.
package auth

import (
	"log"
	"net/http"
)

// NewLocalTenantMiddleware creates middleware that injects a fixed tenant ID
// into every request's context. Used in local/development mode when auth is
// disabled. Optionally runs provisioning on the fixed tenant at startup.
func NewLocalTenantMiddleware(tenantID string, provisioner TenantProvisioner) func(http.Handler) http.Handler {
	// Use the same collision-resistant mapping as the local-mode store.
	sanitized := SanitizeTenantID(tenantID)

	check := ProvisionCheck(provisioner)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if check != nil {
				if err := check(r.Context(), sanitized); err != nil {
					log.Printf("auth: local tenant provisioning failed: %v", err)
					writeAuthError(w, http.StatusServiceUnavailable, "tenant provisioning failed")
					return
				}
			}
			ctx := SetTenantID(r.Context(), sanitized)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
