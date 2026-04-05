# Implementation Plan: Schema Visual Builder

**Branch**: `004-schema-visual-builder` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-schema-visual-builder/spec.md`

## Summary

Build a split-pane schema editor: a visual form builder on the left
(add/edit/reorder fields, live preview) and a CodeMirror JSON editor on
the right (syntax highlighting, graceful error handling). Both panes
share a single reactive draft schema object. A Go binding writes the
schema to disk and hot-reloads the module list. Reinforces Constitution
Principle II: users create schemas without writing code.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 4.6+, Vue 3.2+
**Primary Dependencies**: vue-codemirror + @codemirror/lang-json (code editor), existing Pinia/Wails stack
**Storage**: Writes JSON files to `~/.omnicollect/modules/`
**Testing**: Manual verification via `wails dev`
**Target Platform**: macOS, Linux, Windows (desktop)
**Project Type**: Desktop application (Wails: Go backend + Vue 3 frontend)
**Performance Goals**: <500ms bidirectional sync, <1s save-to-available
**Constraints**: 100% offline, graceful JSON parse error handling
**Scale/Scope**: Single user, schemas with up to ~50 attributes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Schemas saved to local filesystem. No network. |
| II. Schema-Driven UI | PASS | This feature empowers users to create the schemas that drive the UI. |
| III. Flat Data Architecture | N/A | No database changes. |
| IV. Performance & Memory | N/A | No media handling. |
| V. Type-Safe IPC | PASS | SaveCustomModule and LoadModuleFile via Wails bindings. |
| VI. Documentation | PASS | All components and bindings documented. |

**Post-Phase 1 re-check**: All gates pass. The builder creates
standard ModuleSchema JSON files consumed by the existing schema-driven
form renderer.

## Project Structure

### Documentation (this feature)

```text
specs/004-schema-visual-builder/
  plan.md
  research.md
  data-model.md
  quickstart.md
  contracts/
    wails-bindings.md
    component-contracts.md
  checklists/
    requirements.md
```

### Source Code (new and modified files)

```text
# Go backend (project root)
app.go               # MODIFIED: add SaveCustomModule, LoadModuleFile
modules.go           # MODIFIED: add reloadModules(), findModuleFile()

# Vue frontend (frontend/src/)
frontend/src/
  components/
    SchemaBuilder.vue        # NEW: top-level split-pane layout
    SchemaVisualEditor.vue   # NEW: visual field builder
    SchemaCodeEditor.vue     # NEW: CodeMirror JSON editor wrapper
    SchemaFormPreview.vue    # NEW: live form preview
  App.vue                    # MODIFIED: add Schema Builder nav entry
```

**Structure Decision**: Four new Vue components for the builder. Two
Go method additions. No new Go files needed -- the module management
logic extends `modules.go` and `app.go`.

## Complexity Tracking

No constitution violations to justify.
