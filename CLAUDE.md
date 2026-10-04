# OmniCollect Development Guidelines

## Project Overview

Schema-driven desktop collection manager. Go backend + Vue 3 frontend
via Wails v2. See [Constitution](.specify/memory/constitution.md) for
immutable engineering principles.

## Tech Stack

- **Go 1.26+** (toolchain 1.26.8): Backend (SQLite, image processing, backup, Wails bindings)
- **Vue 3 + TypeScript**: Frontend (Composition API, Pinia stores)
- **Wails v2**: Desktop shell, type-safe IPC, AssetServer
- **SQLite**: Local database via `modernc.org/sqlite` (CGO-free)
- **disintegration/imaging**: Thumbnail generation (CGO-free)
- **vue-codemirror**: CodeMirror 6 JSON editor for schema builder

## Project Structure

```
main.go          # Entry point: --serve, --migrate, or default Wails desktop
config.go        # Environment-based config (DATABASE_URL, S3_*, PORT, TENANT_ID, AI_*)
server.go        # HTTP server setup, routing, CORS middleware
handlers.go      # REST endpoint handlers wrapping App methods
app.go           # Core business logic with Store/MediaStore interfaces
db.go            # Legacy helpers (dbFilePath for backup)
imaging.go       # Image validation, thumbnail generation (returns bytes)
backup.go        # ZIP archive export (database + media + modules)
import.go        # ZIP backup import: format detection, Replace/Merge modes
modules.go       # Legacy helpers (modulesDir for backup)
settings.go      # Wails-bound settings methods delegating to Store
models.go        # Shared Go types (Item, ModuleSchema, ProcessImageResult)
Dockerfile       # Multi-stage build (Go + Node -> alpine)
docker-compose.yml # Dev stack (app + postgres + minio)
ai/
  provider.go    # AIProvider interface + factory (NewAIProvider)
  anthropic.go   # Anthropic Messages API client (direct)
  openai_compat.go # OpenAI-compatible client (OpenRouter, Google, etc.)
  prompt.go      # Schema-to-prompt builder + response parser/validator
showcase/
  handler.go     # Public gallery HTTP handler (slug lookup, item loading, pagination)
  templates.go   # go:embed template rendering (RenderGallery, RenderUnavailable)
  templates/
    gallery.html     # Server-rendered gallery page (CSS :target detail overlay, zero JS)
    unavailable.html # Friendly "no longer available" page
auth/
  context.go   # Tenant ID context helpers (SetTenantID, TenantIDFromContext, SanitizeTenantID)
  middleware.go # JWT validation middleware (Auth0 JWKS), ExemptPaths, provisioning cache
  local.go     # Local-mode middleware (fixed tenant ID bypass)
storage/
  db.go          # Store interface (database abstraction, Showcase type, slug generation)
  media.go       # MediaStore interface (object storage abstraction)
  sqlite.go      # SQLiteStore: local SQLite with FTS5
  sqlite_test.go # Storage layer unit tests (in-memory SQLite)
  postgres.go    # PostgresStore: PostgreSQL with schema-per-tenant, tsvector, ProvisionTenant
  local.go       # LocalMediaStore: local filesystem
  s3.go          # S3MediaStore: S3-compatible object store
  migrate.go     # SQLite-to-PostgreSQL migration tool
  testdata/      # Test fixtures (test-module.json, test-image.jpg)
frontend/src/
  auth/
    plugin.ts    # Auth0 Vue plugin config + token injection wiring
    guard.ts     # AuthGuard component (loading/redirect/render)
  api/
    client.ts    # Centralized fetch-based HTTP client with optional Bearer token
    types.ts     # TypeScript interfaces mirroring Go structs (replaces Wails bindings)
  stores/        # Pinia: moduleStore, collectionStore, selectionStore,
                 #   toastStore, smartFolderStore
  components/    # AppSidebar (extracted sidebar with navigation),
                 #   DynamicForm, FormField, ItemList, CollectionGrid,
                 #   ModuleSelector, ImageAttach, ImageLightbox,
                 #   SchemaBuilder, SchemaVisualEditor,
                 #   SchemaCodeEditor, SchemaFormPreview,
                 #   ItemDetail, SettingsPage, ToastProvider,
                 #   ContextMenu, CommandPalette, FilterBar,
                 #   MarkdownEditor, MarkdownRenderer,
                 #   BulkActionBar, TagInput, TagFilter,
                 #   TagManager, ImportDialog,
                 #   DashboardView, DashboardMetricCard,
                 #   SmartFolders, ComparisonView
  composables/   # useDashboardMetrics (computed insights from items)
```

## Commands

```bash
wails dev        # Desktop development with hot reload
wails build      # Production desktop binary to build/bin/
go run . --serve # Standalone HTTP server mode (port 8080)
go run . --serve --port 3001  # Custom port
go run . --migrate --source /path/to/collection.db --tenant default  # SQLite->PG migration
go vet ./...     # Lint Go code
go mod tidy      # Resolve dependencies
go test ./...    # Run all Go tests (storage + handler)
go test ./storage/... -cover  # Storage tests with coverage
cd frontend && npm test       # Run frontend tests (Vitest)
docker build -t omnicollect .   # Build Docker image
docker-compose up               # Run full cloud stack (app + postgres + minio)
```

