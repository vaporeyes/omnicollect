# Implementation Plan: Schema-Driven Faceted Filtering

**Branch**: `007-faceted-filtering` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-faceted-filtering/spec.md`

## Summary

Add a dynamically generated filter bar derived from the active module's schema. Enum attributes become multi-select pills (OR within, AND across), boolean attributes become tri-state toggles (off/true/false), and number attributes become inline min/max range inputs. The backend `GetItems` binding is extended to accept a JSON filter payload that combines with existing FTS5 search. The filter bar is collapsible to manage visual density.

## Technical Context

**Language/Version**: Go 1.25+ (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
**Storage**: SQLite via modernc.org/sqlite (FTS5 full-text search, JSON attributes column)
**Testing**: Manual acceptance testing
**Target Platform**: macOS desktop (Wails v2)
**Project Type**: Desktop application (Go + Vue via Wails)
**Performance Goals**: Filter results within 500ms of filter change
**Constraints**: Offline-capable (local SQLite); Constitution III (flat data, no JOINs); filter JSON attributes in-query
**Scale/Scope**: Hundreds to low-thousands of items; single-user desktop app

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | All filtering is local SQLite queries; no network dependency |
| II. Schema-Driven UI | PASS | Filter bar is generated dynamically from JSON schema at runtime; no hardcoded type templates |
| III. Flat Data Architecture | PASS | Filters query the existing flat JSON `attributes` column using SQLite JSON functions; no new tables or JOINs |
| IV. Performance & Memory | PASS | Filter results use same thumbnail paths; no original images loaded |
| V. Type-Safe IPC | PASS | Extended `GetItems` binding will have Wails-generated TypeScript types |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update required at completion |

All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/007-faceted-filtering/
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
app.go                   # Modified: update GetItems signature to accept filters JSON
db.go                    # Modified: extend queryItems to parse filter payload and build
                         #   WHERE clauses using json_extract on attributes column

# Frontend (Vue/TypeScript)
frontend/src/
  components/
    FilterBar.vue        # New: collapsible filter bar with enum pills, boolean toggles,
                         #   number range inputs, clear-all action
  stores/
    collectionStore.ts   # Modified: add filters state, pass filter payload to GetItems
  App.vue                # Modified: render FilterBar, wire filter state, clear on module switch
```

**Structure Decision**: One new component (FilterBar.vue) plus modifications to existing backend and frontend files. The filter bar is a single component shared between list and grid views via the parent App.vue, which manages filter state in the collection store.

## Complexity Tracking

No constitution violations. No complexity justification needed.
