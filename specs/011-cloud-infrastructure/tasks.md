# Tasks: Cloud Infrastructure Migration

**Input**: Design documents from `/specs/011-cloud-infrastructure/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/infrastructure-contract.md, quickstart.md

**Tests**: No automated test framework. Manual + curl + docker-compose integration testing.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Install dependencies, define interfaces, create configuration

- [x] T001 Add Go dependencies: run `go get github.com/lib/pq github.com/aws/aws-sdk-go-v2 github.com/aws/aws-sdk-go-v2/config github.com/aws/aws-sdk-go-v2/credentials github.com/aws/aws-sdk-go-v2/service/s3`
- [x] T002 Create `config.go`: define `Config` struct with fields for DATABASE_URL, S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, S3_REGION, PORT, TENANT_ID; implement `LoadConfig()` that reads from environment variables with defaults; add `IsCloudDB()` and `IsCloudStorage()` helper methods that check if DATABASE_URL / S3_ENDPOINT are set
- [x] T003 [P] Create `storage/db.go`: define `Store` interface with methods `QueryItems`, `InsertItem`, `UpdateItem`, `DeleteItem`, `DeleteItems`, `BulkUpdateModule`, `ExportItemsCSV`, `GetModules`, `SaveModule`, `LoadModuleFile`, `GetSettings`, `SaveSettings`, `Close`
- [x] T004 [P] Create `storage/media.go`: define `MediaStore` interface with methods `SaveOriginal(filename string, data []byte) error`, `SaveThumbnail(filename string, data []byte) error`, `OriginalURL(filename string) string`, `ThumbnailURL(filename string) string`

**Checkpoint**: Interfaces defined, config loads from env vars, dependencies installed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extract current SQLite and filesystem implementations into interface-conforming types

- [x] T005 Create `storage/sqlite.go`: extract all item/module/settings CRUD logic from current `db.go` into a `SQLiteStore` struct implementing `Store` interface; preserve all existing query logic (FTS5, json_extract, attribute filters, batch operations, CSV export)
- [x] T006 Create `storage/local.go`: extract local filesystem media operations into a `LocalMediaStore` struct implementing `MediaStore`; `SaveOriginal` writes to `~/.omnicollect/media/originals/`, `SaveThumbnail` writes to `thumbnails/`, URL methods return `/originals/{filename}` and `/thumbnails/{filename}`
- [x] T007 Modify `imaging.go`: refactor `processImage` to return processed image bytes (original bytes + thumbnail bytes + metadata) instead of writing directly to filesystem; let the caller use `MediaStore` to persist
- [x] T008 Modify `app.go`: replace direct `*sql.DB` field with `Store` interface field; replace direct filesystem calls with `MediaStore` interface; update `Init()` to accept `Config` and instantiate the appropriate Store and MediaStore based on `IsCloudDB()` / `IsCloudStorage()`
- [x] T009 Modify `handlers.go`: update all handlers to use `Store` and `MediaStore` via the App struct instead of direct `*sql.DB` or filesystem; update image upload handler to use MediaStore for persistence; update media serving routes to proxy through MediaStore
- [x] T010 Modify `server.go`: update media routes to serve via MediaStore (local serves from filesystem; cloud proxies from S3)
- [x] T011 Verify local mode still works: run `go run . --serve` without any cloud env vars; verify all existing functionality works identically through the SQLiteStore + LocalMediaStore path

**Checkpoint**: Application works identically to before using interface-based storage; all business logic decoupled from storage backends

---

## Phase 3: User Story 1 - Cloud Database (Priority: P1) MVP

**Goal**: PostgreSQL with schema-per-tenant, tsvector FTS, JSONB attributes. All queries work.

**Independent Test**: Start PostgreSQL, run migration, verify all CRUD/search/filter operations work via curl against the PG-backed server.

### Implementation for User Story 1

- [x] T012 [US1] Create `storage/postgres.go`: implement `PostgresStore` struct with `*sql.DB` connection and `tenantSchema` string field; implement `initTenantSchema()` that runs `CREATE SCHEMA IF NOT EXISTS` + DDL (items table with JSONB attributes and tsvector search_vector, modules table, settings table, GIN index, search trigger)
- [x] T013 [US1] Implement item CRUD in `storage/postgres.go`: `InsertItem` (INSERT with RETURNING), `UpdateItem` (UPDATE with RETURNING), `DeleteItem` (DELETE), `DeleteItems` (batch in transaction), `BulkUpdateModule` (batch in transaction) -- all scoped to tenant schema via `SET search_path`
- [x] T014 [US1] Implement `QueryItems` in `storage/postgres.go`: translate SQLite FTS5 MATCH to `search_vector @@ plainto_tsquery(?)`; translate `json_extract(attributes, '$.field')` to `attributes->>'field'`; translate attribute filter operators (`IN`, `=`, `>=`, `<=`) to JSONB equivalents; combine with module filter and ORDER BY `ts_rank` or `updated_at`
- [x] T015 [US1] Implement module operations in `storage/postgres.go`: `GetModules` queries the `modules` table and unmarshals `schema_json` JSONB into `ModuleSchema` structs; `SaveModule` upserts into the `modules` table; `LoadModuleFile` returns `schema_json` as a string
- [x] T016 [US1] Implement settings and CSV export in `storage/postgres.go`: `GetSettings` / `SaveSettings` operate on the `settings` table; `ExportItemsCSV` queries items by ID and generates CSV (same logic as SQLiteStore, different SQL syntax)
- [x] T017 [US1] Create `storage/migrate.go`: implement `MigrateToPostgres(sqlitePath, pgStore, modulesDir)` that opens the SQLite DB, creates the tenant schema, copies all items (adapting JSON columns to JSONB), reads module JSON files from disk and inserts into modules table, reports counts
- [x] T018 [US1] Modify `main.go`: add `--migrate` flag with `--source` and `--tenant` parameters; when set, run migration and exit; initialize PostgresStore when `DATABASE_URL` is set
- [x] T019 [US1] Integration test: start PostgreSQL (via docker-compose or local), run migration, start server with `DATABASE_URL`, verify all endpoints work via curl (list modules, CRUD items, search, filter, batch delete, CSV export, settings)

**Checkpoint**: Full PostgreSQL backend functional with schema-per-tenant isolation, tsvector FTS, JSONB queries

---

## Phase 4: User Story 2 - Cloud Object Storage (Priority: P2)

**Goal**: Images stored in S3-compatible object store. Frontend loads images from backend proxy.

**Independent Test**: Start MinIO, configure S3 env vars, upload an image, verify it appears in MinIO bucket, verify it loads in the browser.

### Implementation for User Story 2

- [x] T020 [US2] Create `storage/s3.go`: implement `S3MediaStore` struct using AWS SDK v2; `SaveOriginal` uploads to `s3://{bucket}/originals/{filename}`; `SaveThumbnail` uploads to `s3://{bucket}/thumbnails/{filename}`; URL methods return the local proxy paths (`/originals/{filename}`, `/thumbnails/{filename}`)
- [x] T021 [US2] Add S3 proxy handler in `handlers.go`: when `MediaStore` is `S3MediaStore`, the `/thumbnails/{filename}` and `/originals/{filename}` routes fetch from S3 and stream to the response (instead of serving from local filesystem)
- [x] T022 [US2] Update `server.go` media routes: detect whether MediaStore is local or S3; local mode uses `http.FileServer`; cloud mode uses the S3 proxy handler
- [x] T023 [US2] Verify image flow end-to-end: upload image via `POST /api/v1/images/upload`, verify original and thumbnail appear in MinIO bucket, verify grid view loads thumbnails via proxy, verify detail view loads originals via proxy

