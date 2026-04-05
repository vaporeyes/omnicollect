# Implementation Plan: Backup Export & Sync Preparation

**Branch**: `005-backup-export-sync-prep` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-backup-export-sync-prep/spec.md`

## Summary

Add a ZIP export function to the Go backend that bundles the SQLite
database, media directories, and module schemas into a single archive.
Harden timestamp tracking across all item modification paths. Add an
"Export Backup" button to the frontend. No new dependencies -- uses
Go standard library `archive/zip`. Prepares the data model for future
sync by ensuring all modifications carry accurate UTC timestamps.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 4.6+, Vue 3.2+
**Primary Dependencies**: Go standard library `archive/zip` (no new deps)
**Storage**: Reads from SQLite DB + media dirs, writes ZIP to user-chosen path
**Testing**: Manual verification via `wails dev`
**Target Platform**: macOS, Linux, Windows (desktop)
**Project Type**: Desktop application (Wails: Go backend + Vue 3 frontend)
**Performance Goals**: <30s export for 1000 items + 500 images
**Constraints**: Streaming compression (not in-memory), no UI freeze
**Scale/Scope**: Up to multi-GB media directories

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Export is local. No network. Archive is self-contained. |
| II. Schema-Driven UI | N/A | No schema changes. |
| III. Flat Data Architecture | PASS | Database exported as-is. |
| IV. Performance & Memory | PASS | Streaming ZIP (no full in-memory buffering). |
| V. Type-Safe IPC | PASS | ExportBackup exposed via Wails binding. |
| VI. Documentation | PASS | All methods documented. |

## Project Structure

### Documentation (this feature)

```text
specs/005-backup-export-sync-prep/
  plan.md
  research.md
  data-model.md
  quickstart.md
  contracts/
    wails-bindings.md
  checklists/
    requirements.md
```

### Source Code (new and modified files)

```text
# Go backend (project root)
backup.go            # NEW: ZIP archive generation
app.go               # MODIFIED: add ExportBackup binding

# Vue frontend (frontend/src/)
frontend/src/
  App.vue            # MODIFIED: add Export Backup button in sidebar
```

**Structure Decision**: One new Go file (`backup.go`) for all export
logic. Minimal frontend change -- just a button + progress state.

## Complexity Tracking

No constitution violations to justify.
