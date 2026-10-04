# OmniCollect

A schema-driven desktop collection manager built with Go, Vue 3, and
Wails. Track any type of collection (coins, books, stamps, etc.) by
creating or importing JSON schemas through the Schema Builder. No code changes
are needed to add new collection types.

## Architecture

- **Backend**: Go with SQLite (local) or PostgreSQL (cloud)
- **Frontend**: Vue 3 (Composition API, TypeScript) with Pinia stores
- **Desktop Shell**: Wails v2 (embeds frontend in native webview)
- **Image Processing**: `disintegration/imaging` for thumbnail generation
- **Cloud Storage**: S3-compatible object store (AWS S3, MinIO, R2)
- **Containerization**: Docker multi-stage build, docker-compose for dev

## Core Principles

See [Constitution](.specify/memory/constitution.md) for the full set.
Key rules:

1. **Local-First**: SQLite is the source of truth in local mode. Cloud mode uses PostgreSQL + S3.
2. **Schema-Driven UI**: Forms are generated at runtime from JSON
   schemas. No hardcoded collection-type templates.
3. **Flat Data Architecture**: Single `items` table with JSON
   `attributes` blob. No JOINs for item data.
4. **Performance Protection**: Grid views use compressed thumbnails
   only. Full-resolution images load on demand.
5. **Typed API**: TypeScript interfaces mirror Go REST payloads;
   desktop and standalone modes use the same HTTP client.
6. **Documentation is Paramount**: README, CLAUDE.md, and spec
   artifacts MUST be updated with every iteration.

## Reliability hardening

The ongoing hardening work is tracked in
[the remediation checklist](specs/021-reliability-hardening/tasks.md).
Local modules and settings now live in SQLite alongside items, allowing atomic
metadata restore. In-memory test stores are entirely filesystem-free.
PostgreSQL operations qualify tenant tables explicitly rather than relying on
connection-pool search paths. Tenant names use a full SHA-256 digest encoded
within PostgreSQL's identifier limit, with separate local and issuer-bound JWT
namespaces. This pre-release layout does not migrate old tenant schemas.
Import uploads use tenant-owned opaque handles.

Standalone servers bind to loopback by default. Set `HOST=0.0.0.0` deliberately
for network access; Docker sets this automatically. Cross-origin frontends must
be listed in the comma-separated `CORS_ALLOWED_ORIGINS` environment variable.
Desktop REST requests use the Wails asset handler without a separate TCP port.
Settings updates preserve unrelated top-level keys on both database backends,
with a 1 MB limit on the merged result. Module schemas are limited to 256 KB.
Request cancellation reaches SQL reads/writes and provisioning through immutable
store copies. Standard requests have a 15-second context deadline; upload/import/
backup requests allow three minutes and AI requests allow 90 seconds.

Batch delete, reassignment and CSV export accept at most 500 distinct IDs.
Reassignment requires an existing target schema and rejects the entire batch if
an item is missing or incompatible. Schema edits cannot invalidate existing items:
add new required fields as optional first, populate values, then make them required.
Restore validates items against the resulting schemas; merge also protects retained
items from incompatible schema changes.

CSV export prefixes formula-looking headers and cells with an apostrophe, including
formulas hidden behind whitespace or control characters. This may be visible in
plain CSV readers; stored values are unchanged. Use ZIP backup for lossless transfer.

## Dependency and bundle checks

Use Node.js 24+ and install frontend packages with `npm ci --ignore-scripts`.
The lockfile uses patched dependencies and Vite 7, Vue tooling 3 and TypeScript 5.9.
At the hardening checkpoint, `npm audit` reported zero vulnerabilities.

`npm run build` includes `check:bundle`: every JavaScript chunk and the static entry
dependency graph must stay below 500,000 bytes. Editors, detail/settings views,
comparison and charts load lazily with visible loading/error feedback. Current
static entry JS totals about 433 kB (not including views loaded on demand); the
largest chunk is about 431 kB. The default Insights view still loads its chart
chunk when displayed. These are build sizes, not browser performance measurements.

Go image and crypto dependencies are patched. Run
`go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` to repeat the host audit.
The current scan reports no reachable or imported-package vulnerabilities.
Its remaining module-only OpenPGP advisory concerns a package the app does not
import; this is not a claim that every package in the dependency graph is safe.
Audit evidence is saved in `specs/021-reliability-hardening/`.

## Prerequisites

- Go 1.26+ (the module selects the verified Go 1.26.8 toolchain)
- Node.js 24+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Quick Start

```bash
# Install dependencies
go mod tidy
cd frontend && npm install && cd ..

# Desktop mode (Wails hot reload)
wails dev

# Standalone HTTP server mode
go run . --serve                # localhost:8080
go run . --serve --port 3001    # custom port

# Build production desktop binary
wails build
```

The production binary is at `build/bin/omnicollect.app` (macOS).