## Key Conventions

- All Go files start with two-line ABOUTME comments
- REST API: all operations exposed as HTTP endpoints under /api/v1/
- Frontend uses fetch-based client in api/client.ts (no Wails IPC)
- TypeScript types in api/types.ts mirror Go structs (replaces Wails codegen)
- App struct methods remain as business logic; handlers.go wraps them as HTTP
- Private media served at /thumbnails/ and /originals/ with tenant/auth checks.
  Frontend images MUST use MediaImage.vue (Bearer fetch + revocable blob URLs).
- Two modes: `--serve` for standalone HTTP, default for Wails desktop shell
- Grid views MUST use thumbnails only (Constitution Principle IV)
- No hardcoded collection-type templates (Constitution Principle II)
- README and CLAUDE.md MUST be updated every iteration (Principle VI)
- Module schemas and settings are database-backed in both modes (SQLite/PostgreSQL).
- Database at user config dir (`os.UserConfigDir()`)
- Global shortcuts: Cmd/Ctrl+K (command palette), Cmd/Ctrl+F (search),
  Cmd/Ctrl+N (new item), Escape (close overlays)
- Command palette (`CommandPalette.vue`): cross-module search via
  `collectionStore.searchAllItems()`, quick actions via keyword matching
- Toast notifications via `useToastStore` (replace all alert() calls)
- Context menus on items in list/grid views via `ContextMenu.vue`
- Faceted filtering via `FilterBar.vue`: collapsible bar generated from
  active module schema; enum pills (multi-select OR), boolean tri-state
  toggles (off/true/false), inline number range min/max inputs
- `GetItems(query, moduleID, filtersJSON, tagsJSON)` accepts JSON filter payload;
  backend uses `json_extract()` for attribute filters, direct column for
  `purchasePrice`; `collectionStore.activeFilters` manages filter state
- Cross-collection tags: JSON array on items (`tags TEXT` SQLite / `tags JSONB` PG).
  Tags normalized lowercase on save, max 50 chars. Store methods: `GetAllTags`,
  `RenameTag`, `DeleteTag`. SQLite filter: `json_each` + EXISTS; PG: `?|` + GIN.
  Frontend: `TagInput.vue` (form input with autocomplete), `TagFilter.vue`
  (clickable chips above collection views), `TagManager.vue` (rename/delete).
  Tags included in FTS5/tsvector search index and CSV export.
- Markdown support: `widget: "textarea"` schema attributes use
  `MarkdownEditor.vue` (CodeMirror + `@codemirror/lang-markdown`) in forms
  and `MarkdownRenderer.vue` (marked + DOMPurify) in detail views.
  Global `.prose` class in `style.css` styles rendered Markdown.
  Dependencies: `@codemirror/lang-markdown`, `marked`, `dompurify`
- Auth: `auth/` package handles JWT validation (cloud) and local bypass.
  When AUTH_ISSUER_URL is set, `NewJWTMiddleware` validates Auth0 tokens,
  extracts `sub` claim, derives a collision-safe issuer-bound ID via `tenantid.Subject`,
  provisions schema on first request. When empty, `NewLocalTenantMiddleware`
  injects TENANT_ID env var directly. Frontend uses `@auth0/auth0-vue` SDK
  with AuthGuard wrapper and Bearer token injection in `api/client.ts`.
- AI metadata extraction: `ai/` package with `AIProvider` interface (Anthropic +
  OpenAI-compatible implementations). Prompt dynamically built from module schema
  via `BuildPrompt()`; response validated against schema via `ParseAndValidateResponse()`.
  Feature hidden when `AI_PROVIDER` is empty. `DynamicForm.vue` checks `/api/v1/ai/status`
  on mount, shows "Analyze with AI" button below images, fills only empty fields,
  shows title suggestion if title already has a value. No new Go or npm dependencies.
- Public showcases: `showcase/` package renders server-side HTML galleries via
  Go `html/template` (zero JS). Templates embedded via `//go:embed`. Gallery
  uses CSS `:target` for item detail overlay. Slugs: `{module-name}-{8-hex}`,
  stable across toggles. `showcases` table: SQLite main DB / PostgreSQL
  `public` schema (cross-tenant slug lookup). 24 items/page server-side
  pagination. Feature disabled in local/desktop mode (requires cloud DB).
  Route `/showcase/{slug}` registered OUTSIDE auth middleware.
- Insights dashboard: `DashboardView.vue` renders as default "All Types" landing
  page with glassmorphism summary cards (Total Value, Total Items, Most Valuable
  Item) and two Chart.js charts (doughnut: value-by-module, bar: acquisitions-over-time).
  All data computed client-side via `useDashboardMetrics` composable from existing
  collectionStore items. `showDashboard` ref in App.vue (session-only, defaults true).
  View toggle: Insights/List/Grid when "All Types" active. Charts react to theme
  changes via CSS variable reads on dark mode toggle. Doughnut groups 7+ modules
  into "Other". Dependencies: `chart.js`, `vue-chartjs`.
