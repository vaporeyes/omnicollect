# Tasks: REST API Migration

**Input**: Design documents from `/specs/010-rest-api-migration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/rest-api.md, quickstart.md

**Tests**: No automated test framework. Manual + curl testing per quickstart.md.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Shared infrastructure for both backend and frontend

- [ ] T001 [P] Create `frontend/src/api/types.ts`: define TypeScript interfaces mirroring Go structs -- Item, ModuleSchema, AttributeSchema, DisplayHints, ProcessImageResult, BulkDeleteResult, BulkUpdateResult, and API error response type `{error: string}`
- [ ] T002 [P] Create `frontend/src/api/client.ts`: centralized fetch-based HTTP client with configurable base URL (defaults to `window.location.origin`), helper functions `get<T>(path)`, `post<T>(path, body)`, `put<T>(path, body)`, `del(path)`, `postFile<T>(path, file)`, `downloadFile(path, body?)` -- all handle JSON serialization, error responses (throw with parsed error message), and return typed promises

**Checkpoint**: Frontend API client and types defined; no backend or store changes yet

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: HTTP server infrastructure that all endpoints depend on

- [ ] T003 Create `server.go`: define `Server` struct holding `*App` reference and `http.ServeMux`; implement `NewServer(app *App) *Server` that registers all routes per the REST contract; implement CORS middleware function; implement `Start(port int) error` that listens on the given port
- [ ] T004 Create `handlers.go`: implement all HTTP handler functions wrapping existing App methods -- `handleGetItems`, `handleSaveItem`, `handleDeleteItem`, `handleGetModules`, `handleSaveModule`, `handleLoadModuleFile`, `handleUploadImage`, `handleExportBackup`, `handleExportCSV`, `handleDeleteItems`, `handleBulkUpdateModule`, `handleGetSettings`, `handleSaveSettings`; each parses request params/body, calls the App method, writes JSON response with appropriate status code
- [ ] T005 Modify `main.go`: add a `--serve` flag (or detect absence of Wails) to start the HTTP server in standalone mode on port 8080; serve frontend static files from `frontend/dist/` at root path; serve media files at `/thumbnails/` and `/originals/`; keep Wails launch path that starts embedded server on random port
- [ ] T006 Modify `app.go`: remove `context.Context` field from App struct and the `startup` method's Wails-specific context usage; extract DB/module initialization into a standalone `Init()` method callable from both Wails startup and HTTP server startup

**Checkpoint**: `go run . --serve` starts a standalone HTTP server; all endpoints respond to curl; `go vet ./...` passes

---

## Phase 3: User Story 1 - Backend REST Endpoints (Priority: P1) MVP

**Goal**: All existing operations accessible via REST. Testable with curl.

**Independent Test**: Start server, exercise every endpoint with curl, verify correct responses and status codes per the REST contract.

### Implementation for User Story 1

- [ ] T007 [US1] Implement items handlers in `handlers.go`: `handleGetItems` parses query/moduleId/filters from URL params and calls `queryItems`; `handleSaveItem` reads JSON body and calls `SaveItem`; `handleDeleteItem` extracts `{id}` path param and calls `DeleteItem`; `handleDeleteItems` reads `{"ids":[...]}` and calls `DeleteItems`; `handleBulkUpdateModule` reads `{"ids":[...],"newModuleId":"..."}` and calls `BulkUpdateModule`
- [ ] T008 [US1] Implement modules handlers in `handlers.go`: `handleGetModules` calls `GetActiveModules`; `handleSaveModule` reads JSON body string and calls `SaveCustomModule`; `handleLoadModuleFile` extracts `{id}` path param and calls `LoadModuleFile`
- [ ] T009 [US1] Implement image upload handler in `handlers.go`: `handleUploadImage` parses multipart form, reads the `image` file field, saves to a temp file, calls `ProcessImage` with the temp path, returns `ProcessImageResult` as JSON
- [ ] T010 [US1] Implement export handlers in `handlers.go`: `handleExportBackup` calls `createBackupArchive` to a temp file, serves it with `Content-Disposition: attachment` header; `handleExportCSV` reads `{"ids":[...]}`, calls `exportItemsCSV`, serves with `Content-Disposition: attachment` header as `text/csv`
- [ ] T011 [US1] Implement settings handlers in `handlers.go`: `handleGetSettings` reads settings file and returns JSON; `handleSaveSettings` reads JSON body and writes settings file
- [ ] T012 [US1] Register all routes in `server.go` `NewServer` function using `mux.HandleFunc` with method+path patterns per the REST contract (e.g., `"GET /api/v1/items"`, `"POST /api/v1/items"`, `"DELETE /api/v1/items/{id}"`)
- [ ] T013 [US1] Verify all endpoints with curl: test each endpoint per quickstart.md steps 3-7, verify JSON responses and status codes

**Checkpoint**: Full REST API functional; all operations work via curl; no frontend changes yet

---

## Phase 4: User Story 2 - Frontend HTTP Client Migration (Priority: P2)

**Goal**: All frontend Wails imports replaced with HTTP client calls. Zero `wailsjs/go/main/App` references remain.

**Independent Test**: Run frontend dev server against standalone backend. All operations (CRUD, search, filter, bulk, export) work identically.

### Implementation for User Story 2

- [ ] T014 [US2] Migrate `frontend/src/stores/collectionStore.ts`: replace `import {GetItems, SaveItem, DeleteItem} from '../../wailsjs/go/main/App'` with imports from `../api/client`; update `fetchItems` to call `client.get('/api/v1/items?...')`, `saveItem` to call `client.post('/api/v1/items', item)`, `deleteItem` to call `client.del('/api/v1/items/' + id)`, `searchAllItems` to call `client.get`, `DeleteItems` to call `client.post('/api/v1/items/batch-delete', {ids})`, etc.
- [ ] T015 [US2] Migrate `frontend/src/stores/moduleStore.ts`: replace Wails imports with `client.get('/api/v1/modules')` for `fetchModules`
- [ ] T016 [US2] Migrate `frontend/src/App.vue`: replace `LoadModuleFile` with `client.get('/api/v1/modules/{id}/file')`, replace `LoadSettings` with `client.get('/api/v1/settings')`, replace `ExportBackup` with `client.downloadFile('/api/v1/export/backup')` that triggers browser download, replace `ExportItemsCSV` with `client.downloadFile('/api/v1/export/csv', {ids})`, replace `BulkUpdateModule` with `client.post`, remove `WindowSetSystemDefaultTheme` import from Wails runtime
- [ ] T017 [US2] Migrate `frontend/src/components/ImageAttach.vue`: replace `SelectImageFile` + `ProcessImage` Wails calls with an `<input type="file" accept="image/*">` element and `client.postFile('/api/v1/images/upload', file)` call
- [ ] T018 [US2] Migrate `frontend/src/components/CommandPalette.vue`: update `searchAllItems` call if it directly imported from Wails (should already go through collectionStore; verify and fix if needed)
- [ ] T019 [US2] Remove all `wailsjs/go/main/App` imports: search all `.ts` and `.vue` files for remaining Wails imports; remove or replace any found; verify `grep -r "wailsjs/go/main/App" frontend/src/` returns zero matches
- [ ] T020 [US2] Remove Wails runtime imports from frontend: replace `WindowSetSystemDefaultTheme` and `SaveFileDialog` usage with no-ops or browser equivalents; verify `grep -r "wailsjs/runtime" frontend/src/` returns zero matches (except legitimate Wails shell integration in main.ts if needed)

**Checkpoint**: Frontend operates entirely over HTTP; zero Wails binding imports in stores/components

---

## Phase 5: User Story 3 - Desktop App Continuity (Priority: P3)

**Goal**: Wails desktop shell works by embedding the HTTP server.

**Independent Test**: `wails build` produces a working desktop app. All features function identically.

### Implementation for User Story 3

- [ ] T021 [US3] Modify `main.go` Wails launch path: on desktop startup, start the HTTP server on port 0 (random available port), configure Wails to load `http://localhost:{port}` in the webview instead of using the embedded AssetServer
- [ ] T022 [US3] Verify native dialogs are no longer needed: export/backup now use Content-Disposition downloads which work in the Wails webview; image upload uses file input; remove any remaining Wails `runtime.SaveFileDialog` / `runtime.OpenFileDialog` calls from Go handlers
- [ ] T023 [US3] Build and test desktop app: run `wails build`, launch the binary, verify all features work (create item, search, filter, images, export, bulk actions, command palette, settings)

