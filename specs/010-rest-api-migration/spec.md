# Feature Specification: REST API Migration

**Feature Branch**: `010-rest-api-migration`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Decouple IPC by replacing Wails bindings with standard HTTP REST endpoints on the Go backend and an HTTP client on the Vue frontend. Enables running the backend as a standalone web service."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Serves REST Endpoints for Core CRUD (Priority: P1)

The Go backend exposes all existing item and module operations as standard HTTP REST endpoints under a versioned path prefix (e.g., `/api/v1/`). The existing logic (SQLite queries, FTS5 search, attribute filtering, image processing) remains unchanged -- only the transport layer changes from Wails IPC to HTTP. The backend can be started as a standalone server independent of the Wails desktop shell.

**Why this priority**: The REST server is the foundation of the entire migration. Without it, the frontend has nothing to call. This story delivers a fully functional, testable API that can be exercised with curl or any HTTP client.

**Independent Test**: Start the backend as a standalone HTTP server, use curl or a tool to call each endpoint (create item, list items, get modules, delete item), verify correct JSON responses and status codes.

**Acceptance Scenarios**:

1. **Given** the backend is running as an HTTP server, **When** a client sends `GET /api/v1/modules`, **Then** it receives a JSON array of all loaded module schemas with a 200 status.
2. **Given** items exist in the database, **When** a client sends `GET /api/v1/items?query=coins&moduleId=abc&filters=[...]`, **Then** it receives filtered items as a JSON array with a 200 status.
3. **Given** valid item data, **When** a client sends `POST /api/v1/items` with a JSON body, **Then** the item is created/updated and the response contains the saved item with a 200 status.
4. **Given** an item exists, **When** a client sends `DELETE /api/v1/items/:id`, **Then** the item is removed and a 204 status is returned.
5. **Given** invalid input (missing required fields), **When** a client sends a request, **Then** the server returns a structured error response with a 400 status and a descriptive error message.
6. **Given** the backend is started without the Wails shell, **Then** it runs as a standalone HTTP server listening on a configurable port.

---

### User Story 2 - Frontend Calls REST Endpoints Instead of Wails Bindings (Priority: P2)

The Vue frontend replaces all Wails binding imports (`wailsjs/go/main/App`) with HTTP calls to the REST endpoints. The Pinia stores continue to manage loading/error state exactly as they do now. All existing user-facing functionality (search, filter, create, edit, delete, bulk actions, CSV export, command palette) works identically through the HTTP transport.

**Why this priority**: The frontend migration completes the decoupling. Without it, the REST API exists but the app still uses Wails bindings. This story makes the entire application operate over HTTP.

**Independent Test**: Run the backend as an HTTP server and the frontend as a standard web app (via Vite dev server). Verify all operations (create, search, filter, edit, delete, bulk actions) work identically to the current Wails-based app.

**Acceptance Scenarios**:

1. **Given** the frontend is running against the REST backend, **When** the user performs any CRUD operation, **Then** the behavior is identical to the current Wails-based app.
2. **Given** the frontend API client, **When** a request fails (server error, network timeout), **Then** the Pinia store captures the error and the UI displays it via the existing error/toast system.
3. **Given** the frontend is loaded, **When** it fetches modules and items on startup, **Then** the data loads and displays within the same performance envelope as the current Wails app.
4. **Given** any Pinia store action, **Then** zero references to `wailsjs/go/main/App` remain in the store files; all calls go through the HTTP client.

---

### User Story 3 - Desktop App Continues to Work via Embedded Server (Priority: P3)

The Wails desktop shell continues to function by starting the embedded HTTP server internally and pointing the webview at it. The desktop user experience is unchanged -- the app launches, the UI loads, and all features work as before. The Wails shell simply becomes a thin wrapper that starts the server and opens a webview.

**Why this priority**: Desktop continuity ensures no regression for existing desktop users. This story is lower priority because it's an integration/packaging concern -- the core REST API and frontend migration must work first.

**Independent Test**: Build and launch the Wails desktop app. Verify it starts, loads the UI, and all features (CRUD, search, filter, images, export) work identically to the pre-migration version.

**Acceptance Scenarios**:

1. **Given** the desktop app is built with `wails build`, **When** the user launches it, **Then** the app starts with the embedded server and loads the UI identically to the pre-migration version.
2. **Given** the desktop app is running, **When** the user uses any feature (create item, search, filter, bulk delete, export), **Then** it works identically -- no regressions.
3. **Given** the desktop app, **Then** native dialogs (file save, file open) continue to work for image selection, backup export, and CSV export.

---

### Edge Cases

