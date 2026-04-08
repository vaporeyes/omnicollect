# Feature Specification: JWT Authentication Middleware

**Feature Branch**: `013-jwt-auth`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Integrate external identity provider for SaaS auth. Add JWT middleware to validate tokens, extract user identity, inject tenant context for database scoping."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - API Requests Require Valid Authentication (Priority: P1)

All API endpoints (except health check) require a valid authentication token in the request. When a user's frontend makes an API call, it includes the token obtained from Auth0. The backend validates the token, extracts the user's identity, and uses it to scope all database queries to that user's tenant. Requests without a valid token receive a clear 401 Unauthorized response.

**Why this priority**: Without authentication, any user can access any other user's data. This is the fundamental security gate for multi-tenant SaaS. Nothing else matters if data isolation is broken.

**Independent Test**: Send a request without a token -- verify 401. Send a request with an invalid token -- verify 401. Send a request with a valid token -- verify 200 and data scoped to the authenticated user's tenant.

**Acceptance Scenarios**:

1. **Given** no Authorization header, **When** a client sends any API request, **Then** the server responds with 401 Unauthorized and a JSON error message.
2. **Given** an expired or malformed token, **When** a client sends an API request, **Then** the server responds with 401 Unauthorized.
3. **Given** a valid token for user A, **When** the client sends `GET /api/v1/items`, **Then** only user A's items are returned (tenant isolation enforced).
4. **Given** a valid token, **When** the client creates an item via `POST /api/v1/items`, **Then** the item is stored under the authenticated user's tenant.
5. **Given** a health check endpoint, **When** a client sends `GET /api/v1/health`, **Then** it responds 200 without requiring authentication (public endpoint).

---

### User Story 2 - Frontend Obtains and Sends Tokens Automatically (Priority: P2)

The Vue frontend integrates with Auth0's client SDK. When the app loads, it checks if the user has an active session. If not, the user is redirected to Auth0's login page. Once authenticated, the frontend stores the token and includes it in every API request via the Authorization header. Token refresh happens automatically before expiration.

**Why this priority**: The frontend integration completes the auth loop. Without it, the user has no way to obtain a valid token. This story makes authentication seamless from the user's perspective.

**Independent Test**: Load the app without a session -- verify redirect to login. Log in -- verify the app loads with the user's data. Make API calls -- verify the Authorization header is included. Wait for token to approach expiration -- verify it refreshes without interrupting the user.

**Acceptance Scenarios**:

1. **Given** a user visits the app without an active session, **When** the app loads, **Then** the user is redirected to Auth0's login page.
2. **Given** the user completes login at Auth0, **When** redirected back to the app, **Then** the app loads their collection data immediately.
3. **Given** an authenticated session, **When** the frontend makes any API call, **Then** the Authorization header contains a valid Bearer token.
4. **Given** a token that will expire within a configured threshold, **When** the frontend detects the approaching expiration, **Then** it refreshes the token automatically without user interruption.
5. **Given** an authenticated user, **When** they click a "Sign Out" button, **Then** the session is terminated, the token is cleared, and the user is redirected to the login page.

---

### User Story 3 - Automatic Tenant Provisioning for First-Time Users (Priority: P3)

When a user logs in for the first time, the system automatically creates their tenant (database schema) and sets up any default resources. Subsequent logins detect the existing tenant and skip provisioning. The user does not need to take any manual setup steps after their first login.

**Why this priority**: Without auto-provisioning, new users would see an empty or broken app on first login. This story ensures the onboarding is frictionless. It depends on auth being in place (US1 + US2).

**Independent Test**: Create a new user in Auth0. Log in as that user. Verify the app loads with an empty but functional collection (tenant schema created, empty item list, no errors).

**Acceptance Scenarios**:

1. **Given** a new user who has never logged in before, **When** they authenticate and make their first API request, **Then** the system creates their tenant schema and returns a successful response.
2. **Given** a returning user with an existing tenant, **When** they authenticate, **Then** the system detects the existing tenant and serves their data without re-provisioning.
3. **Given** tenant provisioning fails (e.g., database error), **Then** the user receives a clear error message and the system does not leave partially created state.

---

### Edge Cases

