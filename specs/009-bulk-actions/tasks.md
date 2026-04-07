# Tasks: Multi-Select and Bulk Actions

**Input**: Design documents from `/specs/009-bulk-actions/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ipc-contract.md, quickstart.md

**Tests**: No automated test framework in this project. Tests not requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Backend bindings and selection store shared by all user stories

- [ ] T001 [P] Implement `deleteItems(db, ids)` in `db.go`: accept `[]string` of IDs, wrap `DELETE FROM items WHERE id IN (...)` in a single transaction, return count of rows affected
- [ ] T002 [P] Implement `exportItemsCSV(db, ids, modules)` in `db.go`: query items by IDs, compute union of all attribute keys, generate CSV string with header (id, title, module, purchasePrice, createdAt, updatedAt, then attribute keys alphabetically), return CSV string
- [ ] T003 [P] Implement `bulkUpdateModule(db, ids, newModuleID)` in `db.go`: wrap `UPDATE items SET module_id=?, updated_at=? WHERE id IN (...)` in a single transaction, return count of rows affected
- [ ] T004 Add `DeleteItems(ids)`, `ExportItemsCSV(ids)`, `BulkUpdateModule(ids, newModuleID)`, and `WriteFile(path, content)` bindings to `app.go`: validate inputs, delegate to db.go helpers, `WriteFile` uses `os.WriteFile`
- [ ] T005 Create `frontend/src/stores/selectionStore.ts`: Pinia store with `selectedIds` (reactive Set<string>), `lastClickedIndex` ref, actions `toggle(id, index)`, `shiftSelect(index, items)`, `selectAll(items)`, `clear()`, computed `count`, `hasSelection`, method `isSelected(id)`

**Checkpoint**: Backend accepts batch delete/export/update; selection store manages IDs; no UI yet

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: BulkActionBar component shared by all user stories

- [ ] T006 Create `frontend/src/components/BulkActionBar.vue`: floating glassmorphism bar at bottom center, shows item count (e.g., "3 items selected"), buttons for "Delete Selected", "Export CSV", "Bulk Edit Module", "Deselect All"; uses Teleport to body; emits `delete`, `export`, `editModule`, `deselectAll`; styled with `backdrop-filter: blur()`, `var(--bg-secondary)`, z-index 2500

**Checkpoint**: Action bar component renders with correct styling; emits events on button clicks

---

## Phase 3: User Story 1 - Select and Delete Multiple Items (Priority: P1) MVP

**Goal**: Users can select items via checkboxes/overlays, see the action bar, and bulk delete with confirmation.

**Independent Test**: Click checkboxes on 3 items, verify action bar shows count, click "Delete Selected", confirm, verify items removed.

### Implementation for User Story 1

- [ ] T007 [US1] Add checkbox column to `frontend/src/components/ItemList.vue`: insert a `<th>` with select-all checkbox in header and `<td>` with per-row checkbox in body; checkbox click calls `selectionStore.toggle(item.id, index)` (with Shift detection for range); stop propagation so row click still navigates; style checkbox column narrow (32px)
- [ ] T008 [US1] Add selection overlay to `frontend/src/components/CollectionGrid.vue`: add a checkmark badge in top-left corner of each card, visible on hover and always visible when selected; clicking the badge toggles selection via `selectionStore.toggle(item.id, index)`; stop propagation so card click still navigates
- [ ] T009 [US1] Render `BulkActionBar` in `frontend/src/App.vue`: import BulkActionBar and selectionStore; render bar when `selectionStore.hasSelection` is true; pass `selectionStore.count` as prop
- [ ] T010 [US1] Implement bulk delete handler in `frontend/src/App.vue`: on `@delete` from BulkActionBar, show confirmation dialog with count, on confirm call `DeleteItems([...selectionStore.selectedIds])`, then `selectionStore.clear()`, `collectionStore.fetchItems()`, show toast with deleted count
- [ ] T011 [US1] Clear selection on navigation in `frontend/src/App.vue`: call `selectionStore.clear()` in `onItemSelect`, `onModuleSelect`, `onFilterChange`, `openSettings`, `openNewSchemaBuilder`, and `openEditSchemaBuilder`

**Checkpoint**: Multi-select works in both views; bulk delete with confirmation; selection clears on navigate

---

## Phase 4: User Story 2 - Shift-Click Range Selection (Priority: P2)

**Goal**: Shift-clicking selects all items between anchor and target.

**Independent Test**: Click one checkbox, Shift-click another 5 rows away, verify all items in range are selected.

### Implementation for User Story 2

- [ ] T012 [US2] Implement Shift-click detection in `frontend/src/components/ItemList.vue`: on checkbox click, check `event.shiftKey`; if true, call `selectionStore.shiftSelect(index, sortedItems)` instead of `toggle`; update `lastClickedIndex` on every non-Shift click
- [ ] T013 [US2] Implement Shift-click detection in `frontend/src/components/CollectionGrid.vue`: same Shift-click logic for card selection overlay, using the items array for index resolution
- [ ] T014 [US2] Implement `shiftSelect(targetIndex, items)` logic in `frontend/src/stores/selectionStore.ts`: compute range from `lastClickedIndex` to `targetIndex`, add all item IDs in that range to `selectedIds`

**Checkpoint**: Shift-click selects contiguous ranges in both views

---

## Phase 5: User Story 3 - Export Selected Items as CSV (Priority: P3)

**Goal**: Selected items export to a CSV file via save dialog.

**Independent Test**: Select 5 items, click "Export CSV", save file, verify CSV contains 5 data rows with correct columns.

### Implementation for User Story 3

- [ ] T015 [US3] Implement CSV export handler in `frontend/src/App.vue`: on `@export` from BulkActionBar, call `ExportItemsCSV([...selectionStore.selectedIds])` to get CSV string, then open `SaveFileDialog` with default filename (e.g., "omnicollect-export-5-items.csv"), if path chosen call `WriteFile(path, csvContent)`, show success toast; if cancelled do nothing
- [ ] T016 [US3] Implement `exportItemsCSV` result formatting in `db.go`: ensure module displayName resolution by accepting modules list, handle empty attribute values as empty CSV cells, properly escape commas and quotes in CSV values

**Checkpoint**: CSV export produces valid file with correct data for selected items

---

## Phase 6: User Story 4 - Bulk Edit Module (Priority: P4)

**Goal**: Selected items can be reassigned to a different module via dialog.

**Independent Test**: Select 3 items, click "Bulk Edit Module", choose new module, confirm, verify items moved.

### Implementation for User Story 4

- [ ] T017 [US4] Implement bulk module edit handler in `frontend/src/App.vue`: on `@editModule` from BulkActionBar, show a dialog with dropdown of available modules (from moduleStore), on confirm call `BulkUpdateModule([...selectionStore.selectedIds], selectedModuleId)`, then `selectionStore.clear()`, `collectionStore.fetchItems()`, show toast with updated count
- [ ] T018 [US4] Create module selection dialog markup in `frontend/src/App.vue` (or inline in BulkActionBar): modal overlay with module dropdown, Cancel and Confirm buttons, styled consistently with existing confirmation dialogs

**Checkpoint**: Bulk module reassignment works; items appear under new module after refresh

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Select All, documentation, edge cases

- [ ] T019 Implement "Select All" checkbox in `frontend/src/components/ItemList.vue`: header checkbox that toggles `selectionStore.selectAll(sortedItems)` / `selectionStore.clear()`, reflects indeterminate state when some but not all items are selected
- [ ] T020 Handle edge case: selection persists across list/grid view switches in `frontend/src/App.vue` -- verify selectionStore is not cleared when `viewMode` changes (should work by default since store is independent)
- [ ] T021 [P] Update `CLAUDE.md` to document selectionStore, BulkActionBar, new bindings (DeleteItems, ExportItemsCSV, BulkUpdateModule, WriteFile)
- [ ] T022 [P] Update project `README.md` to document multi-select, bulk actions, Shift-click, CSV export, and iteration history entry
- [ ] T023 Run quickstart.md acceptance test flow (all 16 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (BulkActionBar needs no backend but logically follows)
- **User Story 1 (Phase 3)**: Depends on Phase 1 + Phase 2 (needs store + bar + delete binding)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (adds Shift to existing checkbox/overlay click handlers)
- **User Story 3 (Phase 5)**: Depends on Phase 1 (needs ExportItemsCSV + WriteFile bindings); independent of US1/US2 for backend, but needs bar from Phase 2
- **User Story 4 (Phase 6)**: Depends on Phase 1 (needs BulkUpdateModule binding); independent of US1/US2 for backend, but needs bar from Phase 2
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Setup + Foundational. This is the MVP.
- **US2 (P2)**: Depends on US1 (extends checkbox/overlay click handlers with Shift detection)
- **US3 (P3)**: Depends on Setup + Foundational. Can parallel with US1 for backend, but UI depends on action bar from Foundational.
- **US4 (P4)**: Depends on Setup + Foundational. Can parallel with US1 for backend.

### Parallel Opportunities

- T001, T002, T003 (db.go functions) can run in parallel -- different functions in same file but independent
- T005 (selectionStore) and T001-T004 (backend) target different languages -- can run in parallel
- T007 (ItemList) and T008 (CollectionGrid) modify different files -- can run in parallel
- T021 and T022 (documentation) modify different files -- can run in parallel

---

## Parallel Example: Phase 1 Backend

```bash
# These are independent functions, can develop in parallel:
Task: "Implement deleteItems in db.go" (T001)
Task: "Implement exportItemsCSV in db.go" (T002)
Task: "Implement bulkUpdateModule in db.go" (T003)
Task: "Create selectionStore.ts" (T005)
```

## Parallel Example: Phase 3 View Integration

```bash
# Different files:
Task: "Add checkboxes to ItemList.vue" (T007)
Task: "Add selection overlay to CollectionGrid.vue" (T008)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001, T004, T005) -- just deleteItems + store
2. Complete Phase 2: Foundational (T006) -- action bar
3. Complete Phase 3: User Story 1 (T007-T011) -- select + delete
4. **STOP and VALIDATE**: Select items, delete, verify
5. This alone eliminates the biggest UX pain point (one-at-a-time deletion)

### Incremental Delivery

1. Phase 1 + Phase 2 + Phase 3 (US1) = MVP: multi-select + bulk delete
2. Add Phase 4 (US2) = Shift-click range selection (power user upgrade)
3. Add Phase 5 (US3) = CSV export
4. Add Phase 6 (US4) = Bulk module reassignment
5. Phase 7 (Polish) = Select All, documentation, final validation
6. Each phase adds actions to the existing selection system

---

## Notes

- Three new Go bindings: DeleteItems, ExportItemsCSV, BulkUpdateModule -- all use SQLite transactions for atomicity
- WriteFile utility binding for saving CSV content to user-chosen path
- Selection store (selectionStore) is separate from collectionStore -- manages ephemeral UI state
- Selection persists across list/grid view switches but clears on navigation
- Shift-click uses lastClickedIndex anchor pattern (standard OS range selection)
- CSV uses union of all attribute columns; missing attributes get empty cells
- Action bar z-index 2500 (above main content, below command palette 3000)
