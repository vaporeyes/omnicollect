# Tasks: Schema-Driven Faceted Filtering

**Input**: Design documents from `/specs/007-faceted-filtering/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ipc-contract.md, quickstart.md

**Tests**: No automated test framework in this project. Tests not requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Shared infrastructure needed before any user story

- [x] T001 Define Go `AttributeFilter` struct and JSON parsing helper in `db.go`: struct with Field, Op, Value, Values fields; `parseFilters(filtersJSON string) ([]AttributeFilter, error)` function
- [x] T002 Update `GetItems` binding signature in `app.go` to accept third parameter `filtersJSON string`, pass it through to `queryItems`
- [x] T003 Update `queryItems` in `db.go` to accept `filtersJSON string` parameter, call `parseFilters`, and build dynamic WHERE clauses using `json_extract(attributes, '$.field')` for attribute filters and direct column access for `purchasePrice`; combine with existing FTS5 and module filter logic
- [x] T004 Update all existing frontend callers of `GetItems` to pass empty string for new third parameter: `fetchItems` and `searchAllItems` in `frontend/src/stores/collectionStore.ts`, and `searchAllItems` call in `frontend/src/components/CommandPalette.vue`

**Checkpoint**: Backend accepts filter payloads; existing functionality unchanged with empty filters

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Filter state management in the store, shared by all user stories

- [x] T005 Add `activeFilters` reactive state to `frontend/src/stores/collectionStore.ts`: `ref<Record<string, AttributeFilter[]>>({})`, `setFilters(filters)` action that updates state and re-fetches, `clearFilters()` action, serialize filters to JSON array for `GetItems` calls in `fetchItems`
- [x] T006 Clear `activeFilters` when module changes in `setFilter` method of `frontend/src/stores/collectionStore.ts`
- [x] T007 Create `frontend/src/components/FilterBar.vue` scaffold: props (`schema: ModuleSchema | null`, `filters: Record<string, AttributeFilter[]>`), emits (`update`, `clear`), collapsible container with expand/collapse toggle, active filter count summary when collapsed, "Clear all" button

**Checkpoint**: Store manages filter state; FilterBar renders empty collapsible shell; module switch clears filters

---

## Phase 3: User Story 1 - Filter Items by Enum Attributes (Priority: P1) MVP

**Goal**: Enum attributes from the active schema appear as multi-select pill groups. Clicking pills filters items via backend query.

**Independent Test**: Select a module with enum attributes, expand filter bar, click an enum pill, verify filtered results. Click second pill in same group for OR. Deselect to remove.

### Implementation for User Story 1

- [x] T008 [US1] Implement enum facet rendering in `FilterBar.vue`: iterate schema attributes where `type === 'enum'`, render a labeled group with one pill per `options[]` value, style active pills with accent color
- [x] T009 [US1] Implement enum pill click handler in `FilterBar.vue`: toggle value in/out of the `in` filter for that attribute, emit `update` with full filter state including the `{field, op: 'in', values: [...]}` object
- [x] T010 [US1] Wire `FilterBar` into `frontend/src/App.vue`: import component, render above collection views (below view-controls div), pass `activeSchema` and `collectionStore.activeFilters` as props, handle `@update` by calling `collectionStore.setFilters()` and `@clear` by calling `collectionStore.clearFilters()`
- [x] T011 [US1] Verify enum filter + text search combination: ensure `fetchItems` in collectionStore serializes both `searchQuery` and `activeFilters` to `GetItems` call, confirming AND logic between FTS5 and attribute filters

**Checkpoint**: Enum filtering works end-to-end: pills render from schema, clicks filter items via backend, OR within attribute, AND with text search

---

## Phase 4: User Story 2 - Filter Items by Boolean Attributes (Priority: P2)

**Goal**: Boolean attributes appear as tri-state toggle pills cycling off -> true -> false -> off.

**Independent Test**: Select a module with boolean attributes, click a boolean toggle, verify items filter to true. Click again for false. Click again to clear.

### Implementation for User Story 2

- [x] T012 [US2] Implement boolean facet rendering in `FilterBar.vue`: iterate schema attributes where `type === 'boolean'`, render a single pill per attribute with visual state indicator (off/Yes/No)
- [x] T013 [US2] Implement tri-state click handler in `FilterBar.vue`: cycle through off (remove filter) -> true (`{field, op: 'eq', value: true}`) -> false (`{field, op: 'eq', value: false}`) -> off, emit `update`
- [x] T014 [US2] Style boolean pills with tri-state visual feedback in `FilterBar.vue`: distinct colors or labels for Yes/No states, muted appearance when off

**Checkpoint**: Boolean tri-state toggles work alongside enum filters with AND logic across attributes

---

## Phase 5: User Story 3 - Filter Items by Number Range (Priority: P3)

**Goal**: Number attributes and purchasePrice show inline min/max inputs. Range constrains filtered results.

**Independent Test**: Select a module with number attributes, enter min/max values, verify items filter to range.

### Implementation for User Story 3

- [x] T015 [US3] Implement number facet rendering in `FilterBar.vue`: iterate schema attributes where `type === 'number'`, plus always include `purchasePrice` when a module is active; render labeled group with inline min and max number inputs
- [x] T016 [US3] Implement debounced number input handler in `FilterBar.vue`: on input change, debounce 400ms, then build `gte`/`lte` filter objects for that field (omit if input is empty), emit `update`
- [x] T017 [US3] Handle `purchasePrice` as a special field in backend `queryItems` in `db.go`: when filter field is `purchasePrice`, apply WHERE clause on the `purchase_price` column directly instead of using `json_extract`

**Checkpoint**: Number range filtering works for both schema attributes and purchasePrice; combines with enum and boolean filters

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, edge cases, and refinements

- [x] T018 Handle edge case in `FilterBar.vue`: hide the entire filter bar when `schema` is null or schema has no filterable attributes (no enum, boolean, or number fields)
- [x] T019 Add "No items match filters" empty state in `frontend/src/App.vue` or collection views: display when items array is empty and activeFilters has entries, distinct from the "collection is empty" state
- [x] T020 [P] Update `CLAUDE.md` to document FilterBar component, activeFilters store state, updated GetItems signature with filtersJSON parameter
- [x] T021 [P] Update project `README.md` to document faceted filtering feature, supported filter types, and user interaction
- [x] T022 Run quickstart.md acceptance test flow (all 14 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (store needs updated GetItems signature)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs filter state + FilterBar scaffold)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (adds boolean to existing FilterBar with enum)
- **User Story 3 (Phase 5)**: Depends on Phase 2 (can parallel with US1/US2 but T017 extends backend)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. This is the MVP.
- **US2 (P2)**: Depends on US1 (adds boolean facets to the FilterBar that already renders enum facets)
- **US3 (P3)**: Depends on Foundational. Can run in parallel with US1/US2 for backend work (T017), but FilterBar UI tasks (T015, T016) depend on the FilterBar scaffold from Phase 2.

### Within Each Phase

- T001 before T002 before T003 (struct -> binding -> query logic)
- T004 after T002 (needs new signature)
- T005 after T004 (store needs to call updated GetItems)
- T007 scaffold before T008-T009 (FilterBar must exist before adding facets)

### Parallel Opportunities

- T001 and T004 could overlap if T002 is done first
- T020 and T021 (Phase 6) modify different files -- can run in parallel
- US3 backend work (T017) can parallel with US1/US2 frontend work

---

## Parallel Example: Phase 1 Setup

```bash
# After T002 (binding updated), these can run in parallel:
Task: "Build dynamic WHERE clauses in db.go" (T003)
Task: "Update frontend callers for new GetItems signature" (T004)
```

## Parallel Example: Phase 6 Polish

```bash
Task: "Update CLAUDE.md with FilterBar docs" (T020)
Task: "Update README with faceted filtering docs" (T021)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004) -- backend accepts filters
2. Complete Phase 2: Foundational (T005-T007) -- store + FilterBar scaffold
3. Complete Phase 3: User Story 1 (T008-T011) -- enum filtering works
4. **STOP and VALIDATE**: Test enum filter end-to-end
5. This alone delivers the highest-value filtering capability

### Incremental Delivery

1. Phase 1 + Phase 2 + Phase 3 (US1) = MVP: enum faceted filtering
2. Add Phase 4 (US2) = boolean tri-state toggles
3. Add Phase 5 (US3) = number range inputs + purchasePrice
4. Phase 6 (Polish) = edge cases, documentation, final validation
5. Each phase adds filter types without breaking previous ones

---

## Notes

- Backend uses `json_extract(attributes, '$.field')` for attribute filters; `purchase_price` column directly for purchasePrice
- Filter payload is a JSON array of `{field, op, value/values}` objects
- Empty `filtersJSON` string = no filters (backward compatible)
- Filter state lives in collectionStore, shared between list and grid views
- Number inputs debounced at 400ms to avoid excessive queries during typing
- Filter bar collapsible by default; shows active count when collapsed
