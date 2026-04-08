# Auth Contract: JWT Authentication

**Branch**: `013-jwt-auth` | **Date**: 2026-04-08

## HTTP Authentication Protocol

### Request Format

All protected endpoints require:
```
Authorization: Bearer <jwt-token>
```

### Error Responses

| Status | Condition | Response Body |
|--------|-----------|---------------|
| 401 | Missing Authorization header | `{"error": "missing authorization header"}` |
| 401 | Invalid/malformed token | `{"error": "invalid token"}` |
| 401 | Expired token | `{"error": "token expired"}` |
| 401 | Wrong issuer/audience | `{"error": "invalid token claims"}` |

### Public Endpoints (No Auth Required)

| Path | Method | Reason |
|------|--------|--------|
| /api/v1/health | GET | Health check for load balancers |
| OPTIONS * | OPTIONS | CORS preflight requests |

All other `/api/v1/*` endpoints require valid authentication.

## Auth0 Configuration Requirements

### Backend (Go)

| Setting | Source | Description |
|---------|--------|-------------|
| Domain | `AUTH_DOMAIN` env | e.g., `omnicollect.us.auth0.com` |
| Audience | `AUTH_AUDIENCE` env | e.g., `https://api.omnicollect.com` |
| Issuer URL | `AUTH_ISSUER_URL` env | e.g., `https://omnicollect.us.auth0.com/` |

### Frontend (Vue)

| Setting | Source | Description |
|---------|--------|-------------|
| Domain | Build-time env `VITE_AUTH0_DOMAIN` | Same as backend AUTH_DOMAIN |
| Client ID | Build-time env `VITE_AUTH0_CLIENT_ID` | Auth0 SPA application ID |
| Audience | Build-time env `VITE_AUTH0_AUDIENCE` | Same as backend AUTH_AUDIENCE |
| Redirect URI | `window.location.origin` | Callback after login |

### Auth0 Dashboard Setup Required

1. Create an **API** with identifier = AUTH_AUDIENCE
2. Create a **Single Page Application** with allowed callback/logout URLs
3. Enable the **Authorization Code Flow with PKCE** grant type
4. Configure allowed origins for CORS

## JWT Token Claims (Auth0 Standard)

| Claim | Type | Used For |
|-------|------|----------|
| `sub` | string | Unique user ID -> tenant ID mapping |
| `iss` | string | Validated against AUTH_ISSUER_URL |
| `aud` | string[] | Validated to contain AUTH_AUDIENCE |
| `exp` | number | Token expiration timestamp |
| `iat` | number | Token issued-at timestamp |

## Middleware Pipeline

```
Request -> CORS -> Auth Middleware -> Route Handler
                      |
                      +-> Extract Bearer token
                      +-> Validate signature (cached JWKS)
                      +-> Validate exp, iss, aud
                      +-> Extract sub -> sanitize -> tenant ID
                      +-> Check/provision tenant schema
                      +-> Inject tenant ID into context
                      +-> Next handler
```

## Frontend Auth Flow

```
App loads -> Check isAuthenticated
  |
  +-> Not authenticated -> loginWithRedirect() -> Auth0 Universal Login
  |                                                    |
  |                                                    +-> User logs in
  |                                                    +-> Redirect back with code
  |                                                    +-> SDK exchanges code for tokens (PKCE)
  |
  +-> Authenticated -> getAccessTokenSilently() -> Bearer token
                       |
                       +-> API requests include Authorization header
                       +-> Token auto-refreshes before expiry
```
