// ABOUTME: Tests startup configuration safety boundaries.
// ABOUTME: SQLite cannot be deployed as a JWT-authenticated multi-tenant backend.
package main

import "testing"

func TestConfigRequiresTenantIsolationForAuth(t *testing.T) {
	cfg := Config{Port: 8080, AuthIssuer: "https://issuer.example/", AuthAudience: "test"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("accepted multi-user SQLite")
	}
	cfg.DatabaseURL = "postgres://example/app"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg.AuthAudience = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("accepted auth without audience")
	}
}