- Smart Folders (Saved Views): `smartFolderStore` (Pinia) persists named view
  state snapshots (module + search + filters + tags) as `smartFolders` key in
  the existing settings JSON blob. `SmartFolders.vue` sidebar section with inline
  naming (Enter to save, Escape to cancel), click-to-apply, right-click context
  menu for Rename/Delete. Active folder highlighted; clears on manual filter change.
  View mode (dashboard/list/grid) not included in saved state.
- Multi-select via `selectionStore` (Pinia): Set<string> of selected IDs,
  Shift-click range, select-all. `BulkActionBar.vue` floating bar with
  bulk delete, CSV export, module reassignment, and Compare (when exactly
  2 items selected). Bindings: `DeleteItems`, `ExportItemsCSV`,
  `BulkUpdateModule`
- Masonry grid: `CollectionGrid.vue` uses CSS `column-count` layout for
  variable-height cards based on image aspect ratio. No forced cropping.
  Frosted glass captions and hover effects preserved.
- Item comparison mode: `ComparisonView.vue` displays two selected items
  side-by-side with synchronized image galleries (shared activeImageIndex
  with per-side clamping) and a diff table showing core fields (title,
  price, tags) plus union of schema attributes. Differing values
  highlighted with amber background. Responsive: stacks vertically
  below 768px. No backend changes; client-side only.

## REST API Endpoints

| Endpoint                            | Method  | Purpose                                                      |
| ----------------------------------- | ------- | ------------------------------------------------------------ |
| `/api/v1/items`                     | GET     | List/search items with query, moduleId, filters, tags params |
| `/api/v1/items`                     | POST    | Create or update an item                                     |
| `/api/v1/items/{id}`                | DELETE  | Delete a single item                                         |
| `/api/v1/items/batch-delete`        | POST    | Atomic batch delete                                          |
| `/api/v1/items/batch-update-module` | POST    | Bulk module reassignment                                     |
| `/api/v1/tags`                      | GET     | List all tags with item counts                               |
| `/api/v1/tags/rename`               | POST    | Rename tag across all items                                  |
| `/api/v1/tags/{name}`               | DELETE  | Remove tag from all items                                    |
| `/api/v1/modules`                   | GET     | List active module schemas                                   |
| `/api/v1/modules`                   | POST    | Save custom module schema                                    |
| `/api/v1/modules/{id}/file`         | GET     | Load module schema JSON                                      |
| `/api/v1/images/upload`             | POST    | Multipart image upload + processing                          |
| `/api/v1/export/backup`             | GET     | Download backup ZIP                                          |
| `/api/v1/export/csv`                | POST    | Download CSV for selected items                              |
| `/api/v1/import/analyze`            | POST    | Upload + analyze backup ZIP (multipart)                      |
| `/api/v1/import/execute`            | POST    | Execute import (tempId + replace/merge mode)                 |
| `/api/v1/ai/analyze`                | POST    | Analyze item image with AI vision model                      |
| `/api/v1/ai/status`                 | GET     | Check AI availability (enabled/provider/model)               |
| `/api/v1/showcases`                 | GET     | List showcases for current tenant                            |
| `/api/v1/showcases/toggle`          | POST    | Toggle module public/private, returns Showcase with URL      |
| `/showcase/{slug}`                  | GET     | Public gallery page (no auth, server-rendered HTML)          |
| `/api/v1/settings`                  | GET/PUT | Load/save app settings                                       |
| `/api/v1/health`                    | GET     | Database and storage connectivity check                      |

## Cloud Configuration (Environment Variables)

| Variable        | Default                    | Description                                          |
| --------------- | -------------------------- | ---------------------------------------------------- |
| DATABASE_URL    | (empty = local SQLite)     | PostgreSQL connection string                         |
| S3_ENDPOINT     | (empty = local filesystem) | S3-compatible endpoint URL                           |
| S3_BUCKET       | (empty)                    | Bucket name for media storage                        |
| S3_ACCESS_KEY   | (empty)                    | S3 access key                                        |
| S3_SECRET_KEY   | (empty)                    | S3 secret key                                        |
| S3_REGION       | us-east-1                  | S3 region                                            |
| PORT            | 8080                       | HTTP server listen port                              |
| TENANT_ID       | default                    | PostgreSQL schema-per-tenant isolation (local mode)  |
| AUTH_DOMAIN     | (empty)                    | Auth0 tenant domain                                  |
| AUTH_AUDIENCE   | (empty)                    | Auth0 API audience identifier                        |
| AUTH_ISSUER_URL | (empty = no auth)          | Auth0 issuer URL; enables JWT auth when set          |
| AUTH_CLIENT_ID  | (empty)                    | Auth0 SPA client ID (frontend build-time)            |
| AI_PROVIDER     | (empty = disabled)         | AI provider: "anthropic" or "openai-compatible"      |
| AI_API_KEY      | (empty)                    | API key for the AI provider                          |
| AI_MODEL        | (empty)                    | Model identifier (e.g. "claude-sonnet-4-6-20250514") |
| AI_BASE_URL     | (empty)                    | Custom endpoint URL (required for openai-compatible) |

## Data Locations

### Local Mode (no env vars)
- Database: `~/Library/Application Support/OmniCollect/collection.db`
- Modules: SQLite `modules` table
- Settings: SQLite `settings` table
- Media: `~/.omnicollect/media/originals/` and `thumbnails/`

