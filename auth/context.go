// ABOUTME: Context helpers for tenant ID propagation through HTTP request contexts.
// ABOUTME: Provides SetTenantID, TenantIDFromContext, and SanitizeTenantID for JWT sub claims.
package auth

import (
	"context"
	"omnicollect/tenantid"
)

type contextKey string

const tenantIDKey contextKey = "tenantID"

// SetTenantID returns a new context with the tenant ID attached.
func SetTenantID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, tenantIDKey, id)
}

// TenantIDFromContext extracts the tenant ID from the request context.
// Returns empty string if not set.
func TenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantIDKey).(string); ok {
		return v
	}
	return ""
}

// SanitizeTenantID returns a collision-resistant schema for an opaque local ID.
// JWT middleware uses issuer-bound identities through tenantid.Subject instead.
func SanitizeTenantID(id string) string {
	return tenantid.Local(id)
}
