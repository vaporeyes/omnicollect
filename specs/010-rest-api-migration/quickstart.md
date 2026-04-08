# Quickstart: REST API Migration

**Branch**: `010-rest-api-migration` | **Date**: 2026-04-07

## Prerequisites

- Wails v2 development environment
- Existing codebase on the `010-rest-api-migration` branch
- Items in the database for testing

## Files to Create

1. **`server.go`** -- HTTP server setup: `net/http.ServeMux` router, CORS middleware, static file serving for media + frontend, server start function
2. **`handlers.go`** -- HTTP handler functions for all endpoints: wraps existing App methods, parses request params/body, returns JSON responses with status codes
3. **`frontend/src/api/client.ts`** -- Centralized fetch client: base URL config, JSON helpers, error handling, typed API functions for every endpoint
4. **`frontend/src/api/types.ts`** -- TypeScript interfaces mirroring Go structs: Item, ModuleSchema, AttributeSchema, DisplayHints, ProcessImageResult, BulkDeleteResult, BulkUpdateResult

## Files to Modify

1. **`main.go`** -- Start HTTP server (standalone or embedded in Wails shell)
2. **`app.go`** -- Remove Wails-specific context usage; App struct methods remain but are called by handlers
3. **`frontend/src/stores/collectionStore.ts`** -- Replace Wails imports with api/client calls
4. **`frontend/src/stores/moduleStore.ts`** -- Replace Wails imports with api/client calls
5. **`frontend/src/components/ImageAttach.vue`** -- Use file input + multipart upload
6. **`frontend/src/App.vue`** -- Replace Wails runtime imports with api/client calls for export/settings

## Implementation Order

1. Create `frontend/src/api/types.ts` (TypeScript interfaces)
2. Create `frontend/src/api/client.ts` (HTTP client)
3. Create `server.go` (HTTP server + router + CORS)
4. Create `handlers.go` (all endpoint handlers)
5. Modify `main.go` (server startup)
6. Modify all frontend stores to use api/client (big bang swap)
7. Modify ImageAttach.vue for multipart upload
8. Modify App.vue for export/settings
9. Verify desktop mode still works via Wails
10. Update CLAUDE.md and README

## Acceptance Test Flow

### Standalone Server Mode
1. Start the Go server: `go run . --serve` (or equivalent)
2. Verify it listens on port 8080
3. `curl http://localhost:8080/api/v1/modules` -- verify JSON array of modules
4. `curl http://localhost:8080/api/v1/items` -- verify JSON array of items
5. `curl -X POST http://localhost:8080/api/v1/items -d '{"moduleId":"coins","title":"Test"}'` -- verify item created
6. `curl -X DELETE http://localhost:8080/api/v1/items/{id}` -- verify 204
7. Open `http://localhost:8080` in a browser -- verify frontend loads and works

### Frontend Migration
8. Run the frontend dev server (`npm run dev`) against the standalone backend
9. Create an item -- verify POST /api/v1/items is called
10. Search items -- verify GET /api/v1/items?query=... is called
11. Apply faceted filters -- verify filters param is sent
12. Upload an image -- verify multipart POST to /api/v1/images/upload
13. Export backup -- verify ZIP downloads via browser
14. Export CSV -- verify CSV downloads via browser
15. Bulk delete -- verify POST /api/v1/items/batch-delete
16. Open command palette, search, select -- verify works over HTTP

### Desktop Mode
17. Build with `wails build` and launch the app
18. Verify all features work identically to pre-migration
19. Verify thumbnails and originals load correctly
20. Verify no console errors related to Wails bindings

### Verification
21. `grep -r "wailsjs/go/main/App" frontend/src/` -- verify zero matches
22. `grep -r "WindowSetSystemDefaultTheme\|SaveFileDialog" frontend/src/` -- verify zero Wails runtime imports in stores/components
