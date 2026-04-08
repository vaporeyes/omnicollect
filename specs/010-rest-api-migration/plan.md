# Implementation Plan: REST API Migration

**Branch**: `010-rest-api-migration` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-rest-api-migration/spec.md`

## Summary

Replace Wails IPC bindings with a standard HTTP REST API. The Go backend becomes a standalone HTTP server exposing all operations under `/api/v1/`. The Vue frontend replaces all `wailsjs/go/main/App` imports with a centralized `fetch`-based API client. Migration is big-bang. Exports use Content-Disposition downloads. Image upload uses a single multipart POST endpoint. The Wails desktop shell embeds the HTTP server.

## Technical Context

**Language/Version**: Go 1.25+ (backend -- HTTP server + router), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Go `net/http` + lightweight router, Pinia (state), `fetch` API (no Axios needed)
**Storage**: SQLite via modernc.org/sqlite (unchanged)
**Testing**: Manual acceptance + curl endpoint testing
**Target Platform**: macOS desktop (Wails v2) + standalone web server mode
**Project Type**: Desktop application transitioning to hybrid desktop/web
**Performance Goals**: HTTP round-trip under 200ms on localhost (matching current Wails IPC)
**Constraints**: Local-first (no auth in v1); CORS for dev; big-bang migration
**Scale/Scope**: Single-user; all existing bindings converted to REST endpoints

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | Server runs locally; SQLite remains source of truth; no cloud dependency |
| II. Schema-Driven UI | PASS | No UI template changes; only transport layer changes |
| III. Flat Data Architecture | PASS | Database unchanged; JSON attribute queries unchanged |
| IV. Performance & Memory | PASS | Same thumbnail/original media serving; HTTP adds negligible overhead on localhost |
| V. Type-Safe IPC | ATTENTION | Wails generated TypeScript bindings are removed. The API client must define equivalent TypeScript types manually or generate them from the REST contract. Type safety is maintained through explicit interface definitions. |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update at completion |

Constitution Principle V requires attention: Wails auto-generated TypeScript bindings go away. Mitigation: the frontend API client module defines TypeScript interfaces for all request/response types, matching the Go structs. These are checked at compile time by the TypeScript compiler.

## Project Structure

### Documentation (this feature)

```text
specs/010-rest-api-migration/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  contracts/           # Phase 1 output (REST endpoint specs)
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Backend (Go)
server.go              # New: HTTP server setup, router, CORS middleware, startup
handlers.go            # New: HTTP handler functions wrapping existing App methods
app.go                 # Modified: remove Wails-specific context; App struct becomes
                       #   a dependency injected into handlers
main.go                # Modified: start HTTP server; optionally start Wails shell
                       #   that embeds the server

# Frontend (Vue/TypeScript)
frontend/src/
  api/
    client.ts          # New: centralized fetch-based HTTP client with base URL config
    types.ts           # New: TypeScript interfaces for all API request/response types
  stores/
    collectionStore.ts # Modified: replace Wails imports with api/client calls
    moduleStore.ts     # Modified: replace Wails imports with api/client calls
  components/
    ImageAttach.vue    # Modified: use file input + multipart upload instead of
                       #   SelectImageFile + ProcessImage
  App.vue              # Modified: replace Wails runtime imports; use api/client
                       #   for export/backup/settings
```

**Structure Decision**: Two new Go files (server.go, handlers.go) isolate HTTP concerns from business logic. Two new frontend files (api/client.ts, api/types.ts) centralize HTTP calls and type definitions. All existing stores and components switch from Wails imports to the api client. The `wailsjs/go/main/App` import path is eliminated entirely.

## Complexity Tracking

| Concern | Justification | Simpler Alternative Rejected Because |
|---------|--------------|-------------------------------------|
| Constitution V (Type-Safe IPC) | TypeScript interfaces in api/types.ts replace Wails-generated bindings | Wails bindings are auto-generated; manual types require discipline but are standard practice for REST APIs. The TypeScript compiler catches mismatches at build time. |
