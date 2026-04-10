# OmniCollect Development Guidelines

## Project Overview

Schema-driven desktop collection manager. Go backend + Vue 3 frontend
via Wails v2. See [Constitution](.specify/memory/constitution.md) for
immutable engineering principles.

## Tech Stack

- **Go 1.25+**: Backend (SQLite, image processing, backup, Wails bindings)
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
                 #   toastStore
  components/    # DynamicForm, FormField, ItemList, CollectionGrid,
                 #   ModuleSelector, ImageAttach, ImageLightbox,
                 #   SchemaBuilder, SchemaVisualEditor,
                 #   SchemaCodeEditor, SchemaFormPreview,
                 #   ItemDetail, SettingsPage, ToastProvider,
                 #   ContextMenu, CommandPalette, FilterBar,
                 #   MarkdownEditor, MarkdownRenderer,
                 #   BulkActionBar, TagInput, TagFilter,
                 #   TagManager, ImportDialog,
                 #   DashboardView, DashboardMetricCard
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
- Media served via HTTP server at /thumbnails/ and /originals/
- Two modes: `--serve` for standalone HTTP, default for Wails desktop shell
- Grid views MUST use thumbnails only (Constitution Principle IV)
- No hardcoded collection-type templates (Constitution Principle II)
- README and CLAUDE.md MUST be updated every iteration (Principle VI)
- Module schemas in `~/.omnicollect/modules/*.json`
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
  extracts `sub` claim, sanitizes to tenant ID (`auth0|abc` -> `tenant_auth0_abc`),
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
- Multi-select via `selectionStore` (Pinia): Set<string> of selected IDs,
  Shift-click range, select-all. `BulkActionBar.vue` floating bar with
  bulk delete, CSV export, module reassignment. Bindings: `DeleteItems`,
  `ExportItemsCSV`, `BulkUpdateModule`

## REST API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/items` | GET | List/search items with query, moduleId, filters, tags params |
| `/api/v1/items` | POST | Create or update an item |
| `/api/v1/items/{id}` | DELETE | Delete a single item |
| `/api/v1/items/batch-delete` | POST | Atomic batch delete |
| `/api/v1/items/batch-update-module` | POST | Bulk module reassignment |
| `/api/v1/tags` | GET | List all tags with item counts |
| `/api/v1/tags/rename` | POST | Rename tag across all items |
| `/api/v1/tags/{name}` | DELETE | Remove tag from all items |
| `/api/v1/modules` | GET | List active module schemas |
| `/api/v1/modules` | POST | Save custom module schema |
| `/api/v1/modules/{id}/file` | GET | Load module schema JSON |
| `/api/v1/images/upload` | POST | Multipart image upload + processing |
| `/api/v1/export/backup` | GET | Download backup ZIP |
| `/api/v1/export/csv` | POST | Download CSV for selected items |
| `/api/v1/import/analyze` | POST | Upload + analyze backup ZIP (multipart) |
| `/api/v1/import/execute` | POST | Execute import (tempId + replace/merge mode) |
| `/api/v1/ai/analyze` | POST | Analyze item image with AI vision model |
| `/api/v1/ai/status` | GET | Check AI availability (enabled/provider/model) |
| `/api/v1/showcases` | GET | List showcases for current tenant |
| `/api/v1/showcases/toggle` | POST | Toggle module public/private, returns Showcase with URL |
| `/showcase/{slug}` | GET | Public gallery page (no auth, server-rendered HTML) |
| `/api/v1/settings` | GET/PUT | Load/save app settings |
| `/api/v1/health` | GET | Database and storage connectivity check |

## Cloud Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | (empty = local SQLite) | PostgreSQL connection string |
| S3_ENDPOINT | (empty = local filesystem) | S3-compatible endpoint URL |
| S3_BUCKET | (empty) | Bucket name for media storage |
| S3_ACCESS_KEY | (empty) | S3 access key |
| S3_SECRET_KEY | (empty) | S3 secret key |
| S3_REGION | us-east-1 | S3 region |
| PORT | 8080 | HTTP server listen port |
| TENANT_ID | default | PostgreSQL schema-per-tenant isolation (local mode) |
| AUTH_DOMAIN | (empty) | Auth0 tenant domain |
| AUTH_AUDIENCE | (empty) | Auth0 API audience identifier |
| AUTH_ISSUER_URL | (empty = no auth) | Auth0 issuer URL; enables JWT auth when set |
| AUTH_CLIENT_ID | (empty) | Auth0 SPA client ID (frontend build-time) |
| AI_PROVIDER | (empty = disabled) | AI provider: "anthropic" or "openai-compatible" |
| AI_API_KEY | (empty) | API key for the AI provider |
| AI_MODEL | (empty) | Model identifier (e.g. "claude-sonnet-4-6-20250514") |
| AI_BASE_URL | (empty) | Custom endpoint URL (required for openai-compatible) |

## Data Locations

### Local Mode (no env vars)
- Database: `~/Library/Application Support/OmniCollect/collection.db`
- Modules: `~/.omnicollect/modules/`
- Media: `~/.omnicollect/media/originals/` and `thumbnails/`

### Cloud Mode
- Database: PostgreSQL (schema-per-tenant, tsvector FTS, JSONB attributes)
- Modules: PostgreSQL `modules` table (schema_json JSONB)
- Media: S3-compatible object store (originals/ and thumbnails/ prefixes)
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

## Recent Changes
- 017-public-showcase: Added public showcase URLs with server-rendered HTML galleries (zero JS), CSS :target detail overlay, stable slug generation, toggle public/private from ModuleSelector, showcases table in SQLite/PostgreSQL, 24-item pagination, cloud-mode-only feature
- 006-command-palette: Added Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
