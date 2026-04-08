# Feature Specification: Comprehensive Test Coverage

**Feature Branch**: `012-test-coverage`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Ensure we have adequate testing across both the Go backend and the Vue frontend."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Storage Layer Tests (Priority: P1)

A developer makes changes to the database queries or storage logic and needs confidence that existing behavior is preserved. Unit tests for the storage layer verify that all CRUD operations, text search, attribute filtering, batch operations, and CSV export produce correct results against a real test database. Tests run locally without external dependencies.

**Why this priority**: The storage layer is the foundation of the application. Bugs here cause data loss or incorrect query results. This is where test coverage provides the highest return on investment.

**Independent Test**: Run `go test ./storage/...` and verify all tests pass with zero failures. Tests use an in-memory or temporary SQLite database, requiring no external setup.

**Acceptance Scenarios**:

1. **Given** the test suite is run, **When** `go test ./storage/...` executes, **Then** all tests pass and report coverage for the storage package.
2. **Given** a test for item creation, **When** an item is inserted and queried back, **Then** all fields (title, moduleId, attributes, images, price) match the original.
3. **Given** a test for text search, **When** items are inserted and searched by keyword, **Then** only matching items are returned in ranked order.
4. **Given** a test for attribute filters, **When** items with different attribute values are inserted and filtered, **Then** enum (IN), boolean (EQ), and number range (GTE/LTE) filters return correct subsets.
5. **Given** a test for batch delete, **When** multiple items are deleted in one call, **Then** all specified items are removed and the count matches.
6. **Given** a test for CSV export, **When** items spanning multiple modules are exported, **Then** the CSV contains correct headers (union of attributes) and data rows.

---

### User Story 2 - Backend HTTP Handler Tests (Priority: P2)

A developer modifies an API endpoint and needs to verify it returns correct status codes, response bodies, and handles error cases properly. Integration tests for the HTTP handlers exercise the full request-response cycle against a test server with a real database, validating JSON serialization, error responses, and content types.

**Why this priority**: HTTP handlers are the API contract boundary. Testing them catches serialization bugs, incorrect status codes, and missing error handling that unit tests on the storage layer alone would miss.

**Independent Test**: Run `go test ./...` (excluding storage, which runs separately) and verify handler tests pass. Tests start a test HTTP server backed by a temporary database.

**Acceptance Scenarios**:

1. **Given** a test for `GET /api/v1/items`, **When** items exist, **Then** the response is 200 with a JSON array of items.
2. **Given** a test for `POST /api/v1/items` with valid data, **Then** the response is 200 with the created item including a generated ID.
3. **Given** a test for `POST /api/v1/items` with missing title, **Then** the response is 400 with a JSON error message.
4. **Given** a test for `DELETE /api/v1/items/{id}` with a non-existent ID, **Then** the response is 404.
5. **Given** a test for `POST /api/v1/images/upload` with a valid image file, **Then** the response is 200 with processing result metadata.
6. **Given** a test for `GET /api/v1/export/backup`, **Then** the response has Content-Type `application/zip` and Content-Disposition header.

---

### User Story 3 - Frontend Store Tests (Priority: P3)

A developer modifies a Pinia store and needs to verify that state management, API call orchestration, and error handling work correctly. Unit tests for the key stores (collectionStore, moduleStore, selectionStore, toastStore) validate reactive state changes, action behavior, and edge cases using mocked API responses.

**Why this priority**: Stores contain the frontend's business logic and API orchestration. Testing them in isolation (with mocked HTTP) catches logic bugs without requiring a running backend.

**Independent Test**: Run `npm test` in the frontend directory. Tests use a test runner with mocked fetch responses.

**Acceptance Scenarios**:

1. **Given** a test for collectionStore.fetchItems, **When** the API returns items, **Then** the store's items ref is populated and loading/error states are correct.
2. **Given** a test for collectionStore.fetchItems with API error, **Then** the error ref is set and items remain unchanged.
3. **Given** a test for selectionStore.toggle, **When** toggling an item ID, **Then** selectedIds contains/removes it and count updates.
4. **Given** a test for selectionStore.shiftSelect, **When** a range is selected, **Then** all items in the range are in selectedIds.
5. **Given** a test for toastStore.show, **When** a toast is added, **Then** it appears in the toasts array and auto-dismisses after the specified duration.

---

### Edge Cases

- What happens when tests run in parallel and share database state? Each test should use its own isolated database instance (temp file or in-memory) to prevent cross-test contamination.
- What happens when a test depends on timing (e.g., toast auto-dismiss)? Use fake timers or assertion helpers that handle async timing rather than real delays.
- What happens when the API client tests need network calls? All HTTP calls in frontend tests should be mocked -- no real backend required.
- What happens when Go tests need the Wails runtime? Tests should not depend on Wails; the storage and handler layers are Wails-independent.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend MUST have unit tests for all Store interface methods (QueryItems, InsertItem, UpdateItem, DeleteItem, DeleteItems, BulkUpdateModule, ExportItemsCSV, GetModules, SaveModule, GetSettings, SaveSettings).
- **FR-002**: Storage tests MUST use a temporary database that is created fresh for each test and cleaned up afterward.
- **FR-003**: The backend MUST have HTTP handler tests that exercise request parsing, JSON response serialization, status codes, and error responses for all API endpoints.
- **FR-004**: Handler tests MUST use a test HTTP server (`httptest.NewServer`) backed by a real temporary database.
- **FR-005**: The frontend MUST have unit tests for the core Pinia stores (collectionStore, moduleStore, selectionStore, toastStore).
- **FR-006**: Frontend store tests MUST mock all HTTP API calls and not require a running backend.
- **FR-007**: All tests MUST be runnable with standard commands (`go test ./...` for backend, `npm test` for frontend) with no external service dependencies.
- **FR-008**: Backend test coverage for the storage package MUST be reported and target at least 80% line coverage.
- **FR-009**: Tests MUST not depend on the Wails runtime or desktop environment.
- **FR-010**: Each test MUST be independent and runnable in isolation (no shared mutable state between tests).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go test ./...` passes with zero failures.
- **SC-002**: Backend storage package achieves at least 80% line coverage.
- **SC-003**: `npm test` passes with zero failures in the frontend.
- **SC-004**: All 13 Store interface methods have at least one test each.
- **SC-005**: All REST API endpoints (items CRUD, modules, images, export, settings) have at least one handler test each.
- **SC-006**: All 4 core Pinia stores have at least one test per exported action.
- **SC-007**: Tests complete in under 30 seconds total (backend + frontend combined).

## Assumptions

- Backend tests use Go's standard `testing` package and `net/http/httptest` -- no third-party test frameworks needed.
- Frontend tests use Vitest (the standard test runner for Vite-based Vue projects). The `vitest` dependency will be added to the frontend.
- Storage tests use temporary SQLite databases (`:memory:` or temp files), not PostgreSQL. PostgreSQL-specific tests are out of scope for this iteration (they require a running PostgreSQL instance).
- Image processing tests use small test fixture images included in the test directory.
- Frontend component rendering tests (testing Vue component output in a DOM) are out of scope for this iteration. This iteration focuses on store logic and API integration, not visual rendering.
- The `wailsjs/` directory and Wails runtime are not required by any test. All testable code has been decoupled from Wails in prior iterations.
- Test fixtures (sample module schemas, test images) will be stored in a `testdata/` directory following Go conventions.
