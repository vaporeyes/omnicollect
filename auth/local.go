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
	// Sanitize to a valid PostgreSQL schema name (e.g., "default" -> "tenant_default")
	sanitized := SanitizeTenantID(tenantID)

	if provisioner != nil {
		if err := provisioner(sanitized); err != nil {
			log.Printf("auth: warning: failed to provision local tenant %s: %v", sanitized, err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := SetTenantID(r.Context(), sanitized)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
