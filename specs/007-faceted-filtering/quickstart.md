# Quickstart: Schema-Driven Faceted Filtering

**Branch**: `007-faceted-filtering` | **Date**: 2026-04-07

## Prerequisites

- Wails v2 development environment (`wails dev` works)
- Existing codebase on the `007-faceted-filtering` branch
- At least one module schema with enum, boolean, and/or number attributes
- Some items saved in that module for testing filter results

## Files to Create

1. **`frontend/src/components/FilterBar.vue`** -- Collapsible filter bar containing:
   - Expand/collapse toggle with active filter count summary
   - Enum attribute groups with multi-select pills
   - Boolean attribute tri-state toggles (off/true/false)
   - Number attribute inline min/max inputs with debounce
   - "Clear all filters" button

## Files to Modify

1. **`db.go`** -- Extend `queryItems` function to:
   - Accept a filters JSON string parameter
   - Parse filter objects from JSON
   - Build dynamic WHERE clauses using `json_extract()` for attributes and direct column access for `purchasePrice`
   - Combine with existing FTS5 search and module filter

2. **`app.go`** -- Update `GetItems` binding signature:
   - Add third parameter `filtersJSON string`
   - Pass through to `queryItems`

3. **`frontend/src/stores/collectionStore.ts`** -- Add:
   - `activeFilters` reactive state (map of attribute name to filter objects)
   - Serialize filters to JSON for `GetItems` calls
   - `setFilters` and `clearFilters` actions
   - Clear filters on module switch in `setFilter`

4. **`frontend/src/App.vue`** -- Add:
   - Import and render `FilterBar` component above collection views
   - Wire `@update` and `@clear` events to collectionStore actions
   - Pass active schema and filters as props

5. **`frontend/src/components/CommandPalette.vue`** -- Update:
   - `searchAllItems` call to pass empty string for third `filtersJSON` parameter

## Implementation Order

1. Backend: extend `queryItems` and `GetItems` (Go changes)
2. Store: add filter state and serialization (collectionStore.ts)
3. FilterBar component (bulk of frontend work)
4. Wire into App.vue
5. Update CommandPalette for new GetItems signature
6. Update CLAUDE.md and README

## Acceptance Test Flow

1. Select a module with enum attributes in the sidebar
2. Verify filter bar appears (collapsed) below view controls
3. Expand the filter bar -- enum pills, boolean toggles, and number inputs visible
4. Click an enum pill (e.g., "Condition: Mint") -- list filters to matching items
5. Click a second enum pill in the same group -- both values match (OR)
6. Click an active pill to deselect -- filter removed
7. Click a boolean toggle once -- filters to true; click again -- filters to false; click again -- off
8. Enter a min value in a number range -- items below that value disappear
9. Enter a max value -- items above disappear; only range items remain
10. Click "Clear all filters" -- all filters removed, full list returns
11. Switch to "All Types" in sidebar -- filter bar disappears
12. Switch to a module with no filterable attributes -- filter bar does not appear
13. Combine text search with filters -- both apply simultaneously
14. Verify grid view shows same filtered results as list view
