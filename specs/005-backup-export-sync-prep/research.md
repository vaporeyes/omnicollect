# Research: Backup Export & Sync Preparation

**Date**: 2026-04-05
**Feature**: 005-backup-export-sync-prep

## R1: ZIP Archive Strategy

**Decision**: Use Go standard library `archive/zip` for streaming
compression. Write directly to a file (not buffered in memory).

**Rationale**: `archive/zip` is battle-tested, requires no external
dependencies, and supports streaming writes. For large media
directories, streaming avoids memory issues -- each file is compressed
and written to the output file incrementally.

**Pattern**: Walk the media directory with `filepath.Walk`, add each
file to the zip writer. Copy the SQLite database file directly (after
ensuring WAL is checkpointed). Include module schemas.

**Alternatives considered**:
- `archive/tar` + gzip: Slightly better compression ratios but ZIP
  is more universally readable on all platforms. Rejected for
  cross-platform accessibility.
- Third-party compression (zstd, lz4): Better performance but adds
  a dependency for marginal benefit. Rejected.

## R2: SQLite Backup During Export

**Decision**: Use SQLite `PRAGMA wal_checkpoint(TRUNCATE)` before
copying the database file, or use Go's `database/sql` to copy via
the `backup` API.

**Rationale**: The database uses WAL mode. Simply copying the .db
file while WAL data is outstanding could produce a corrupt backup.
The simplest approach: checkpoint the WAL, then copy the .db file.
This is safe because OmniCollect is single-user with no concurrent
writers during export.

**Simpler alternative**: Close the database briefly, copy the file,
reopen. This is unnecessary since WAL checkpoint achieves the same
result without interrupting queries.

## R3: Timestamp Verification

**Decision**: Audit existing `insertItem` and `updateItem` in `db.go`
to confirm they correctly set UTC timestamps. No changes expected --
the existing code uses `time.Now().UTC().Format(time.RFC3339)`.

**Rationale**: Iteration 1 already implements UTC timestamps. This
iteration verifies the behavior is correct and consistent, rather
than reimplementing it. The verification covers:
- `insertItem`: sets both `created_at` and `updated_at`
- `updateItem`: sets `updated_at` only, preserves `created_at`
- All timestamps use `time.RFC3339` (ISO 8601 compatible)

## R4: Export Filename Convention

**Decision**: `omnicollect-backup-{YYYYMMDD-HHMMSS}.zip`

**Rationale**: Timestamp in the filename allows multiple backups
without collision. Human-readable format. UTC time used to avoid
timezone ambiguity.

## Summary of New Dependencies

None. Go standard library `archive/zip` and `io` cover all needs.
