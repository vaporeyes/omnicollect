# Quickstart: JWT Authentication Middleware

**Branch**: `013-jwt-auth` | **Date**: 2026-04-08

## Prerequisites

- Auth0 account with a tenant configured
- Auth0 API created (identifier = audience)
- Auth0 SPA Application created (client ID available)
- Existing codebase on the `013-jwt-auth` branch
- PostgreSQL running (for tenant provisioning tests)

## New Dependencies

```bash
# Backend
go get github.com/auth0/go-jwt-middleware/v2
go get gopkg.in/go-jose/go-jose.v2

# Frontend
cd frontend && npm install @auth0/auth0-vue
```

## Files to Create

### Backend
1. **`auth/middleware.go`** -- JWT validation middleware: JWKS fetching/caching, token parsing, claim extraction, tenant ID injection into context
2. **`auth/context.go`** -- Helper functions: `TenantIDFromContext(ctx)`, `SetTenantID(ctx, id)`, tenant ID sanitization

### Frontend
3. **`frontend/src/auth/plugin.ts`** -- Auth0 Vue plugin configuration with domain, clientId, audience from env
4. **`frontend/src/auth/guard.ts`** -- Auth guard that checks `isAuthenticated` and redirects to login if needed

## Files to Modify

### Backend
1. **`config.go`** -- Add AUTH_DOMAIN, AUTH_AUDIENCE, AUTH_ISSUER_URL, AUTH_CLIENT_ID env vars
2. **`server.go`** -- Wrap API routes with auth middleware; exempt health check
3. **`handlers.go`** -- Extract tenant ID from context instead of App's fixed tenant
4. **`app.go`** -- Use tenant ID from context for store operations; add auto-provisioning

### Frontend
5. **`frontend/src/main.ts`** -- Register Auth0 plugin
6. **`frontend/src/api/client.ts`** -- Inject Authorization header from Auth0 token
7. **`frontend/src/App.vue`** -- Wrap with auth loading/redirect check; add Sign Out button

## Implementation Order

1. Install Go + npm dependencies
2. Create auth/context.go (context helpers + sanitization)
3. Create auth/middleware.go (JWT validation + JWKS caching)
4. Update config.go with auth env vars
5. Update server.go to apply middleware
6. Update handlers.go to use tenant from context
7. Add auto-provisioning in app.go
8. Create frontend auth plugin + guard
9. Update main.ts to register Auth0
10. Update client.ts to inject Bearer token
11. Update App.vue with auth check + sign out
12. Test locally (auth disabled, verify no regressions)
13. Test with Auth0 (auth enabled, verify full flow)
14. Update CLAUDE.md and README

## Acceptance Test Flow

### Local Mode (No Auth, Backward Compatible)
1. Start server without AUTH_ISSUER_URL env var
2. `curl http://localhost:8080/api/v1/items` -- verify 200 (no auth required)
3. Verify all features work as before

### Cloud Mode (Auth0 Enabled)
4. Set AUTH_DOMAIN, AUTH_AUDIENCE, AUTH_ISSUER_URL, AUTH_CLIENT_ID
5. `curl http://localhost:8080/api/v1/items` -- verify 401 (no token)
6. `curl -H "Authorization: Bearer invalid" http://localhost:8080/api/v1/items` -- verify 401
7. `curl http://localhost:8080/api/v1/health` -- verify 200 (public)
8. Obtain a valid token from Auth0 (via test app or curl)
9. `curl -H "Authorization: Bearer {token}" http://localhost:8080/api/v1/items` -- verify 200
10. Open browser, load app -- verify redirect to Auth0 login
11. Log in at Auth0 -- verify redirect back to app with data loaded
12. Create items -- verify they're scoped to the authenticated user's tenant
13. Open incognito, log in as a different user -- verify separate tenant data
14. Click "Sign Out" -- verify session cleared, redirect to login
15. Wait for token to near expiry -- verify silent refresh (no interruption)

### First-Login Provisioning
16. Create a new user in Auth0
17. Log in as the new user -- verify tenant schema is auto-created
18. Verify empty but functional collection (no errors)
