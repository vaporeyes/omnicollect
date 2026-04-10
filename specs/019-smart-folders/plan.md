# Implementation Plan: Smart Folders (Saved Views)

**Branch**: `019-smart-folders` | **Date**: 2026-04-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/019-smart-folders/spec.md`

## Summary

Add Smart Folders to the sidebar that let users save, restore, rename, and delete named view state snapshots (module + search + filters + tags). Saved as a JSON array within the existing settings storage mechanism. New sidebar section with inline naming input, context menu for management, and visual highlight for the active Smart Folder.

## Technical Context

**Language/Version**: TypeScript 4.6+ (frontend), Go 1.25+ (backend -- settings storage only, no new endpoints)
**Primary Dependencies**: Vue 3.2+, Pinia 3.0+ (new store: smartFolderStore)
**Storage**: Existing settings JSON blob (SQLite for local, PostgreSQL for cloud) -- Smart Folders stored as `smartFolders` key
**Testing**: Vitest (frontend unit tests for store logic)
**Target Platform**: Desktop (Wails) + standalone HTTP server (browser)
**Project Type**: Desktop app (Wails v2 shell)
**Performance Goals**: Smart Folder apply within 1 second; sidebar usable with 50+ folders
**Constraints**: No new backend API endpoints; persistence via existing GET/PUT /api/v1/settings
**Scale/Scope**: Up to 50 Smart Folders per user

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First Mandate | PASS | Smart Folders stored in local settings. No network dependency. |
| II. Schema-Driven UI | PASS | Smart Folders are navigation shortcuts, not type-specific UI. |
| III. Flat Data Architecture | PASS | No new database tables. Data stored as JSON within existing settings blob. |
| IV. Performance & Memory Protection | PASS | Smart Folders store serialized state (small JSON), not media references. |
| V. Type-Safe IPC | PASS | Uses existing typed settings GET/PUT endpoints. No new IPC calls. |
| VI. Documentation is Paramount | REQUIRES | CLAUDE.md and README must be updated after implementation. |

## Project Structure

### Documentation (this feature)

```text
specs/019-smart-folders/
  plan.md              # This file
  research.md          # Persistence and UX decisions
  data-model.md        # SmartFolder entity definition
  quickstart.md        # Setup and dev guide
  tasks.md             # Implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
frontend/src/
  stores/
    smartFolderStore.ts     # New Pinia store: CRUD, persistence, active state
  components/
    SmartFolders.vue        # New sidebar section: folder list, save button, inline naming
    App.vue                 # Modified: integrate SmartFolders section, wire apply handler
```

**Structure Decision**: One new Pinia store for Smart Folder state management and persistence, one new Vue component for the sidebar section. App.vue wires the apply action to set collection store state. Existing ContextMenu component reused for right-click rename/delete.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