**Checkpoint**: Desktop app works identically to pre-migration via embedded HTTP server

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, verification

- [ ] T024 Add `--serve` flag documentation to `main.go`: print usage help showing `--serve` starts standalone HTTP mode and `--port` sets the listen port
- [ ] T025 [P] Update `CLAUDE.md` to document REST API architecture: new files (server.go, handlers.go, api/client.ts, api/types.ts), endpoint list, standalone server mode, frontend migration
- [ ] T026 [P] Update project `README.md` to document REST API migration, standalone server mode, development setup (backend + frontend dev servers), endpoint reference, iteration history entry
- [ ] T027 Run quickstart.md full acceptance test flow (all 22 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 (Go and TS files are independent)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs server.go + handlers.go infrastructure)
- **User Story 2 (Phase 4)**: Depends on Phase 1 (needs api/client.ts + types.ts) AND Phase 3 (needs working endpoints to call)
- **User Story 3 (Phase 5)**: Depends on Phase 3 + Phase 4 (needs both backend and frontend working over HTTP)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. This is the MVP -- a working REST API.
- **US2 (P2)**: Depends on US1 (frontend needs working endpoints)
- **US3 (P3)**: Depends on US1 + US2 (desktop mode needs everything working over HTTP)

### Within Each Phase

- Phase 1: T001 and T002 are parallel (different files)
- Phase 2: T003 before T004 (handlers need server struct); T005 and T006 after T003-T004
- Phase 3: T007-T011 can be implemented in any order within handlers.go; T012 after all handlers; T013 after T012
- Phase 4: T014-T018 can be done in any order (different files); T019-T020 after all migrations
- Phase 6: T025 and T026 are parallel

