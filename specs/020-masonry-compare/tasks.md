# Tasks: Masonry Grid & Item Comparison

**Input**: Design documents from `/specs/020-masonry-compare/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in the spec. Test tasks omitted.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No new project setup needed. This feature modifies existing files and adds one new component. No new dependencies.

- [x] T001 Verify existing dev environment starts cleanly with `wails dev` or `cd frontend && npm run dev`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational/blocking tasks. All three user stories operate on independent files (CollectionGrid.vue, BulkActionBar.vue, ComparisonView.vue) with only App.vue as a shared integration point handled in each story's final task.

**Checkpoint**: No blocking prerequisites; user story implementation can begin immediately.

---

## Phase 3: User Story 1 - Masonry Grid Layout (Priority: P1) MVP

**Goal**: Replace the uniform-height CSS grid in CollectionGrid.vue with a masonry layout where card heights adapt to each image's natural aspect ratio, while preserving hover states, frosted glass captions, selection badges, and staggered entrance animations.

**Independent Test**: Load a collection with mixed portrait, landscape, and square images. Cards should display at varying heights. Hover, selection, context menus, and captions must work identically to the current grid.

### Implementation for User Story 1

- [x] T002 [US1] Replace CSS grid layout with column-count masonry in `frontend/src/components/CollectionGrid.vue`: change `.grid` from `display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr))` to `column-count` with `column-gap: 16px` and responsive `column-width: 240px`
- [x] T003 [US1] Remove `aspect-ratio: 1` from `.card-image` in `frontend/src/components/CollectionGrid.vue`: replace with `width: 100%; height: auto; display: block` so images render at their natural proportions
- [x] T004 [US1] Add `break-inside: avoid` to `.grid-card` in `frontend/src/components/CollectionGrid.vue` to prevent cards from splitting across columns; add `margin-bottom: 16px` to replace the grid gap for vertical spacing
- [x] T005 [US1] Verify frosted glass caption (`.card-caption` with `backdrop-filter: blur`) still positions correctly at image bottom, hover scale/shadow effects still work, select badge visibility toggles on hover, and staggered animation delays still apply

**Checkpoint**: Masonry grid is fully functional. All existing grid interactions (hover, selection, context menu, keyboard shortcuts) work identically.

---

## Phase 4: User Story 2 - Compare Button Activation (Priority: P2)

**Goal**: Add a "Compare" button to the bulk action bar that appears only when exactly two items are selected.

**Independent Test**: Select 0, 1, 2, and 3+ items. Compare button should appear only at exactly 2. Clicking it should emit a `compare` event.

### Implementation for User Story 2

- [x] T006 [US2] Add `compare` emit to `frontend/src/components/BulkActionBar.vue`: add a Compare button (`.bar-btn` style, matching existing button aesthetics) that renders conditionally when `props.count === 2` and emits `compare` on click
- [x] T007 [US2] Wire Compare emit in `frontend/src/App.vue`: add `showComparison` ref and `comparisonItems` ref; handle `@compare` from BulkActionBar by resolving the two selected item IDs from `selectionStore.selectedIdArray()` against `collectionStore.items`, setting `comparisonItems` and `showComparison = true`

**Checkpoint**: Compare button appears at exactly 2 selections. Clicking it sets the comparison state (view not yet rendered).

---

## Phase 5: User Story 3 - Side-by-Side Comparison View (Priority: P2)

**Goal**: Create ComparisonView.vue with synchronized image galleries and an attribute diff table (core fields + schema attributes) with difference highlighting. Integrate into App.vue view routing.

**Independent Test**: Enter comparison mode with two items from the same module that have differing attribute values. Galleries sync, diff rows highlight differences. Close returns to grid.

### Implementation for User Story 3

- [x] T008 [P] [US3] Create `frontend/src/components/ComparisonView.vue` scaffold: two-column CSS grid layout (`1fr 1fr`), props for `itemA`, `itemB`, `schemaA`, `schemaB`, emit `close`; responsive stacking at `max-width: 768px`; add ABOUTME comment header
- [x] T009 [US3] Implement synchronized image galleries in `frontend/src/components/ComparisonView.vue`: shared `activeImageIndex` ref, prev/next controls, per-side clamping to `min(activeImageIndex, images.length - 1)`, thumbnail paths via `/originals/` for main display, placeholder when no images, counter pill showing position
- [x] T010 [US3] Implement diff table in `frontend/src/components/ComparisonView.vue`: compute `diffRows` from core fields (title, purchasePrice, tags) and union of both schemas' attributes; each row has label, valueA, valueB, isDifferent boolean; tags compared as sorted comma-joined strings; null-aware price comparison
- [x] T011 [US3] Add diff highlighting CSS in `frontend/src/components/ComparisonView.vue`: `.diff-row` class with subtle background color (light amber in light mode, desaturated amber in dark mode via CSS variables); apply to rows where `isDifferent === true`
- [x] T012 [US3] Integrate ComparisonView into `frontend/src/App.vue` view routing: render `ComparisonView` when `showComparison === true` (highest priority in the view conditional chain); pass resolved items and schemas as props; handle `@close` to set `showComparison = false`; pass module schemas by looking up `moduleStore` for each item's `moduleId`

**Checkpoint**: Full comparison mode works end-to-end. Synced galleries, diff table with highlighting, close returns to grid with selection preserved.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T013 Verify all existing grid interactions still work after masonry change: selection checkboxes, shift-click range select, context menus, Cmd/Ctrl+N shortcut, add-item empty state CTA
- [x] T014 Verify comparison mode with edge cases: items with no images (both sides), items from different modules (union of attributes), items with identical values (no highlighting), narrow viewport stacking
- [x] T015 Update `CLAUDE.md` with new component (`ComparisonView.vue`), modified components, and comparison mode documentation
- [x] T016 Update project README to document masonry grid and comparison mode features

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: N/A (no blocking prerequisites)
- **US1 Masonry (Phase 3)**: Can start immediately
- **US2 Compare Button (Phase 4)**: Can start immediately (independent of US1)
- **US3 Comparison View (Phase 5)**: Depends on US2 (T007 creates the `showComparison` state that T012 uses)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent. Only touches CollectionGrid.vue CSS.
- **User Story 2 (P2)**: Independent of US1. Touches BulkActionBar.vue + App.vue.
- **User Story 3 (P2)**: Depends on US2 (needs `showComparison` and `comparisonItems` refs in App.vue from T007). Touches ComparisonView.vue (new) + App.vue.

### Parallel Opportunities

- US1 (T002-T005) and US2 (T006-T007) can proceed in parallel since they modify different files
- Within US3, T008 (scaffold) can start in parallel with US1/US2 since it creates a new file
- T013 and T014 (polish) can run in parallel

---

## Parallel Example: User Stories 1 & 2

```bash
# These can run simultaneously (different files):
Task: "T002 [US1] Replace CSS grid with column-count masonry in CollectionGrid.vue"
Task: "T006 [US2] Add compare emit to BulkActionBar.vue"

# After both complete, US3 can begin:
Task: "T008 [US3] Create ComparisonView.vue scaffold"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete T001 (verify environment)
2. Complete T002-T005 (masonry grid)
3. **STOP and VALIDATE**: Load collection with mixed images, verify masonry layout
4. This alone delivers significant visual improvement

### Incremental Delivery

1. US1: Masonry grid -> Test independently -> Commit (visual upgrade ships)
2. US2: Compare button -> Test independently -> Commit (entry point ready)
3. US3: Comparison view -> Test independently -> Commit (full feature complete)
4. Polish: Edge cases, docs -> Commit (release ready)

---

## Notes

- No new npm dependencies required
- No backend/Go changes
- All changes are in `frontend/src/` (components + App.vue)
- Constitution Principle IV satisfied: grid uses thumbnails, comparison galleries use originals (matching existing ItemDetail pattern)
- Commit after each user story checkpoint for clean git history
