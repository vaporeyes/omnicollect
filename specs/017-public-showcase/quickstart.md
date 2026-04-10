# Quickstart: Public Showcase URLs

**Branch**: `017-public-showcase` | **Date**: 2026-04-10

## Prerequisites

- Existing codebase on the `017-public-showcase` branch
- Cloud mode deployment (showcase requires public-facing URLs)
- Items with images for a visually compelling showcase

## No New Dependencies

Go `html/template` (standard library). No new npm packages.

## Files to Create

### Backend
1. **`showcase/handler.go`** -- Showcase HTTP handlers (gallery page, toggle API, list API)
2. **`showcase/templates.go`** -- Template loading/rendering via `//go:embed`
3. **`showcase/templates/gallery.html`** -- Gallery page template (header, grid, detail overlay, pagination, CSS)
4. **`showcase/templates/unavailable.html`** -- Revoked showcase page

## Files to Modify

### Backend
1. **`storage/db.go`** -- Add Showcase type + Store interface methods
2. **`storage/sqlite.go`** -- Add showcases table DDL + CRUD methods
3. **`storage/postgres.go`** -- Add showcases table (public schema) + CRUD methods
4. **`server.go`** -- Register public showcase route + authenticated showcase API routes
5. **`handlers.go`** -- Add showcase toggle/list handlers (or delegate to showcase package)

### Frontend
6. **`frontend/src/api/types.ts`** -- Add Showcase type
7. **`frontend/src/api/client.ts`** -- Add toggleShowcase, listShowcases functions
8. **`frontend/src/components/ModuleSelector.vue`** -- Add public/private toggle per module with copy-link action
9. **`frontend/src/App.vue`** -- Wire showcase toggle (minor)

## Implementation Order

1. Backend: Showcase type + Store methods (db.go, sqlite.go, postgres.go)
2. Backend: Go HTML templates (gallery.html, unavailable.html)
3. Backend: Showcase handlers (handler.go, templates.go)
4. Backend: Register routes (server.go) -- public gallery route outside auth
5. Frontend: API client + types
6. Frontend: Toggle UI in ModuleSelector
7. Test: toggle collection public, visit showcase URL, verify gallery
8. Update CLAUDE.md and README

## Acceptance Test Flow

### Make Public + View Gallery
1. Toggle a collection (e.g., "Coins") to public via the UI
2. Copy the generated showcase URL
3. Open the URL in an incognito browser (no auth cookies)
4. Verify: gallery page shows collection name, item count, responsive grid
5. Verify: item cards show thumbnails, titles, attributes
6. Click an item card -- verify detail overlay shows full image + attributes
7. Click close (or browser back) -- verify overlay closes
8. Verify: no sidebar, toolbar, edit buttons, or login prompts visible

### Toggle Private
9. Toggle the collection back to "Private"
10. Refresh the showcase URL -- verify "no longer available" page
11. Toggle back to "Public" -- verify same URL works again

### Edge Cases
12. Make a collection with zero items public -- verify empty state message
13. View showcase on mobile -- verify responsive layout
14. Test pagination: create 30+ items, verify page 2 link works
15. Verify showcase does not expose other collections or owner identity
