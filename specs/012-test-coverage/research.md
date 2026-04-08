# Research: Comprehensive Test Coverage

**Branch**: `012-test-coverage` | **Date**: 2026-04-08

## R1: Go Test Strategy for SQLiteStore

**Decision**: Use `:memory:` SQLite databases via the existing `initDB`-style setup, creating a fresh database per test function. Each test calls a helper that returns a configured `SQLiteStore` with the schema already created.

**Rationale**: In-memory databases are fast (no disk I/O), automatically cleaned up when the connection closes, and provide complete isolation between tests. The `modernc.org/sqlite` driver supports `:memory:` connections.

**Alternatives considered**:
- Temp file databases: Slightly more realistic but slower and require cleanup. Rejected for unit tests.
- Shared test database with cleanup: Risk of cross-test contamination. Rejected.

## R2: Go Handler Test Strategy

**Decision**: Use `httptest.NewServer` wrapping the real `Server` struct, backed by a fresh `:memory:` SQLiteStore per test. Tests make real HTTP requests to the test server and assert on response status codes, headers, and JSON bodies.

**Rationale**: Integration tests at the HTTP boundary catch serialization bugs, routing errors, and middleware issues that unit tests miss. Using a real database (not mocks) ensures the full stack is exercised.

**Alternatives considered**:
- Mock Store interface: Would test handler logic in isolation but miss serialization and real query issues. Rejected for integration tests (could add later for unit-level handler tests).

## R3: Frontend Test Runner

**Decision**: Use Vitest as the test runner. Add `vitest` and `@vue/test-utils` as dev dependencies.

**Rationale**: Vitest is the standard test runner for Vite-based projects. It shares the same config, transforms, and module resolution as the dev server. Zero additional configuration beyond a `vitest.config.ts` file.

**Alternatives considered**:
- Jest: Would work but requires separate Babel/TypeScript configuration. Vitest integrates natively with Vite.

## R4: Frontend Fetch Mocking

**Decision**: Mock the global `fetch` function in tests using `vi.fn()` from Vitest. Each test sets up mock responses for the specific API calls it needs.

**Rationale**: The API client (`api/client.ts`) uses native `fetch`. Mocking `fetch` at the global level is the simplest approach -- no additional mocking library needed. Each test controls exactly what responses are returned.

**Alternatives considered**:
- MSW (Mock Service Worker): More realistic but adds a dependency and complexity. Overkill for store-level unit tests.
- Inject mock client into stores: Would require refactoring stores to accept a client parameter. Too invasive for this iteration.

## R5: Test Fixture Strategy

**Decision**: Go test fixtures in `storage/testdata/` (a small JPEG and a sample module JSON). Frontend tests use inline mock data objects.

**Rationale**: Go convention is `testdata/` directories. Frontend tests don't need file fixtures -- mock API responses are defined inline for clarity.