- What happens when the REST server is not reachable (frontend loaded but backend down)? The frontend should display a clear "Cannot connect to server" error, not silently fail.
- What happens with media file serving (thumbnails, originals)? The REST server must serve static media files at the same paths (/thumbnails/, /originals/) as the current Wails AssetServer.
- What happens with native dialogs (file save/open) when running as a web app vs desktop? Native dialogs only work in the Wails shell. When running as a web app, file operations should use browser-native alternatives (download links for export, file input for import).
- What happens with CORS when the frontend dev server (Vite) and backend run on different ports? The backend must include appropriate CORS headers for development.
- What happens when the API version path changes? The frontend API client should reference the base URL from a single configuration point.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend MUST expose all existing item operations (create, read, update, delete, list, search, filter) as REST endpoints under a versioned path prefix.
- **FR-002**: The backend MUST expose module schema operations (list active modules, save custom module, load module file) as REST endpoints.
- **FR-003**: The backend MUST expose bulk operations (batch delete, CSV export, bulk module update) as REST endpoints.
- **FR-004**: The backend MUST expose a single image upload endpoint that accepts a multipart file upload, processes the image (validate, copy original, generate thumbnail), and returns the result metadata. This replaces the current two-step SelectImageFile + ProcessImage flow.
- **FR-005**: The backend MUST serve static media files (thumbnails, originals) via HTTP, maintaining the same URL paths as the current Wails AssetServer.
- **FR-006**: The backend MUST be startable as a standalone HTTP server (without the Wails shell) on a configurable port.
- **FR-007**: All REST endpoints MUST return structured JSON responses with appropriate HTTP status codes (200 for success, 400 for validation errors, 404 for not found, 500 for server errors).
- **FR-008**: The frontend MUST replace all Wails binding imports with HTTP calls to the REST endpoints.
- **FR-009**: The frontend MUST use a centralized HTTP client configuration with a single base URL setting.
- **FR-010**: The frontend MUST handle HTTP errors (network failures, server errors) and surface them through the existing Pinia error/toast system.
- **FR-011**: The Wails desktop shell MUST continue to function by embedding the HTTP server and pointing the webview at it.
- **FR-012**: The backend MUST include CORS headers to support frontend development on a different port.
- **FR-013**: The backup export and CSV export endpoints MUST return file content with a `Content-Disposition: attachment` header so the browser triggers a native download. No server-side save dialogs.
- **FR-014**: The settings load/save operations MUST be exposed as REST endpoints.
- **FR-015**: The migration MUST be performed as a single complete pass (big bang) -- all Wails binding imports are replaced with HTTP calls at once, and the `wailsjs/go/main/App` imports are fully removed.

### Key Entities

- **REST Endpoint**: An HTTP route that maps to an existing backend operation. Has a method (GET/POST/PUT/DELETE), path, request format, and response format.
- **API Client**: A frontend service layer that makes HTTP requests to the REST endpoints and returns typed responses to the Pinia stores.
- **Server Configuration**: The HTTP server's runtime settings (port, CORS origins, media directory paths).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing Wails binding functionality is accessible via REST endpoints, verified by exercising every endpoint.
- **SC-002**: The frontend operates identically over HTTP as it did over Wails bindings -- zero user-visible behavioral regressions.
- **SC-003**: The backend can be started as a standalone server and serve the frontend as a static web application without the Wails shell.
- **SC-004**: The desktop app built with Wails continues to launch and function with zero regressions.
- **SC-005**: Frontend API calls complete within the same performance envelope as the current Wails IPC (under 200ms for typical operations on localhost).
- **SC-006**: Zero references to `wailsjs/go/main/App` remain in Pinia store files or component files after migration.

## Clarifications

### Session 2026-04-07

- Q: Should migration be big-bang, incremental via shim, or permanent dual-mode? -> A: Big bang -- replace all Wails imports with HTTP calls in one pass.
- Q: How should file exports (CSV, backup) work in web mode without native dialogs? -> A: HTTP download -- export endpoints return file content with Content-Disposition attachment header; browser handles save.
- Q: How should image upload work without native file dialogs? -> A: Single multipart upload endpoint -- frontend uses file input, POSTs multipart form data; backend receives bytes, processes, generates thumbnail, and stores. Replaces both SelectImageFile and ProcessImage.

## Assumptions

- REST is chosen over GraphQL for simplicity and alignment with the existing procedural binding pattern. Each Wails binding maps directly to one REST endpoint.
- The HTTP server uses the Go standard library (`net/http`) with a lightweight router. No heavy web framework is needed for this scope.
- Authentication is not required in v1. The server runs locally (localhost) for both desktop and development use cases. Authentication can be added later when multi-user or remote access is needed.
- The API version prefix is `/api/v1/`. This allows future versioning without breaking existing clients.
- Image upload uses a single multipart POST endpoint that replaces both `SelectImageFile` and `ProcessImage`. The frontend uses a standard `<input type="file">` element to select the image, then uploads it. The backend receives the file bytes, validates, copies the original, generates a thumbnail, and returns the result metadata.
- Export operations (CSV, backup) use HTTP download responses with `Content-Disposition: attachment` headers. This works universally in both web and desktop modes, replacing the current Wails native save dialogs.
- The Wails desktop shell starts the embedded HTTP server on a random available port and configures the webview to connect to it. No user-visible change in the desktop experience.
- The frontend API client base URL defaults to the current origin (same-origin when served by the backend) and can be overridden via environment variable for development.
- Settings (theme configuration) are stored server-side (same file-based approach) and accessed via REST endpoints.
