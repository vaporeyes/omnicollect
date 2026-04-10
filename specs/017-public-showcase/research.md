# Research: Public Showcase URLs

**Branch**: `017-public-showcase` | **Date**: 2026-04-10

## R1: Server-Rendered HTML Strategy

**Decision**: Per clarification, use Go's `html/template` package for showcase pages. Templates are embedded in the binary via `//go:embed`. Zero JavaScript required for visitors.

**Rationale**: Fastest possible page load for external visitors. No JS bundle to download, parse, or execute. Better SEO (full HTML content on first response). Simpler deployment (no separate frontend build for showcase pages).

**Template structure**:
- `gallery.html` -- grid layout with header, item cards, pagination
- `detail.html` -- full item view with image, attributes, tags (or inline overlay via CSS-only `:target` selector)
- `unavailable.html` -- friendly message for revoked showcases

**Alternatives considered**:
- Separate Vue SPA entry: Rejected per clarification -- adds JS bundle weight for visitors.
- Same Vue app with route check: Rejected -- ships the entire application to visitors.

## R2: Showcase Slug Generation

**Decision**: Slug format: `{slugified-module-name}-{8-char-random-hex}`. Example: `rare-coins-a3f7b2c1`.

**Generation**: On first toggle to public, generate the slug and store it. The slug is stable -- toggling private/public reuses the same slug. This means the same URL works if the collection is re-enabled.

**Slugification**: Module display name -> lowercase, replace spaces with hyphens, remove non-alphanumeric (except hyphens), truncate to 30 chars. Random suffix: 8 hex characters from `crypto/rand`.

**Rationale**: Human-readable part gives context ("rare-coins"). Random suffix prevents URL guessing (8 hex = 4 billion possibilities). Stable slug avoids broken links when toggling.

## R3: Showcase Data Model

**Decision**: New `showcases` table with columns: `id`, `slug` (unique), `tenant_id`, `module_id`, `enabled` (boolean), `created_at`, `updated_at`.

**SQLite DDL**:
```sql
CREATE TABLE IF NOT EXISTS showcases (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  tenant_id TEXT NOT NULL,
  module_id TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_showcases_slug ON showcases(slug);
```

**PostgreSQL**: Same structure in each tenant schema, but `tenant_id` is implicit (schema-per-tenant). Index on `slug`.

**Rationale**: One row per collection that has ever been public. `enabled` toggles visibility without deleting the row (preserves slug stability).

## R4: Showcase Route Architecture

**Decision**: Showcase routes are registered OUTSIDE the auth middleware chain (public, no auth required).

**Routes**:
- `GET /showcase/{slug}` -- serves the gallery HTML page
- `GET /showcase/{slug}/item/{id}` -- serves the item detail HTML page (or inline on gallery page via anchor)

**Data flow**:
1. Look up showcase by slug
2. If not found or not enabled -> render `unavailable.html`
3. Load items for the showcase's tenant + module (set search_path, query items)
4. Render template with items + module schema

**Rationale**: Public routes must be outside the auth middleware. The showcase handler directly queries the store with the showcase's tenant context.

## R5: Item Detail on Showcase

**Decision**: For v1, use a CSS-only approach with `:target` pseudo-class. Each item card links to `#item-{id}`. A hidden detail section with `id="item-{id}"` becomes visible when targeted. This avoids JavaScript entirely while providing an in-page detail overlay.

**Rationale**: Zero JS requirement from the spec. CSS `:target` is well-supported and provides a native "modal" experience. The browser's back button naturally closes the overlay (removes the hash).

**Alternative for v2**: Progressive enhancement with a small JS file for smoother transitions.

## R6: Pagination Strategy

**Decision**: Server-side pagination via query parameters. `GET /showcase/{slug}?page=2`. Default 24 items per page (4x6 grid).

**Rationale**: Server-rendered pages need server-side pagination. 24 items is a good balance for a gallery grid. Pagination links are standard HTML anchor tags.

## R7: Local Mode Detection

**Decision**: The showcase feature is disabled when the application is in local/desktop mode (no `DATABASE_URL` set, or standalone Wails). The toggle UI is hidden. This is checked in the frontend via a new field in the settings/status response.

**Rationale**: Local mode has no public-facing URL. The showcase page requires a server accessible to external visitors.
