# Data Model: REST API Migration

**Branch**: `010-rest-api-migration` | **Date**: 2026-04-07

## No Data Model Changes

This feature changes the transport layer (Wails IPC to HTTP REST) but makes zero changes to the data model, database schema, or storage format. All existing tables, columns, indexes, FTS5 triggers, and JSON attribute structures remain identical.

## New Configuration Entity

### ServerConfig (runtime, not persisted)

| Field | Type | Description |
|-------|------|-------------|
| port | int | HTTP server listen port (default: 8080, or 0 for random in desktop mode) |
| corsOrigins | string[] | Allowed CORS origins (default: ["*"] in dev, empty in production) |
| mediaDir | string | Path to media directory (~/.omnicollect/media/) |
| staticDir | string | Path to built frontend (frontend/dist/) |

## TypeScript API Types (frontend/src/api/types.ts)

These mirror the existing Go structs (already defined in models.go) and replace the auto-generated Wails TypeScript bindings.

| Type | Fields | Mirrors Go Struct |
|------|--------|-------------------|
| Item | id, moduleId, title, purchasePrice, images, attributes, createdAt, updatedAt | models.go:Item |
| ModuleSchema | id, displayName, description, attributes | models.go:ModuleSchema |
| AttributeSchema | name, type, required, options, display | models.go:AttributeSchema |
| DisplayHints | label, placeholder, widget, group, order | models.go:DisplayHints |
| ProcessImageResult | filename, originalPath, thumbnailPath, width, height, format | models.go:ProcessImageResult |
| BulkDeleteResult | deleted | app.go:BulkDeleteResult |
| BulkUpdateResult | updated | app.go:BulkUpdateResult |
