# Tasks: Comprehensive Test Coverage

**Input**: Design documents from `/specs/012-test-coverage/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: This feature IS the tests. All tasks create test files.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Install dependencies, create test fixtures and helpers

- [x] T001 Install Vitest: run `npm install -D vitest` in `frontend/`
- [x] T002 Create `frontend/vitest.config.ts`: configure Vitest with Vue plugin, resolve aliases matching vite.config.ts, set test globals
- [x] T003 Add `"test": "vitest run"` script to `frontend/package.json`
- [x] T004 [P] Create `storage/testdata/test-module.json`: sample module schema with enum (condition: Mint/Fine/Poor), number (year), and boolean (isGraded) attributes
- [x] T005 [P] Create `storage/testdata/test-image.jpg`: minimal valid JPEG file (can be a 1x1 pixel image) for image processing tests

**Checkpoint**: `npx vitest run` executes (no tests yet); test fixtures exist

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Test helpers used by all test files

- [x] T006 Create `newTestStore` helper at top of `storage/sqlite_test.go`: function that creates an in-memory SQLiteStore with schema initialized, registers `t.Cleanup` to close it, and returns the store ready for use
- [x] T007 Create `newTestServer` helper at top of `handlers_test.go`: function that creates an App with in-memory SQLiteStore + LocalMediaStore (temp dir), creates a Server, starts `httptest.NewServer`, registers cleanup, returns the server URL and App

**Checkpoint**: Helpers compile; no test functions yet

---

## Phase 3: User Story 1 - Backend Storage Layer Tests (Priority: P1) MVP

**Goal**: All 13 Store interface methods tested against SQLite. 80%+ coverage.

**Independent Test**: `go test ./storage/... -cover` passes with 80%+ line coverage.

### Implementation for User Story 1

- [x] T008 [US1] Write item CRUD tests in `storage/sqlite_test.go`: TestInsertItem (insert + query back, verify all fields), TestUpdateItem (update title/attributes, verify changes), TestDeleteItem (delete + verify gone), TestDeleteItem_NotFound (returns error)
- [x] T009 [US1] Write search tests in `storage/sqlite_test.go`: TestQueryItems_TextSearch (insert 3 items, search by keyword, verify only matching returned), TestQueryItems_ModuleFilter (filter by moduleId), TestQueryItems_NoResults (empty result set)
- [x] T010 [US1] Write attribute filter tests in `storage/sqlite_test.go`: TestQueryItems_EnumFilter (IN operator), TestQueryItems_BooleanFilter (EQ operator), TestQueryItems_NumberRangeFilter (GTE/LTE), TestQueryItems_CombinedFilters (enum + boolean + range together)
- [x] T011 [US1] Write batch operation tests in `storage/sqlite_test.go`: TestDeleteItems_Batch (delete 3 items at once, verify count), TestBulkUpdateModule (change moduleId for 2 items, verify)
- [x] T012 [US1] Write CSV export tests in `storage/sqlite_test.go`: TestExportItemsCSV (export 2 items from different modules, verify header has union of attribute columns, verify data rows match), TestExportItemsCSV_Empty (empty IDs returns empty string)
- [x] T013 [US1] Write module/settings tests in `storage/sqlite_test.go`: TestGetModules_Empty, TestSaveAndGetModules (save + retrieve), TestLoadModuleFile, TestGetSettings_Empty, TestSaveAndGetSettings
- [x] T014 [US1] Run `go test ./storage/... -cover -v` and verify all tests pass with 80%+ line coverage

**Checkpoint**: `go test ./storage/...` passes; coverage target met

---

## Phase 4: User Story 2 - Backend HTTP Handler Tests (Priority: P2)

**Goal**: Integration tests for all REST API endpoints via httptest.

**Independent Test**: `go test . -run TestHandler -v` passes for all handler tests.

### Implementation for User Story 2

- [x] T015 [US2] Write item endpoint tests in `handlers_test.go`: TestHandlerGetItems (200 + JSON array), TestHandlerSaveItem (POST valid item, 200 + created item), TestHandlerSaveItem_Invalid (missing title, 400 + error), TestHandlerDeleteItem (204), TestHandlerDeleteItem_NotFound (404)
- [x] T016 [US2] Write batch endpoint tests in `handlers_test.go`: TestHandlerDeleteItems (POST batch delete, 200 + count), TestHandlerBulkUpdateModule (POST module update, 200 + count)
- [x] T017 [US2] Write module endpoint tests in `handlers_test.go`: TestHandlerGetModules (200 + JSON array), TestHandlerSaveModule (POST valid schema, 200), TestHandlerLoadModuleFile (GET file content, 200)
- [x] T018 [US2] Write export endpoint tests in `handlers_test.go`: TestHandlerExportBackup (200, Content-Type application/zip, Content-Disposition header), TestHandlerExportCSV (POST with IDs, 200, Content-Type text/csv)
- [x] T019 [US2] Write settings + health endpoint tests in `handlers_test.go`: TestHandlerGetSettings (200), TestHandlerSaveSettings (PUT, 200), TestHandlerHealth (200, JSON with status)
- [x] T020 [US2] Run `go test . -run TestHandler -v` and verify all handler tests pass

**Checkpoint**: All REST endpoints have handler tests; `go test ./...` passes

---

## Phase 5: User Story 3 - Frontend Store Tests (Priority: P3)

**Goal**: Unit tests for all 4 Pinia stores + API client with mocked fetch.

**Independent Test**: `cd frontend && npx vitest run` passes.

### Implementation for User Story 3

- [x] T021 [P] [US3] Write API client tests in `frontend/src/api/client.test.ts`: test `get` (success + error), `post` (JSON body sent + response parsed), `del` (204 handling), `postFile` (FormData sent), `downloadFile` (blob + anchor creation); mock `fetch` with `vi.fn()`
- [x] T022 [P] [US3] Write collectionStore tests in `frontend/src/stores/collectionStore.test.ts`: test `fetchItems` (sets items on success, sets error on failure), `saveItem` (calls POST + re-fetches), `deleteItem` (calls DELETE + re-fetches), `searchAllItems` (returns results), `setFilter` (clears activeFilters), `setActiveFilters` (triggers re-fetch), `clearFilters`; mock fetch globally
- [x] T023 [P] [US3] Write moduleStore tests in `frontend/src/stores/moduleStore.test.ts`: test `fetchModules` (sets modules on success, sets error on failure); mock fetch
- [x] T024 [P] [US3] Write selectionStore tests in `frontend/src/stores/selectionStore.test.ts`: test `toggle` (add/remove ID, updates count), `shiftSelect` (range selection with mock items), `selectAll` (all IDs added), `clear` (empty set), `isSelected`; no fetch mocking needed (pure state)
- [x] T025 [P] [US3] Write toastStore tests in `frontend/src/stores/toastStore.test.ts`: test `show` (adds toast to array), `dismiss` (removes by ID), auto-dismiss via `vi.useFakeTimers` + `vi.advanceTimersByTime`
- [x] T026 [US3] Run `cd frontend && npx vitest run` and verify all frontend tests pass

**Checkpoint**: All store tests pass; `npm test` works

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, coverage verification, CI readiness

- [x] T027 [P] Update `CLAUDE.md` to document test commands (`go test ./...`, `npm test`), test file locations, and coverage targets
- [x] T028 [P] Update `README.md` to add Testing section with commands for running backend/frontend tests and viewing coverage
- [x] T029 Verify full suite: run `go test ./... && cd frontend && npx vitest run` -- all pass, total time under 30 seconds

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs fixtures + vitest config)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs newTestStore helper)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (needs newTestServer helper)
- **User Story 3 (Phase 5)**: Depends on Phase 1 (needs vitest config); independent of US1/US2
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- storage tests.
- **US2 (P2)**: Depends on Foundational. Independent of US1 (different test file).
- **US3 (P3)**: Depends on Setup only. Fully independent of US1/US2 (frontend vs backend).

### Parallel Opportunities

- T004 and T005 (Phase 1 fixtures) -- different files
- US1 (Phase 3), US2 (Phase 4), and US3 (Phase 5) -- backend and frontend tests are fully independent
- T021-T025 (Phase 5 frontend stores) -- all different test files
- T027 and T028 (Phase 6 docs) -- different files

---

## Parallel Example: All Three User Stories

```bash
# After Phase 2, all three can run in parallel:
# Backend storage tests (US1):
Task: "Write item CRUD tests" (T008-T013)
# Backend handler tests (US2):
Task: "Write endpoint tests" (T015-T019)
# Frontend store tests (US3):
Task: "Write store tests" (T021-T025)
```

## Parallel Example: Frontend Store Tests

```bash
# All five test files are independent:
Task: "API client tests" (T021)
Task: "collectionStore tests" (T022)
Task: "moduleStore tests" (T023)
Task: "selectionStore tests" (T024)
Task: "toastStore tests" (T025)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (Setup) + Phase 2 (Foundational)
2. Phase 3 (US1) -- storage layer tests with 80%+ coverage
3. **STOP and VALIDATE**: `go test ./storage/... -cover` passes
4. This delivers the highest-value test coverage (data layer correctness)

### Incremental Delivery

1. Phase 1 + Phase 2 = test infrastructure ready
2. Phase 3 (US1) = storage tests (most critical)
3. Phase 4 (US2) = handler tests (API contract validation)
4. Phase 5 (US3) = frontend store tests (UI logic validation)
5. Phase 6 = docs + final verification

---

## Notes

- All Go tests use `:memory:` SQLite -- no external database needed
- All frontend tests mock `fetch` -- no running backend needed
- `go test ./...` covers both storage and handler tests
- `npm test` (or `npx vitest run`) covers all frontend tests
- Test files follow naming conventions: `*_test.go` for Go, `*.test.ts` for TypeScript
- No Wails runtime dependency in any test
