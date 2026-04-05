# Tasks: Schema Visual Builder

**Input**: Design documents from `/specs/004-schema-visual-builder/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go backend**: project root (`app.go`, `modules.go`)
- **Vue frontend**: `frontend/src/components/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install CodeMirror dependencies and add Go backend methods.

- [x] T001 Install frontend dependencies: `vue-codemirror` and `@codemirror/lang-json` via `npm install` in `frontend/`
- [x] T002 [P] Add `SaveCustomModule` method on App struct in `app.go`: accepts JSON string, parses into ModuleSchema, validates (non-empty ID/displayName, valid attribute types, unique attribute names, enum attributes have options), writes formatted JSON to `~/.omnicollect/modules/{id}.json`, reloads in-memory module schemas by calling `loadModuleSchemas()`, returns parsed ModuleSchema. Check for ID conflicts (same ID but different filename).
- [x] T003 [P] Add `LoadModuleFile` method on App struct in `app.go`: accepts moduleID string, finds the corresponding `.json` file in `~/.omnicollect/modules/` by scanning for matching `id` field, returns raw JSON string content. Return error if module not found.
- [x] T004 [P] Add `reloadModules` helper in `modules.go`: calls `loadModuleSchemas()` and updates the App's modules slice. Add `findModuleFile(moduleID string)` that returns the file path for a given module ID by scanning the modules directory.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core builder components that all user stories depend on.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T005 Create `frontend/src/components/SchemaCodeEditor.vue`: wrapper around `vue-codemirror` with `@codemirror/lang-json` extension. Accepts `modelValue: string` prop (JSON text) and `error: string | null` prop. Emits `update:modelValue` on text change. Shows error indicator bar when error prop is set. Basic styling with border and min-height.
- [x] T006 [P] Create `frontend/src/components/SchemaFormPreview.vue`: accepts `schema` prop (draft schema object). Loops over `schema.attributes` and renders `FormField` components (from Iteration 2) in read-only/disabled mode. Shows "Add fields to see preview" when attributes array is empty. Shows base fields (title, purchase price) at the top as disabled inputs.

**Checkpoint**: CodeMirror editor renders with JSON highlighting. Form preview renders fields from a schema object.

---

## Phase 3: User Story 1 - Create a New Schema Visually (Priority: P1)

**Goal**: User opens the builder, adds fields visually, sees live preview, saves to disk.

**Independent Test**: Open builder, enter name, add 3 fields (string, number, enum), save. Verify schema file written and collection type appears in module selector.

### Implementation for User Story 1

- [x] T007 [US1] Create `frontend/src/components/SchemaVisualEditor.vue`: accepts `schema` prop (reactive draft object), emits `update:schema`. Renders editable inputs for `displayName` and `description` at the top. Auto-generates `id` as slug from displayName (lowercase, hyphens, no special chars). "Add Field" button appends a new attribute with defaults (name: "", type: "string", required: false). Each field row shows: name text input, type dropdown (string/number/boolean/date/enum), required checkbox toggle, remove button, move up/move down buttons. For enum fields, shows options sub-editor: list of option strings with add/remove. Display hints section (collapsible): label, placeholder, widget inputs.
- [x] T008 [US1] Create `frontend/src/components/SchemaBuilder.vue`: top-level split-pane layout. Manages central `draftSchema` reactive ref (deep watched). Left pane: SchemaVisualEditor (top) + SchemaFormPreview (bottom). Right pane: SchemaCodeEditor. Bidirectional sync: visual changes serialize to JSON and update code editor; code editor changes attempt `JSON.parse` -- on success update draft object, on failure set error string and keep last valid state. Save button: validates schema (non-empty displayName, unique attribute names, enum options present), calls `SaveCustomModule(JSON.stringify(schema))` binding, emits `saved`. Cancel button: if unsaved changes, shows `confirm()` dialog, emits `close`. Toolbar with save/cancel buttons.
- [x] T009 [US1] Integrate SchemaBuilder into `frontend/src/App.vue`: add "Schema Builder" button in the sidebar below the module selector. When clicked, show SchemaBuilder (replacing main content). On `saved`, refresh `moduleStore.fetchModules()` and close the builder. On `close`, return to collection view.

**Checkpoint**: User can create a schema visually, see live preview, save to disk. New collection type appears immediately in module selector.

---

## Phase 4: User Story 2 - JSON Code Editor Editing (Priority: P2)

**Goal**: Power user edits schema JSON directly with graceful error handling.

**Independent Test**: Open builder, edit JSON in code editor, verify visual preview updates. Introduce syntax error, verify preview holds last valid state and recovers.

### Implementation for User Story 2