### Standalone Server Mode

Run `go run . --serve` to start OmniCollect as a standalone HTTP server.
The REST API is available at `http://localhost:8080/api/v1/` and the
frontend is served at the root. No Wails desktop shell required.

### Cloud Deployment

Set environment variables to enable cloud backends:

```bash
# PostgreSQL + S3 (cloud mode)
export DATABASE_URL=postgres://user:pass@host:5432/omnicollect
export S3_ENDPOINT=https://s3.amazonaws.com
export S3_BUCKET=omnicollect-media
export S3_ACCESS_KEY=AKIA...
export S3_SECRET_KEY=secret...
go run . --serve
```

Or use Docker:

```bash
docker build -t omnicollect .
docker-compose up   # Runs app + PostgreSQL + MinIO
```

### Authentication (Auth0 JWT)

OmniCollect supports optional Auth0 JWT authentication for multi-tenant
cloud deployments. When configured, all API endpoints (except health)
require a valid Bearer token. Each user gets an isolated PostgreSQL
schema provisioned automatically on first login.

**Local mode** (default, no auth): Leave `AUTH_ISSUER_URL` empty. All
endpoints work without tokens using the `TENANT_ID` env var.

**Cloud mode** (Auth0 enabled): Set the auth env vars and configure
Auth0 dashboard (API + SPA application).

```bash
# Backend env vars
export AUTH_DOMAIN=your-tenant.us.auth0.com
export AUTH_AUDIENCE=https://api.your-app.com
export AUTH_ISSUER_URL=https://your-tenant.us.auth0.com/
export AUTH_CLIENT_ID=your-spa-client-id

# Frontend build-time env vars (in frontend/.env or as build args)
# VITE_AUTH0_DOMAIN, VITE_AUTH0_CLIENT_ID, VITE_AUTH0_AUDIENCE
```

| Variable        | Default           | Description                            |
| --------------- | ----------------- | -------------------------------------- |
| AUTH_DOMAIN     | (empty)           | Auth0 tenant domain                    |
| AUTH_AUDIENCE   | (empty)           | Auth0 API audience identifier          |
| AUTH_ISSUER_URL | (empty = no auth) | Token issuer URL; enables JWT when set |
| AUTH_CLIENT_ID  | (empty)           | Auth0 SPA client ID                    |

Auth0 dashboard setup:
1. Create an **API** with identifier = AUTH_AUDIENCE
2. Create a **Single Page Application** with callback/logout URLs
3. Enable **Authorization Code Flow with PKCE**
4. Configure allowed origins for CORS

### AI Metadata Extraction

OmniCollect can analyze photos of collection items using AI vision models
and auto-fill form fields. Set the AI environment variables to enable:

```bash
# Anthropic direct
export AI_PROVIDER=anthropic
export AI_API_KEY=sk-ant-...
export AI_MODEL=claude-sonnet-4-6-20250514

# Or OpenRouter (OpenAI-compatible)
export AI_PROVIDER=openai-compatible
export AI_BASE_URL=https://openrouter.ai/api/v1
export AI_API_KEY=sk-or-...
export AI_MODEL=anthropic/claude-sonnet-4.6
```

| Variable    | Default            | Description                                          |
| ----------- | ------------------ | ---------------------------------------------------- |
| AI_PROVIDER | (empty = disabled) | "anthropic" or "openai-compatible"                   |
| AI_API_KEY  | (empty)            | API key for the provider                             |
| AI_MODEL    | (empty)            | Model identifier                                     |
| AI_BASE_URL | (empty)            | Custom endpoint URL (required for openai-compatible) |

When enabled, an "Analyze with AI" button appears in the item form below
the image attachment section. It sends the primary image to the AI model
with a prompt built from the active module schema, then populates empty
fields with the AI response. Fields that already have values are preserved.

### Data Migration (SQLite to PostgreSQL)

```bash
export DATABASE_URL=postgres://user:pass@host:5432/omnicollect
go run . --migrate --source ~/.omnicollect/collection.db --tenant default
```

## Adding a Collection Type

Use the built-in **Schema Builder** (click "+ New Schema" in the
sidebar) to create collection types visually with a live form preview.
Or paste a JSON schema into the builder's code editor (or submit it to
`POST /api/v1/modules`). Schemas are persisted in the database:

```json
{
  "id": "coins",
  "displayName": "Coins",
  "description": "Coin collection",
  "attributes": [
    {
      "name": "year",
      "type": "number",
      "required": true,
      "display": { "label": "Mint Year", "widget": "text" }
    },
    {
      "name": "country",
      "type": "string",
      "required": true,
      "display": { "label": "Country of Origin" }
    },
    {
      "name": "condition",
      "type": "enum",
      "options": ["Poor", "Fair", "Good", "Fine", "Very Fine", "Uncirculated"],
      "display": { "label": "Condition", "widget": "dropdown" }
    }
  ]
}
```