### Cloud Mode
- Database: PostgreSQL (schema-per-tenant, tsvector FTS, JSONB attributes)
- Modules: PostgreSQL `modules` table (schema_json JSONB)
- Media: S3 or local fallback under `tenants/{tenant-id}/originals/` and `thumbnails/`.
  S3 buckets must remain private; all HTTP reads go through authorization.
- Settings: PostgreSQL `settings` table

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

## Active Technologies
- Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API (006-command-palette)
- SQLite via modernc.org/sqlite (existing FTS5 full-text search) (006-command-palette)
- SQLite via modernc.org/sqlite (FTS5 full-text search, JSON attributes column) (007-faceted-filtering)
- Go 1.25+ (backend, no changes needed), TypeScript + Vue 3 (frontend) + Wails v2, Pinia, vue-codemirror (existing), CodeMirror markdown extensions (new), marked (new), DOMPurify (new) (008-markdown-textarea)
- SQLite (no changes -- raw Markdown stored as string in existing JSON attributes) (008-markdown-textarea)
- Go 1.25+ (backend -- new bindings), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API (009-bulk-actions)
- SQLite via modernc.org/sqlite (batch delete in transaction, CSV query) (009-bulk-actions)
- Go 1.25+ (backend -- HTTP server + router), TypeScript + Vue 3 (frontend) + Go `net/http` + lightweight router, Pinia (state), `fetch` API (no Axios needed) (010-rest-api-migration)
- SQLite via modernc.org/sqlite (unchanged) (010-rest-api-migration)
- Go 1.25+ (backend), TypeScript + Vue 3 (frontend -- minimal changes) + `database/sql` + `lib/pq` (PostgreSQL driver), AWS SDK v2 for Go (S3), Docker multi-stage build (011-cloud-infrastructure)
- PostgreSQL (cloud) / SQLite (local fallback); S3-compatible object store (cloud) / local filesystem (fallback) (011-cloud-infrastructure)
- Go 1.25+ (backend tests), TypeScript + Vue 3 (frontend tests) + Go `testing` + `net/http/httptest` (backend), Vitest (frontend) (012-test-coverage)
- Temporary SQLite `:memory:` databases for test isolation (012-test-coverage)
- Go 1.25+ (backend middleware), TypeScript + Vue 3 (frontend Auth0 SDK) + `github.com/auth0/go-jwt-middleware/v2` + `gopkg.in/go-jose/go-jose.v2` (Go JWT validation), `@auth0/auth0-vue` (frontend SDK) (013-jwt-auth)
- PostgreSQL schema-per-tenant (existing); tenant provisioning reuses existing PostgresStore.initTenantSchema() (013-jwt-auth)
- Go 1.25+ (backend -- Store interface + handlers), TypeScript + Vue 3 (frontend) + Existing stack (no new dependencies) (014-cross-collection-tags)
- Tags stored as JSON array on items (`tags TEXT/JSONB`); both SQLite and PostgreSQL Store implementations updated (014-cross-collection-tags)
- Go 1.25+ (backend import logic), TypeScript + Vue 3 (frontend upload + progress UI) + Go `archive/zip` (existing), existing Store and MediaStore interfaces (015-backup-import)
- Imports into whatever backend is active (SQLite or PostgreSQL, local filesystem or S3) (015-backup-import)
- Go 1.25+ (backend -- showcase package + handlers), TypeScript + Vue 3 (frontend -- toggle UI) + Go `html/template` (server-rendered galleries), `//go:embed` (binary-embedded templates) (017-public-showcase)
- Showcases table in SQLite main DB / PostgreSQL `public` schema; items queried via existing Store interface (017-public-showcase)
- Go 1.25+ (backend AI client + handler), TypeScript + Vue 3 (frontend button + form integration) + No new Go dependencies (uses `net/http` for AI API calls + `encoding/json`); no new frontend dependencies (016-ai-metadata-extraction)
- No database changes; AI results populate existing Item attributes (016-ai-metadata-extraction)
- Go 1.25+ (backend -- templates, handlers, database), TypeScript + Vue 3 (frontend -- toggle UI only) + Go `html/template` (standard library); no new frontend dependencies (017-public-showcase)
- New `showcases` table in both SQLite and PostgreSQL; both Store implementations extended (017-public-showcase)
- TypeScript 4.6+ (frontend), Go 1.25+ (backend -- no changes) + Vue 3.2+, Pinia 3.0+, Chart.js 4.x (new), vue-chartjs 5.x (new) (018-insights-dashboard)
- N/A (no backend changes; all computation is client-side from existing store data) (018-insights-dashboard)
- TypeScript 4.6+ (frontend), Go 1.25+ (backend -- settings storage only, no new endpoints) + Vue 3.2+, Pinia 3.0+ (new store: smartFolderStore) (019-smart-folders)
- Existing settings JSON blob (SQLite for local, PostgreSQL for cloud) -- Smart Folders stored as `smartFolders` key (019-smart-folders)
- Go 1.25+ (backend, no changes), TypeScript + Vue 3 (frontend) + Vue 3 Composition API, Pinia stores, existing CSS variable system (020-masonry-compare)
- N/A (no backend or database changes) (020-masonry-compare)