- [x] T010 [US2] Enhance bidirectional sync in `frontend/src/components/SchemaBuilder.vue`: ensure code editor text is debounced (300ms) before parsing to avoid excessive updates during rapid typing. Add `parseError` ref that holds the error message when JSON is invalid. Pass `parseError` to SchemaCodeEditor's `error` prop. When parse succeeds after a failure, clear the error immediately. Verify the visual preview never shows a blank/crashed state during parse errors.
- [x] T011 [US2] Add error indicator to `frontend/src/components/SchemaCodeEditor.vue`: when `error` prop is set, show a red bar below the editor with the error message text. Style it to be non-intrusive (small font, below the editor, not overlaying content).

**Checkpoint**: JSON edits update the visual preview. Syntax errors show error indicator. Preview holds last valid state. Recovery is immediate when JSON becomes valid.

---

## Phase 5: User Story 3 - Edit Existing Schema (Priority: P3)

**Goal**: User loads an existing schema into the builder, modifies it, saves. Existing items unaffected.

**Independent Test**: Load "Coins" schema in builder, add a field, save. Verify file updated. Verify existing coin items still display.

### Implementation for User Story 3

- [x] T012 [US3] Add "Edit" button to each module in `frontend/src/components/ModuleSelector.vue`: small edit icon/button next to each module displayName. Emits `edit(module)` with the ModuleSchema.
- [x] T013 [US3] Wire edit flow in `frontend/src/App.vue`: when ModuleSelector emits `edit(module)`, call `LoadModuleFile(module.id)` to get the raw JSON, open SchemaBuilder with the parsed schema as initial state. On save, refresh moduleStore. SchemaBuilder's `moduleId` prop (non-null) indicates edit mode.
- [x] T014 [US3] Update `frontend/src/components/SchemaBuilder.vue` to accept optional `moduleId: string` and `initialJSON: string` props. When provided, initialize the draft schema from `initialJSON` instead of an empty template. In edit mode, the ID field is read-only (prevent accidental ID changes that would orphan items).

**Checkpoint**: Existing schemas load in the builder. Edits save correctly. Existing items under the schema are unaffected.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Styling, layout polish, and end-to-end validation.

- [x] T015 [P] Add split-pane CSS to `frontend/src/components/SchemaBuilder.vue`: flexbox layout with left pane (60%) and right pane (40%), resizable divider (optional stretch goal), proper scrolling in each pane, toolbar styling
- [x] T016 [P] Add visual editor CSS to `frontend/src/components/SchemaVisualEditor.vue`: field row layout, add/remove/move button styling, enum options editor styling, collapsible display hints section
- [x] T017 Run `quickstart.md` validation: follow all verification steps (create new schema, code editor sync, edit existing, validation errors, unsaved changes prompt)
- [x] T018 Run `wails build` and verify production binary includes the schema builder with CodeMirror

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Setup -- BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (T005-T006)
- **User Story 2 (Phase 4)**: Depends on US1 (enhances SchemaBuilder from T008)
- **User Story 3 (Phase 5)**: Depends on US1 (uses SchemaBuilder + App.vue integration)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2. Independent of other stories.
- **User Story 2 (P2)**: Requires US1 (enhances bidirectional sync in SchemaBuilder).
- **User Story 3 (P3)**: Requires US1 (adds edit mode to existing builder).

### Within Each User Story

- Backend bindings before frontend components
- Inner components before container components
- Container before App.vue integration

### Parallel Opportunities

- T002, T003, T004 can all run in parallel (different methods/files)
- T005 and T006 can run in parallel (different files)
- T015 and T016 can run in parallel (different files)
- US2 and US3 are independent of each other (both depend on US1 only)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T006)
3. Complete Phase 3: User Story 1 (T007-T009)
4. **STOP and VALIDATE**: Create schema visually, verify save + hot reload
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Editor components ready
2. Add User Story 1 -> Visual builder works -> Demo (MVP!)
3. Add User Story 2 -> Graceful JSON editing -> Demo
4. Add User Story 3 -> Edit existing schemas -> Demo
5. Polish -> Production-ready

### Recommended Execution Order

```
Phase 1 (Setup: deps + Go bindings)
  |
Phase 2 (Foundational: CodeEditor + FormPreview)
  |
Phase 3 (US1: VisualEditor + Builder + App.vue)
  |
  +-- Phase 4 (US2: error handling) [can parallel with US3]
  |
  +-- Phase 5 (US3: edit existing) [can parallel with US2]
  |
Phase 6 (Polish: CSS + validation)
```

US2 and US3 can run in parallel since they modify different aspects
of the SchemaBuilder (sync logic vs edit mode props).

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable at its checkpoint
- Commit after each task or logical group
- The builder reuses FormField from Iteration 2 for live preview