Restart the app (or save via the Schema Builder for instant reload).
The new collection type appears in the sidebar.

### Supported Attribute Types

| Type      | Input Control | Notes                    |
| --------- | ------------- | ------------------------ |
| `string`  | Text input    | Default widget           |
| `number`  | Number input  |                          |
| `boolean` | Checkbox      |                          |
| `date`    | Date picker   |                          |
| `enum`    | Dropdown      | Requires `options` array |

### Display Hints

Each attribute can include a `display` object:
- `label`: Override the display name
- `placeholder`: Input placeholder text
- `widget`: Force a specific control (e.g., `"textarea"` for Markdown editor)
- `group`: Group attributes into form sections
- `order`: Sort priority within a group

## Project Structure

```
omnicollect/
  main.go              # Entry point: --serve, --migrate, or Wails desktop
  config.go            # Environment-based config (DATABASE_URL, S3_*, AUTH_*, AI_*)
  app.go               # Core business logic with Store/MediaStore/AIProvider
  ai/
    provider.go        # AIProvider interface + factory function
    anthropic.go       # Anthropic Messages API client
    openai_compat.go   # OpenAI-compatible client (OpenRouter, Google, etc.)
    prompt.go          # Schema-to-prompt builder + response validator
  auth/
    context.go         # Tenant ID context helpers and sanitization
    middleware.go      # JWT validation middleware with JWKS caching
    local.go           # Local-mode middleware (no auth bypass)
  db.go                # Legacy helpers (dbFilePath for backup)
  imaging.go           # Image validation, thumbnail generation (returns bytes)
  backup.go            # ZIP archive export (database + media + modules)
  import.go            # ZIP backup import: format detection, Replace/Merge modes
  modules.go           # Legacy helpers (modulesDir for backup)
  settings.go          # Settings methods delegating to Store
  models.go            # Shared types (Item, ModuleSchema, etc.)
  Dockerfile           # Multi-stage build (Go + Node -> alpine)
  docker-compose.yml   # Dev stack (app + postgres + minio)
  storage/
    db.go              # Store interface (database abstraction)
    media.go           # MediaStore interface (storage abstraction)
    sqlite.go          # SQLiteStore: local SQLite with FTS5
    postgres.go        # PostgresStore: PostgreSQL schema-per-tenant
    local.go           # LocalMediaStore: local filesystem
    s3.go              # S3MediaStore: S3-compatible object store
    migrate.go         # SQLite-to-PostgreSQL migration tool
  wails.json           # Wails project config
  frontend/
    src/
      main.ts          # Vue app entry, Pinia setup
      App.vue          # Root layout: sidebar + main content
      stores/
        moduleStore.ts     # Module schema cache
        collectionStore.ts # Item cache with search/filter
      components/
        DynamicForm.vue      # Schema-driven form renderer
        FormField.vue        # Type-dispatched field input
        ModuleSelector.vue   # Collection type picker
        ItemList.vue         # List view with search + context menu
        CollectionGrid.vue   # Masonry grid view with lazy thumbnails + context menu
        ComparisonView.vue   # Side-by-side item comparison with synced galleries
        ItemDetail.vue       # Premium split-layout item detail view
        ImageAttach.vue      # Image file picker + attachment
        ImageLightbox.vue    # Full-resolution image overlay
        SchemaBuilder.vue    # Split-pane schema editor
        SchemaVisualEditor.vue # Visual field builder
        SchemaCodeEditor.vue # CodeMirror JSON editor
        SchemaFormPreview.vue # Live form preview
        SettingsPage.vue     # Theme configuration
        CommandPalette.vue   # Spotlight-style search overlay (Cmd/Ctrl+K)
        ContextMenu.vue      # Right-click context menu
        ToastProvider.vue    # Global toast notifications
        TagInput.vue         # Tag input with autocomplete chips
        TagFilter.vue        # Tag filter chips for collection views
        TagManager.vue       # Tag rename/delete management panel
        ImportDialog.vue     # Backup import multi-step modal
      stores/
        toastStore.ts        # Toast notification queue
    wailsjs/           # Auto-generated Wails bindings (do not edit)
```

## Data Storage

### Local Mode (default, no env vars)

| Data            | Location                                                          |
| --------------- | ----------------------------------------------------------------- |
| Database        | `~/Library/Application Support/OmniCollect/collection.db` (macOS) |
| Module schemas  | SQLite `modules` table                                            |
| Settings        | SQLite `settings` table                                           |
| Original images | `~/.omnicollect/media/originals/`                                 |
| Thumbnails      | `~/.omnicollect/media/thumbnails/`                                |

### Cloud Mode (env vars set)

