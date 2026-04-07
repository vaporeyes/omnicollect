# Implementation Plan: Multi-Select and Bulk Actions

**Branch**: `009-bulk-actions` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-bulk-actions/spec.md`

## Summary

Add multi-select capability to list and grid views with checkboxes/overlays, Shift-click range selection, and a floating glassmorphism action bar. Three bulk actions: atomic batch delete (new Go binding), CSV export (new Go binding), and bulk module reassignment (new Go binding). Selection state managed in a Pinia store, shared between views.

## Technical Context

**Language/Version**: Go 1.25+ (backend -- new bindings), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
**Storage**: SQLite via modernc.org/sqlite (batch delete in transaction, CSV query)
**Testing**: Manual acceptance testing
**Target Platform**: macOS desktop (Wails v2)
**Project Type**: Desktop application (Go + Vue via Wails)
**Performance Goals**: Action bar appears within 100ms; batch delete of 100 items under 1s
**Constraints**: Offline-capable; atomic delete transactions; CSV generated server-side
**Scale/Scope**: Single-user; selections up to hundreds of items

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | All operations are local SQLite; CSV generated locally |
| II. Schema-Driven UI | PASS | No type-specific templates; selection/action bar are generic |
| III. Flat Data Architecture | PASS | Batch delete/update operates on existing items table; CSV reads flat JSON attributes |
| IV. Performance & Memory | PASS | Selection is a set of IDs (lightweight); no media loaded |
| V. Type-Safe IPC | PASS | Three new Wails bindings with generated TypeScript types |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update at completion |

All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/009-bulk-actions/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  contracts/           # Phase 1 output
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Backend (Go)
app.go                   # Modified: add DeleteItems, ExportItemsCSV, BulkUpdateModule bindings
db.go                    # Modified: add deleteItems (batch in transaction),
                         #   exportItemsCSV (query + CSV generation),
                         #   bulkUpdateModule (batch update in transaction)

# Frontend (Vue/TypeScript)
frontend/src/
  stores/
    selectionStore.ts    # New: selected IDs set, toggle, range select, clear, select all
  components/
    BulkActionBar.vue    # New: floating glassmorphism bar with count + action buttons
    ItemList.vue         # Modified: add checkboxes, wire selection, select-all header
    CollectionGrid.vue   # Modified: add selection overlay on cards
  App.vue                # Modified: render BulkActionBar, wire actions, clear on navigate
```

**Structure Decision**: One new store (selectionStore), one new component (BulkActionBar), modifications to list/grid views and App.vue. Three new Go backend bindings for atomic batch operations.

## Complexity Tracking

No constitution violations. No complexity justification needed.
