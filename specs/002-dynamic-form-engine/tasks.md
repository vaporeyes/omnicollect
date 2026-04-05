# Tasks: Dynamic Form Engine

**Input**: Design documents from `/specs/002-dynamic-form-engine/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/component-contracts.md

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Wails desktop app**: Vue 3 frontend under `frontend/src/`
- Stores in `frontend/src/stores/`, components in `frontend/src/components/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install Pinia and configure the Vue app to use it.

- [x] T001 Install Pinia dependency via `npm install pinia` in `frontend/`
- [x] T002 Register Pinia in the Vue app entry point in `frontend/src/main.ts`: import `createPinia`, call `app.use(createPinia())` before mount

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create Pinia stores that all user stories depend on.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Create `frontend/src/stores/moduleStore.ts`: define `useModuleStore` with state (modules array, loading boolean, error string), action `fetchModules()` that calls `GetActiveModules()` from Wails bindings, getter `getModuleById(id)` that returns a single schema
- [x] T004 [P] Create `frontend/src/stores/collectionStore.ts`: define `useCollectionStore` with state (items array, loading boolean, error string, activeModuleId string, searchQuery string), actions `fetchItems()` that calls `GetItems(searchQuery, activeModuleId)`, `saveItem(item)` that calls `SaveItem(item)` then refreshes, `setFilter(moduleId)` and `setSearch(query)` that update state and re-fetch
- [x] T005 Create `frontend/src/components/FormField.vue`: single field renderer that dispatches on `attribute.type` prop -- "string" renders text input, "number" renders number input, "boolean" renders checkbox, "date" renders date input, "enum" renders select with options from `attribute.options`. Apply `attribute.display.label` as label (fallback to `attribute.name`), `attribute.display.placeholder` as placeholder, `attribute.display.widget` to override control type. Support v-model via `modelValue` prop and `update:modelValue` emit. Fall back to text input for unrecognized types with console warning.

**Checkpoint**: Stores fetch data from backend. FormField renders all five attribute types correctly.

---

## Phase 3: User Story 1 - Add a New Collection Item (Priority: P1)

**Goal**: Collector selects a module type, fills a dynamically generated form, and saves a new item.

**Independent Test**: Select "Coins" module, fill all fields, save. Verify item appears in list with correct data.

### Implementation for User Story 1

- [x] T006 [US1] Create `frontend/src/components/ModuleSelector.vue`: renders list of available modules from `useModuleStore().modules`, shows displayName for each, emits `select(module)` on click. Shows "No collection types available" message when modules array is empty.
- [x] T007 [US1] Create `frontend/src/components/DynamicForm.vue`: accepts `schema: ModuleSchema` and `item: Item | null` props. Renders base fields (title as text input required, purchasePrice as number input optional) at the top. Loops over `schema.attributes` and renders a `FormField` for each. Maintains reactive `baseFields` and `attributes` objects. On submit: validates required fields (title always, plus any attributes marked `required: true`), shows inline errors for invalid fields, constructs Item payload with attributes nested in `attributes` object, emits `save(item)`. Emits `cancel()` on cancel button. When `item` prop is provided, pre-populates all fields from it (edit mode).
- [x] T008 [US1] Update `frontend/src/App.vue`: replace template HelloWorld content with app layout. Import and use `useModuleStore` and `useCollectionStore`. Call `fetchModules()` and `fetchItems()` on mount. Render `ModuleSelector` in sidebar area. When a module is selected, show `DynamicForm` with that module's schema. On form `save` emit, call `collectionStore.saveItem(item)`. Show error messages from store errors. Show loading state while stores are fetching.

**Checkpoint**: Collector can select a module, fill in the dynamic form, save, and the item persists in the backend. Form validation blocks submission when required fields are empty.

---

## Phase 4: User Story 2 - Browse and Search Collection Items (Priority: P2)

**Goal**: Collector views a list of items, filters by module, and searches by keyword.

**Independent Test**: Save items across two module types. Browse all, filter by one, search by keyword. Verify correct results.

### Implementation for User Story 2

