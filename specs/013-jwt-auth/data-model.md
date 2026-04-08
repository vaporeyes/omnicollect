# Data Model: JWT Authentication Middleware

**Branch**: `013-jwt-auth` | **Date**: 2026-04-08

## No Database Schema Changes

Authentication does not add tables or columns. Tenant scoping already exists via schema-per-tenant (from feature 011). The only change is how the tenant ID is determined:

| Mode | Before | After |
|------|--------|-------|
| Local | TENANT_ID env var | TENANT_ID env var (unchanged) |
| Cloud | TENANT_ID env var | Extracted from JWT `sub` claim |

## New Configuration

| Variable | Type | Required (Cloud) | Default | Description |
|----------|------|-----------------|---------|-------------|
| AUTH_DOMAIN | string | yes | (empty) | Auth0 tenant domain |
| AUTH_AUDIENCE | string | yes | (empty) | Auth0 API audience |
| AUTH_ISSUER_URL | string | yes | (empty = no auth) | Token issuer URL |
| AUTH_CLIENT_ID | string | yes (frontend) | (empty) | Auth0 SPA client ID |

## Request Context

Each authenticated request has a tenant ID injected into the Go `context.Context`:

| Context Key | Type | Source (Cloud) | Source (Local) |
|-------------|------|----------------|----------------|
| tenantID | string | JWT `sub` claim, sanitized | TENANT_ID env var |

## Tenant ID Sanitization

Auth0 `sub` format: `auth0|64abc123def`
Sanitized for PostgreSQL schema name: `tenant_auth0_64abc123def`

Rules: replace all non-alphanumeric characters with `_`, prefix with `tenant_`.

## In-Memory Provisioning Cache

A `map[string]bool` tracks which tenant IDs have been provisioned during this server instance's lifetime. On first request for an unknown tenant, the middleware provisions the schema (idempotent CREATE SCHEMA IF NOT EXISTS + DDL).
