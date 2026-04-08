# Quickstart: Cross-Collection Tags

**Branch**: `014-cross-collection-tags` | **Date**: 2026-04-08

## Prerequisites

- Existing codebase on the `014-cross-collection-tags` branch
- Items in the database for testing

## No New Dependencies

This feature uses only existing packages.

## Files to Create

1. **`frontend/src/components/TagInput.vue`** -- Tag input with autocomplete + removable chips
2. **`frontend/src/components/TagFilter.vue`** -- Tag filter control for collection views
3. **`frontend/src/components/TagManager.vue`** -- Tag management (list, rename, delete)

## Files to Modify

### Backend
1. **`storage/db.go`** -- Add `tags` to Item struct; add TagCount type; add GetAllTags, RenameTag, DeleteTag to Store interface
2. **`storage/sqlite.go`** -- Add tags column DDL; update FTS triggers; implement tag query filter; implement GetAllTags, RenameTag, DeleteTag
3. **`storage/postgres.go`** -- Same for PostgreSQL (JSONB array, GIN index, tsvector)
4. **`handlers.go`** -- Add tag endpoints (GET /tags, POST /tags/rename, DELETE /tags/{name}); update GetItems handler to accept tags param
5. **`server.go`** -- Register new tag routes

### Frontend
6. **`frontend/src/api/types.ts`** -- Add tags to Item; add TagCount type
7. **`frontend/src/api/client.ts`** -- Add getAllTags, renameTag, deleteTag functions
8. **`frontend/src/components/DynamicForm.vue`** -- Include TagInput in item forms
9. **`frontend/src/components/ItemDetail.vue`** -- Display tags as chips
10. **`frontend/src/stores/collectionStore.ts`** -- Add activeTags filter; pass tags param to queries
11. **`frontend/src/App.vue`** -- Add TagFilter above collection views; add TagManager access

## Implementation Order

1. Backend: add tags to Item struct + DDL migration
2. Backend: tag filter in GetItems query (both SQLite + PG)
3. Backend: GetAllTags, RenameTag, DeleteTag store methods
4. Backend: REST endpoints for tags
5. Frontend: TagInput component
6. Frontend: wire TagInput into DynamicForm + ItemDetail
7. Frontend: TagFilter component + wire into App.vue
8. Frontend: TagManager component
9. Frontend: collectionStore tag filter integration
10. Update CLAUDE.md and README

## Acceptance Test Flow

1. Edit an item -- type "gift" in tag input, press Enter -- tag chip appears
2. Add a second tag "rare" -- both chips visible
3. Save the item -- reopen, verify both tags persist
4. Add "gift" tag to an item in a different module
5. Select "All Types" in sidebar -- activate "gift" tag filter -- verify both items appear
6. Combine tag filter with text search -- verify AND logic
7. Combine tag filter with module filter -- verify items filtered correctly
8. Search for "gift" in command palette -- verify tagged items appear
9. Open tag management -- verify "gift" shows count of 2, "rare" shows count of 1
10. Rename "gift" to "presents" -- verify both items now show "presents"
11. Delete "rare" tag -- verify removed from the item that had it
12. Export items as CSV -- verify "tags" column with comma-separated values
