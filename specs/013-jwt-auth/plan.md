# Implementation Plan: JWT Authentication Middleware

**Branch**: `013-jwt-auth` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/013-jwt-auth/spec.md`

## Summary

Add Auth0-based JWT authentication to the REST API. Go middleware validates tokens using Auth0's JWKS endpoint, extracts the user's `sub` claim as tenant ID, and injects it into the request context for database scoping. The Vue frontend integrates Auth0's SPA SDK for Universal Login (redirect flow), automatic token injection, and session management. Auto-provisioning creates tenant schemas on first login. Local mode bypasses auth when no Auth0 config is present.

## Technical Context

**Language/Version**: Go 1.25+ (backend middleware), TypeScript + Vue 3 (frontend Auth0 SDK)
**Primary Dependencies**: `github.com/auth0/go-jwt-middleware/v2` + `gopkg.in/go-jose/go-jose.v2` (Go JWT validation), `@auth0/auth0-vue` (frontend SDK)
**Storage**: PostgreSQL schema-per-tenant (existing); tenant provisioning reuses existing PostgresStore.initTenantSchema()
**Testing**: Go test with mock JWKS server; frontend tests with mocked Auth0 SDK
**Target Platform**: Docker container (cloud) + macOS desktop (local, auth bypassed)
**Project Type**: Multi-tenant SaaS with external identity provider
**Performance Goals**: Token validation under 10ms (cached JWKS); auth overhead imperceptible
**Constraints**: No custom auth UI; Auth0 Universal Login handles all credential flows; local mode bypass preserves backward compatibility

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS (mitigated) | Auth bypassed in local mode (no AUTH_ISSUER_URL); desktop users unaffected |
| II. Schema-Driven UI | PASS | No UI template changes; Auth0 SDK wraps existing app |
| III. Flat Data Architecture | PASS | No database schema changes; tenant scoping already exists |
| IV. Performance & Memory | PASS | JWKS cached; token validation is CPU-only (no I/O per request) |
| V. Type-Safe IPC | PASS | REST API unchanged; Authorization header is standard HTTP |
| VI. Documentation | PASS | Spec artifacts produced; Auth0 setup documented |

All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/013-jwt-auth/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  contracts/           # Phase 1 output
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Backend (Go)
auth/
  middleware.go        # New: JWT validation middleware, JWKS caching, context injection
  context.go           # New: tenant ID extraction from request context
config.go              # Modified: add AUTH_DOMAIN, AUTH_AUDIENCE, AUTH_ISSUER_URL env vars
server.go              # Modified: wrap API routes with auth middleware; exempt health check
handlers.go            # Modified: extract tenant ID from context instead of env var
app.go                 # Modified: tenant provisioning on first request for unknown tenant

# Frontend (Vue/TypeScript)
frontend/
  src/
    auth/
      plugin.ts        # New: Auth0 Vue plugin configuration
      guard.ts         # New: route guard / auth check for app initialization
    api/
      client.ts        # Modified: inject Authorization header from Auth0 token
    App.vue            # Modified: wrap with auth check; show login redirect or app
    main.ts            # Modified: register Auth0 plugin
```

**Structure Decision**: Backend auth logic isolated in `auth/` package (middleware + context helpers). Frontend auth in `src/auth/` directory (plugin config + guard). Minimal changes to existing files -- middleware wraps the existing router; frontend client adds one header.

## Complexity Tracking

No constitution violations. No complexity justification needed.
