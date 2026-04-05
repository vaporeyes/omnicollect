# Tasks: Core Engine (Data & IPC)

**Input**: Design documents from `/specs/001-core-engine-data-ipc/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/wails-bindings.md

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Wails desktop app**: Go backend at project root, Vue 3 frontend under `frontend/`
- Generated bindings in `frontend/wailsjs/` (auto-generated, never edit)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Wails project and Go module with all dependencies.

- [x] T001 Initialize Wails v2 project with Vue 3 + TypeScript template via `wails init -n omnicollect -t vue-ts` (or scaffold manually if repo already has structure) and configure `wails.json`
- [x] T002 Add Go dependencies: `modernc.org/sqlite` and `github.com/google/uuid` in `go.mod`
- [x] T003 [P] Define shared Go types (Item, ModuleSchema, AttributeSchema, DisplayHints) in `models.go` with JSON struct tags per contracts/wails-bindings.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database initialization and module loader -- MUST be complete before ANY user story can be implemented.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Implement SQLite database initialization in `db.go`: open database file at user config dir (`~/Library/Application Support/OmniCollect/collection.db` on macOS, equivalent on other platforms), create `items` table with schema from data-model.md, enable WAL mode and busy timeout via DSN pragmas
- [x] T005 Implement FTS5 virtual table creation in `db.go`: create `items_fts` table and INSERT/UPDATE/DELETE triggers that flatten `attributes` JSON via `json_each()` per research.md R2
- [x] T006 Implement module schema loader in `modules.go`: scan `~/.omnicollect/modules/` directory at startup, create directory if missing, parse each `.json` file into `ModuleSchema` struct, validate required fields (non-empty ID, displayName, unique attribute names), skip and log malformed files, reject duplicate module IDs
- [x] T007 Wire App struct in `app.go`: create `App` struct with `*sql.DB` and `[]ModuleSchema` fields, implement `startup` lifecycle hook that calls db init (T004/T005) and module loader (T006), register App in `main.go` via `options.Bind`
- [x] T008 Configure `main.go` entry point: set up `wails.Run()` with App binding, embed directive for `frontend/dist`, window title "OmniCollect", and `OnStartup` hook

**Checkpoint**: Database initializes on launch, module schemas load from disk, App struct is bound to Wails runtime.

---

## Phase 3: User Story 1 - Save and Retrieve a Collection Item (Priority: P1)

**Goal**: A collector can save a new item with title, price, and custom attributes, then retrieve it by ID or browse by module.

**Independent Test**: Save an item via `SaveItem()`, then call `GetItems("", moduleId)` and verify the item appears with all fields intact.

### Implementation for User Story 1

- [x] T009 [US1] Implement `SaveItem` method on App struct in `app.go`: generate UUID v4 for new items (empty ID), set `created_at`/`updated_at` timestamps, upsert into `items` table, return saved Item with populated fields. Validate `module_id` and `title` are non-empty, return error otherwise.
- [x] T010 [US1] Implement `GetItems` method on App struct in `app.go` (browse mode only, no search): accept `query` and `moduleID` string params, when both empty return all items ordered by `updated_at` DESC, when `moduleID` is non-empty filter by `module_id` column. FTS search wired in US2.
- [x] T011 [US1] Implement helper SQL functions in `db.go`: `insertItem`, `updateItem`, `queryItems` with parameterized queries. Use `json.Marshal`/`json.Unmarshal` to convert between Go maps and JSON text columns (`images`, `attributes`).

**Checkpoint**: `SaveItem` creates/updates items in SQLite, `GetItems` returns items filtered by module. Round-trip under 1 second verified manually via `wails dev`.

---

## Phase 4: User Story 2 - Search Across Collections (Priority: P2)

**Goal**: A collector can search by keyword and find items matching in title or any attribute value.

**Independent Test**: Insert items with known attribute values, call `GetItems("keyword", "")`, verify correct items returned.

### Implementation for User Story 2

- [x] T012 [US2] Extend `GetItems` in `app.go` to support FTS5 search: when `query` is non-empty, join `items` with `items_fts` using `MATCH` clause, rank results by FTS5 relevance. Combine with `moduleID` filter when both provided.
- [x] T013 [US2] Update `queryItems` in `db.go`: add FTS5 query path that searches `items_fts` for the query term, join back to `items` table by rowid, return results ordered by FTS5 rank.

**Checkpoint**: FTS5 search finds items by title keywords and attribute values. Empty results return `[]` without error. Search under 500ms for 10k items verified with seed data.

---

## Phase 5: User Story 3 - Discover Available Collection Types (Priority: P3)

**Goal**: The application loads module schemas at startup and exposes them to the frontend so the collector can see which collection types are available.

**Independent Test**: Place test JSON schema files in `~/.omnicollect/modules/`, start app, call `GetActiveModules()`, verify all valid schemas returned.

### Implementation for User Story 3

- [x] T014 [US3] Implement `GetActiveModules` method on App struct in `app.go`: return the `[]ModuleSchema` slice loaded during startup.
- [x] T015 [US3] Create sample module schema file at `~/.omnicollect/modules/coins.json` per quickstart.md for manual verification.

**Checkpoint**: `GetActiveModules()` returns parsed schemas. Malformed files are skipped with logged errors. Empty modules directory returns `[]`.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories.

- [x] T016 [P] Add ABOUTME comments to all Go source files (`main.go`, `app.go`, `db.go`, `modules.go`, `models.go`) per project guidelines
- [x] T017 [P] Add Go doc comments to all exported types and methods in `models.go` and `app.go`
- [x] T018 Run `quickstart.md` validation: follow all steps end-to-end, verify SaveItem/GetItems/GetActiveModules/FTS5 search work as documented
- [x] T019 Run `wails build` and verify single binary output in `build/bin/` launches correctly

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion -- BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (T004-T008)
- **User Story 2 (Phase 4)**: Depends on User Story 1 (extends GetItems from T010)
- **User Story 3 (Phase 5)**: Depends on Foundational only (T006 already loads modules). Can run in parallel with US1 if desired.
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2 complete. No dependency on other stories.
- **User Story 2 (P2)**: Requires US1 complete (extends `GetItems` with FTS path).
- **User Story 3 (P3)**: Requires Phase 2 only. Can start in parallel with US1.

### Within Each User Story

- Models/helpers before service methods
- Service methods before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T002 and T003 can run in parallel (different files)
- T016 and T017 can run in parallel (different files)
- US1 and US3 can run in parallel after Phase 2 (independent concerns)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008)
3. Complete Phase 3: User Story 1 (T009-T011)
4. **STOP and VALIDATE**: Save and retrieve items via `wails dev`
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add User Story 1 -> CRUD works -> Demo (MVP!)
3. Add User Story 2 -> Search works -> Demo
4. Add User Story 3 -> Module discovery exposed -> Demo
5. Polish -> Production-ready

### Recommended Execution Order

Since US2 depends on US1 but US3 is independent:

```
Phase 1 (Setup)
  |
Phase 2 (Foundational)
  |
  +-- Phase 3 (US1: CRUD) ---> Phase 4 (US2: Search)
  |
  +-- Phase 5 (US3: Modules) [parallel with US1]
  |
Phase 6 (Polish)
```

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable at its checkpoint
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