- [x] T009 [US2] Create `frontend/src/components/ItemList.vue`: accepts `items: Item[]` and `modules: ModuleSchema[]` props. Renders each item as a row with title, module displayName (looked up from modules), and formatted updatedAt timestamp. Provides module filter dropdown (all types + each module) that emits `filterChange(moduleId)`. Provides search text input that emits `search(query)` with debounce. Shows empty state message when no items match. Emits `select(item)` when user clicks an item row.
- [x] T010 [US2] Integrate `ItemList.vue` into `frontend/src/App.vue`: render ItemList in main content area alongside the form. Wire `filterChange` to `collectionStore.setFilter()`, `search` to `collectionStore.setSearch()`. Display items from `collectionStore.items`. Show ItemList when no form is active, or alongside the form in a split layout.

**Checkpoint**: Item list displays saved items. Filtering by module shows only matching items. Search returns results from titles and attributes. Empty state shown when no matches.

---

## Phase 5: User Story 3 - Edit an Existing Item (Priority: P3)

**Goal**: Collector clicks an item in the list to edit it in the dynamic form.

**Independent Test**: Save an item, click it in the list, change a field, save. Verify the change persists.

### Implementation for User Story 3

- [x] T011 [US3] Wire item editing in `frontend/src/App.vue`: when ItemList emits `select(item)`, look up the item's module schema from moduleStore, set DynamicForm's `item` prop to the selected item and `schema` to the matching schema. On save, call `collectionStore.saveItem(item)` and clear the editing state. On cancel, clear the editing state without saving.
- [x] T012 [US3] Handle orphaned items (schema removed) in `frontend/src/App.vue`: when an item's moduleId has no matching schema in moduleStore, show the item in the list but display a warning when the user attempts to edit ("Collection type schema not available"). Do not crash.

**Checkpoint**: Clicking an item opens the form pre-populated with its values. Saving updates the item. Canceling discards changes. Orphaned items show a warning instead of crashing.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, cleanup, and validation.

- [x] T013 [P] Add minimal CSS styling in `frontend/src/style.css`: basic layout (sidebar + main content), form field spacing, list item hover state, error message styling, empty state styling. Functional appearance only -- visual polish deferred.
- [x] T014 [P] Remove the old Wails template components: delete `frontend/src/assets/images/logo-universal.png` and any unused template files. Clean up unused CSS.
- [x] T015 Run `quickstart.md` validation: follow all verification steps end-to-end (module loading, form rendering, save, search, edit, multiple collection types).
- [x] T016 Run `wails build` and verify the production binary includes the complete dynamic form frontend.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion -- BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (T003-T005)
- **User Story 2 (Phase 4)**: Depends on US1 (uses App.vue layout from T008)
- **User Story 3 (Phase 5)**: Depends on US2 (uses ItemList select from T009)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2. No dependency on other stories.
- **User Story 2 (P2)**: Requires US1 (App.vue layout and store wiring).
- **User Story 3 (P3)**: Requires US2 (item selection from list).

### Within Each User Story

- Store/data layer before components
- Components before App.vue integration
- Story complete before moving to next priority

### Parallel Opportunities

- T003 and T004 can run in parallel (different files)
- T013 and T014 can run in parallel (different files)
- T005 (FormField) can run in parallel with T003/T004 (different file)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T005)
3. Complete Phase 3: User Story 1 (T006-T008)
4. **STOP and VALIDATE**: Add item via dynamic form, verify it saves
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Stores and FormField ready
2. Add User Story 1 -> Dynamic form works -> Demo (MVP!)
3. Add User Story 2 -> List + search works -> Demo
4. Add User Story 3 -> Edit works -> Demo
5. Polish -> Production-ready

### Recommended Execution Order

```
Phase 1 (Setup: Pinia)
  |
Phase 2 (Foundational: Stores + FormField)
  |
Phase 3 (US1: DynamicForm + ModuleSelector + App.vue)
  |
Phase 4 (US2: ItemList + App.vue integration)
  |
Phase 5 (US3: Edit wiring + orphan handling)
  |
Phase 6 (Polish: CSS + cleanup + validation)
```

US2 and US3 are sequential in this iteration because each builds
on the previous App.vue layout changes.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable at its checkpoint
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
