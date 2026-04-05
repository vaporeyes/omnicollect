# Implementation Plan: Image Processing & Grid Display

**Branch**: `003-image-processing-grid` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-image-processing-grid/spec.md`

## Summary

Add image processing to the Go backend (thumbnail generation via
`disintegration/imaging`, original archival), configure Wails AssetServer
to serve local media files to the frontend, and build a Vue 3 collection
grid component with lazy-loaded thumbnails. Enforces Constitution
Principle IV: the grid view MUST use compressed thumbnails only, never
original full-resolution images.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript 4.6+, Vue 3.2+
**Primary Dependencies**: disintegration/imaging (Go thumbnails), golang.org/x/image/webp (WebP decode)
**Storage**: Local filesystem (`~/.omnicollect/media/originals/` and `thumbnails/`)
**Testing**: Manual verification via `wails dev`
**Target Platform**: macOS, Linux, Windows (desktop)
**Project Type**: Desktop application (Wails: Go backend + Vue 3 frontend)
**Performance Goals**: <3s image attach-to-thumbnail, smooth grid scroll at 100+ items
**Constraints**: 100% offline, thumbnails under 100KB, no full-res in grid
**Scale/Scope**: Up to 20 images per item, hundreds of items in grid

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | All media stored locally. No network. |
| II. Schema-Driven UI | N/A | No schema changes in this iteration. |
| III. Flat Data Architecture | PASS | Images stored as filename array in items JSON. |
| IV. Performance & Memory | PASS | Grid uses thumbnails only. Full-res on demand only. Core principle for this feature. |
| V. Type-Safe IPC | PASS | ProcessImage exposed via Wails binding. Media served via AssetServer handler. |
| VI. Documentation | PASS | All new methods documented. |

**Post-Phase 1 re-check**: All gates pass. The AssetServer handler
serves only from the thumbnails directory for grid views. Full-res
originals are served on a separate path, loaded only on explicit user
click.

## Project Structure

### Documentation (this feature)

```text
specs/003-image-processing-grid/
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
imaging.go             # NEW: ProcessImage, thumbnail generation, validation
main.go                # MODIFIED: AssetServer Handler for local media

# Vue frontend (frontend/src/)
frontend/src/
  components/
    CollectionGrid.vue   # NEW: grid view with lazy thumbnails
    ImageAttach.vue      # NEW: file picker for image attachment
    ImageLightbox.vue    # NEW: full-resolution image viewer
    DynamicForm.vue      # MODIFIED: add image attachment section
  App.vue                # MODIFIED: grid/list view toggle
```

**Structure Decision**: One new Go file (`imaging.go`) for all image
processing. Three new Vue components. Media files live outside the
project in the user's home directory.

## Complexity Tracking

No constitution violations to justify.
