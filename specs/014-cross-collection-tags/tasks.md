# Tasks: Cross-Collection Tags

**Input**: Design documents from `/specs/014-cross-collection-tags/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tags-contract.md, quickstart.md

**Tests**: Not explicitly requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Data model changes shared by all user stories

- [x] T001 Add `Tags []string` field to `Item` struct in `storage/db.go`; add `TagCount` struct (`Name string`, `Count int`); add `GetAllTags`, `RenameTag`, `DeleteTag` methods to the `Store` interface
- [x] T002 Add `tags: string[]` field to the `Item` interface in `frontend/src/api/types.ts`; add `TagCount` interface (`name: string`, `count: number`)

**Checkpoint**: Types defined in both Go and TypeScript; compilation passes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema changes and tag query support in both storage backends

- [x] T003 Update SQLite DDL in `storage/sqlite.go`: add `tags TEXT NOT NULL DEFAULT '[]'` column to CREATE TABLE items; update FTS5 insert/update/delete triggers to include tag values in the indexed text (concatenate tag array elements into `attrs_text`)
- [x] T004 Update PostgreSQL DDL in `storage/postgres.go`: add `tags JSONB NOT NULL DEFAULT '[]'` column to CREATE TABLE items; add `CREATE INDEX idx_items_tags ON items USING GIN(tags)`; update the tsvector search trigger to include tag values
- [x] T005 Update `QueryItems` in `storage/sqlite.go`: accept new `tags` parameter (JSON array string); when present, add `WHERE EXISTS (SELECT 1 FROM json_each(items.tags) WHERE value IN (?,...))` clause; combine with existing filters via AND
- [x] T006 Update `QueryItems` in `storage/postgres.go`: accept `tags` parameter; when present, add `WHERE tags ?| array[?,...] ` clause; combine with existing filters via AND
- [x] T007 Update `InsertItem` and `UpdateItem` in both `storage/sqlite.go` and `storage/postgres.go`: serialize `item.Tags` to JSON for storage; deserialize on read in `scanItems`
- [x] T008 Update `ExportItemsCSV` in both storage backends: add a "tags" column to the CSV output containing comma-separated tag values

**Checkpoint**: Tags persist on items; tag filter works in queries; `go test ./storage/...` passes

---

## Phase 3: User Story 1 - Add and Remove Tags on Items (Priority: P1) MVP

**Goal**: Users can add/remove free-form tags on items via the edit form and detail view.

**Independent Test**: Edit an item, add tags, save, reopen, verify tags persist. Remove a tag, save, verify removed.

### Implementation for User Story 1

- [x] T009 [US1] Create `frontend/src/components/TagInput.vue`: props `modelValue: string[]` and `allTags: TagCount[]`; text input that adds a tag on Enter (lowercase, trimmed, max 50 chars, no duplicates); renders existing tags as removable chips; autocomplete dropdown filtering `allTags` as user types; emits `update:modelValue`
- [x] T010 [US1] Modify `frontend/src/components/DynamicForm.vue`: add `TagInput` below the existing form fields; bind to the item's `tags` array; pass `allTags` (fetched from API)
- [x] T011 [US1] Modify `frontend/src/components/ItemDetail.vue`: display tags as styled chips in the metadata section (below attributes, above provenance); read-only (editing happens in the form)
- [x] T012 [US1] Add `getAllTags` function to `frontend/src/api/client.ts`: calls `GET /api/v1/tags`, returns `TagCount[]`
- [x] T013 [US1] Add tag REST endpoints in `handlers.go`: `handleGetAllTags` (GET /api/v1/tags), wire to Store.GetAllTags
- [x] T014 [US1] Register tag routes in `server.go`: `GET /api/v1/tags`, `POST /api/v1/tags/rename`, `DELETE /api/v1/tags/{name}`
- [x] T015 [US1] Implement `GetAllTags` in `storage/sqlite.go`: `SELECT value, COUNT(*) FROM items, json_each(items.tags) GROUP BY value ORDER BY value`
- [x] T016 [US1] Implement `GetAllTags` in `storage/postgres.go`: `SELECT tag, COUNT(*) FROM items, jsonb_array_elements_text(tags) AS tag GROUP BY tag ORDER BY tag`

**Checkpoint**: Tags can be added/removed on items; tags persist; autocomplete works; detail view shows tags

---

## Phase 4: User Story 2 - Filter Items by Tag Across All Modules (Priority: P2)

**Goal**: Tag filter in collection views shows items matching selected tags across all modules.

**Independent Test**: Tag items in different modules with "gift", activate tag filter, verify cross-module results.

### Implementation for User Story 2

- [x] T017 [US2] Create `frontend/src/components/TagFilter.vue`: props `allTags: TagCount[]` and `selectedTags: string[]`; renders clickable tag chips (active state when selected); clicking toggles selection; emits `update` with new tag array
- [x] T018 [US2] Modify `frontend/src/stores/collectionStore.ts`: add `activeTags` ref (string[]); add `setTags(tags)` action that updates activeTags and re-fetches; include `tags` param in `buildQueryString` when activeTags is non-empty (JSON-encoded array)
- [x] T019 [US2] Modify `frontend/src/App.vue`: render `TagFilter` above collection views (below FilterBar); pass `allTags` (fetched on mount and after tag changes); wire `@update` to `collectionStore.setTags`; clear activeTags on module switch
- [x] T020 [US2] Update `handlers.go` `handleGetItems`: parse `tags` query parameter (JSON array); pass to `Store.QueryItems`
- [x] T021 [US2] Update the `Store.QueryItems` signature in `storage/db.go` to accept a `tagsJSON` parameter (or extend the existing `filtersJSON` to include tags)

**Checkpoint**: Tag filtering works across all modules; combines with text search and attribute filters

---

## Phase 5: User Story 3 - Tag Autocomplete and Management (Priority: P3)

**Goal**: Autocomplete from existing tags; rename and delete tags in bulk.

**Independent Test**: Type partial tag, select from autocomplete. Rename a tag, verify all items updated. Delete a tag, verify removed.

### Implementation for User Story 3

- [x] T022 [US3] Implement `RenameTag` in `storage/sqlite.go`: for each item containing the old tag, rebuild the tags JSON array with the new name; use a transaction
- [x] T023 [US3] Implement `RenameTag` in `storage/postgres.go`: `UPDATE items SET tags = (tags - $1) || to_jsonb($2::text) WHERE tags ? $1`
- [x] T024 [US3] Implement `DeleteTag` in `storage/sqlite.go`: for each item containing the tag, rebuild the tags JSON array without it; use a transaction
- [x] T025 [US3] Implement `DeleteTag` in `storage/postgres.go`: `UPDATE items SET tags = tags - $1 WHERE tags ? $1`
- [x] T026 [US3] Add rename and delete handlers in `handlers.go`: `handleRenameTag` (POST /api/v1/tags/rename) and `handleDeleteTag` (DELETE /api/v1/tags/{name})
- [x] T027 [US3] Add `renameTag` and `deleteTag` functions to `frontend/src/api/client.ts`
- [x] T028 [US3] Create `frontend/src/components/TagManager.vue`: list all tags with item counts; inline rename (click tag name to edit); delete button per tag with confirmation; emits `rename` and `delete` events
- [x] T029 [US3] Wire TagManager into `frontend/src/App.vue`: accessible from settings or a sidebar button; on rename/delete, call API, refresh tag list and collection

**Checkpoint**: Full tag management works -- autocomplete, rename, delete

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Command palette integration, documentation

- [x] T030 Update `frontend/src/components/CommandPalette.vue`: include tag values in search (when searching items, match against item tags in addition to title)
- [x] T031 [P] Update `CLAUDE.md`: document TagInput, TagFilter, TagManager components; tag storage (JSON array); new REST endpoints; Store interface additions
- [x] T032 [P] Update `README.md`: add Tags section describing the feature; update iteration history
- [x] T033 Run quickstart.md acceptance test flow (all 12 steps) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs tags on Item struct)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs DDL + query changes)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (needs tag filter in queries); can parallel with US1 frontend work
- **User Story 3 (Phase 5)**: Depends on Phase 2 (needs tag management store methods)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- add/remove tags + autocomplete.
- **US2 (P2)**: Depends on Foundational. Can develop in parallel with US1 (different components).
- **US3 (P3)**: Depends on Foundational. Can develop in parallel with US1/US2 for backend; TagManager UI depends on TagInput existing.

### Parallel Opportunities

- T001 (Go types) and T002 (TS types) -- different languages
- T003 (SQLite DDL) and T004 (PostgreSQL DDL) -- different files
- T005 (SQLite query) and T006 (PG query) -- different files
- T015 (SQLite GetAllTags) and T016 (PG GetAllTags) -- different files
- T022/T024 (SQLite rename/delete) and T023/T025 (PG rename/delete) -- different files
- US1 frontend (T009-T012) and US2 frontend (T017-T019) -- different components
- T031 and T032 (docs) -- different files

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 + Phase 2 = schema changes + query support
2. Phase 3 (US1) = add/remove tags on items + autocomplete
3. **STOP and VALIDATE**: Tag items, verify persistence, verify autocomplete
4. This delivers the core tagging capability

### Incremental Delivery

1. Phase 1 + Phase 2 = data model + queries ready
2. Phase 3 (US1) = tag input/display (MVP)
3. Phase 4 (US2) = tag filtering across modules
4. Phase 5 (US3) = tag management (rename/delete)
5. Phase 6 = command palette + docs

---

## Notes

- Tags stored as JSON array on item -- no junction table (Constitution III)
- SQLite: `json_each()` + EXISTS for tag queries; PostgreSQL: `?|` + GIN index
- Tags normalized to lowercase on save; 50-char limit
- Tag filter is separate from schema-driven faceted filter bar
- Tags included in FTS5/tsvector for text search
- Tags included in CSV export as comma-separated column
- No new dependencies needed