- What happens when Auth0 is unreachable? The backend should fail gracefully -- if it cannot verify tokens because Auth0's public keys are cached, it continues to work with cached keys. If no cached keys exist, requests fail with 503 Service Unavailable.
- What happens in local/desktop mode (no identity provider configured)? Authentication middleware should be bypassed entirely, using the existing TENANT_ID environment variable for tenant scoping. This preserves backward compatibility.
- What happens when a token is valid but the user's tenant doesn't exist yet? The middleware detects this and triggers auto-provisioning before processing the request (US3).
- What happens with concurrent first-login requests from the same user? Tenant provisioning should be idempotent -- if the schema already exists, provisioning is a no-op.
- What happens when Auth0's token format changes? The backend should validate tokens using Auth0's published public keys (JWKS), which auto-rotate. No hardcoded secrets.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All API endpoints (except health check and any designated public endpoints) MUST require a valid Bearer token in the Authorization header.
- **FR-002**: The backend MUST validate the token's signature, expiration, and issuer against Auth0's published OIDC configuration (JWKS endpoint at `{domain}/.well-known/jwks.json`).
- **FR-003**: The backend MUST extract the user identifier from the validated token and use it as the tenant identifier for database query scoping.
- **FR-004**: Requests without a token or with an invalid/expired token MUST receive a 401 Unauthorized response with a JSON error body.
- **FR-005**: The health check endpoint (`GET /api/v1/health`) MUST remain publicly accessible without authentication.
- **FR-006**: The frontend MUST integrate with Auth0's client SDK to manage user sessions, obtain tokens, and handle login/logout flows.
- **FR-007**: The frontend MUST include the Bearer token in the Authorization header of every API request.
- **FR-008**: The frontend MUST handle token refresh automatically before the token expires.
- **FR-009**: The frontend MUST use Auth0's Universal Login (redirect flow) for authentication. Unauthenticated users are redirected to Auth0's hosted login page and returned to the app with an authorization code after successful login.
- **FR-010**: The frontend MUST provide a "Sign Out" action that terminates the session and redirects to login.
- **FR-011**: When a user authenticates for the first time, the system MUST automatically provision their tenant (create database schema with tables, indexes, and default configuration).
- **FR-012**: Tenant provisioning MUST be idempotent -- repeated provisioning attempts for the same user MUST NOT fail or duplicate data.
- **FR-013**: Authentication MUST be optional: when no identity provider is configured (local/desktop mode), the middleware MUST be bypassed and the existing TENANT_ID environment variable used for tenant scoping.
- **FR-014**: The backend MUST cache Auth0's public keys (JWKS) to avoid calling the provider on every request.

### Key Entities

- **Authenticated User**: A user whose identity has been verified by the external identity provider. Has a unique user ID (from the provider) that maps to a tenant.
- **Tenant**: An isolated database schema containing one user's collections, modules, and settings. Created automatically on first login.
- **Bearer Token**: A signed token (JWT) issued by Auth0, containing the user's identity claims. Sent in the Authorization header of every API request.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of protected API endpoints reject requests without valid authentication (zero unauthorized data access).
- **SC-002**: Token validation adds less than 10ms overhead per request (cached key validation).
- **SC-003**: New users can sign up and see a functional (empty) collection within 30 seconds of completing identity provider registration.
- **SC-004**: Token refresh occurs seamlessly -- zero user-visible session interruptions during normal usage.
- **SC-005**: The application continues to function in local mode (no auth configured) with zero regressions.
- **SC-006**: Two authenticated users operating simultaneously see only their own data with zero cross-tenant data leakage.

## Clarifications

### Session 2026-04-08

- Q: Which Auth0 login flow should the frontend use? -> A: Universal Login (redirect flow) -- user redirected to Auth0's hosted page, returned with auth code. Most secure, recommended by Auth0.

## Assumptions

- Auth0 is Auth0. It handles user registration, login, password management, MFA, and social login. The OmniCollect application does not implement any of these directly.
- The identity provider issues standard JWTs with a `sub` (subject) claim containing the user's unique ID. This `sub` value becomes the tenant identifier.
- The provider publishes a JWKS (JSON Web Key Set) endpoint for token signature verification. The backend fetches and caches these keys.
- The frontend uses Auth0's SPA SDK (`@auth0/auth0-spa-js` or `@auth0/auth0-vue`) for session management. The SDK handles the Universal Login redirect flow (Authorization Code + PKCE), token storage in memory, and silent refresh via hidden iframe.
- CORS configuration already exists (from the REST API migration). The auth middleware operates after CORS handling.
- Local/desktop mode detection: if the `AUTH_ISSUER_URL` environment variable is empty, authentication is disabled and the existing `TENANT_ID` env var is used directly.
- User profile data (name, email, avatar) is not stored in OmniCollect's database in v1. The identity provider is the source of truth for user profile information.
- Rate limiting on auth endpoints is handled by Auth0, not by OmniCollect.