## Reliability Hardening (In Progress)

- Store.QueryItemPage applies validated SQL limit/offset (1–200, 0–1,000,000)
  with one lookahead row. queryItems is the shared internal implementation; legacy
  QueryItems reads at most 201 rows and rejects >200 matches with ErrPaginationRequired
  rather than returning a partial array. Legacy HTTP GET /items maps this to 422. Stable ID tie-breakers and updated_at/id indexes
  support non-search pagination. Offset pages are not a multi-request snapshot.
- Store.CollectionSummary uses one read transaction (repeatable-read on PG), never
  QueryItems. Totals are tenant-wide; module groups cap at 1,000 with truncation
  metadata. Price sums remain JSON-safe even on overflow; PG sums textual round-trip
  floats as NUMERIC to avoid intermediate float overflow/precision loss.
- read_handlers.go exposes additive GET /api/v1/items/page and /items/summary.
  Page parameters are strict; summary rejects filters. Frontend types mirror them.
  summaryStore/CollectionSummary consume independent global snapshots; sidebar counts
  use the same store. Loading/errors hide stale totals; module-group truncation is explicit.
  Initial load, acknowledged mutations/imports/recovery and explicit refresh update totals.
  Generation/abort/disposal guards apply, and account changes drop summary state.
  AuthGuard uses an explicit store-disposal list: add every new account-scoped store
  there and extend its account-switch regression; clearing Pinia state alone is insufficient.
- read_budget.go preflights snapshot items (100,000 rows / 32 MiB including 512
  bytes/row allowance), modules (1,000 / 8 MiB, 256 KiB/row), and settings (1 MiB)
  within the reading transaction. PG uses repeatable-read. Schema compatibility
  validation streams records with a 1 MiB row preflight; module writes/restores
  check resulting catalog budgets before commit. GetModules now rejects corrupt
  PostgreSQL schema JSON rather than falling back to incomplete row metadata.
- csv_read.go preflights 8 MiB selected data; buildCSV returns an error on more
  than 1,024 total columns or 8 MiB output and uses sort.Strings. Reassignment and
  recovery capture also preflight 8 MiB selections. Tags use a 10,001-row lookahead
  and reject more than 10,000 names without partial results. These limits budget
  encoded data, not exact Go heap allocation. ListShowcases also has a 10,000-record
  budget with a SQL lookahead and explicit overflow error. SQLite tag mutations are
  single atomic SQL statements, not collection-sized Go buffers. PostgreSQL rename
  uses ordinality/grouping to deduplicate destinations and preserve first-occurrence
  order, matching SQLite.
- ReadGalleryPage returns only 24 title/cover projections, name and count within one
  read transaction (PG repeatable-read). It never calls QueryItems/GetModules. A
  module/updated_at/id index supports the bounded query. Requests cap at 41,667 pages;
  the gallery discloses truncation while preserving complete total counts.
- collectionStore loads one 100-item page via /items/page, validates page shape and
  never appends pages. fetchItems() deliberately restarts at offset zero; goToPage()
  preserves scope and replaces the page. Query/mutation/refresh flows reset pagination.
  Generation/abort guards and disposal cover page metadata as well as item arrays.
- CollectionPaging supplies Previous/Next/restart controls, offset-limit disclosure
  and an honest empty-later-page state. Selection clears across pages. Client column
  sorting and page Insights are explicitly page-local; sidebar/summary counts are
  independent global snapshots. Do not derive global totals from the visible page.
- searchAllItems uses /items/page with limit 20 and no current-view filters. Palette
  discloses this cap; use collection search to browse further results. Do not reintroduce
  all-item loads or silently interpret a malformed page as an empty collection.

- ContextMenu uses menu/menuitem roles, explicit focus entry/return, roving keyboard
  focus and viewport clamping. Escape does not bubble into underlying editors.
  menuAnchor supports pointer and keyboard-opened menus from named action buttons.
- ImageLightbox uses ModalSurface; zoom/pan/reset are keyboard reachable and reset
  on filename/visibility changes. Do not infer native browser focus correctness
  from the stubbed DOM tests.
- Collection/sidebar/saved-view navigation uses native buttons rather than
  click-only labels. Selection inputs have names, list sort headers expose
  aria-sort, and focus-within reveals hover actions. Saved-view active state only
  changes after App's navigation guard consents.

- Smart Folder baselineReady defaults false. Only a validated settings response
  enables writes; invalidate before reloading. Missing smartFolders means empty;
  malformed arrays/entries are an error, not permission to overwrite stored views.
  App exposes settings retry, gates appearance editing and reloads settings after
  import. Import entry waits for outstanding local settings writes.
- Collection/module stores track successful loading separately from empty data.
  App hides unavailable rows, selection actions and dashboard totals, but keeps
  list search available. Retries do not turn errors into empty-success screens.
- Dashboard/sidebar metrics describe loaded query results, not global totals.
  Price coverage distinguishes missing/invalid/zero; sum overflow is unavailable.
  The timeline is record creation month, not purchase/acquisition date.
  Pagination and independent global aggregates are still outstanding.

