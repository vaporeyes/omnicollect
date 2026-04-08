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
main.go          # Entry point: --serve for standalone HTTP, default for Wails desktop
server.go        # HTTP server setup, routing, CORS middleware
handlers.go      # REST endpoint handlers wrapping App methods
app.go           # Core business logic: SaveItem, GetItems, etc.
db.go            # SQLite init, DDL, FTS5 triggers, CRUD helpers
imaging.go       # Image validation, copy, thumbnail generation
backup.go        # ZIP archive export (database + media + modules)
modules.go       # Module schema loader, save, find helpers
models.go        # Shared Go types (Item, ModuleSchema, ProcessImageResult)
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
go vet ./...     # Lint Go code
go mod tidy      # Resolve dependencies
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

## Data Locations

- Database: `~/Library/Application Support/OmniCollect/collection.db`
- Modules: `~/.omnicollect/modules/`
- Media: `~/.omnicollect/media/originals/` and `thumbnails/`

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

## Recent Changes
- 006-command-palette: Added Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
