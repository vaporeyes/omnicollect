// ABOUTME: Environment-based configuration for OmniCollect cloud deployment.
// ABOUTME: Reads DATABASE_URL, S3_*, PORT, TENANT_ID from env vars with local-mode defaults.
package main

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL  string
	S3Endpoint   string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	S3Region     string
	Port         int
	TenantID     string
	AuthDomain   string
	AuthAudience string
	AuthIssuer   string
	AuthClientID string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
// Empty DATABASE_URL means local SQLite mode. Empty S3_ENDPOINT means local filesystem.
func LoadConfig() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = "default"
	}

	return Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		S3Endpoint:   os.Getenv("S3_ENDPOINT"),
		S3Bucket:     os.Getenv("S3_BUCKET"),
		S3AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:  os.Getenv("S3_SECRET_KEY"),
		S3Region:     region,
		Port:         port,
		TenantID:     tenantID,
		AuthDomain:   os.Getenv("AUTH_DOMAIN"),
		AuthAudience: os.Getenv("AUTH_AUDIENCE"),
		AuthIssuer:   os.Getenv("AUTH_ISSUER_URL"),
		AuthClientID: os.Getenv("AUTH_CLIENT_ID"),
	}
}

// IsCloudDB returns true if a PostgreSQL DATABASE_URL is configured.
func (c Config) IsCloudDB() bool {
	return c.DatabaseURL != ""
}

// IsCloudStorage returns true if an S3-compatible endpoint is configured.
func (c Config) IsCloudStorage() bool {
	return c.S3Endpoint != ""
}

// IsAuthEnabled returns true if Auth0 JWT validation is configured.
func (c Config) IsAuthEnabled() bool {
	return c.AuthIssuer != ""
}
