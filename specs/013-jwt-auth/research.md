# Research: JWT Authentication Middleware

**Branch**: `013-jwt-auth` | **Date**: 2026-04-08

## R1: Go JWT Validation Library

**Decision**: Use `github.com/auth0/go-jwt-middleware/v2` with the `jwtvalidator` from `go-jose`.

**Rationale**: This is Auth0's official Go middleware library. It handles JWKS fetching, caching, key rotation, token parsing, signature validation, claims extraction, and audience/issuer checks. Using the official library ensures compatibility with Auth0's token format and reduces custom code.

**Alternatives considered**:
- Manual JWKS fetch + `golang-jwt/jwt`: More control but requires writing caching, key rotation, and validation logic manually. Rejected -- the official library does this correctly out of the box.
- `lestrrat-go/jwx`: Excellent general JWKS/JWT library. Rejected -- Auth0's own library is purpose-built for this exact integration.

## R2: Auth0 Frontend SDK

**Decision**: Use `@auth0/auth0-vue` -- Auth0's official Vue 3 SDK.

**Rationale**: Provides a Vue plugin with `useAuth0()` composable, `loginWithRedirect()`, `getAccessTokenSilently()`, `isAuthenticated`, `isLoading`, and `logout()`. Handles the full Universal Login redirect flow, PKCE exchange, token storage in memory, and silent refresh via hidden iframe. Purpose-built for Vue 3 Composition API.

**Alternatives considered**:
- `@auth0/auth0-spa-js` (vanilla JS): Would work but requires manual Vue integration (wrapping in composables, managing reactive state). The Vue SDK does this already.
- Custom OAuth implementation: Rejected -- insecure and unnecessary when an official SDK exists.

## R3: Token-to-Tenant Mapping

**Decision**: Use the JWT `sub` (subject) claim directly as the tenant identifier. The tenant schema name becomes `tenant_{sub}` (with non-alphanumeric characters replaced by underscores).

**Rationale**: Auth0's `sub` claim is a stable, unique identifier for each user (format: `auth0|64abc123`). Using it directly avoids maintaining a separate user-to-tenant mapping table. The `|` character is sanitized to `_` for PostgreSQL schema naming.

**Alternatives considered**:
- Separate users table mapping auth0_id -> tenant_id: Adds a database lookup per request. Rejected for v1 -- direct mapping is simpler and sufficient.
- Use `email` claim as tenant ID: Rejected -- email can change; `sub` is immutable.

## R4: Middleware Architecture

**Decision**: The auth middleware is an `http.Handler` wrapper that:
1. Skips validation for exempt paths (health check, CORS preflight)
2. Extracts the Bearer token from the Authorization header
3. Validates signature, expiration, issuer, and audience using cached JWKS
4. Extracts the `sub` claim
5. Sanitizes it into a valid PostgreSQL schema name
6. Injects the tenant ID into the request context via `context.WithValue`
7. Calls the next handler

Handlers extract the tenant ID from context using a helper function.

**Rationale**: Standard Go middleware pattern. Wraps the existing router so all routes are protected by default. Exempt paths are configured explicitly (whitelist approach -- secure by default).

## R5: Local Mode Bypass

**Decision**: When `AUTH_ISSUER_URL` is empty, the auth middleware is not registered. Instead, the existing `TENANT_ID` env var is injected into every request's context directly. This preserves full backward compatibility for desktop/local mode.

**Rationale**: Clean separation -- in cloud mode, auth middleware determines the tenant from the token. In local mode, the env var determines the tenant. The rest of the application doesn't know or care which path provided the tenant ID.

## R6: Auto-Provisioning Strategy

**Decision**: On each authenticated request, after extracting the tenant ID, the middleware checks if the tenant schema exists. If not, it provisions it (CREATE SCHEMA + DDL). A simple in-memory cache (map of provisioned tenant IDs) avoids checking the database on every request.

**Rationale**: First-request provisioning is the simplest onboarding flow. The in-memory cache means only the first request per tenant per server instance incurs the check. The cache rebuilds on server restart (acceptable -- provisioning is idempotent and fast).

**Alternatives considered**:
- Provision via webhook on Auth0 user creation: More complex (requires webhook handler, Auth0 Actions configuration). Rejected for v1.
- Provision lazily on first database error: Fragile -- the error could be something else. Rejected.

## R7: Auth0 Configuration

**Decision**: Three new environment variables:

| Variable | Example | Description |
|----------|---------|-------------|
| AUTH_DOMAIN | `omnicollect.us.auth0.com` | Auth0 tenant domain |
| AUTH_AUDIENCE | `https://api.omnicollect.com` | API audience identifier (registered in Auth0) |
| AUTH_ISSUER_URL | `https://omnicollect.us.auth0.com/` | Token issuer URL (empty = local mode, no auth) |

Frontend also needs: `AUTH_CLIENT_ID` (Auth0 application client ID) for the SPA SDK.

**Rationale**: Standard Auth0 configuration pattern. The domain and audience are used by both backend (validation) and frontend (token request). Empty `AUTH_ISSUER_URL` triggers local mode bypass.
