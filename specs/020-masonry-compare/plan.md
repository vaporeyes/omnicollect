# Implementation Plan: Masonry Grid & Item Comparison

**Branch**: `020-masonry-compare` | **Date**: 2026-04-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/020-masonry-compare/spec.md`

## Summary

Upgrade the collection grid from a uniform-height CSS grid to a masonry layout that respects each item's image aspect ratio, and add an item comparison mode where exactly two selected items are displayed side-by-side with synchronized image galleries and an attribute diff table (including core fields: title, price, tags).

## Technical Context

**Language/Version**: Go 1.25+ (backend, no changes), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Vue 3 Composition API, Pinia stores, existing CSS variable system
**Storage**: N/A (no backend or database changes)
**Testing**: Vitest (frontend unit tests)
**Target Platform**: Desktop (Wails) and web browser (standalone HTTP mode)
**Project Type**: Desktop app with web frontend
**Performance Goals**: Masonry reflow within 200ms of viewport resize; 60fps scroll
**Constraints**: Grid views MUST use thumbnails only (Constitution Principle IV); no new npm dependencies preferred
**Scale/Scope**: Collections of 10-1000 items typical; comparison always exactly 2 items

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First Mandate | PASS | No network dependency; all client-side |
| II. Schema-Driven UI | PASS | Comparison diff table uses module schema for field labels; no hardcoded forms |
| III. Flat Data Architecture | PASS | No database changes |
| IV. Performance & Memory Protection | PASS | Grid continues using thumbnails; originals only in comparison detail galleries |
| V. Type-Safe IPC | PASS | No new IPC calls; frontend-only feature |
| VI. Documentation is Paramount | PASS | Spec artifacts produced; README and CLAUDE.md updates required at completion |

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/020-masonry-compare/
├── plan.md              # This file
├── research.md          # Phase 0: CSS masonry approaches, diff highlighting patterns
├── data-model.md        # Phase 1: No new entities; documents comparison view data flow
├── quickstart.md        # Phase 1: Developer setup and testing guide
├── contracts/           # Phase 1: Component prop/emit interfaces
└── tasks.md             # Phase 2: Task breakdown (via /speckit.tasks)
```

### Source Code (repository root)

```text
frontend/src/
├── components/
│   ├── CollectionGrid.vue      # MODIFY: Replace CSS grid with masonry layout
│   ├── BulkActionBar.vue       # MODIFY: Add Compare button (count === 2)
│   └── ComparisonView.vue      # NEW: Side-by-side comparison with synced galleries
├── stores/
│   └── selectionStore.ts       # EXISTING: No changes (already provides selectedIds)
└── App.vue                     # MODIFY: Add comparison view routing and state
```

**Structure Decision**: All changes are in the existing frontend/src/ tree. One new component (ComparisonView.vue), three modified files. No new directories, stores, or backend changes.