| Data            | Location                                          |
| --------------- | ------------------------------------------------- |
| Database        | PostgreSQL (schema-per-tenant: `tenant_{digest}`) |
| Module schemas  | PostgreSQL `modules` table (JSONB)                |
| Original images | S3 `tenants/{tenant-id}/originals/` prefix        |
| Thumbnails      | S3 `tenants/{tenant-id}/thumbnails/` prefix       |
| Settings        | PostgreSQL `settings` table                       |

## Media privacy and processing

Cloud media is isolated by tenant in S3 keys or equivalent local directories.
Keep the S3 bucket private. Original and thumbnail routes require authentication
when JWT is enabled; the frontend fetches blobs with authorization headers, not
tokens in URLs. Public showcases expose only titles and cover images through
reference-checked URLs; prices, tags, additional images, and schema attributes
remain private. Disabling a showcase revokes new requests to its media routes.

Uploads accept JPEG, PNG, GIF, and WebP, up to 30 MB and 24 megapixels. Originals
retain their format; JPEG thumbnails fit within 400 × 400 without cropping.
The server detects response MIME from bytes, limits concurrent decoding to two
images, and bounds upload/archive concurrency. AI receives the actual image MIME,
with a 60-second provider timeout and a 2 MB response limit.

## Dependencies

### Go

| Package                                 | Purpose                         |
| --------------------------------------- | ------------------------------- |
| `github.com/wailsapp/wails/v2`          | Desktop framework + IPC         |
| `modernc.org/sqlite`                    | CGO-free SQLite driver          |
| `github.com/lib/pq`                     | PostgreSQL driver               |
| `github.com/aws/aws-sdk-go-v2`          | S3-compatible object storage    |
| `github.com/google/uuid`                | UUID v4 generation              |
| `github.com/disintegration/imaging`     | Image resize/crop               |
| `golang.org/x/image/webp`               | WebP format support             |
| `github.com/auth0/go-jwt-middleware/v2` | Auth0 JWT validation middleware |
| `gopkg.in/go-jose/go-jose.v2`           | JWKS key handling for JWT       |

### Frontend

| Package                 | Purpose                                 |
| ----------------------- | --------------------------------------- |
| `vue`                   | UI framework                            |
| `pinia`                 | State management                        |
| `vue-codemirror`        | CodeMirror 6 editor wrapper             |
| `@codemirror/lang-json` | JSON syntax highlighting                |
| `@auth0/auth0-vue`      | Auth0 Vue 3 SDK (login, token, session) |

## Keyboard Shortcuts

| Shortcut   | Action                                                                     |
| ---------- | -------------------------------------------------------------------------- |
| Cmd/Ctrl+K | Toggle command palette (search items + quick actions)                      |
| Cmd/Ctrl+F | Focus search bar (switches to list view)                                   |
| Cmd/Ctrl+N | New item for active collection module                                      |
| Escape     | Close topmost overlay (palette, lightbox, form, detail, builder, settings) |

The **Command Palette** provides instant access to any item across all
modules. Type keywords to surface quick actions: "new" (create item/schema),
"settings", "backup"/"export". Navigate results with arrow keys and Enter.

Right-click any item in list or grid view for a context menu with
View, Edit, and Delete actions.

## Multi-Select & Bulk Actions

Click checkboxes in list view or selection badges in grid view to select
multiple items. Shift-click to select a contiguous range. When items are
selected, a floating action bar appears at the bottom with:

- **Delete Selected**: Removes all selected items in one atomic operation
- **Export CSV**: Generates a CSV file with all selected items' data
- **Bulk Edit Module**: Reassigns selected items to a different collection type

Selection persists across list/grid view switches and clears on navigation.

## Rich Text / Markdown

Schema attributes with `widget: "textarea"` render a Markdown editor
with a formatting toolbar (bold, italic, heading, lists, links). Content
is stored as raw Markdown in the database and rendered as formatted HTML
in the item detail view. All rendered HTML is sanitized to prevent
script injection.

## Faceted Filtering

When a specific collection type is selected, a collapsible **Filter Bar**
appears above the item list/grid. It dynamically generates filter controls
from the module's JSON schema:

- **Enum attributes**: Multi-select pills (OR logic within the same field)
- **Boolean attributes**: Tri-state toggle (off -> Yes -> No -> off)
- **Number attributes**: Inline min/max range inputs (400ms debounce)
- **Purchase Price**: Always available as a number range filter

Filters combine with AND logic across attributes and with text search.
A "Clear all" button removes all active filters at once.

## Search and selection

List search is shared application state, so it survives switching views and saved
folders. Requests are debounced and stale responses are ignored. Selections are
cleared when search/filter scope changes and pruned when items disappear;
Shift-click anchors follow item IDs rather than obsolete row positions.

The command palette distinguishes loading, search failures, and empty results.
Closing it cancels pending work. Import and palette overlays use native modal
dialogs; remaining dialog/control accessibility work is tracked in the checklist.

## Keyboard navigation

