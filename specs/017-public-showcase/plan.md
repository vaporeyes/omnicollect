# Implementation Plan: Public Showcase URLs

**Branch**: `017-public-showcase` | **Date**: 2026-04-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/017-public-showcase/spec.md`

## Summary

Add public showcase URLs that let users share read-only collection galleries. A `showcases` table stores slug/tenant/module/enabled state. The showcase page is server-rendered HTML via Go templates (zero JS for visitors). Users toggle collections public via the app UI, get a shareable link, and can revoke access instantly. Disabled in local mode.

## Technical Context

**Language/Version**: Go 1.25+ (backend -- templates, handlers, database), TypeScript + Vue 3 (frontend -- toggle UI only)
**Primary Dependencies**: Go `html/template` (standard library); no new frontend dependencies
**Storage**: New `showcases` table in both SQLite and PostgreSQL; both Store implementations extended
**Testing**: Manual + curl testing of public showcase routes
**Target Platform**: Cloud deployment (public URLs require a server); disabled in local/desktop mode
**Project Type**: Multi-tenant SaaS
**Performance Goals**: Showcase page renders under 500ms; supports 500+ concurrent visitors
**Constraints**: Server-rendered HTML (zero JS); no auth for visitors; stable slugs across toggles

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS (mitigated) | Feature disabled in local mode; app works fully without showcases |
| II. Schema-Driven UI | PASS | Showcase renders attributes dynamically from module schema |
| III. Flat Data Architecture | PASS | New `showcases` table is simple key-value; items queried with existing flat architecture |
| IV. Performance & Memory | PASS | Server-rendered HTML; thumbnails only in grid; originals on detail click |
| V. Type-Safe IPC | PASS | New REST endpoints with typed responses; Go templates are compile-time checked |
| VI. Documentation | PASS | Spec artifacts produced |

All gates pass.

## Project Structure

### Source Code (repository root)

```text
# Backend (Go)
showcase/
  handler.go           # New: showcase HTTP handlers (gallery page, item detail, toggle API)
  templates.go         # New: Go template rendering for gallery + detail pages
  templates/
    gallery.html       # New: gallery page template (header, grid, responsive CSS)
    detail.html        # New: item detail template (full image, attributes, tags)
    unavailable.html   # New: "no longer available" page template
storage/
  db.go                # Modified: add Showcase type + ShowcaseStore methods to Store interface
  sqlite.go            # Modified: add showcases table DDL + CRUD
  postgres.go          # Modified: add showcases table DDL + CRUD
handlers.go            # Modified: add toggle showcase endpoint
server.go              # Modified: register showcase routes (public, no auth)

# Frontend (Vue/TypeScript)
frontend/src/
  api/types.ts         # Modified: add Showcase type
  api/client.ts        # Modified: add toggleShowcase, getShowcase functions
  components/
    ModuleSelector.vue # Modified: add public/private toggle indicator per module
  App.vue              # Modified: showcase toggle UI + copy link action
```

**Structure Decision**: A new `showcase/` package handles all showcase rendering. Go `html/template` for server-rendered pages. Templates embedded via `//go:embed`. Showcase CRUD in the existing Store interface. Frontend changes are minimal -- just the toggle UI.

## Complexity Tracking

No constitution violations. No complexity justification needed.
