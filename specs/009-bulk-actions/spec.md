# Feature Specification: Multi-Select and Bulk Actions

**Feature Branch**: `009-bulk-actions`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Add multi-select with checkboxes in list view and selection overlay in grid view. Shift-click for range selection. Floating action bar with bulk edit module, export CSV, and delete selected actions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Select and Delete Multiple Items (Priority: P1)

A collector wants to clean up their collection by removing several items at once. They click checkboxes next to items in the list view (or click card overlays in grid view) to select them. A floating action bar appears at the bottom showing "5 items selected" with a "Delete Selected" button. They click it, confirm the destructive action, and all selected items are removed.

**Why this priority**: Multi-select with bulk delete is the highest-value combination. It solves the most painful workflow gap -- currently users must delete items one at a time through the detail view.

**Independent Test**: Select multiple items via checkboxes, verify action bar appears with count, click "Delete Selected", confirm, verify all selected items are removed from the collection.

**Acceptance Scenarios**:

1. **Given** the list view is showing items, **When** the user clicks a checkbox next to an item, **Then** the item is visually marked as selected and the floating action bar appears showing "1 item selected".
2. **Given** the grid view is showing items, **When** the user clicks a card's selection overlay, **Then** the card shows a selected state and the action bar appears.
3. **Given** 3 items are selected, **When** the user clicks "Delete Selected" on the action bar, **Then** a confirmation dialog appears showing the count of items to be deleted.
4. **Given** the confirmation dialog is showing, **When** the user confirms, **Then** all selected items are deleted, the selection is cleared, the action bar disappears, and a toast notification confirms the deletion count.
5. **Given** items are selected, **When** the user clicks a selected item's checkbox/overlay again, **Then** it deselects and the count updates. If count reaches zero, the action bar disappears.
6. **Given** items are selected, **When** the user navigates away (opens detail view, switches module, etc.), **Then** the selection is cleared.

---

### User Story 2 - Shift-Click Range Selection (Priority: P2)

A collector wants to quickly select a contiguous range of items. They click one item's checkbox, then Shift-click another item further down the list. All items between the two (inclusive) become selected at once.

**Why this priority**: Range selection is a standard power-user pattern that dramatically speeds up selection of adjacent items. Without it, users must click each item individually.

**Independent Test**: Click one checkbox, Shift-click another checkbox several rows away, verify all items in the range are selected.

**Acceptance Scenarios**:

1. **Given** no items are selected, **When** the user clicks item A's checkbox, **Then** item A is selected and becomes the range anchor.
2. **Given** item A is selected, **When** the user Shift-clicks item D's checkbox, **Then** items A through D (inclusive) are all selected.
3. **Given** items A-D are selected via range, **When** the user Shift-clicks item B (within the range), **Then** the range adjusts to A-B (shrinks to new endpoint).
4. **Given** items are selected, **When** the user clicks (without Shift) on a new item, **Then** previous selection is cleared and only the new item is selected (new anchor set).
5. **Given** the grid view is active, **When** the user Shift-clicks, **Then** range selection works based on the visual order of items in the grid.

---

### User Story 3 - Export Selected Items as CSV (Priority: P3)

A collector wants to export a subset of their collection to a spreadsheet. They select the desired items, then click "Export Selected as CSV" on the action bar. A save dialog opens and a CSV file is generated containing the selected items' data (title, module, purchase price, and all attributes).

**Why this priority**: CSV export is valuable but less frequently needed than delete. It extends the selection system with a non-destructive action.

**Independent Test**: Select multiple items, click "Export Selected as CSV", choose a save location, verify a valid CSV file is created with correct data for each selected item.

**Acceptance Scenarios**:

1. **Given** 5 items are selected, **When** the user clicks "Export Selected as CSV", **Then** a native save dialog opens with a default filename (e.g., "omnicollect-export-5-items.csv").
2. **Given** the save dialog is confirmed, **Then** a CSV file is written with a header row and one data row per selected item, including: title, module name, purchase price, and each attribute as a column.
3. **Given** selected items span multiple modules with different schemas, **Then** the CSV includes the union of all attribute columns; items missing an attribute have empty cells for those columns.
4. **Given** the export completes, **Then** a toast notification confirms success with the filename.
5. **Given** the user cancels the save dialog, **Then** no file is created and the selection remains intact.

---

### User Story 4 - Bulk Edit Module (Priority: P4)

A collector imported items under the wrong collection type and wants to move them to a different module. They select the items, click "Bulk Edit Module" on the action bar, choose the target module from a dropdown, confirm, and all selected items' module assignment is updated.

**Why this priority**: Module reassignment is a niche but painful manual task. It rounds out the bulk action set but is the least commonly needed.

**Independent Test**: Select items, click "Bulk Edit Module", choose a different module, confirm, verify all selected items now belong to the new module.

**Acceptance Scenarios**:

1. **Given** 3 items are selected (all from module "Coins"), **When** the user clicks "Bulk Edit Module", **Then** a dialog appears with a dropdown of all available modules.
2. **Given** the module dialog is showing, **When** the user selects "Stamps" and confirms, **Then** all 3 items' module assignment updates to "Stamps", the list refreshes, and a toast confirms the change.
3. **Given** the user cancels the module dialog, **Then** no changes are made and the selection remains.

---

### Edge Cases

- What happens when the user selects all items and deletes them? The collection should empty and show the standard empty state. The action bar disappears.
- What happens when an item in the selection was already deleted by another action (e.g., context menu delete while selected)? The bulk action should skip missing items gracefully and report the actual count affected.
- What happens when the user switches between list and grid view while items are selected? The selection should persist across view mode changes (same item IDs).
- What happens with "Select All"? A "Select All" checkbox in the list header (or action) should select/deselect all currently visible items.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The list view MUST show a checkbox on the left of each item row for selection.
- **FR-002**: The grid view MUST show a selection overlay (e.g., checkmark badge) on each card when hovered or selected.
- **FR-003**: Clicking a checkbox or card overlay MUST toggle that item's selection state.
- **FR-004**: Shift-clicking MUST select all items between the last-clicked item and the current item (inclusive range).
- **FR-005**: When one or more items are selected, a floating action bar MUST appear at the bottom center of the screen.
- **FR-006**: The action bar MUST display the count of selected items (e.g., "3 items selected").
- **FR-007**: The action bar MUST use a frosted glass visual style consistent with the application's design language.
- **FR-008**: The action bar MUST offer a "Delete Selected" action that shows a confirmation dialog before deleting.
- **FR-009**: The action bar MUST offer an "Export Selected as CSV" action that opens a native save dialog.
- **FR-010**: The action bar MUST offer a "Bulk Edit Module" action that shows a module selection dialog.
- **FR-011**: The CSV export MUST include a header row and one data row per item with title, module, price, and all attribute columns.
- **FR-012**: For items spanning multiple modules, the CSV MUST include the union of all attribute columns with empty cells for missing attributes.
- **FR-013**: The "Delete Selected" action MUST remove all selected items in a single atomic operation and show a toast confirming the count deleted.
- **FR-014**: The "Bulk Edit Module" action MUST update the module assignment of all selected items and refresh the view.
- **FR-015**: Selection MUST be cleared when the user navigates away (detail view, form, builder, settings, module switch).
- **FR-016**: Selection MUST persist when switching between list and grid view modes.
- **FR-017**: A "Select All" control MUST be available to select or deselect all currently visible items.
- **FR-018**: The action bar MUST include a "Deselect All" button to clear the selection.

### Key Entities

- **Selection State**: A set of item IDs currently selected by the user. Ephemeral (not persisted). Shared between list and grid views.
- **Action Bar**: A floating UI element that appears when the selection is non-empty, providing bulk action buttons and selection count.
- **CSV Export**: A file containing selected items' data in comma-separated format with a header row derived from the union of all relevant schemas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can select 20 items and delete them in under 10 seconds (vs. 60+ seconds deleting one at a time).
- **SC-002**: Shift-click range selection correctly selects all items in the visual range with zero missed or extra items.
- **SC-003**: CSV export produces a valid file that opens correctly in standard spreadsheet applications.
- **SC-004**: Bulk module reassignment updates all selected items without data loss in any field.
- **SC-005**: The action bar appears within 100ms of the first item being selected and disappears immediately when selection is cleared.
- **SC-006**: Selection state is consistent between list and grid view modes (same items remain selected when switching).

## Clarifications

### Session 2026-04-07

- Q: Should bulk delete call DeleteItem per item (sequential) or use a batch binding (atomic)? -> A: Add a new batch DeleteItems(ids) backend binding that deletes all in one transaction.
- Q: Should CSV export happen in the frontend or via a backend binding? -> A: New Go backend binding accepts item IDs, queries DB, generates CSV string, returns to frontend.

## Assumptions

- Selection state is ephemeral and stored only in the frontend. It is not persisted across app restarts or page reloads.
- The "Delete Selected" action uses a dedicated batch deletion backend binding that accepts a list of item IDs and deletes them in a single atomic transaction. This ensures no partial-delete scenarios.
- The CSV export is handled by a new Go backend binding that accepts a list of item IDs, queries the database, generates the CSV string, and returns it to the frontend. The frontend then uses the Wails native save dialog for the file location and writes the content.
- "Bulk Edit Module" updates only the `moduleId` field of each item. It does not modify attributes, images, or other fields. Attributes that exist in the old schema but not the new one remain in the JSON blob (no data loss).
- The selection overlay on grid cards appears as a small checkmark in the top-left corner of the card, visible on hover and always visible when selected.
- "Select All" selects all items currently visible (after any active search/filter), not all items in the database.
