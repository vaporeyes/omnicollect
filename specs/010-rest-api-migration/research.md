# Research: REST API Migration

**Branch**: `010-rest-api-migration` | **Date**: 2026-04-07

## R1: Go HTTP Router Choice

**Decision**: Use Go standard library `net/http` with `http.ServeMux` (Go 1.22+ enhanced pattern matching).

**Rationale**: Go 1.22+ `ServeMux` supports method-based routing (`GET /api/v1/items`, `POST /api/v1/items`) and path parameters (`DELETE /api/v1/items/{id}`). This covers all our routing needs without any third-party dependency. The project already uses Go 1.25+ which includes these features.

**Alternatives considered**:
- Chi router: Lightweight, popular. Rejected -- standard library now covers the same patterns; no need for an extra dependency.
- Echo/Gin: Full frameworks with middleware. Rejected -- too heavy for this use case; we only need routing and CORS.
- gorilla/mux: Widely used but archived. Rejected -- maintenance concerns.

## R2: CORS Middleware

**Decision**: Implement a simple CORS middleware function in server.go that adds `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers` headers. In development mode, allow `*` origin. In production (desktop), CORS is irrelevant (same-origin).

**Rationale**: A custom middleware is ~15 lines of Go code. No need for a third-party CORS package for a single-user local app.

**Alternatives considered**:
- rs/cors package: Feature-complete CORS middleware. Rejected -- overkill for a localhost-only app.

## R3: Frontend HTTP Client

**Decision**: Use the native `fetch` API wrapped in a thin `api/client.ts` module. No Axios.

**Rationale**: `fetch` is built into all modern browsers and Wails webview. The client module provides: base URL configuration, JSON serialization/deserialization, error handling (non-2xx responses throw with parsed error message), and typed response generics. Axios would add a dependency for features we don't need.

**Alternatives considered**:
- Axios: Popular HTTP client. Rejected -- adds ~15KB dependency; fetch covers our needs.
- ofetch/ky: Lightweight fetch wrappers. Rejected -- marginal benefit over a custom 50-line wrapper.

## R4: Endpoint Mapping (Wails Bindings to REST)

**Decision**: Direct 1:1 mapping of each Wails binding to a REST endpoint.

| Wails Binding | REST Endpoint | Method |
|---------------|---------------|--------|
| GetItems(query, moduleId, filtersJSON) | /api/v1/items?query=...&moduleId=...&filters=... | GET |
| SaveItem(item) | /api/v1/items | POST |
| DeleteItem(id) | /api/v1/items/{id} | DELETE |
| DeleteItems(ids) | /api/v1/items/batch-delete | POST |
| GetActiveModules() | /api/v1/modules | GET |
| SaveCustomModule(json) | /api/v1/modules | POST |
| LoadModuleFile(moduleId) | /api/v1/modules/{id}/file | GET |
| ProcessImage(path) + SelectImageFile() | /api/v1/images/upload | POST (multipart) |
| ExportBackup() | /api/v1/export/backup | GET (returns ZIP) |
| ExportItemsCSV(ids) | /api/v1/export/csv | POST (returns CSV) |
| BulkUpdateModule(ids, newModuleID) | /api/v1/items/batch-update-module | POST |
| LoadSettings() | /api/v1/settings | GET |
| SaveSettings(json) | /api/v1/settings | PUT |

**Rationale**: 1:1 mapping minimizes risk. Each endpoint wraps the exact same Go function that the Wails binding called. No business logic changes.

## R5: Static File Serving for Media and Frontend

**Decision**: The HTTP server serves:
1. `/thumbnails/*` and `/originals/*` from the media directory (same as current Wails AssetServer)
2. `/` serves the built frontend (index.html + assets from `frontend/dist/`)
3. `/api/v1/*` routes to the API handlers

**Rationale**: Consolidating all serving into one HTTP server simplifies deployment. The frontend is a single-page app; all non-API, non-media routes serve index.html for client-side routing.

## R6: Wails Desktop Integration

**Decision**: In desktop mode, `main.go` starts the HTTP server on a random available port, then launches the Wails webview pointed at `http://localhost:{port}`. The Wails `AssetServer` is no longer used; the HTTP server handles everything.

**Rationale**: This is the simplest integration path. The Wails shell becomes a thin window wrapper. All logic goes through HTTP, ensuring identical behavior in desktop and web modes.

**Alternatives considered**:
- Keep Wails AssetServer for frontend, HTTP for API: Rejected -- creates two serving paths to maintain; increases complexity.

## R7: Type Safety Strategy (Constitution V)

**Decision**: Create `frontend/src/api/types.ts` with TypeScript interfaces mirroring every Go struct used in API responses (Item, ModuleSchema, AttributeSchema, ProcessImageResult, etc.). The API client functions return typed promises (e.g., `getItems(): Promise<Item[]>`).

**Rationale**: The TypeScript compiler catches type mismatches at build time. This replaces the Wails auto-generated bindings with equivalent manually-defined types. The types are a direct copy of the Go struct JSON tags.

**Alternatives considered**:
- OpenAPI code generation: Generate TypeScript from an OpenAPI spec. Rejected for v1 -- adds tooling complexity; manual types are sufficient for ~15 endpoints.
- No types (use `any`): Rejected -- violates Constitution Principle V.
