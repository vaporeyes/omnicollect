# Implementation Plan: Spotlight-style Command Palette

**Branch**: `006-command-palette` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-command-palette/spec.md`

## Summary

Add a Spotlight-style command palette overlay (Cmd/Ctrl+K) that lets users search collection items across all modules and trigger quick actions via keyword matching. The palette features a frosted glass input, rich results with thumbnails and module badges, full keyboard navigation, and instant routing to item detail views.

## Technical Context

**Language/Version**: Go 1.25+ (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
**Storage**: SQLite via modernc.org/sqlite (existing FTS5 full-text search)
**Testing**: Manual acceptance testing (no automated test framework in project currently)
**Target Platform**: macOS desktop (Wails v2)
**Project Type**: Desktop application (Go + Vue via Wails)
**Performance Goals**: Search results within 300ms of debounced input; palette open/close under 100ms
**Constraints**: Offline-capable (local SQLite); must use thumbnails only in results (Constitution IV)
**Scale/Scope**: Hundreds to low-thousands of items; single-user desktop app

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Palette queries local SQLite FTS5; no network dependency |
| II. Schema-Driven UI | PASS | No type-specific templates added; palette renders generic item results |
| III. Flat Data Architecture | PASS | Uses existing `items` table + FTS5; no new tables or JOINs |
| IV. Performance & Memory | PASS | Results display thumbnails only; originals loaded only in detail view |
| V. Type-Safe IPC | PASS | Uses existing `GetItems` Wails binding; no new untyped calls needed |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update required at completion |

All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/006-command-palette/
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
frontend/src/
  components/
    CommandPalette.vue     # New: palette overlay component
  stores/
    collectionStore.ts     # Modified: add searchAllItems action (unfiltered search)
  App.vue                  # Modified: Cmd/Ctrl+K handler, palette state, wiring

(No backend changes needed -- existing GetItems binding with empty moduleID
already searches across all modules)
```

**Structure Decision**: Single new component plus minor modifications to existing files. The palette is entirely frontend -- the existing `GetItems("query", "")` backend binding already provides cross-module FTS5 search. A new store action wraps this with debouncing for the palette's use case.

## Complexity Tracking

No constitution violations. No complexity justification needed.
