# Research: Multi-Select and Bulk Actions

**Branch**: `009-bulk-actions` | **Date**: 2026-04-07

## R1: Batch Delete Strategy

**Decision**: New `DeleteItems(ids []string)` Go binding that wraps all deletes in a single SQLite transaction. FTS5 cleanup triggers fire per-row automatically.

**Rationale**: Per the clarification session, atomic batch delete prevents partial-failure scenarios. SQLite transactions are fast for bulk operations (hundreds of items in <100ms). The existing `items_ad` FTS5 trigger handles index cleanup per row within the transaction.

**Alternatives considered**:
- Sequential DeleteItem calls: Rejected per clarification -- slow for large selections, non-atomic.
- Soft-delete (mark as deleted): Rejected -- adds complexity; OmniCollect is single-user with no undo requirement beyond the confirmation dialog.

## R2: CSV Export Architecture

**Decision**: New `ExportItemsCSV(ids []string)` Go binding that queries the database for the specified items, builds a CSV string with union of all attribute columns, and returns the string. Frontend uses Wails `SaveFileDialog` to get the file path, then writes the string via a second binding or Wails file API.

**Rationale**: Per clarification, backend handles CSV generation. This ensures correct attribute column union computation (multiple modules may have different schemas) and avoids large data transfer of raw items to the frontend for processing.

**CSV column order**: id, title, module, purchasePrice, createdAt, updatedAt, then all unique attribute keys alphabetically sorted.

**Alternatives considered**:
- Frontend CSV generation: Rejected per clarification -- backend has direct DB access for efficient querying.

## R3: Bulk Module Update Strategy

**Decision**: New `BulkUpdateModule(ids []string, newModuleID string)` Go binding that updates all specified items' `module_id` in a single transaction. Attributes are preserved (no data loss).

**Rationale**: Atomic transaction ensures consistency. Attributes remain in the JSON blob even if the new module's schema doesn't reference them -- this matches Constitution Principle III (flat data, no migrations).

**Alternatives considered**:
- Sequential SaveItem calls: Rejected -- same atomicity concerns as delete.
- Strip attributes not in new schema: Rejected -- data loss risk; attributes should be preserved for potential future schema changes.

## R4: Selection State Architecture

**Decision**: Dedicated `selectionStore` (Pinia) holding a `Set<string>` of selected item IDs, plus a `lastClickedIndex` ref for Shift-click range tracking. The store provides `toggle(id)`, `selectRange(fromIndex, toIndex, items)`, `selectAll(items)`, `clear()`, and computed `count`/`hasSelection`.

**Rationale**: A dedicated store keeps selection logic isolated from the collection store. Using a Set ensures O(1) lookup for `isSelected(id)` checks needed by every row/card render. The store is shared between list and grid views.

**Alternatives considered**:
- Selection state in collectionStore: Rejected -- separation of concerns; selection is transient UI state, not data management.
- Selection state in component: Rejected -- would not persist across list/grid view switches (FR-016).

## R5: Shift-Click Range Selection

**Decision**: Track `lastClickedIndex` in selectionStore. On Shift-click, compute the range between `lastClickedIndex` and the clicked index in the current items array, and add all IDs in that range to the selection set. Without Shift, clear selection and set new anchor.

**Rationale**: Standard OS-level range selection pattern (Finder, Windows Explorer). Uses the visible items array order, which works for both list and grid views.

**Alternatives considered**:
- Shift adds to existing selection (no clear): More complex but some apps do this. Rejected for simplicity -- standard pattern is Shift replaces range from anchor.

## R6: CSV File Writing

**Decision**: The backend `ExportItemsCSV` returns the CSV string. The frontend calls `SaveFileDialog` for the path, then passes both path and CSV content to a new `WriteFile(path, content)` utility binding (or reuses an existing file-write mechanism).

**Rationale**: Wails SaveFileDialog returns a path but doesn't write content. The backend must write the file since the frontend webview can't write to arbitrary filesystem paths. A simple `WriteFile` binding keeps this generic and reusable.

**Alternatives considered**:
- Backend writes directly after generating CSV: Would need the save dialog result passed to the backend, which is the same flow but combined into one call. Keeping them separate is cleaner.
