# Research: Smart Folders (Saved Views)

**Branch**: `019-smart-folders` | **Date**: 2026-04-10

## R1: Persistence Mechanism

**Decision**: Store Smart Folders as a JSON array under the `smartFolders` key within the existing settings blob.

**Rationale**:
- The app already has `GET/PUT /api/v1/settings` endpoints that read/write a JSON object. The frontend loads settings on mount (`App.vue` line 258) and saves via `api.put('/api/v1/settings', data)`.
- Adding a `smartFolders` array to this JSON object requires zero backend changes. The Go backend stores settings as an opaque JSON string (`GetSettings`/`SaveSettings` in Store interface).
- Both SQLite (local) and PostgreSQL (cloud) implementations already handle this.
- Smart Folder data is small (each folder is ~200 bytes of serialized state), so even 50 folders add negligible overhead to the settings blob.

**Alternatives considered**:
- **Separate database table**: Would require new Go handlers, Store interface methods, and migrations for both SQLite and PostgreSQL. Overkill for a simple list of small JSON objects.
- **LocalStorage**: Would not sync with cloud mode settings. Would lose data if browser storage is cleared.
- **Separate settings key per folder**: More complex to manage than a single array. No benefit.

## R2: Smart Folder Identity

**Decision**: Generate a random ID (8-character hex string) for each Smart Folder at creation time.

**Rationale**:
- Need a stable identifier for active-state tracking and context menu operations (rename/delete target).
- Random hex is simple, collision-resistant for the expected scale (< 50 folders), and does not leak creation order (timestamp is a separate field).
- Same pattern used by the showcase slug generation in the existing codebase.

## R3: Active State Detection

**Decision**: Track active Smart Folder by ID in the store, not by deep-comparing current view state against all saved folders.

**Rationale**:
- Deep comparison of module + search + filters + tags against every Smart Folder on every state change would be expensive and fragile (filter object ordering, floating point values, etc.).
- Instead: when a Smart Folder is clicked, set `activeSmartFolderId`. When the user manually changes any filter/search/module/tags, clear `activeSmartFolderId`. This is simple, predictable, and matches the spec (FR-005, FR-006).

## R4: Inline Rename UX

**Decision**: Reuse the same inline text field pattern from "Save Current View" for rename. Clicking "Rename" in context menu replaces the folder name with an editable text field; Enter confirms, Escape cancels.

**Rationale**:
- Consistent UX with the save flow (both use inline text fields in the sidebar).
- No modal needed for a single-field edit.
- Matches the clarification decision from the spec session.

## R5: Settings Load/Save Race Condition

**Decision**: Smart Folder store loads from settings on mount and saves the full settings blob (merging its `smartFolders` key) on every mutation.

**Rationale**:
- Settings are loaded once on mount (`App.vue` onMounted). Smart Folder store initializes from that loaded data.
- On each Smart Folder mutation (create/rename/delete), the store re-saves the full settings blob. This is the same pattern used by the theme settings.
- Risk of race condition is minimal: settings are only written by user-initiated actions (theme change, smart folder change), not by automated processes.
