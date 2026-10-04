// ABOUTME: Collision-resistant tenant schema names shared by authentication and storage.
// ABOUTME: Names preserve the full SHA-256 digest while staying below PostgreSQL's 63-byte limit.
package tenantid

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"strings"
)

// Schema binds an opaque identity to its namespace without lossy sanitization.
// Keep the namespace stable: local IDs and issuer-bound JWT subjects are distinct.
func Schema(namespace, identity string) string {
	framed, _ := json.Marshal([]string{namespace, identity})
	sum := sha256.Sum256(framed)
	digest := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum[:])
	return "tenant_" + strings.ToLower(digest)
}

// Local returns the schema for a configured local-mode tenant ID.
func Local(id string) string { return Schema("local", id) }

// Subject returns the schema for a JWT subject within its trusted issuer.
func Subject(issuer, sub string) string { return Schema("jwt:"+issuer, sub) }