Collection names, saved views, item titles, selection controls and row/grid action
menus are keyboard reachable. List sorting uses header buttons with announced sort
state. Action menus focus their first option, support arrow keys/Home/End, and
dismiss with Escape or Tab. Named action buttons provide an alternative to
right-click, including on saved views.

The image viewer uses the shared native modal. Focus the image and press Enter or
Space to toggle zoom; use +/− to adjust zoom and arrow keys to pan. Close and reset
are named buttons. The image-viewer workflow still needs browser and assistive-
technology verification. Chrome recovery-modal evidence covers the shared surface,
not every overlay; automated DOM tests do not substitute for native verification.

## Loading and result scope

Saved-view edits stay disabled until settings load successfully. Loading failures
show a retry action; malformed saved views are not silently replaced with an empty
list. Import refresh also reloads settings and saved views before enabling edits.

Loading and failed collection requests no longer display empty-collection prompts,
stale rows or dashboard totals. List search remains available, and failed item or
schema loads have separate retry actions. Filtered empty results are distinct from
an empty collection.

Sidebar counts and the **Collection-wide totals** panel use an independent server
snapshot of all active records, unaffected by filters or paging. They distinguish
missing prices, recorded zero, invalid prices and overflowing totals. Truncated
module groups are disclosed; overall totals stay complete. Refresh failures hide
stale numbers and offer an independent retry. These are snapshots, not live updates
from other clients.

The separate page Insights charts describe only the current page (at most 100 items). Insights reports recorded purchase prices, not
market valuations, and states how many items have valid prices. Missing prices are
not recorded zeroes; invalid prices are excluded and overflowing totals are marked
unavailable. Global totals refresh on initial load and after acknowledged saves,
deletions, recovery, bulk/tag changes and imports, or via the totals refresh button.

Browsing now uses 100-item SQL-backed pages. Previous/Next controls replace the
visible page rather than appending records. Selection and column sorting apply to
that page only; changing pages clears selection. Search/filter changes, writes and
refreshes restart from page one because concurrent changes can shift offsets.
An empty later page asks you to restart instead of claiming the collection is empty.
After the maximum offset, narrow your query to find remaining records.
The command palette requests at most 20 matches across collections and discloses
that limit; use collection search and page navigation to browse further matches.

### Bounded browsing and read API

- GET /api/v1/items/page accepts query, moduleId, filters and tags, plus limit
  (default 100, range 1–200) and offset (default 0, maximum 1,000,000). It returns
  items, limit, offset and hasMore. Limits are applied in SQL with one lookahead row.
  Tied sort values use item IDs. Concurrent edits can shift offset pages; refresh
  from the beginning rather than treating multiple requests as one frozen snapshot.
- GET /api/v1/items/summary accepts no filters and reports tenant-wide item counts,
  valid-price coverage and purchase totals, plus at most 1,000 module counts.
  modulesTruncated explicitly signals omitted groups. Missing prices yield a null
  purchaseTotal; numeric overflow yields null with valueAvailable false. Each
  summary uses one database read snapshot.
- These additive endpoints share normal authentication and request cancellation.
  The legacy GET /api/v1/items array returns complete results only up to 200 matches.
  Larger results return HTTP 422 directing callers to /api/v1/items/page; it never
  silently truncates. The collection UI and palette already use the paged API.
- Public galleries now fetch 24 title/cover projections in SQL, with count and
  collection name in the same read transaction. Page parameters are strict and
  capped at 41,667; larger collections disclose the browsing limit while retaining
  their full count. No full item or module-list read is needed for a gallery page.
- Snapshot reads preflight at most 100,000 items / 32 MiB of encoded item data
  (including a 512-byte allowance per record), 1,000 modules / 8 MiB of module
  JSON, and 1 MiB settings within one transaction. Individual item reads in these
  bulk paths are capped at 1 MiB; schemas at 256 KiB. Oversized exports fail rather
  than silently dropping records. These are encoded-data limits, not heap benchmarks.
- Module saves/restores enforce the module catalog budget transactionally. Schema
  compatibility scans process one record at a time. Tags list at most 10,000
  distinct names; exceeding this returns an error, not a partial list.
- CSV export reads at most 8 MiB of selected encoded records and limits output to
  1,024 columns / 8 MiB. Reassignment and recovery capture also preflight selected
  records against 8 MiB. Choose smaller selections when a bulk budget is exceeded.
- SQLite tag rename/delete runs as one atomic SQL statement without retaining all
  matching rows in Go. Both databases preserve first-occurrence tag order and merge
  existing destinations without duplicates. Showcase lists return at most 10,000
  records, rejecting larger results rather than silently truncating them. These
  are application resource bounds, not universal performance guarantees.


## Account changes

Authentication is checked before the application is created. Account changes,
sign-out and unavailable authentication clear in-memory collection, search,
selection, saved-view, toast and editor state. Save drafts before signing out;
manual sign-out still observes editor guards. A different account gets fresh stores.

