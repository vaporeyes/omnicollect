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
config.go        # Environment-based config (DATABASE_URL, S3_*, PORT, TENANT_ID)
server.go        # HTTP server setup, routing, CORS middleware
handlers.go      # REST endpoint handlers wrapping App methods
app.go           # Core business logic with Store/MediaStore interfaces
db.go            # Legacy helpers (dbFilePath for backup)
imaging.go       # Image validation, thumbnail generation (returns bytes)
backup.go        # ZIP archive export (database + media + modules)
modules.go       # Legacy helpers (modulesDir for backup)
settings.go      # Wails-bound settings methods delegating to Store
models.go        # Shared Go types (Item, ModuleSchema, ProcessImageResult)
Dockerfile       # Multi-stage build (Go + Node -> alpine)
docker-compose.yml # Dev stack (app + postgres + minio)
storage/
  db.go          # Store interface (database abstraction)
  media.go       # MediaStore interface (object storage abstraction)
  sqlite.go      # SQLiteStore: local SQLite with FTS5
  sqlite_test.go # Storage layer unit tests (in-memory SQLite)
  postgres.go    # PostgresStore: PostgreSQL with schema-per-tenant, tsvector
  local.go       # LocalMediaStore: local filesystem
  s3.go          # S3MediaStore: S3-compatible object store
  migrate.go     # SQLite-to-PostgreSQL migration tool
  testdata/      # Test fixtures (test-module.json, test-image.jpg)
frontend/src/
  api/
    client.ts    # Centralized fetch-based HTTP client (base URL, typed helpers)
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
                 #   BulkActionBar
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
- `GetItems(query, moduleID, filtersJSON)` accepts JSON filter payload;
  backend uses `json_extract()` for attribute filters, direct column for
  `purchasePrice`; `collectionStore.activeFilters` manages filter state
- Markdown support: `widget: "textarea"` schema attributes use
  `MarkdownEditor.vue` (CodeMirror + `@codemirror/lang-markdown`) in forms
  and `MarkdownRenderer.vue` (marked + DOMPurify) in detail views.
  Global `.prose` class in `style.css` styles rendered Markdown.
  Dependencies: `@codemirror/lang-markdown`, `marked`, `dompurify`
- Multi-select via `selectionStore` (Pinia): Set<string> of selected IDs,
  Shift-click range, select-all. `BulkActionBar.vue` floating bar with
  bulk delete, CSV export, module reassignment. Bindings: `DeleteItems`,
  `ExportItemsCSV`, `BulkUpdateModule`

## REST API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/items` | GET | List/search items with query, moduleId, filters params |
| `/api/v1/items` | POST | Create or update an item |
| `/api/v1/items/{id}` | DELETE | Delete a single item |
| `/api/v1/items/batch-delete` | POST | Atomic batch delete |
| `/api/v1/items/batch-update-module` | POST | Bulk module reassignment |
| `/api/v1/modules` | GET | List active module schemas |
| `/api/v1/modules` | POST | Save custom module schema |
| `/api/v1/modules/{id}/file` | GET | Load module schema JSON |
| `/api/v1/images/upload` | POST | Multipart image upload + processing |
| `/api/v1/export/backup` | GET | Download backup ZIP |
| `/api/v1/export/csv` | POST | Download CSV for selected items |
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
| TENANT_ID | default | PostgreSQL schema-per-tenant isolation |

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

## Recent Changes
- 006-command-palette: Added Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
