# Implementation Plan: Cloud Infrastructure Migration

**Branch**: `011-cloud-infrastructure` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-cloud-infrastructure/spec.md`

## Summary

Migrate OmniCollect from local-only infrastructure (SQLite + filesystem) to cloud-ready: PostgreSQL with schema-per-tenant isolation using tsvector/tsquery full-text search, S3-compatible object storage for images, and Docker containerization. Local mode preserved as fallback. Database abstraction layer enables runtime switching between SQLite and PostgreSQL backends.

## Technical Context

**Language/Version**: Go 1.25+ (backend), TypeScript + Vue 3 (frontend -- minimal changes)
**Primary Dependencies**: `database/sql` + `lib/pq` (PostgreSQL driver), AWS SDK v2 for Go (S3), Docker multi-stage build
**Storage**: PostgreSQL (cloud) / SQLite (local fallback); S3-compatible object store (cloud) / local filesystem (fallback)
**Testing**: Manual + curl + docker compose for integration testing
**Target Platform**: Docker container (cloud), macOS desktop (local/Wails)
**Project Type**: Hybrid desktop/web transitioning to cloud-deployable service
**Performance Goals**: Query latency under 200ms; image upload under 2s including S3 round-trip
**Constraints**: Schema-per-tenant PostgreSQL; stateless container; 12-factor env config; backward compatible local mode
**Scale/Scope**: Multi-tenant; hundreds of users; thousands of items per tenant

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | ATTENTION | Cloud mode shifts to cloud-first. Mitigated: local SQLite mode preserved as fallback when cloud config absent. Constitution Principle I is satisfied in local mode; cloud mode is a new deployment target. |
| II. Schema-Driven UI | PASS | No UI changes; module schemas stored in DB table instead of files, but still loaded as JSON |
| III. Flat Data Architecture | PASS | Same flat JSON attributes column; PostgreSQL JSONB replaces SQLite JSON; no JOINs for attributes |
| IV. Performance & Memory | PASS | Thumbnails served from S3 presigned URLs or proxied; same thumbnail-only pattern in views |
| V. Type-Safe IPC | PASS | REST API unchanged; TypeScript types unchanged |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update at completion |

Constitution Principle I requires attention: cloud mode is not strictly "local-first." Mitigation: the application detects whether `DATABASE_URL` is set. If absent, it falls back to local SQLite + filesystem (existing behavior). Cloud mode is an additive deployment option, not a replacement for local mode.

## Project Structure

### Documentation (this feature)

```text
specs/011-cloud-infrastructure/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  contracts/           # Phase 1 output
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Backend (Go)
storage/
  db.go                # New: database interface (Store) with SQLite and PostgreSQL implementations
  sqlite.go            # New: SQLite Store implementation (extracted from current db.go)
  postgres.go          # New: PostgreSQL Store implementation (schema-per-tenant, tsvector FTS)
  migrate.go           # New: data migration tool (SQLite -> PostgreSQL)
  media.go             # New: media storage interface with local filesystem and S3 implementations
  s3.go                # New: S3 media storage implementation
  local.go             # New: local filesystem media storage (extracted from current imaging.go)
config.go              # New: environment-based configuration (DATABASE_URL, S3_*, PORT)
imaging.go             # Modified: use media storage interface instead of direct filesystem
handlers.go            # Modified: use Store interface instead of direct *sql.DB
app.go                 # Modified: initialize Store and MediaStore based on config
Dockerfile             # New: multi-stage build (Go + Node -> minimal runtime)
docker-compose.yml     # New: development stack (app + postgres + minio)
```

**Structure Decision**: A `storage/` package introduces interfaces for both database operations (`Store`) and media storage (`MediaStore`). Two implementations each: SQLite+LocalFS for desktop/local mode, PostgreSQL+S3 for cloud mode. The rest of the application programs against the interfaces. This is the minimal abstraction needed to support both backends without duplicating business logic.

## Complexity Tracking

| Concern | Justification | Simpler Alternative Rejected Because |
|---------|--------------|-------------------------------------|
| Constitution I (Local-First) | Cloud mode is additive; local mode fully preserved | Removing local mode would break desktop users and violate the constitution |
| Storage interface abstraction | Two implementations (SQLite/PG, Local/S3) behind interfaces | Without interfaces, every handler would need if/else for backend type; interfaces are cleaner and standard Go practice |