**Checkpoint**: Images stored in S3; frontend loads them via backend proxy; zero frontend changes

---

## Phase 5: User Story 3 - Containerization (Priority: P3)

**Goal**: Docker image packages the app; stateless container runs with env-var config.

**Independent Test**: Build image, run container with docker-compose, verify all features work.

### Implementation for User Story 3

- [x] T024 [US3] Create `Dockerfile`: multi-stage build -- stage 1: `golang:1.25-alpine` builds Go binary; stage 2: `node:18-alpine` runs `npm install && npm run build` in frontend/; stage 3: `alpine:3.19` copies binary + `frontend/dist/` + `ca-certificates`, sets ENTRYPOINT
- [x] T025 [US3] Create `docker-compose.yml`: three services (app, postgres, minio) per the infrastructure contract; app depends on postgres and minio; includes init commands for MinIO bucket creation
- [x] T026 [US3] Add health check endpoint `GET /api/v1/health` in `handlers.go` + `server.go`: returns `{"status":"ok","database":"connected","storage":"connected"}` after pinging DB and S3; returns error status if either is unreachable
- [x] T027 [US3] Build and test: `docker build -t omnicollect .` then `docker-compose up`; verify app starts, frontend loads, all features work through the container

**Checkpoint**: Containerized app runs with external PostgreSQL + S3; fully stateless

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, migration from existing db.go, cleanup

