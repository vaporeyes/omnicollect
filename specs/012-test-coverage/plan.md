# Implementation Plan: Comprehensive Test Coverage

**Branch**: `012-test-coverage` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-test-coverage/spec.md`

## Summary

Add test suites for Go backend (storage unit tests + HTTP handler integration tests) and Vue frontend (Pinia store unit tests). Backend uses Go's `testing` package with temporary SQLite databases. Frontend uses Vitest with mocked `fetch`. Target 80%+ storage package coverage, all endpoints tested, all store actions tested.

## Technical Context

**Language/Version**: Go 1.25+ (backend tests), TypeScript + Vue 3 (frontend tests)
**Primary Dependencies**: Go `testing` + `net/http/httptest` (backend), Vitest (frontend)
**Storage**: Temporary SQLite `:memory:` databases for test isolation
**Testing**: Go `testing` (backend), Vitest (frontend) -- no third-party test frameworks
**Target Platform**: Local development (tests run without external services)
**Project Type**: Desktop/web hybrid with REST API
**Performance Goals**: Full test suite under 30 seconds
**Constraints**: No Wails dependency; no PostgreSQL required; no running backend for frontend tests

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Tests use local temp databases; no cloud dependencies |
| II. Schema-Driven UI | N/A | Tests validate existing behavior, don't add UI |
| III. Flat Data Architecture | PASS | Tests verify JSON attribute queries work correctly |
| IV. Performance & Memory | PASS | No media changes |
| V. Type-Safe IPC | PASS | Handler tests validate REST API contract |
| VI. Documentation | PASS | Spec artifacts produced; test conventions documented |

All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/012-test-coverage/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Backend tests
storage/
  sqlite_test.go       # New: unit tests for all SQLiteStore methods
  testdata/
    test-image.jpg     # New: small test fixture image
    test-module.json   # New: sample module schema for tests
handlers_test.go       # New: HTTP handler integration tests
imaging_test.go        # New: image processing tests

# Frontend tests
frontend/
  vitest.config.ts     # New: Vitest configuration
  src/
    stores/
      collectionStore.test.ts  # New: collectionStore unit tests
      moduleStore.test.ts      # New: moduleStore unit tests
      selectionStore.test.ts   # New: selectionStore unit tests
      toastStore.test.ts       # New: toastStore unit tests
    api/
      client.test.ts           # New: API client unit tests
```

**Structure Decision**: Go tests follow convention (`_test.go` alongside source). Frontend tests use `.test.ts` suffix alongside source files. Test fixtures in `testdata/` per Go convention.

## Complexity Tracking

No constitution violations. No complexity justification needed.