- main.ts places AuthGuard OUTSIDE App setup. The guard keys the whole App by
  session epoch, disposes private Pinia stores and clears state before remounting.
  Include any future private store in the guard's disposal list.
- auth/session.ts invalidates pending token/fetch/body work on principal changes.
  API perform combines caller/session cancellation and checks before downloads.
  Do not treat cancellation as proof of rollback.
- Store scope cleanup clears private data and timers; Smart Folder queued writes
  must check disposal before invoking the API. App initial settings/tag callbacks
  check their captured session before applying global theme or starting more work.
- Explicit sign-out suspends the session immediately after draft consent, preventing
  competing automatic login redirects; login/logout failures have visible recovery.

- Frontend toolchain: Node >=24, Vite 7/plugin-vue 6, vue-tsc 3, TypeScript 5.9.
  Lockfile installs with npm ci --ignore-scripts. npm audit is clear at this
  checkpoint; audit results are time-dependent, not a permanent guarantee.
- App uses lazyComponent for expensive views, with loading/error UI. Unloaded
  editor refs do not block navigation. vite.config.ts splits editor core/languages,
  charts and auth chunks, retains ES2020 output and generates a manifest.
  npm run build enforces 500,000-byte per-chunk and static-entry budgets via
  scripts/check-bundle.mjs; preserve dist/.vite/manifest.json for reruns.
- Go selects toolchain 1.26.8; x/image 0.45 and x/crypto 0.56 fix audited issues.
  govulncheck v1.8.0 reports zero reachable/package findings on darwin/arm64.
  Module-only GO-2026-5932 concerns unimported x/crypto/openpgp; audit evidence
  and this residual qualification are preserved under the reliability spec.

- Scope and verification live in `specs/021-reliability-hardening/`.
- In-memory SQLite stores are fully filesystem-free; modules and settings are
  stored in their database. Use `NewSQLiteStoreAt` for isolated file-backed tests.
  Never point tests at user data. PostgreSQL integration tests require
  `OMNICOLLECT_TEST_POSTGRES_URL` with a database name ending in `_test`.
- `Store.Snapshot(ctx)` reads one consistent items/modules/settings state;
  `Store.Restore(ctx, snapshot, mode)` preserves IDs/timestamps and commits all
  metadata atomically. Replace also disables existing showcases. HTTP restore
  stages validated originals under content-addressed names first, regenerates
  thumbnails, and retains handles for retries. Never delete failed staged content
  immediately: a concurrent import/upload could already reference it.
- Both backends export portable ZIPs using one DB snapshot plus referenced media.
  No live SQLite file copying or global-directory backup reads. Export checks
  close errors and atomically publishes the final archive. Limits: 256 MB
  compressed, 1 GB expanded, 10,000 entries. Import checks paths, duplicates,
  symlinks, every file CRC, versions, and actual expanded bytes. JSON metadata:
  32 MB per entry; settings: 1 MB. Legacy SQLite and JSON imports are supported.
- MediaStore reads AND writes take context; derive cloud namespaces with ForTenant.
  Never use the unscoped app media store from tenant handlers (including AI/backups).
  mediahttp serves sniffed image MIME with no-store/nosniff. Public gallery media
  requires an enabled showcase and a current cover-image reference.
- Image processing: 30 MB, 24 MP, two decode workers; thumbnails fit 400×400.
  Original names are content hashes with correct format extensions; thumbnail
  bytes are JPEG even when sharing an original's filename. AI MIME comes from
  actual bytes; provider clients have 60-second deadlines and 2 MB response limits.
- Public galleries expose only titles and cover images by default, not tags,
  purchase prices, additional images, or arbitrary schema attributes.
  Sharing disclosure/field controls are still pending frontend work.
- `tenantid.Local` and `tenantid.Subject` use namespaced SHA-256/base32 identities.
  JWT mode requires PostgreSQL. There is no automatic legacy tenant migration.
- PostgreSQL queries must use `table()` qualification. Provisioning uses a scoped
  store and a transaction-local search path protected by an advisory lock.
- Validate filenames at storage boundaries. Import IDs are opaque, tenant-owned,
  expiring handles, not filesystem paths.
- Store.WithContext(ctx) returns an immutable borrowed handle; its Close is a no-op.
  Use requestStore(r) in HTTP handlers. SQL operations, auth provisioning, health
  and public galleries inherit request cancellation. Standard context deadline:
  15s; archive/upload: 3m; AI: 90s. Explicit-context SaveItem/Snapshot/Restore
  continue taking their caller's context; startup/CLI work is not request-scoped.
- Batch mutation/export IDs are unique, nonblank, <=128 bytes, maximum 500.
  Reassignment locks the target schema and selected items and validates all before
  updating. Schema edits validate existing items; restore validates incoming items
  and merge validates retained items against incoming schemas. PostgreSQL metadata
  locks always acquire modules before items to match SaveItem/reassignment.
- ValidateQuery rejects malformed tag JSON instead of silently broadening a query.
  JSON request bodies accept one value only. Schemas are capped at 256 KB and
  settings at 1 MB including cumulative merges. Tag normalization never byte-truncates
  Unicode; validate normalized tags before writing.
