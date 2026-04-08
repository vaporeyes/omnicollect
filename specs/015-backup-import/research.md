# Research: Backup Import and Restore

**Branch**: `015-backup-import` | **Date**: 2026-04-08

## R1: Backup Format Detection

**Decision**: Detect format by inspecting ZIP contents. If the ZIP contains `collection.db`, it's a local-mode backup. If it contains `items.json`, it's a cloud-mode backup. If neither, reject as unrecognized.

**Rationale**: The export feature produces deterministic file names. No format version header needed -- file presence is sufficient.

**Detection logic**:
1. Open ZIP reader
2. Scan entries for `collection.db` -> local format
3. Scan entries for `items.json` -> cloud/JSON format
4. Neither found -> error: unrecognized format

## R2: Local Format Import Strategy

**Decision**: Open the embedded `collection.db` as a read-only SQLite database. Query all items from it. Insert each item into the active Store via the Store interface. Extract media files from `media/originals/` and `media/thumbnails/` paths in the ZIP and save via MediaStore. Read module schemas from `modules/*.json` entries and save via Store.

**Rationale**: Reading from the embedded SQLite database gives access to items in their native format (including all fields, attributes, tags). Using the Store interface ensures items are written correctly to whatever backend is active (SQLite or PostgreSQL).

**Alternatives considered**:
- Copy the database file directly: Only works for SQLite-to-SQLite (no cross-mode support). Rejected.

## R3: Cloud/JSON Format Import Strategy

**Decision**: Parse `items.json` as a JSON array of Item objects. Parse `modules.json` as a JSON array of ModuleSchema objects. Insert via Store interface. Extract media files from `media/` paths (if present) and save via MediaStore.

**Rationale**: JSON format is straightforward to parse and insert. Same Store interface path as local format, just different source parsing.

## R4: Replace Mode Atomicity

**Decision**: For Replace mode:
1. Begin a transaction (or create a savepoint)
2. Delete all existing items
3. Delete all existing modules
4. Insert all items from backup
5. Insert all modules from backup
6. Commit transaction

If any step fails, rollback. Existing data is preserved.

For media: images are copied after the database transaction commits. If image copy fails, the items exist but some images may be missing (acceptable -- images can be re-imported).

**Rationale**: Full atomicity for database operations prevents data loss. Images are best-effort after the DB commit because they're stored externally (filesystem/S3) and can't participate in a DB transaction.

## R5: Merge Mode Strategy

**Decision**: For each item in the backup:
1. Check if an item with the same ID exists
2. If yes, update it with the backup's version
3. If no, insert it

For modules: same upsert logic (match by ID).

Items NOT in the backup are preserved (this is the key difference from Replace).

**Rationale**: Merge is useful for restoring specific items or combining collections from different devices. "Backup wins" on ID conflicts is the safe default -- the user explicitly chose to import this backup.

## R6: Pre-Import Analysis (Summary)

**Decision**: A separate endpoint (`POST /api/v1/import/analyze`) accepts the ZIP file, scans its contents without modifying any data, and returns counts:
- Item count
- Image count (originals)
- Module count
- Format detected (local/cloud)
- Warnings (e.g., missing module schemas)

The frontend shows this summary and lets the user choose Replace or Merge before confirming.

**Rationale**: A separate analyze step avoids the need to re-upload the file. The file can be uploaded once and the server can keep it in a temp location for the actual import.

**Alternative approach**: Upload the file once to a temp endpoint. Return a temporary import ID + summary. The confirm step sends the import ID + chosen mode. This avoids re-uploading large files.

**Decision**: Use the two-step approach (upload -> analyze -> confirm with temp ID) to avoid double-uploading.

## R7: Progress Reporting

**Decision**: The import endpoint returns immediately with progress available via polling or Server-Sent Events (SSE). For v1, use a simple polling approach: the import runs synchronously and returns the final result. The frontend shows an indeterminate progress spinner during the request.

**Rationale**: True progress reporting requires async processing + a status endpoint. For v1, the simpler synchronous approach is sufficient -- a 100-item import takes under 10 seconds. If imports grow large enough to need real progress, SSE can be added later.

**Alternatives considered**:
- SSE progress stream: More complex but provides real-time updates. Deferred to v2.
- WebSocket: Overkill for a one-shot operation. Rejected.