Requests are bound to a session generation, including time spent obtaining a token
and reading response bodies. Old requests and downloads are cancelled or ignored,
and queued saved-view writes are discarded. Cancellation does not undo a server
operation that already committed. These lifecycle checks have automated coverage;
live Auth0 lifecycle verification remains outstanding. The native smoke below uses
local mode, not a live authenticated account.

## Sharing and motion preferences

Publishing requires explicit consent: collection names, item titles and cover
images become accessible to anyone with the link, including original images and
embedded metadata such as GPS coordinates. Future items and cover changes are
included too. Disabling sharing blocks future requests but cannot recall copies.
Visibility changes and clipboard failures are reported rather than silently ignored.

The app and public galleries honor reduced-motion preferences. Dashboard canvas
charts also react to changes in that preference. Browser-level visual verification
remains outstanding.

## Editing safely

Item saves validate declared schema types, required fields, enum choices, dates,
prices, record sizes, and tenant-owned image references. Unknown attributes from
older schemas are retained rather than silently discarded. Forms isolate drafts
from collection state and prevent duplicate submissions or saves during uploads
and AI analysis. Item, schema and settings editors require confirmation before
discarding dirty drafts; pending saves block in-app exits. Schema saves parse the
current JSON immediately rather than saving an older debounced preview. Cancelling
settings restores the last committed theme instead of leaving an unsaved preview.

API clients editing an item must send its unchanged `updatedAt` value from the
last read. Saves use an atomic compare-and-swap: stale edits return HTTP 409;
missing edit versions return HTTP 428. The UI retains the draft on failure.
Preserve timestamp precision when round-tripping this value. Item saves have a
15-second request deadline.

Single and bulk item deletion use an explicit confirmation listing the target
records. Targets do not change with the selection while the dialog is open.
Pending deletion blocks duplicate requests and dismissal. Failures remain visible;
an uncertain outcome asks you to inspect recovery and refresh before retrying.

Open **Deletion recovery** in the sidebar to undo item deletions, including after
restarting the app or losing the deletion response. Each batch shows a count,
deletion timestamp and up to five titles. Recovery preserves IDs, creation times,
attributes, tags and image references, but gives items fresh edit versions. It
never overwrites an existing ID; missing or incompatible schemas block the whole
batch. Recovered items can reappear in enabled public showcases.

Batches do not expire automatically. Limits are 500 items / 8 MiB per batch and
100 batches / 64 MiB of record and title-preview payloads per tenant. At capacity,
further deletions are refused until batches are recovered or explicitly discarded.
Permanent discard requires separate consent and removes recovery records, not
retained media files or copies in earlier backups. Recovery cannot recreate image
files removed outside the app. ZIP exports exclude deleted records; Replace
imports permanently clear recovery, with a warning in the import dialog.
Recovery/discard acknowledgements and later view-refresh errors are separate.

Tag rename/delete, saved-view deletion and bulk moves now use explicit native
confirmations. Tag changes cover every collection; renaming into an existing tag
warns that assignments merge. Rename drafts survive cancelled confirmation.
Saved-view deletion removes only the view and waits for persistence before
reporting success. Bulk moves freeze the selected IDs and destination; schema
incompatibility rejects the whole move. These operations prevent duplicate writes
and block another attempt after an uncertain result until reload. A successful
write followed by a refresh failure is reported separately.
Chrome verifies focus entry, background inertness, Tab navigation, pending Escape
protection and focus restoration for recovery. Other destructive workflows and
broader native desktop behavior still need verification.

## Browser verification

From the repository root, with Go, Node 24+ and a local Chrome installation:

```sh
(cd frontend && npm run build) && node frontend/scripts/browser-smoke.mjs
```

The runner builds a temporary binary and uses fresh application HOME, database,
media and browser profile directories. It blocks external page requests, including
font stylesheets; screenshots therefore use existing fallback fonts. It stops only
its own processes and retains screenshots, logs, accessibility tree, test data and
report in the printed temporary artifact directory. Set CHROME_PATH when Chrome
is not at the default macOS application path.

Iteration 20: 33 assertions pass in Chrome 154 with isolated test data and profiles.
At 390px, navigation collapses behind a named disclosure and main content
retains the full viewport width. Coverage includes recovery consent/focus/pending
guards, detail deletion consent, schema typing and keyboard reordering, native
draft-discard confirmation, named list controls and narrow bulk actions. Dynamic
field labels are inspected through Chrome's accessibility tree; a paused item-save
request verifies that Markdown rejects edits and persists its original snapshot.
External
fonts are blocked. Screen-reader, live Auth0 and other-browser checks remain open.
This Chrome runner does not verify native desktop runtime; see the separate smoke
below.

