# Tasks: Public Showcase URLs

**Input**: Design documents from `/specs/017-public-showcase/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/showcase-contract.md, quickstart.md

**Tests**: Not explicitly requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Data model and types shared by all stories

- [x] T001 Add `Showcase` struct to `storage/db.go`: fields ID, Slug, TenantID, ModuleID, Enabled (bool), CreatedAt, UpdatedAt; add `GetShowcaseBySlug`, `GetShowcaseForModule`, `UpsertShowcase`, `ListShowcases` methods to the Store interface
- [x] T002 [P] Add `Showcase` interface to `frontend/src/api/types.ts`: id, slug, moduleId, enabled, createdAt, updatedAt
- [x] T003 [P] Add `toggleShowcase(moduleId: string, enabled: boolean)` and `listShowcases()` functions to `frontend/src/api/client.ts`

**Checkpoint**: Types defined in Go + TypeScript; API client functions ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema + Store implementations for showcase CRUD

- [x] T004 Add showcases table DDL to `storage/sqlite.go`: `CREATE TABLE IF NOT EXISTS showcases (id TEXT PRIMARY KEY, slug TEXT NOT NULL UNIQUE, tenant_id TEXT NOT NULL, module_id TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`; add unique index on slug; add migration for existing databases (`ALTER TABLE` if not exists pattern)
- [x] T005 Add showcases table DDL to `storage/postgres.go`: create in the `public` schema (not per-tenant) so slugs can be looked up without tenant context; columns: id, slug, tenant_id, module_id, enabled, created_at, updated_at; unique index on slug; add in `initTenantSchema` or a separate public schema init
- [x] T006 Implement `GetShowcaseBySlug`, `GetShowcaseForModule`, `UpsertShowcase`, `ListShowcases` in `storage/sqlite.go`
- [x] T007 Implement `GetShowcaseBySlug`, `GetShowcaseForModule`, `UpsertShowcase`, `ListShowcases` in `storage/postgres.go` -- note: these query the `public.showcases` table, not the tenant schema
- [x] T008 Implement slug generation helper in `storage/db.go` or a shared util: `GenerateShowcaseSlug(moduleName string) string` -- slugify module name (lowercase, hyphens, max 30 chars) + 8-char hex random suffix from `crypto/rand`

**Checkpoint**: Showcase CRUD works in both SQLite and PostgreSQL; `go vet ./...` passes

---

## Phase 3: User Story 1 - Make Public + Shareable Link (Priority: P1) MVP

**Goal**: Toggle a collection public, get a URL, visitors see a read-only gallery.

**Independent Test**: Toggle collection public, copy URL, open in incognito, verify gallery renders.

### Implementation for User Story 1

- [x] T009 [US1] Create `showcase/templates/gallery.html`: Go HTML template with embedded CSS; header (collection name, item count), responsive card grid (thumbnail, title, attributes per schema), CSS `:target` detail overlay per item (full image, all attributes, tags, close link), pagination links; uses Outfit + Instrument Serif fonts via Google Fonts link
- [x] T010 [US1] Create `showcase/templates/unavailable.html`: friendly "This showcase is no longer available" page with minimal styling
- [x] T011 [US1] Create `showcase/templates.go`: embed templates via `//go:embed templates/*.html`; implement `RenderGallery(w, data)` and `RenderUnavailable(w)` functions using `html/template`
- [x] T012 [US1] Create `showcase/handler.go`: implement `HandleShowcase(store, mediaStore)` returning an `http.HandlerFunc` that: parses slug from URL path, calls `store.GetShowcaseBySlug`, if not found/disabled renders unavailable page, otherwise loads items for that tenant+module (set search_path for PG), loads module schema, renders gallery template with items + schema + pagination
- [x] T013 [US1] Add showcase toggle handler in `handlers.go`: `handleToggleShowcase` -- parse `{moduleId, enabled}` from body; if enabling for first time, generate slug via `GenerateShowcaseSlug`; call `store.UpsertShowcase`; return Showcase JSON with the URL
- [x] T014 [US1] Add showcase list handler in `handlers.go`: `handleListShowcases` -- call `store.ListShowcases`, return JSON array
- [x] T015 [US1] Register routes in `server.go`: `GET /showcase/{slug}` OUTSIDE auth middleware (public); `POST /api/v1/showcases/toggle` and `GET /api/v1/showcases` inside auth middleware
- [x] T016 [US1] Modify `frontend/src/components/ModuleSelector.vue`: add a small "Share" toggle icon per module; when clicked, call `toggleShowcase(moduleId, !currentState)`; show the showcase URL in a copyable field when enabled; indicate public state with a small icon/badge

**Checkpoint**: Full flow works: toggle public -> copy URL -> incognito visit -> gallery renders

---

## Phase 4: User Story 2 - Toggle Back to Private (Priority: P2)

**Goal**: Revoking access immediately; stable slug on re-enable.

**Independent Test**: Make public, copy URL, toggle private, verify URL shows "unavailable", toggle public again, verify same URL works.

### Implementation for User Story 2

- [x] T017 [US2] Verify toggle-off behavior in `showcase/handler.go`: `GetShowcaseBySlug` returns the showcase even when disabled; the handler checks `enabled` flag and renders unavailable page
- [x] T018 [US2] Verify slug stability in `storage/sqlite.go` and `storage/postgres.go`: `UpsertShowcase` must update `enabled` and `updated_at` without regenerating the slug; only generate slug on first creation
- [x] T019 [US2] Update `ModuleSelector.vue`: when toggle is OFF, hide the URL field but keep the toggle visible; show visual indicator (e.g., grayed icon) for private state

**Checkpoint**: Toggle private/public works; slug is stable; unavailable page renders for disabled showcases

---

## Phase 5: User Story 3 - Gallery Design Polish (Priority: P3)

**Goal**: Polished, responsive gallery page for external viewers.

**Independent Test**: View gallery on desktop + mobile, verify responsive layout, verify item detail overlay works.

### Implementation for User Story 3

- [x] T020 [US3] Polish `showcase/templates/gallery.html` CSS: responsive grid (`grid-template-columns: repeat(auto-fill, minmax(220px, 1fr))`), card hover effects, smooth detail overlay transition, proper image aspect ratios, attribute key-value layout in detail, tag chips, mobile breakpoints
- [x] T021 [US3] Add pagination to `showcase/handler.go`: parse `?page=N` query param, default 24 items per page, pass page info to template, render previous/next links
- [x] T022 [US3] Handle empty collection in `showcase/templates/gallery.html`: when item count is zero, show "This collection is empty" message instead of the grid
- [x] T023 [US3] Ensure no data leakage in `showcase/handler.go`: verify the handler never includes owner identity, tenant ID, other collections, or any application navigation in the rendered HTML

**Checkpoint**: Gallery is visually polished, responsive, paginated, zero data leakage

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Local mode detection, documentation

- [x] T024 Add cloud-mode check: in the frontend, call `/api/v1/ai/status` or a dedicated config endpoint to determine if showcase feature is available (only in cloud mode); hide the showcase toggle UI in local/desktop mode
- [x] T025 [P] Update `CLAUDE.md`: document showcase/ package, templates, public routes, Store interface additions, showcase table schema
- [x] T026 [P] Update `README.md`: add "Public Showcase" section with usage description, URL format, toggle flow, iteration 17 history
- [x] T027 Run quickstart.md acceptance test flow (all 15 steps) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs Showcase type in Store interface)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs showcase CRUD + slug generation)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (extends toggle behavior)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (polishes the gallery template)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- toggle + gallery rendering.
- **US2 (P2)**: Depends on US1 (extends toggle + slug stability).
- **US3 (P3)**: Depends on US1 (polishes gallery template created in US1).

