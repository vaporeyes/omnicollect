# Data Model: Comprehensive Test Coverage

**Branch**: `012-test-coverage` | **Date**: 2026-04-08

## No Data Model Changes

This feature adds tests only. No database schema, API, or storage changes.

## Test Data Fixtures

### Go Test Helpers

A `newTestStore(t)` helper creates a fresh `:memory:` SQLiteStore with schema initialized, returning it ready for use. The store is automatically closed when the test ends via `t.Cleanup`.

### Sample Module Schema (testdata/test-module.json)

```json
{
  "id": "test-coins",
  "displayName": "Test Coins",
  "attributes": [
    {"name": "year", "type": "number", "required": true},
    {"name": "condition", "type": "enum", "options": ["Mint", "Fine", "Poor"]},
    {"name": "isGraded", "type": "boolean"}
  ]
}
```

### Sample Items for Tests

Tests create items programmatically via the Store interface. Standard test items include:
- Item with all fields populated (title, price, images, attributes)
- Item with minimal fields (title + moduleId only)
- Item with null/missing attributes (for filter edge cases)
- Multiple items for search/filter/batch test scenarios
