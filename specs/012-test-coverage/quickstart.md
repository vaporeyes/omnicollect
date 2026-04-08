# Quickstart: Comprehensive Test Coverage

**Branch**: `012-test-coverage` | **Date**: 2026-04-08

## Prerequisites

- Go 1.25+, Node.js 18+
- Existing codebase on the `012-test-coverage` branch
- No external services needed (no PostgreSQL, no MinIO)

## New Dependencies

```bash
cd frontend && npm install -D vitest @vue/test-utils
```

## Files to Create

### Backend
1. **`storage/sqlite_test.go`** -- Unit tests for all SQLiteStore methods
2. **`storage/testdata/test-image.jpg`** -- Small test fixture image
3. **`storage/testdata/test-module.json`** -- Sample module schema
4. **`handlers_test.go`** -- HTTP handler integration tests
5. **`imaging_test.go`** -- Image processing tests

### Frontend
6. **`frontend/vitest.config.ts`** -- Vitest configuration
7. **`frontend/src/stores/collectionStore.test.ts`** -- collectionStore tests
8. **`frontend/src/stores/moduleStore.test.ts`** -- moduleStore tests
9. **`frontend/src/stores/selectionStore.test.ts`** -- selectionStore tests
10. **`frontend/src/stores/toastStore.test.ts`** -- toastStore tests
11. **`frontend/src/api/client.test.ts`** -- API client tests

## Implementation Order

1. Install Vitest dependency
2. Create test fixtures (testdata/)
3. Create Go test helpers (newTestStore)
4. Write storage unit tests (sqlite_test.go)
5. Write handler integration tests (handlers_test.go)
6. Configure Vitest (vitest.config.ts)
7. Write frontend store tests
8. Write API client tests
9. Verify coverage targets
10. Update CLAUDE.md and README

## Running Tests

```bash
# Backend (from project root)
go test ./... -v
go test ./storage/... -coverprofile=coverage.out
go tool cover -func=coverage.out   # view coverage

# Frontend (from frontend/)
npx vitest run
npx vitest run --coverage          # with coverage report
```

## Acceptance Test Flow

1. `go test ./storage/...` -- all storage tests pass
2. `go test ./storage/... -cover` -- reports 80%+ line coverage
3. `go test .` -- handler tests pass (root package)
4. `go test ./...` -- all Go tests pass, under 15 seconds
5. `cd frontend && npx vitest run` -- all frontend tests pass, under 15 seconds
6. Verify each Store method has at least one test
7. Verify each REST endpoint has at least one handler test
8. Verify each Pinia store action has at least one test
