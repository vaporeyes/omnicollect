# Research: Command Palette

**Branch**: `006-command-palette` | **Date**: 2026-04-07

## R1: Cross-Module Search Without New Backend Binding

**Decision**: Reuse the existing `GetItems(query, moduleID)` binding with an empty `moduleID` to search across all modules.

**Rationale**: The `queryItems` function in `db.go` already handles `moduleID == ""` by omitting the module filter. FTS5 search works identically. No new Go code or Wails binding is needed.

**Alternatives considered**:
- New dedicated `SearchItems(query)` binding: Rejected because it would duplicate existing logic.
- Client-side filtering of already-fetched items: Rejected because the in-memory item list may be filtered by module. Using the backend ensures complete cross-module results.

## R2: Debounce Strategy

**Decision**: Use a 200ms debounce on the palette input before querying the backend.

**Rationale**: 200ms balances responsiveness (feels near-instant) with avoiding excessive queries on rapid typing. The spec requires results within 300ms of the user pausing, leaving 100ms for the actual query round-trip (local SQLite is sub-1ms).

**Alternatives considered**:
- No debounce (query on every keystroke): Rejected due to unnecessary load on rapid typing.
- 500ms debounce: Rejected because it would feel sluggish for a power-user tool.

## R3: Result Rendering -- Thumbnails vs Originals

**Decision**: Display thumbnail images in palette results using the existing `/thumbnails/` AssetServer path.

**Rationale**: Constitution Principle IV prohibits loading originals in list-like views. Thumbnails (300x300 JPEG) are already generated at import time and served via the Wails AssetServer.

**Alternatives considered**: None -- this is a constitutional requirement, not a design choice.

## R4: Quick Action Matching Strategy

**Decision**: Case-insensitive substring match against a hardcoded keyword map. Keywords: "new" (Add New Item, Create New Schema), "settings" (Open Settings), "backup"/"export" (Export Backup).

**Rationale**: Simple and predictable. Users type natural keywords and see relevant actions. No fuzzy matching needed for 4 predefined actions.

**Alternatives considered**:
- Fuzzy matching (e.g., "sett" matching "settings"): Adds complexity with minimal benefit for so few actions. Can be added later if needed.
- Full NLP intent detection: Overkill for a local desktop app.

## R5: Result Count Limit

**Decision**: Cap results at 25 items.

**Rationale**: A command palette should show a focused, scannable list. 25 items fill roughly one viewport. Users refine their query to narrow results further.

**Alternatives considered**:
- Unlimited results: Rejected due to potential DOM bloat and scroll fatigue.
- 10 results: Too few; users with large collections may not find their item without very precise queries.

## R6: Palette Z-Index and Overlay Behavior

**Decision**: The palette renders via `<Teleport to="body">` with a high z-index (3000+), above all other overlays (lightbox at 1000, context menu at 2000, toast at 9999). It does not dismiss other overlays.

**Rationale**: The palette is a global navigation tool. It should always be reachable regardless of what else is open. Closing the palette returns focus to whatever was underneath.

**Alternatives considered**:
- Dismissing other overlays when palette opens: Rejected because the user may want to return to their previous context after a quick palette lookup.
