// ABOUTME: Context helpers for tenant ID propagation through HTTP request contexts.
// ABOUTME: Provides SetTenantID, TenantIDFromContext, and SanitizeTenantID for JWT sub claims.
package auth

import (
	"context"
	"strings"
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

// SanitizeTenantID converts a JWT sub claim into a valid PostgreSQL schema name.
// Replaces non-alphanumeric characters with underscores and adds a "tenant_" prefix.
// Example: "auth0|64abc123" -> "tenant_auth0_64abc123"
func SanitizeTenantID(sub string) string {
	var sb strings.Builder
	for _, c := range sub {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	if result == "" {
		result = "default"
	}
	return "tenant_" + result
}