- [x] T028 Remove legacy code from `db.go`: after SQLiteStore extraction is complete and verified, remove duplicated functions from db.go (keep only what SQLiteStore doesn't cover, if any); ensure db.go is either empty or renamed
- [x] T029 [P] Update `CLAUDE.md`: document storage/ package (Store and MediaStore interfaces, SQLite/Postgres/Local/S3 implementations), config.go env vars, --migrate flag, Dockerfile, docker-compose, health endpoint
- [x] T030 [P] Update `README.md`: add Cloud Deployment section (env vars, Docker, docker-compose, migration CLI), update architecture description, add iteration 11 to history
- [x] T031 Run quickstart.md full acceptance test flow (all 22 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs interfaces + config)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs Store interface + App refactor)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (needs MediaStore interface + handler refactor); independent of US1
- **User Story 3 (Phase 5)**: Depends on Phase 3 + Phase 4 (container needs both backends working)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. This is the MVP -- cloud database with all queries working.
- **US2 (P2)**: Depends on Foundational. Independent of US1 (different storage backend). Can parallel with US1.
- **US3 (P3)**: Depends on US1 + US2 (container packages the full cloud-ready app).

### Parallel Opportunities

- T003 and T004 (Phase 1 interfaces) -- different files
- T005 and T006 (Phase 2 extractions) -- different files, different concerns
- US1 (Phase 3) and US2 (Phase 4) -- different storage backends, can develop in parallel after Phase 2
- T029 and T030 (Phase 6 docs) -- different files

---

## Parallel Example: Phase 1 Interfaces

```bash
Task: "Define Store interface in storage/db.go" (T003)
Task: "Define MediaStore interface in storage/media.go" (T004)
```

## Parallel Example: US1 + US2

```bash
# After Phase 2 complete, these can develop in parallel:
Task: "Implement PostgresStore" (T012-T016)
Task: "Implement S3MediaStore" (T020-T022)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (Setup) + Phase 2 (Foundational) -- interfaces + extraction + refactor
2. Phase 3 (US1) -- PostgreSQL backend with full query support
3. **STOP and VALIDATE**: docker-compose with PostgreSQL; verify all operations via curl
4. This delivers multi-tenant cloud database support -- the hardest part

### Incremental Delivery

1. Phase 1 + Phase 2 = Interface abstraction (app works identically on SQLite via interfaces)
2. Phase 3 (US1) = PostgreSQL backend (cloud database ready)
3. Phase 4 (US2) = S3 storage (cloud images ready)
4. Phase 5 (US3) = Docker container (deployable artifact)
5. Phase 6 (Polish) = documentation, cleanup, final validation

---

## Notes

- The biggest risk is Phase 2 (extraction): moving existing db.go logic into SQLiteStore while keeping everything working. Test thoroughly after T011.
- PostgreSQL query translation (FTS5 -> tsvector, json_extract -> JSONB operators) is the most complex code in T014. Reference data-model.md query translation table.
- S3 media is proxied through the backend -- zero frontend changes needed.
- Container is stateless: all data in PostgreSQL + S3. Container restart loses nothing.
- Local mode (no env vars) continues to work unchanged -- critical for desktop users.
- Module schemas move from filesystem JSON files to a `modules` database table in cloud mode.
