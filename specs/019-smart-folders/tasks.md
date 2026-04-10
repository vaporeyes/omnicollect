# Tasks: Smart Folders (Saved Views)

**Input**: Design documents from `/specs/019-smart-folders/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

---

## Phase 1: Setup

**Purpose**: No new dependencies needed; create store structure

- [x] T001 Create smartFolderStore.ts in frontend/src/stores/smartFolderStore.ts -- Pinia store with SmartFolder type definition, folders ref, activeSmartFolderId ref, CRUD methods (create, rename, delete), loadFromSettings/saveToSettings persistence via api.get/api.put on /api/v1/settings, generateId helper (8-char hex)

---

## Phase 2: Foundational (Store Tests)

**Purpose**: Verify store logic before building UI

- [x] T002 Write Vitest tests for smartFolderStore in frontend/src/stores/__tests__/smartFolderStore.test.ts -- cover: create with name + state, rename, delete, load from settings object, save serialization, empty name rejection, generateId uniqueness, delete active folder clears activeSmartFolderId

**Checkpoint**: Store tests passing before UI work

---

## Phase 3: User Story 1 - Save Current View (Priority: P1)

**Goal**: Users can save the current module/search/filter/tag state as a named Smart Folder

- [x] T003 [US1] Create SmartFolders.vue in frontend/src/components/SmartFolders.vue -- sidebar section with: "Saved Views" header, folder list with bookmark icons, "Save Current View" button, inline text field (appears on click, Enter to save, Escape to cancel), empty state message when no folders exist
- [x] T004 [US1] Wire SmartFolders.vue into App.vue sidebar -- add component between sidebar-scroll and sidebar-bottom sections; pass current view state (activeModuleId, searchQuery, activeFilters, activeTags) as props for save; initialize smartFolderStore from loaded settings on mount

**Checkpoint**: Can save Smart Folders, they appear in sidebar and persist across refresh

---

## Phase 4: User Story 2 - Apply Smart Folder (Priority: P2)

**Goal**: Clicking a Smart Folder restores the saved view state

- [x] T005 [US2] Implement apply handler in App.vue -- on SmartFolders emit "apply" with folder data, set collectionStore activeModuleId/searchQuery/activeFilters/activeTags from saved state, set smartFolderStore.activeSmartFolderId; handle missing module gracefully (fall back to "All Types" + toast message)
- [x] T006 [US2] Clear active Smart Folder on manual state change -- in App.vue onFilterChange, onSearch, and tag change handlers, clear smartFolderStore.activeSmartFolderId; pass activeSmartFolderId to SmartFolders.vue for visual highlight styling

**Checkpoint**: Clicking folder restores view; manual changes deselect highlight

---

## Phase 5: User Story 3 - Rename and Delete (Priority: P3)

**Goal**: Right-click context menu for managing Smart Folders

- [x] T007 [US3] Add context menu support to SmartFolders.vue -- on right-click a folder, show existing ContextMenu component with "Rename" and "Delete" options; Rename activates inline edit field on the target folder; Delete shows confirmation then removes folder via store
- [x] T008 [US3] Implement inline rename in SmartFolders.vue -- when Rename selected, replace folder name text with editable input field pre-filled with current name; Enter confirms (calls store.rename), Escape cancels; empty name rejected with validation

**Checkpoint**: Full CRUD on Smart Folders working

---

## Phase 6: Polish & Documentation

- [x] T009 [P] Update CLAUDE.md -- add SmartFolders.vue and smartFolderStore.ts to component/store lists; document Smart Folder conventions and persistence mechanism
- [x] T010 [P] Run go vet ./... and cd frontend && npm test to verify no regressions

---

## Dependencies & Execution Order

- **Phase 1**: No dependencies
- **Phase 2**: Depends on Phase 1 (store must exist to test)
- **Phase 3**: Depends on Phase 2 (tests pass first); T003 then T004 (component before integration)
- **Phase 4**: Depends on Phase 3 (SmartFolders.vue must exist); T005 then T006
- **Phase 5**: Depends on Phase 4; T007 then T008
- **Phase 6**: Depends on all story phases; T009 and T010 can run in parallel

## Notes

- No backend changes required
- Persistence uses existing GET/PUT /api/v1/settings (smartFolders key in JSON blob)
- Reuse existing ContextMenu component for right-click actions
- View mode (dashboard/list/grid) is NOT included in saved state
