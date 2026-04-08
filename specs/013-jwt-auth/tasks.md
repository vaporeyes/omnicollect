# Tasks: JWT Authentication Middleware

**Input**: Design documents from `/specs/013-jwt-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-contract.md, quickstart.md

**Tests**: Not explicitly requested. Auth middleware test added in foundational phase for security assurance.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Install dependencies, add configuration

- [x] T001 [P] Install Go dependencies: run `go get github.com/auth0/go-jwt-middleware/v2 gopkg.in/go-jose/go-jose.v2`
- [x] T002 [P] Install frontend dependency: run `npm install @auth0/auth0-vue` in `frontend/`
- [x] T003 Update `config.go`: add AUTH_DOMAIN, AUTH_AUDIENCE, AUTH_ISSUER_URL env vars to Config struct and LoadConfig(); add `IsAuthEnabled()` helper that returns true when AUTH_ISSUER_URL is non-empty; add AUTH_CLIENT_ID for frontend build-time reference

**Checkpoint**: Dependencies installed; config reads auth env vars; `IsAuthEnabled()` works

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Auth package with context helpers and JWT middleware

- [x] T004 Create `auth/context.go`: define context key type; implement `SetTenantID(ctx, id) context.Context`, `TenantIDFromContext(ctx) string`, and `SanitizeTenantID(sub string) string` (replace non-alphanumeric with `_`, prefix with `tenant_`)
- [x] T005 Create `auth/middleware.go`: implement `NewJWTMiddleware(issuerURL, audience string) func(http.Handler) http.Handler` that creates a JWKS-validating middleware using `go-jwt-middleware/v2`; extracts `sub` claim; calls `SanitizeTenantID`; injects into context via `SetTenantID`; returns 401 JSON for missing/invalid/expired tokens
- [x] T006 Create `auth/middleware.go` exemption logic: implement `ExemptPaths(paths []string, next http.Handler) http.Handler` that skips auth for configured paths (health check, OPTIONS preflight)
- [x] T007 Create `auth/local.go`: implement `NewLocalTenantMiddleware(tenantID string) func(http.Handler) http.Handler` that injects a fixed tenant ID into every request's context (for local mode when auth is disabled)

**Checkpoint**: Auth package compiles; middleware functions defined; `go vet ./auth/...` passes

---

## Phase 3: User Story 1 - API Requires Valid Authentication (Priority: P1) MVP

**Goal**: All API endpoints (except health) reject unauthenticated requests. Valid tokens scope queries to the authenticated tenant.

**Independent Test**: curl without token -> 401; curl with valid token -> 200 with tenant-scoped data.

### Implementation for User Story 1

- [x] T008 [US1] Modify `server.go`: conditionally apply auth middleware based on `config.IsAuthEnabled()` -- if enabled, wrap API routes with `NewJWTMiddleware` + `ExemptPaths` for health check; if disabled, wrap with `NewLocalTenantMiddleware(config.TenantID)`
- [x] T009 [US1] Modify `handlers.go`: in every handler that accesses the store, extract tenant ID from request context via `auth.TenantIDFromContext(r.Context())` instead of using a fixed App-level tenant; pass tenant ID to store operations
- [x] T010 [US1] Modify `app.go` or handlers: add tenant-aware store resolution -- when using PostgresStore, set the search_path to the tenant's schema before each request (or per-handler); ensure the correct tenant's data is queried
- [x] T011 [US1] Add auto-provisioning check: in the auth middleware (or a separate middleware after auth), check if the tenant schema exists using an in-memory cache (`map[string]bool`); if not cached, call `PostgresStore.ProvisionTenant(tenantID)` (idempotent CREATE SCHEMA IF NOT EXISTS + DDL); add to cache on success
- [x] T012 [US1] Verify 401 responses: test with curl -- no header returns 401 JSON; invalid token returns 401; expired token returns 401; health check returns 200 without token
- [x] T013 [US1] Verify local mode: start server without AUTH_ISSUER_URL; verify all endpoints work without tokens using TENANT_ID env var (zero regressions)

**Checkpoint**: Backend auth gate fully functional; 401 for bad/missing tokens; tenant scoping via JWT sub claim

---

## Phase 4: User Story 2 - Frontend Auth Integration (Priority: P2)

**Goal**: Vue app uses Auth0 SDK for login/logout, token injection, and session management.

**Independent Test**: Load app -> redirect to Auth0 login -> log in -> app loads with user's data -> API calls include Authorization header.

### Implementation for User Story 2

- [x] T014 [US2] Create `frontend/src/auth/plugin.ts`: configure and export Auth0 Vue plugin with domain from `import.meta.env.VITE_AUTH0_DOMAIN`, clientId from `import.meta.env.VITE_AUTH0_CLIENT_ID`, audience from `import.meta.env.VITE_AUTH0_AUDIENCE`, redirect URI from `window.location.origin`, cacheLocation `'memory'`
- [x] T015 [US2] Modify `frontend/src/main.ts`: import and register the Auth0 plugin; only register if `VITE_AUTH0_DOMAIN` is set (skip in local mode)
- [x] T016 [US2] Create `frontend/src/auth/guard.ts`: export an `AuthGuard` component (or composable) that checks `isAuthenticated` and `isLoading` from `useAuth0()`; while loading shows a spinner; if not authenticated calls `loginWithRedirect()`; if authenticated renders the slot/children
- [x] T017 [US2] Modify `frontend/src/App.vue`: wrap the main app layout with the AuthGuard (when Auth0 is configured); add a "Sign Out" button in the sidebar that calls `logout({ logoutParams: { returnTo: window.location.origin } })`
- [x] T018 [US2] Modify `frontend/src/api/client.ts`: before each request, obtain the access token via `getAccessTokenSilently()` from the Auth0 SDK; add `Authorization: Bearer {token}` header to all requests; skip token injection when Auth0 is not configured (local mode)
- [x] T019 [US2] Add `.env.example` file in `frontend/`: document `VITE_AUTH0_DOMAIN`, `VITE_AUTH0_CLIENT_ID`, `VITE_AUTH0_AUDIENCE` with placeholder values

**Checkpoint**: Full auth loop works: login redirect -> Auth0 -> callback -> app loads -> API calls authenticated -> sign out works

---

## Phase 5: User Story 3 - Auto Tenant Provisioning (Priority: P3)

**Goal**: First-time users get their tenant schema auto-created on first API request.

**Independent Test**: Create new Auth0 user, log in, verify app loads with empty but functional collection.

### Implementation for User Story 3

- [x] T020 [US3] Implement `ProvisionTenant(tenantID string)` method on PostgresStore in `storage/postgres.go`: runs `CREATE SCHEMA IF NOT EXISTS {tenantID}` + all DDL (items table, modules table, settings table, indexes, search trigger) within the schema; idempotent (safe to call multiple times)
- [x] T021 [US3] Wire provisioning into the middleware pipeline in `auth/middleware.go` or `server.go`: after successful JWT validation and tenant ID extraction, call the provisioning check (in-memory cache + ProvisionTenant fallback); log when a new tenant is provisioned
- [x] T022 [US3] Handle provisioning errors gracefully: if ProvisionTenant fails (e.g., database error), return 503 Service Unavailable with a JSON error message; do not cache the failed tenant ID

**Checkpoint**: New Auth0 users get auto-provisioned tenants on first request; returning users hit the cache (no DB check)

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, Docker config, testing

- [x] T023 Update `docker-compose.yml`: add Auth0 env vars (AUTH_DOMAIN, AUTH_AUDIENCE, AUTH_ISSUER_URL) to the app service with placeholder values and comments
- [x] T024 Update `Dockerfile`: add `VITE_AUTH0_DOMAIN`, `VITE_AUTH0_CLIENT_ID`, `VITE_AUTH0_AUDIENCE` as build args in the Node builder stage so they're available at frontend build time
- [x] T025 [P] Update `CLAUDE.md`: document auth/ package, auth env vars, local mode bypass, tenant provisioning, Auth0 setup requirements
- [x] T026 [P] Update `README.md`: add Authentication section with Auth0 setup guide, env vars table, local vs cloud mode explanation, iteration 13 history
- [x] T027 Run quickstart.md acceptance test flow (local mode steps 1-3 + cloud mode steps 4-15 if Auth0 configured) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs Go deps + config)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs auth middleware)
- **User Story 2 (Phase 4)**: Depends on Phase 1 (needs @auth0/auth0-vue); independent of US1 for frontend work
- **User Story 3 (Phase 5)**: Depends on Phase 3 (needs middleware pipeline to hook into)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- backend auth gate.
- **US2 (P2)**: Depends on Setup only for frontend SDK. Can develop frontend in parallel with US1 backend, but full integration testing requires US1.
- **US3 (P3)**: Depends on US1 (provisioning hooks into the middleware pipeline).

### Parallel Opportunities

- T001 and T002 (Phase 1) -- Go and npm installs
- T004-T007 (Phase 2) -- all in auth/ package, but T005-T006 depend on T004 (context helpers)
- US1 backend (T008-T013) and US2 frontend (T014-T019) -- different languages, can develop in parallel
- T025 and T026 (Phase 6 docs) -- different files

---

## Parallel Example: Phase 1

```bash
Task: "Install Go auth dependencies" (T001)
Task: "Install @auth0/auth0-vue" (T002)
```

## Parallel Example: US1 Backend + US2 Frontend

```bash
# After Phase 2, both can develop in parallel:
Task: "Wire auth middleware into server.go" (T008-T013)
Task: "Create Auth0 Vue plugin + guard" (T014-T019)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (Setup) + Phase 2 (Foundational)
2. Phase 3 (US1) -- backend auth gate with JWT validation
3. **STOP and VALIDATE**: curl tests prove 401 for bad tokens, 200 for valid ones, local mode works
4. This alone secures the API -- frontend can be added incrementally

### Incremental Delivery

1. Phase 1 + Phase 2 = auth middleware ready
2. Phase 3 (US1) = backend secured (testable with curl + manual tokens)
3. Phase 4 (US2) = frontend integrated (full login/logout flow)
4. Phase 5 (US3) = auto-provisioning (frictionless onboarding)
5. Phase 6 = docs, Docker config, final validation

---

## Notes

- Auth0's official Go middleware handles JWKS caching, key rotation, and claim validation
- Frontend uses Auth0's Vue SDK with Universal Login (redirect + PKCE)
- JWT `sub` claim -> sanitized to PostgreSQL schema name (e.g., `auth0|abc` -> `tenant_auth0_abc`)
- In-memory provisioning cache avoids DB check on every request
- Local mode: empty AUTH_ISSUER_URL = no auth middleware = TENANT_ID env var used directly
- Auth0 dashboard setup (API + SPA application) is a manual prerequisite, not automated