Detail deletion goes directly to the recovery-aware confirmation. Detail and
comparison images have named keyboard controls. Schema fields use stable editor
keys, immutable reordering and move-up/down buttons. Uncertain item saves retain
the draft but lock repeat submissions to avoid duplicate creates after lost
acknowledgements; explicit validation rejections remain correctable. Copy any
unsaved details and inspect the collection before starting another save.

## Native desktop verification

macOS arm64 packaging and ad-hoc signature verification also passed. On this host,
the default 27.0 SDK produced a linker/TAPI `arm64e.x1` error; selecting the installed
15.4 SDK for the command worked without changing system settings:

```sh
SDKROOT=/Library/Developer/CommandLineTools/SDKs/MacOSX15.4.sdk \
  wails build -s -skipbindings -m -nosyncgomod -o omnicollect-hardening-20
codesign --verify --deep --strict build/bin/omnicollect.app
```

Build frontend assets first. Wails packages to `build/bin/omnicollect.app` on this
host despite the output-name flag. Packaging alone does not prove runtime behavior
or establish signing/notarization for distribution.

On a logged-in macOS desktop with Swift/CLT, Node 24+, and existing Accessibility
and Screen Recording permission for the launching terminal, run:

```sh
node frontend/scripts/native-smoke.mjs
```

**Leave the test window and dialogs untouched.** The runner copies the built app,
assigns a unique fixture bundle identity, re-signs it, and launches only that copy.
Fresh HOME/CFFIXED_USER_HOME and config/temp paths isolate application and WebKit
state. Permission and Foundation-path preflight runs before launch; the driver
refuses non-fixture identities and captures only owned windows. Reports, native
accessibility trees, screenshots, logs and isolated data remain in the printed
temporary directory. Unlike the Chrome runner, external font requests are not
intercepted.

Seven checks pass: WebKit rendering, a same-origin summary read, No/Yes close
confirmation, native default No, No retaining an unsaved schema field, normal exit
after Yes, and the database beneath isolated HOME. Accessibility inspection is not
screen-reader verification. Native text entry, Escape/Return shortcuts, concurrent
close requests and broader destructive workflows are not covered by this smoke.

Desktop close requests now always ask a native No/Yes question, even for a clean
window. No is the default/cancel choice; errors and duplicate pending requests keep
the app open. This conservative policy does not depend on browser unload support
or a potentially stale draft flag. Yes explicitly accepts possible draft loss or
interruption: requests already sent may still complete. It does not protect against
force-quit, process crashes or power loss. Policy tests and the scoped isolated
native smoke above pass; they do not establish every native workflow.

Dynamic field labels and validation errors are associated with their controls.
Markdown formatting and editing are disabled during saves, and schema previews
are disabled for keyboard users as well as pointer users.

## Backup & Export

Click "Export Backup" in the sidebar to create a portable ZIP containing a
consistent database snapshot (`items.json`, `modules.json`, `settings.json`),
a version manifest, and every referenced original and thumbnail. Both local and
cloud backends export this format. Missing referenced media fails the export;
a failed export never replaces an existing backup file. Archives are limited to
256 MB compressed, 1 GB expanded, and 10,000 entries.

## Backup Import & Restore

Click "Import Backup" in the sidebar (or use the command palette) to
restore from a previously exported ZIP archive. Results and warnings remain visible
until you choose Done. Active analysis/import cannot be dismissed accidentally;
replace mode requires explicit confirmation. If restoring succeeds but refreshing
the collection fails, retrying the view refresh does not repeat the restore.
Connection loss is reported as an uncertain outcome, not a guaranteed rollback.

The two-step flow:

1. **Analyze**: Upload the ZIP file. OmniCollect scans it and shows a
   summary (item count, image count, module count, detected format).
2. **Import**: Choose a mode and confirm:
   - **Merge** (default): Adds backup items alongside existing data.
     Existing items not in the backup are preserved. Matching IDs are
     updated with the backup version.
   - **Replace**: Replaces items, modules, and settings in one database transaction.
     A database failure leaves existing metadata intact. Existing public showcases
     are disabled. Item IDs and historical timestamps are preserved in both modes.

Originals are validated and staged under content-addressed names before the
database commit; thumbnails are regenerated. Failed staging leaves collection
metadata and existing media references unchanged, and the import handle remains
retryable. A metadata restore error does not prove rollback: a commit acknowledgement
can be lost after success. Refresh and inspect the collection before retrying. Unreferenced staged files may remain after failure; they are not
deleted immediately because concurrent work can reference identical content.

Imports reject duplicate/unsafe ZIP entries, symlinks, bad checksums, unsupported
versions, and oversized content before committing. Legacy SQLite backups
(including those without tags) and legacy JSON backups remain supported. Missing
originals must already exist in the destination; restore never silently succeeds
with broken image references.

Supported backup formats:

- **Legacy local** (SQLite-based): `collection.db` + media + module JSON files.
- **Portable JSON**: `items.json` + `modules.json`, with settings and referenced
  media in current exports. Legacy JSON exports without media/settings remain
  recognizable but may be incomplete.

