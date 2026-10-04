// ABOUTME: Regression tests for tenant identities that previously collided or were truncated.
// ABOUTME: Verifies namespace, case, punctuation, Unicode, and long subjects remain distinct.
package tenantid

import (
	"regexp"
	"strings"
	"testing"
)

func TestTenantIdentities(t *testing.T) {
	ids := []string{"auth0|abc", "auth0_abc", "Auth0|abc", "auth0/abc", "用户", "用戶", "", strings.Repeat("a", 100) + "x", strings.Repeat("a", 100) + "y"}
	seen := map[string]bool{}
	pattern := regexp.MustCompile(`^tenant_[a-z2-7]{52}$`)
	for _, id := range ids {
		schema := Local(id)
		if !pattern.MatchString(schema) || len(schema) > 63 {
			t.Fatalf("unsafe schema: %q", schema)
		}
		if seen[schema] {
			t.Fatalf("identity collision: %q", id)
		}
		seen[schema] = true
		if Local(id) != schema {
			t.Fatal("unstable identity")
		}
	}
	if Subject("https://one/", "abc") == Subject("https://two/", "abc") {
		t.Fatal("issuers collided")
	}
	if Local("abc") == Subject("local", "abc") {
		t.Fatal("local and JWT collided")
	}
}