- Metadata restore failures must not promise unchanged data: Commit may have
  succeeded despite a lost acknowledgement. The HTTP round-trip test injects
  exactly this outcome. Media staging failures precede metadata mutation.
- useShowcaseControls is shared by AppSidebar/ModuleSelector: explicit publication
  consent covers titles, cover originals/embedded metadata, future changes and
  inability to recall downloaded copies. Toggle errors disable controls until
  reload; clipboard feedback waits for writeText success.
- Reduced motion: global CSS bounds animation/transition duration and removes
  delays; public-gallery CSS disables transitions; useReducedMotion supplies
  reactive Chart.js animation settings with listener cleanup.
- CSV quoting does not neutralize formulas. csvRow prefixes risky headers/cells
  with an apostrophe, including whitespace/control/full-width operator prefixes.
- Use Store.SaveItem(ctx, item) for application item saves. It validates the stored
  module schema inside a transaction and compares updatedAt on UPDATE. HTTP 409
  means stale/deleted, 428 means missing version; retain the user's draft.
  Item saves have a 15-second deadline and tenant-scoped original existence checks.
  Legacy InsertItem/UpdateItem helpers remain for lower-level fixture/migration
  work; they are not the authoritative API save path. Other CRUD context and
  validation coverage is still pending.
- Never truncate edit timestamps to seconds: PostgreSQL QueryItems now returns
  RFC3339Nano. Tag rename/delete updates item timestamps to invalidate old drafts.
- DynamicForm uses createItemDraft to deep-clone unknown attributes, retain the
  edit version, and clear keys when changing drafts. App uses keyed editor
  instances, a shared discard policy, and a beforeunload guard. Async saves are
  awaited; uploads/AI block save and cancel on unmount or draft replacement.
- Search text and debounce timers belong to collectionStore, not ItemList.
  Changing queries immediately invalidates prior requests. Fetches prune selection;
  scope changes clear it. Shift-range anchors use item IDs across sorting.
- CommandPalette cancels/invalidate searches on edit/close/unmount, distinguishes
  loading/error/empty states, and stops Escape from closing an underlying editor.
- useEditorGuard protects schema/settings navigation and beforeunload. Parent
  App.canLeaveDraft consults exposed canLeave methods; child close buttons guard
  themselves. Key schema/settings instances when reopening so accepted discard
  cannot reuse a stale draft. The native smoke covers an unsaved schema field's
  close confirmation; broader native workflows remain unverified.
- SchemaBuilder parses current codeContent synchronously before Save, clears pending
  preview timers, validates JSON shape before rendering and blocks duplicate writes.
  SettingsPage saves immutable snapshots and rolls back cancelled live theme previews.
- DeleteConfirmation is shared by single/bulk item deletion. Snapshot target IDs
  at mount, prevent duplicate writes/close/unload while busy, retain errors, and
  never infer rollback from a failed response. Parent refresh failure after deletion
  is reported separately. Acknowledged and uncertain deletions cannot repeat in
  the same dialog.
- HTTP deletion uses Store.DeleteWithRecovery, not the legacy trusted permanent
  DeleteItem/DeleteItems helpers. Capture and deletion commit together in tenant-local
  deletion_batches. RecoveryDialog lists bounded batch metadata and requires separate
  consent for restore (public-showcase reappearance) or permanent discard.
- Recovery keeps IDs/creation times and metadata, regenerates edit versions, validates
  current schemas and inserts without upsert. Any ID conflict rolls back the batch;
  only successful recovery consumes it. PG lock order is modules, items, deletion_batches;
  SQLite reserves the writer before reading. Context cancellation and tenant scoping apply.
- Recovery never expires automatically. Caps: 500 items / 8 MiB record payload per
  batch, 100 batches / 64 MiB record-plus-title-preview payload per tenant. Capacity
  rejects deletion, never evicts old batches. Explicit discard removes only recovery
  records, not media files. Future media GC must protect references in recovery payloads.
- GET /api/v1/recovery lists batches; POST /api/v1/recovery/{id} restores atomically;
  DELETE /api/v1/recovery/{id} permanently discards a batch. Single-item DELETE stays
  204 with X-Recovery-ID; batch deletion returns deleted/recoveryId/createdAt/titles.
  ZIP Snapshot exports exclude recovery; Replace imports clear it in the transaction.
  Read/write acknowledgements can be lost: refresh both collection and recovery list,
  and never promise rollback. DOM tests are not browser/native verification.
- ActionConfirmation snapshots mutation intent for tag rename/delete, saved-view
  deletion and bulk reassignment. Pending writes block close/unload/duplicates;
  errors disable repeat operations until reload. Mutation callbacks perform only
  the write; App refreshes separately after acknowledgement. Never report a failed
  refresh as a failed write. Tags retain rename drafts until success and disclose
  collection-wide scope and merging. Smart Folder remove/removePersisted waits for
  persistence, preserves the local view on uncertain errors, and locks further
  settings writes until reload. Bulk moves capture IDs before destination selection.
