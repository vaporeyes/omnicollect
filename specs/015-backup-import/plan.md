# Implementation Plan: Backup Import and Restore

**Branch**: `015-backup-import` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/015-backup-import/spec.md`

## Summary

Add the ability to import OmniCollect backup ZIP files, completing the backup/restore cycle. Supports two modes (Replace and Merge), two backup formats (SQLite-based from local mode and JSON-based from cloud mode), and restores items, images, and module schemas to whichever storage backend is active. Pre-import summary, progress tracking, and atomic Replace mode.

## Technical Context

**Language/Version**: Go 1.25+ (backend import logic), TypeScript + Vue 3 (frontend upload + progress UI)
**Primary Dependencies**: Go `archive/zip` (existing), existing Store and MediaStore interfaces
**Storage**: Imports into whatever backend is active (SQLite or PostgreSQL, local filesystem or S3)
**Testing**: Extend existing test suite; import/export round-trip tests
**Target Platform**: Docker container (cloud) + macOS desktop (local)
**Project Type**: Multi-tenant SaaS + desktop hybrid
**Performance Goals**: 100-item import under 10 seconds; summary preview under 2 seconds
**Constraints**: Replace mode atomic (transaction); Merge mode per-item; both backup formats supported

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Import works with local SQLite; no network required |
| II. Schema-Driven UI | PASS | Module schemas imported from backup drive the UI |
| III. Flat Data Architecture | PASS | Items imported with same flat JSON attributes; no schema changes |
| IV. Performance & Memory | PASS | ZIP processed as stream; images extracted one at a time |
| V. Type-Safe IPC | PASS | New REST endpoint with typed request/response |
| VI. Documentation | PASS | Spec artifacts produced |

All gates pass.

## Project Structure

### Source Code (repository root)

```text
# Backend (Go)
import.go              # New: ZIP extraction, format detection, import logic (replace/merge)
handlers.go            # Modified: add handleImportBackup (multipart upload), handleAnalyzeBackup
server.go              # Modified: register import routes

# Frontend (Vue/TypeScript)
frontend/src/
  components/
    ImportDialog.vue   # New: import modal with file picker, mode selection, summary, progress
  App.vue              # Modified: add "Import Backup" button, wire import dialog
  api/client.ts        # Modified: add analyzeBackup and importBackup functions
  api/types.ts         # Modified: add ImportSummary and ImportResult types
```

**Structure Decision**: One new Go file (`import.go`) for all import logic. One new Vue component (`ImportDialog.vue`) for the entire import UI. Minimal modifications to existing files (handlers, server, App.vue).

## Complexity Tracking

No constitution violations. No complexity justification needed.
