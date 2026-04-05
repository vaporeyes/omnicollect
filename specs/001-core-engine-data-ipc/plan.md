# Implementation Plan: Core Engine (Data & IPC)

**Branch**: `001-core-engine-data-ipc` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-core-engine-data-ipc/spec.md`

## Summary

Establish the Go backend foundation for OmniCollect: a single SQLite
`items` table with FTS5 full-text search, a module schema loader that
reads JSON definitions from `~/.omnicollect/modules/`, and three Wails
IPC bindings (`SaveItem`, `GetItems`, `GetActiveModules`) exposed to the
Vue 3 frontend via generated TypeScript. Uses `modernc.org/sqlite`
(CGO-free) and `github.com/google/uuid` for item IDs.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Wails v2, modernc.org/sqlite, github.com/google/uuid
**Storage**: SQLite (local, embedded, CGO-free via modernc.org/sqlite)
**Testing**: Go standard `testing` package
**Target Platform**: macOS, Linux, Windows (desktop)
**Project Type**: Desktop application (Wails: Go backend + Vue 3 frontend)
**Performance Goals**: <1s item save/retrieve, <500ms FTS search over 10k items
**Constraints**: 100% offline, no network calls, CGO-free builds
**Scale/Scope**: Single user, up to 50 module schemas, up to 10k+ items

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First Mandate | PASS | SQLite is local. No network calls. Fully offline. |
| II. Schema-Driven UI | PASS | ModuleSchema loaded from JSON files drives frontend forms. No hardcoded type templates. |
| III. Flat Data Architecture | PASS | Single `items` table with JSON `attributes` blob. No JOINs for item data. |
| IV. Performance & Memory Protection | N/A | This iteration stores image paths only. No media loading. Thumbnails deferred. |
| V. Type-Safe IPC | PASS | All three methods exposed via Wails-generated TypeScript bindings. No raw IPC. |
| VI. Documentation is Paramount | PASS | All types documented. Quickstart provided. |

**Post-Phase 1 re-check**: All gates still pass. The FTS5 virtual table
uses triggers (no JOINs on the items table itself). The `items_fts`
table is an index, not a data table, so Principle III is not violated.

## Project Structure

### Documentation (this feature)

```text
specs/001-core-engine-data-ipc/
  plan.md
  research.md
  data-model.md
  quickstart.md
  contracts/
    wails-bindings.md
  checklists/
    requirements.md
```

### Source Code (repository root)

```text
omnicollect/
  wails.json               # Wails project config
  main.go                  # Entry point: wails.Run(), embed directive
  app.go                   # App struct: SaveItem, GetItems, GetActiveModules
  db.go                    # SQLite init, schema DDL, FTS5 triggers
  modules.go               # Module schema loader (scan dir, parse JSON)
  models.go                # Shared types: Item, ModuleSchema, etc.
  go.mod
  go.sum
  build/
    appicon.png
    darwin/
  frontend/
    index.html
    package.json
    vite.config.ts
    tsconfig.json
    src/
      main.ts
      App.vue
      components/
      assets/
    wailsjs/               # Auto-generated (never edit)
```

**Structure Decision**: Standard Wails v2 layout. Go backend files at
project root. Vue 3 frontend under `frontend/`. No sub-packages needed
for this iteration; four Go files (`app.go`, `db.go`, `modules.go`,
`models.go`) plus `main.go` keep the codebase flat and navigable.

## Complexity Tracking

No constitution violations to justify. All principles pass cleanly.