- ModalSurface uses native dialog.showModal and restores focus on removal. After
  mount/busy rendering, it repairs focus lost to disabled/removed controls without
  stealing valid child focus. Chrome can briefly retain a disabled activeElement
  before its later focus-fixup task: check :disabled, not just dialog.contains.
  Explicit first/last Tab boundaries keep sequential navigation inside the modal.
  Unit tests stub native APIs; browser-smoke.mjs covers the real recovery workflow.
- frontend/scripts/browser-smoke.mjs builds/serves with an isolated env allowlist,
  fresh HOME/data/profile, and Chrome CDP. External page requests are blocked (font
  CSS uses fallback fonts). It retains evidence and kills only its own processes.
  Iteration 20 passes 33 Chrome assertions, including paused Markdown saves. Bring
  the CDP page to front and verify visibility; never copy Windows virtual key codes
  into nativeVirtualKeyCode on macOS. Waits have wall-clock deadlines, and failure
  diagnostics sample only owned Chrome processes. Main content is 390px at a 390px
  viewport; mobile navigation is a native expanded-state disclosure. Screen readers,
  live Auth0, other browsers and native desktop runtime are not covered.
- ItemDetail delegates deletion directly to the parent recovery-aware confirmation.
  Detail/comparison images are named buttons; list collection filters have labels
  and reflect activeModuleId. Narrow bulk controls wrap without obscuring the end
  of content. SchemaVisualEditor uses modelValue/update:modelValue (not mutable
  list plus a second manual reorder), stable WeakMap/toRaw field keys across edits,
  keyboard move buttons and focus on the moved field. Never key editing rows by
  their editable names or serialize temporary UI keys into schemas.
- DynamicForm locks uncertain save retries while retaining draft inputs, preventing
  duplicate creates after lost acknowledgement. Explicit known HTTP rejections
  permit corrections; timeouts/unknown/server errors do not prove rollback.
- macOS arm64 packaging passes with the installed 15.4 SDK selected per command:
  SDKROOT=/Library/Developer/CommandLineTools/SDKs/MacOSX15.4.sdk wails build -s
  -skipbindings -m -nosyncgomod -o omnicollect-hardening-20. The default 27.0 SDK
  failed with linker/TAPI arm64e.x1 errors. Artifact: build/bin/omnicollect.app;
  ad-hoc signature verifies. Native smoke launches only a copied fixture bundle,
  not installed app data. No machine-wide SDK configuration was changed.
- main.go wires OnBeforeClose through desktop_close.go. Always ask native No/Yes,
  default/cancel No; dialog errors, cancelled context and concurrent pending close
  requests prevent closure. Do not trust a stale frontend dirty flag or assume
  beforeunload runs in Wails. Yes accepts possible lost drafts/interruption, not
  rollback. This asks even for clean windows; policy tests are not native UI proof.
- frontend/scripts/native-smoke.mjs and native-smoke.swift provide macOS-only
  PID-scoped Accessibility verification. Require existing Accessibility/screen
  capture permissions and a logged-in GUI; never request permissions silently.
  Copy/re-sign the built bundle with a unique org.omnicollect.smoke identity, use
  fresh HOME/CFFIXED_USER_HOME/config/temp paths, and verify Foundation paths before
  launch. Capture only owned windows; retain all evidence and fixture data.
  Seven checks cover native rendering, same-origin summary, default No, unsaved
  schema-field retention after No and normal Yes exit. Leave dialogs untouched:
  operator interaction previously produced misleading premature-exit failures.
  AX press can lose its acknowledgement while opening a modal (AXCannotComplete);
  record that uncertainty and verify the postcondition without blindly retrying.
  Native text input/keyboard shortcuts, concurrent close and destructive flows,
  screen readers and live Auth0 are not established. External fonts are not blocked.
- The resumed darwin/arm64 govulncheck v1.8.0 scan with desktop,production tags
  reports zero reachable/imported-package vulnerabilities; unimported module-only
  GO-2026-5932 remains. Do not generalize this to other platforms or build tags.
- FormField assigns instance-unique label/control/error IDs and aria associations.
  MarkdownEditor applies names/errors to CodeMirror's editable surface; disabled
  state is explicit because a disabled fieldset cannot disable contenteditable.
  DynamicForm passes submitting through FormField; schema previews disable all
  fields. Keep these props wired when adding alternative widgets.
- ImportDialog remains mounted after success, blocks duplicate execution/dismissal
  while busy, requires replace consent, and distinguishes uncertain network
  outcomes. View-refresh retry never repeats executeImport.
- Settings updates merge top-level keys; successful PUT returns HTTP 204.
- Desktop routes REST through the Wails asset handler, without a TCP listener
  or parallel unscoped Wails App bindings.
  Standalone defaults to loopback; `HOST` enables explicit network binding and
  `CORS_ALLOWED_ORIGINS` lists allowed cross-origin frontends.
- The user confirmed pre-release status with no existing data to preserve.
  Secure storage-layout changes are allowed; legacy backup import compatibility
  remains required. Do not run migrations against real user directories.

## Recent Changes
- 017-public-showcase: Added public showcase URLs with server-rendered HTML galleries (zero JS), CSS :target detail overlay, stable slug generation, toggle public/private from ModuleSelector, showcases table in SQLite/PostgreSQL, 24-item pagination, cloud-mode-only feature
- 006-command-palette: Added Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