Cross-format import works: a local backup can be imported into a cloud
deployment, and vice versa.

## Iteration History

1. **Core Engine** (001): Go backend, SQLite schema with FTS5, Wails
   IPC bindings (SaveItem, GetItems, GetActiveModules)
2. **Dynamic Form Engine** (002): Pinia stores, schema-driven form
   renderer, item list with search/filter, edit support
3. **Image Processing & Grid** (003): Thumbnail generation, Wails
   AssetServer for local media, collection grid with lazy loading,
   full-resolution lightbox
4. **Schema Visual Builder** (004): Split-pane schema editor with
   visual drag-and-drop field builder, CodeMirror JSON editor with
   bidirectional sync, live form preview, save-to-disk with hot reload
5. **Backup Export & Sync Prep** (005): ZIP archive export of
   database + media + modules, UTC timestamp hardening for future sync
6. **UX & Power User Features** (006): Command palette (Cmd/Ctrl+K)
   for cross-module item search and quick actions, global keyboard
   shortcuts (Cmd+F/N/Esc), right-click context menus, toast
   notifications, item delete with confirmation, premium split-layout
   item detail view with Instrument Serif/Outfit typography
7. **Faceted Filtering** (007): Schema-driven filter bar with enum
   multi-select pills, boolean tri-state toggles, number range inputs,
   purchasePrice filtering, collapsible UI, backend json_extract queries
8. **Markdown Textarea** (008): CodeMirror Markdown editor with
   formatting toolbar for textarea widgets, safe rendered HTML in detail
   views via marked + DOMPurify, global .prose typography class
9. **Multi-Select & Bulk Actions** (009): Checkboxes in list view,
   selection badges in grid view, Shift-click range select, floating
   glassmorphism action bar, atomic batch delete, CSV export with save
   dialog, bulk module reassignment
10. **REST API Migration** (010): Decoupled Wails IPC to standard HTTP
    REST endpoints under /api/v1/, fetch-based frontend client, standalone
    server mode (--serve), multipart image upload, Content-Disposition
    downloads for export, TypeScript API types replacing Wails codegen
11. **Cloud Infrastructure** (011): Storage abstraction layer (Store +
    MediaStore interfaces), PostgreSQL backend with schema-per-tenant
    isolation and tsvector FTS, S3-compatible media storage, SQLite-to-PG
    migration tool (--migrate), Docker multi-stage build, docker-compose
    dev stack, health check endpoint, config via environment variables
13. **JWT Authentication** (013): Auth0-based JWT authentication
    middleware, JWKS-validating Go middleware with tenant ID extraction
    from sub claim, auto-provisioning of PostgreSQL tenant schemas,
    Vue Auth0 SDK integration (AuthGuard, token injection, sign out),
    local mode bypass for backward compatibility
14. **Cross-Collection Tags** (014): Free-form tag system stored as JSON
    arrays on items (SQLite TEXT / PostgreSQL JSONB with GIN index). Tag
    input with autocomplete, clickable tag filter chips above collection
    views (OR logic, cross-module), tag management panel (rename/delete).
    Tags included in FTS5/tsvector search index and CSV export. REST
    endpoints: GET /api/v1/tags, POST /api/v1/tags/rename,
    DELETE /api/v1/tags/{name}
15. **Backup Import & Restore** (015): Two-step import flow (analyze +
    execute) for backup ZIP files. Supports Replace (atomic) and Merge
    (per-item upsert) modes. Handles both local (SQLite) and cloud (JSON)
    backup formats with cross-format import. ImportDialog.vue multi-step
    modal with file picker, summary, mode selection, progress spinner.
    REST endpoints: POST /api/v1/import/analyze, POST /api/v1/import/execute
16. **AI Metadata Extraction** (016): Vision-model-powered auto-fill of
    item form fields from uploaded photos. Configurable AI provider
    (Anthropic direct or OpenAI-compatible for OpenRouter/Google). Prompt
    built from module schema, response validated against types and enum
    options. "Analyze with AI" button in DynamicForm fills only empty
    fields, suggests title if already set. Feature disabled when
    AI_PROVIDER is empty. No new dependencies.
    REST endpoints: POST /api/v1/ai/analyze, GET /api/v1/ai/status
17. **Public Showcase URLs** (017): Toggle any collection public to get a
    shareable gallery URL. Server-rendered HTML via Go templates (zero JS
    for visitors). CSS :target overlay for item detail with full image.
    Stable slug generation ({name}-{8-hex}), toggle private revokes access
    instantly. Showcases table in SQLite main DB / PostgreSQL public schema
    (cross-tenant slug lookup). 24-item server-side pagination. Feature
    disabled in local/desktop mode.
    Routes: GET /showcase/{slug} (public), POST /api/v1/showcases/toggle,
    GET /api/v1/showcases