### Parallel Opportunities

- T001 and T002 (Phase 1) -- different files
- T001-T002 (frontend) and T003-T004 (backend) -- different languages, full parallel
- T014-T018 (Phase 4 store/component migrations) -- different files
- T025 and T026 (Phase 6 docs) -- different files

---

## Parallel Example: Phase 1 + Phase 2

```bash
# Frontend and backend setup can run fully in parallel:
# Frontend:
Task: "Create api/types.ts" (T001)
Task: "Create api/client.ts" (T002)
# Backend:
Task: "Create server.go" (T003)
Task: "Create handlers.go" (T004)
```

## Parallel Example: Phase 4 Store Migration

```bash
# Different files, can run in parallel:
Task: "Migrate collectionStore.ts" (T014)
Task: "Migrate moduleStore.ts" (T015)
Task: "Migrate App.vue" (T016)
Task: "Migrate ImageAttach.vue" (T017)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002) -- types + client
2. Complete Phase 2: Foundational (T003-T006) -- server infrastructure
3. Complete Phase 3: User Story 1 (T007-T013) -- all endpoints working
4. **STOP and VALIDATE**: curl every endpoint; verify JSON responses
5. This delivers a fully functional REST API testable independently of the frontend

### Incremental Delivery

1. Phase 1 + Phase 2 + Phase 3 (US1) = MVP: REST API works, testable with curl
2. Add Phase 4 (US2) = frontend migrated to HTTP; app works as web app
3. Add Phase 5 (US3) = desktop app continues working via embedded server
4. Phase 6 (Polish) = documentation, final validation
5. Each phase is independently verifiable before proceeding

---

## Notes

- Big-bang migration per clarification: all Wails imports replaced in one pass (Phase 4)
- Go `net/http.ServeMux` with Go 1.22+ method-based routing; no third-party router
- Native `fetch` API; no Axios dependency
- Image upload: single multipart POST replaces SelectImageFile + ProcessImage
- Exports: Content-Disposition attachment headers trigger browser downloads
- Constitution V (Type-Safe IPC): manual TypeScript interfaces in api/types.ts replace Wails auto-generated bindings
- Desktop mode: Wails shell starts embedded HTTP server on random port, webview loads from it
- CORS middleware for development (Vite dev server on different port)
