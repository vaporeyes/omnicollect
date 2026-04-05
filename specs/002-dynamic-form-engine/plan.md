# Implementation Plan: Dynamic Form Engine

**Branch**: `002-dynamic-form-engine` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-dynamic-form-engine/spec.md`

## Summary

Build the Vue 3 frontend for OmniCollect: Pinia stores to cache module
schemas and collection items, a dynamic form component that renders
input controls from JSON schema attribute definitions, an item list view
with filtering and search, and proper payload construction for the Go
backend's `SaveItem` binding. All UI forms are schema-driven per
Constitution Principle II -- no hardcoded collection-type templates.

## Technical Context

**Language/Version**: TypeScript 4.6+, Vue 3.2+ (Composition API)
**Primary Dependencies**: Vue 3, Pinia (state management), Wails runtime (IPC)
**Storage**: N/A (frontend caches only; persistence via Go backend)
**Testing**: Manual verification via `wails dev` (automated frontend tests deferred)
**Target Platform**: Desktop (macOS, Linux, Windows via Wails)
**Project Type**: Desktop application frontend (Vue 3 SPA embedded in Wails)
**Performance Goals**: <1s list refresh, <1s search response, <60s item creation
**Constraints**: 100% offline, schema-driven UI, type-safe IPC only
**Scale/Scope**: Single user, up to a few hundred items, up to 50 module schemas

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | All data via local Wails IPC to SQLite. No network calls. |
| II. Schema-Driven UI | PASS | DynamicForm.vue renders from schema. No type-specific templates. |
| III. Flat Data Architecture | PASS | Frontend constructs flat `attributes` JSON object. No JOINs. |
| IV. Performance & Memory | N/A | No media loading in this iteration. Images field ignored. |
| V. Type-Safe IPC | PASS | All calls via Wails-generated TypeScript bindings. |
| VI. Documentation | PASS | All components and stores documented. |

**Post-Phase 1 re-check**: All gates still pass. The DynamicForm component
is generic and receives schema as a prop -- Constitution Principle II is
directly enforced by the architecture.

## Project Structure

### Documentation (this feature)

```text
specs/002-dynamic-form-engine/
  plan.md
  research.md
  data-model.md
  quickstart.md
  contracts/
    component-contracts.md
  checklists/
    requirements.md
```

### Source Code (frontend/ directory)

```text
frontend/
  src/
    main.ts                    # Vue app entry, Pinia setup
    App.vue                    # Root layout: sidebar + main content
    stores/
      moduleStore.ts           # useModuleStore: fetches/caches schemas
      collectionStore.ts       # useCollectionStore: fetches/caches items
    components/
      DynamicForm.vue          # Schema-driven form renderer
      ItemList.vue             # Item list with filter/search
      ModuleSelector.vue       # Module type picker
      FormField.vue            # Single field renderer (type dispatch)
```

**Structure Decision**: All new code lives under `frontend/src/`. Stores
in `stores/`, UI components in `components/`. No routing needed for this
iteration -- single-page layout with sidebar navigation via component
state.

## Complexity Tracking

No constitution violations to justify.
