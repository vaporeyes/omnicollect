# Tasks: Spotlight-style Command Palette

**Input**: Design documents from `/specs/006-command-palette/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ui-contract.md, quickstart.md

**Tests**: No automated test framework in this project. Tests not requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Shared infrastructure needed before any user story

- [x] T001 Add `searchAllItems(query)` action to `frontend/src/stores/collectionStore.ts` that calls `GetItems(query, "")` and returns results directly without modifying the store's `items` ref
- [x] T002 Add palette visibility state (`showPalette` ref) and Cmd/Ctrl+K toggle handler to the existing `onGlobalKeydown` in `frontend/src/App.vue`

**Checkpoint**: Store has cross-module search capability; Cmd/Ctrl+K toggles a boolean

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational/blocking tasks needed. The palette has no backend changes and builds on existing infrastructure (GetItems binding, Wails AssetServer thumbnails, Pinia stores). User story work can begin immediately after Phase 1.

---

## Phase 3: User Story 1 - Search and Navigate to Any Item (Priority: P1) MVP

**Goal**: User presses Cmd/Ctrl+K, types a query, sees matching items with thumbnails and module badges, selects one to open its detail view.

**Independent Test**: Open palette, type a known item name, verify results appear with thumbnails and module names, select a result, verify item detail view opens.

### Implementation for User Story 1

- [x] T003 [US1] Create `frontend/src/components/CommandPalette.vue` with the component scaffold: props (`visible: boolean`), emits (`close`, `selectItem`, `action`), template with Teleport to body, blurred glass overlay backdrop, centered dialog container, and large auto-focused search input
- [x] T004 [US1] Implement debounced search in `CommandPalette.vue`: watch the query input, call `searchAllItems` from collectionStore after 200ms debounce, store results locally, cap at 25 items
- [x] T005 [US1] Implement result rendering in `CommandPalette.vue`: display each item result with thumbnail image (from `/thumbnails/` path, 40x40px, placeholder SVG for items without images), item title, and module name badge resolved from moduleStore
- [x] T006 [US1] Implement result selection in `CommandPalette.vue`: clicking a result emits `selectItem` with the item; show "No results" message when query has no matches
- [x] T007 [US1] Implement Escape and click-outside-to-close in `CommandPalette.vue`: Escape keydown and clicking the backdrop overlay emit `close`
- [x] T008 [US1] Wire `CommandPalette` into `frontend/src/App.vue`: import component, render with `:visible="showPalette"`, handle `@close` (set showPalette false), handle `@selectItem` (call existing `onItemSelect` then close palette)
- [x] T009 [US1] Style the palette per design language in `CommandPalette.vue`: z-index 3000, `backdrop-filter: blur()` on overlay, frosted glass dialog using `var(--bg-secondary)` and `var(--glass-blur)`, input styled with `font-family: 'Instrument Serif'` at large size, results using `font-family: 'Outfit'`

**Checkpoint**: User can open palette with Cmd/Ctrl+K, search items across all modules, see rich results, click to navigate to detail view, press Escape to close

---

## Phase 4: User Story 2 - Keyboard Navigation of Results (Priority: P2)

**Goal**: User navigates results entirely with arrow keys and selects with Enter.

**Independent Test**: Open palette, type query, use Down/Up arrows to move highlight, press Enter to confirm selection.

### Implementation for User Story 2

- [x] T010 [US2] Add keyboard navigation state to `CommandPalette.vue`: `highlightedIndex` ref initialized to 0, reset to 0 when results change
- [x] T011 [US2] Implement Up/Down/Enter keydown handlers in `CommandPalette.vue`: Down increments index (clamped to results length - 1), Up decrements (clamped to 0), Enter selects the highlighted result
- [x] T012 [US2] Add highlighted styling and scroll-into-view in `CommandPalette.vue`: highlighted result gets distinct background color, use `scrollIntoView({block: 'nearest'})` when highlight changes

**Checkpoint**: Full keyboard-only operation works: open, search, arrow navigate, Enter to select, Escape to close

---

## Phase 5: User Story 3 - Quick Actions via Keywords (Priority: P3)

**Goal**: Typing keywords like "new", "settings", "backup" surfaces quick action entries above item results.

**Independent Test**: Open palette, type "new", verify quick actions appear above item results, select one, verify corresponding app action triggers.

### Implementation for User Story 3

- [x] T013 [US3] Define quick actions data in `CommandPalette.vue`: array of `{label, keywords, action}` objects for "Add New Item" (new/add/create), "Create New Schema" (new/schema/create), "Open Settings" (settings/preferences), "Export Backup" (backup/export)
- [x] T014 [US3] Implement keyword matching in `CommandPalette.vue`: case-insensitive substring match of query against each action's keywords array, filter matching actions, display above item results in a separate "Actions" group with distinct styling
- [x] T015 [US3] Include quick actions in the keyboard navigation index in `CommandPalette.vue`: combined list (quick actions first, then items), highlight and Enter work across both groups
- [x] T016 [US3] Handle `@action` emit in `frontend/src/App.vue`: map action identifiers to existing handlers -- "newItem" calls `onModuleSelect`, "newSchema" calls `openNewSchemaBuilder`, "openSettings" calls `openSettings`, "exportBackup" calls `onExportBackup`

**Checkpoint**: All three user stories functional; quick actions, item search, and keyboard navigation all work together

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and refinements

- [x] T017 [P] Update `CLAUDE.md` to document CommandPalette component, Cmd/Ctrl+K shortcut, and `searchAllItems` store action
- [x] T018 [P] Update project README to document the command palette feature and keyboard shortcut
- [x] T019 Run quickstart.md acceptance test flow (all 10 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **User Story 1 (Phase 3)**: Depends on Phase 1 (store action + palette toggle state)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (needs results rendering to add keyboard nav)
- **User Story 3 (Phase 5)**: Depends on Phase 4 (needs keyboard nav to integrate quick actions into index)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends only on Setup. This is the MVP.
- **US2 (P2)**: Depends on US1 (keyboard navigation operates on the result list built in US1)
- **US3 (P3)**: Depends on US2 (quick actions must integrate into the combined keyboard navigation index from US2)

### Within Each User Story

- T003 (scaffold) before T004-T009
- T004 (search) before T005 (rendering) before T006 (selection)
- T007 (close behavior) can parallel with T004-T006

### Parallel Opportunities

- T001 and T002 (Phase 1) modify different files -- can run in parallel
- T017 and T018 (Phase 6) modify different files -- can run in parallel
- T007 (Escape/click-outside) can run in parallel with T004-T006 (search/render/select)

---

## Parallel Example: Phase 1 Setup

```bash
# Both modify different files, can run simultaneously:
Task: "Add searchAllItems action to frontend/src/stores/collectionStore.ts"
Task: "Add palette visibility state and Cmd/Ctrl+K handler to frontend/src/App.vue"
```

## Parallel Example: Phase 6 Polish

```bash
# Documentation updates in different files:
Task: "Update CLAUDE.md with CommandPalette docs"
Task: "Update README with command palette feature"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001, T002)
2. Complete Phase 3: User Story 1 (T003-T009)
3. **STOP and VALIDATE**: Test palette open/search/select/close flow
4. This alone delivers significant value -- instant cross-module item navigation

### Incremental Delivery

1. Phase 1 (Setup) + Phase 3 (US1) = MVP: searchable palette with click selection
2. Add Phase 4 (US2) = keyboard-native experience
3. Add Phase 5 (US3) = quick actions for power users
4. Phase 6 (Polish) = documentation and final validation
5. Each phase adds value without breaking previous phases

---

## Notes

- No backend changes needed -- existing `GetItems("query", "")` searches all modules
- Palette state lives in the component, not Pinia (open/closed, query, highlight index)
- Results capped at 25 for performance and scannability
- Thumbnails only in results (Constitution Principle IV)
- Z-index 3000 (above lightbox 1000, context menu 2000; below toast 9999)