### Parallel Opportunities

- T002 and T003 (Phase 1 frontend) and T001 (Phase 1 backend) -- different languages
- T004 and T005 (Phase 2 SQLite/PG DDL) -- different files
- T006 and T007 (Phase 2 SQLite/PG CRUD) -- different files
- T025 and T026 (docs) -- different files

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 + Phase 2 = types + CRUD + slug generation
2. Phase 3 (US1) = templates + handlers + routes + toggle UI
3. **STOP and VALIDATE**: toggle public, visit URL in incognito, verify gallery
4. This delivers the full shareable showcase experience

### Incremental Delivery

1. Phase 1 + Phase 2 = data model + Store methods
2. Phase 3 (US1) = gallery rendering + toggle (MVP)
3. Phase 4 (US2) = private toggle + slug stability
4. Phase 5 (US3) = gallery design polish + pagination + responsiveness
5. Phase 6 = local mode detection + docs

---

## Notes

- Gallery is server-rendered HTML via Go `html/template` -- zero JavaScript for visitors
- CSS `:target` pseudo-class provides item detail overlay without JS
- Slugs: `{module-name}-{8-hex}`, stable across public/private toggles
- PostgreSQL: showcases in `public` schema for cross-tenant slug lookup
- SQLite: showcases in main database (single-tenant)
- Public gallery route registered OUTSIDE auth middleware chain
- 24 items per page with server-side pagination
- Templates embedded via `//go:embed`
- Feature disabled in local/desktop mode
