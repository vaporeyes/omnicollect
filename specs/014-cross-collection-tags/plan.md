# Implementation Plan: Cross-Collection Tags

**Branch**: `014-cross-collection-tags` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/014-cross-collection-tags/spec.md`

## Summary

Add a lightweight tagging system that lets users assign free-form labels to items and query across all collection types. Tags are stored as a JSON array on each item (flat architecture). The backend gains a `tags` query parameter for filtering. The frontend adds a tag input component, tag filter control in collection views, autocomplete from existing tags, and a tag management section.

## Technical Context

**Language/Version**: Go 1.25+ (backend -- Store interface + handlers), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Existing stack (no new dependencies)
**Storage**: Tags stored as JSON array on items (`tags TEXT/JSONB`); both SQLite and PostgreSQL Store implementations updated
**Testing**: Extend existing test suite with tag-related tests
**Target Platform**: Docker container (cloud) + macOS desktop (local)
**Project Type**: Multi-tenant SaaS + desktop hybrid
**Performance Goals**: Tag filter queries under 200ms; autocomplete under 200ms
**Constraints**: Flat data (no junction table, Constitution III); tags in FTS/tsvector index; case-insensitive

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Tags stored locally in SQLite; works offline |
| II. Schema-Driven UI | PASS | Tags are universal, not schema-driven; tag input is a generic component |
| III. Flat Data Architecture | PASS | Tags stored as JSON array on item -- no junction table, no JOINs |
| IV. Performance & Memory | PASS | No media changes; tag arrays are small |
| V. Type-Safe IPC | PASS | Tags added to existing REST endpoints and TypeScript types |
| VI. Documentation | PASS | Spec artifacts produced |

All gates pass.

## Project Structure

### Source Code (repository root)

```text
# Backend (Go)
storage/
  db.go              # Modified: add tags field to Item struct
  sqlite.go          # Modified: add tags column, update FTS trigger, tag filter queries,
                     #   GetAllTags, RenameTag, DeleteTag methods
  postgres.go        # Modified: same changes for PostgreSQL (JSONB array, tsvector)
  migrate.go         # Modified: handle tags in SQLite->PG migration

# Frontend (Vue/TypeScript)
frontend/src/
  api/types.ts       # Modified: add tags field to Item interface
  components/
    TagInput.vue     # New: tag input with autocomplete + removable chips
    TagFilter.vue    # New: tag filter control for collection views
    TagManager.vue   # New: tag management view (list, rename, delete)
    DynamicForm.vue  # Modified: include TagInput in item forms
    ItemDetail.vue   # Modified: display tags as chips
  stores/
    collectionStore.ts  # Modified: add tags filter parameter to queries
  App.vue            # Modified: add tag filter control, tag management route
```

**Structure Decision**: Three new Vue components (TagInput, TagFilter, TagManager). The `tags` field added to the existing Item struct/interface. Store interface extended with tag management methods. No new backend files needed -- just extensions to existing storage implementations.

## Complexity Tracking

No constitution violations. No complexity justification needed.
