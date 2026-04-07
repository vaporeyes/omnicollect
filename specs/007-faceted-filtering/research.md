# Research: Schema-Driven Faceted Filtering

**Branch**: `007-faceted-filtering` | **Date**: 2026-04-07

## R1: SQLite JSON Attribute Filtering Strategy

**Decision**: Use SQLite's `json_extract()` function to filter on the `attributes` JSON column in WHERE clauses.

**Rationale**: The `attributes` column stores item metadata as a JSON blob. SQLite's built-in `json_extract(attributes, '$.fieldName')` allows extracting typed values for comparison without schema changes or new tables. This aligns with Constitution Principle III (flat data, no JOINs).

**Alternatives considered**:
- Adding indexed columns per filterable attribute: Rejected -- violates Constitution III (would require migrations per schema change).
- Client-side filtering after fetch: Rejected -- would require fetching all items, defeating the purpose of server-side filtering and harming performance on large collections.
- FTS5 for attribute filtering: Rejected -- FTS5 is optimized for text search, not exact-match or range comparisons on typed values.

## R2: Filter Payload Format for GetItems Binding

**Decision**: Extend `GetItems` to accept a fourth parameter: a JSON string encoding an array of filter objects. Each filter specifies a field name, operator, and value(s).

**Rationale**: A JSON string is the simplest way to pass structured filter data through Wails IPC without changing the binding paradigm. The backend parses it and builds dynamic WHERE clauses. Empty string means no filters (backward compatible).

**Filter payload structure**:
```json
[
  {"field": "condition", "op": "in", "values": ["Mint", "Fine"]},
  {"field": "isGraded", "op": "eq", "value": true},
  {"field": "year", "op": "gte", "value": 1900},
  {"field": "year", "op": "lte", "value": 1950},
  {"field": "purchasePrice", "op": "gte", "value": 50.00}
]
```

Operators: `in` (enum OR match), `eq` (exact match for booleans), `gte` (>=), `lte` (<=).

**Alternatives considered**:
- Separate parameters per filter type: Rejected -- would require changing the binding signature for every new filter type.
- Map[string]interface{} parameter: Rejected -- loses type information for operators and range bounds.

## R3: Filter State Management (Frontend)

**Decision**: Store active filters in the collectionStore as a reactive map keyed by attribute name. The filter bar reads schema to render controls; user interactions update the store; store watches trigger re-fetch with the filter payload.

**Rationale**: Centralizing filter state in the store ensures both list and grid views see the same filtered results (FR-013). The store already manages `activeModuleId` and `searchQuery`; adding `activeFilters` is a natural extension.

**Alternatives considered**:
- Filter state in FilterBar component only: Rejected -- would require prop-drilling or event chains to coordinate between FilterBar, ItemList, CollectionGrid, and App.vue.
- Separate filterStore: Rejected -- unnecessary complexity for state that's tightly coupled to collection queries.

## R4: Collapsible Filter Bar UX

**Decision**: The filter bar renders collapsed by default, showing a single row with: expand/collapse toggle, count of active filters (e.g., "3 filters active"), and a "Clear all" button. Expanding reveals the full filter controls grouped by attribute.

**Rationale**: Prevents the filter bar from dominating the viewport when a schema has many filterable attributes. Power users expand when needed; casual users see a clean interface.

**Alternatives considered**:
- Always expanded: Rejected per clarification session -- too much visual noise.
- Horizontal scroll: Rejected per clarification session -- less discoverable than expand/collapse.

## R5: purchasePrice as a Filterable Field

**Decision**: The `purchasePrice` base field is included as a number range filter when any module is active, rendered as its own group in the filter bar labeled "Purchase Price".

**Rationale**: Every item has `purchasePrice` regardless of module schema. It's a natural filter dimension. Since it's a top-level column (not in the JSON `attributes` blob), the backend handles it with a direct WHERE clause rather than json_extract.

**Alternatives considered**:
- Exclude purchasePrice from faceted filters: Rejected -- it's universally useful and already displayed in list views.

## R6: Number Range Input Debounce

**Decision**: Debounce number input changes by 400ms before triggering a re-fetch.

**Rationale**: Users type digits sequentially (e.g., "1", "19", "190", "1900"). Without debounce, each keystroke triggers a query. 400ms balances responsiveness with avoiding unnecessary backend calls.

**Alternatives considered**:
- No debounce (query on every change): Rejected -- excessive queries during typing.
- Explicit "Apply" button: Rejected -- adds friction to the interaction; debounce is seamless.
