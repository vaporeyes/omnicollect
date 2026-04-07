# Quickstart: Multi-Select and Bulk Actions

**Branch**: `009-bulk-actions` | **Date**: 2026-04-07

## Prerequisites

- Wails v2 development environment (`wails dev` works)
- Existing codebase on the `009-bulk-actions` branch
- Several items in the database across one or more modules

## Files to Create

1. **`frontend/src/stores/selectionStore.ts`** -- Selection state: Set of IDs, toggle, range select, select all, clear
2. **`frontend/src/components/BulkActionBar.vue`** -- Floating glassmorphism action bar with count + action buttons

## Files to Modify (Backend)

1. **`app.go`** -- Add `DeleteItems`, `ExportItemsCSV`, `BulkUpdateModule`, `WriteFile` bindings
2. **`db.go`** -- Add `deleteItems` (batch transaction), `exportItemsCSV` (query + CSV generation), `bulkUpdateModule` (batch transaction)

## Files to Modify (Frontend)

1. **`frontend/src/components/ItemList.vue`** -- Add checkboxes, select-all header, wire selection store
2. **`frontend/src/components/CollectionGrid.vue`** -- Add selection overlay on cards, wire selection store
3. **`frontend/src/App.vue`** -- Render BulkActionBar, wire bulk action handlers, clear selection on navigate

## Implementation Order

1. Backend: `deleteItems`, `exportItemsCSV`, `bulkUpdateModule`, `WriteFile` in db.go + app.go
2. Selection store (selectionStore.ts)
3. BulkActionBar component
4. Wire checkboxes into ItemList.vue
5. Wire selection overlay into CollectionGrid.vue
6. Wire BulkActionBar + action handlers into App.vue
7. Update CLAUDE.md and README

## Acceptance Test Flow

1. Open list view with items
2. Click a checkbox -- item shows selected state, action bar appears at bottom with "1 item selected"
3. Click two more checkboxes -- count updates to "3 items selected"
4. Click a selected checkbox -- it deselects, count drops to 2
5. Shift-click an item 5 rows below the last clicked -- all items in range selected
6. Switch to grid view -- same items remain selected (overlays visible)
7. Click "Select All" in list view header -- all visible items selected
8. Click "Deselect All" on action bar -- selection cleared, bar disappears
9. Select 3 items, click "Delete Selected" -- confirmation dialog shows count
10. Confirm delete -- items removed, toast shows "3 items deleted", bar disappears
11. Select 5 items, click "Export Selected as CSV" -- save dialog opens
12. Save the CSV -- file contains header + 5 rows with correct data
13. Select 2 items, click "Bulk Edit Module" -- module dropdown dialog appears
14. Choose a different module, confirm -- items reassigned, toast confirms
15. Navigate to detail view -- selection clears
16. Navigate back to list -- no items selected, no action bar
