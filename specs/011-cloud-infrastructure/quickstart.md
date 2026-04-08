# Quickstart: Cloud Infrastructure Migration

**Branch**: `011-cloud-infrastructure` | **Date**: 2026-04-08

## Prerequisites

- Go 1.25+, Node.js 18+, Docker
- Existing codebase on the `011-cloud-infrastructure` branch
- PostgreSQL 16+ (or use docker-compose)
- S3-compatible storage (or use MinIO via docker-compose)

## New Dependencies

```bash
go get github.com/lib/pq                    # PostgreSQL driver
go get github.com/aws/aws-sdk-go-v2/...     # S3 SDK
```

## Files to Create

1. **`storage/db.go`** -- `Store` interface definition
2. **`storage/sqlite.go`** -- SQLite `Store` implementation (extracted from current db.go)
3. **`storage/postgres.go`** -- PostgreSQL `Store` implementation (schema-per-tenant, tsvector FTS, JSONB)
4. **`storage/migrate.go`** -- SQLite-to-PostgreSQL migration tool
5. **`storage/media.go`** -- `MediaStore` interface definition
6. **`storage/s3.go`** -- S3 `MediaStore` implementation
7. **`storage/local.go`** -- Local filesystem `MediaStore` implementation (extracted from imaging.go)
8. **`config.go`** -- Environment-based configuration with mode detection
9. **`Dockerfile`** -- Multi-stage build (Go + Node -> alpine runtime)
10. **`docker-compose.yml`** -- Dev stack (app + postgres + minio)

## Files to Modify

1. **`app.go`** -- Use `Store` and `MediaStore` interfaces instead of direct `*sql.DB` and filesystem
2. **`handlers.go`** -- Use `Store` interface; add health check endpoint; add S3 proxy for media
3. **`server.go`** -- Register health check route; media routes proxy to MediaStore
4. **`imaging.go`** -- Return processed bytes instead of writing to filesystem; let MediaStore handle persistence
5. **`main.go`** -- Add `--migrate` flag; initialize Store/MediaStore based on config

## Implementation Order

1. Define interfaces (storage/db.go, storage/media.go)
2. Extract SQLite implementation (storage/sqlite.go from db.go)
3. Extract local filesystem implementation (storage/local.go from imaging.go)
4. Implement PostgreSQL store (storage/postgres.go)
5. Implement S3 media store (storage/s3.go)
6. Add configuration (config.go)
7. Wire interfaces into app.go + handlers.go
8. Implement migration tool (storage/migrate.go)
9. Create Dockerfile + docker-compose.yml
10. Add health check endpoint
11. Test with docker-compose
12. Update CLAUDE.md and README

## Acceptance Test Flow

### Local Mode (Regression)
1. Start without any DATABASE_URL or S3 config
2. Verify app works identically to pre-migration (SQLite + local files)
3. Create items, upload images, search, filter, export -- all work

### Cloud Mode (docker-compose)
4. `docker-compose up -d postgres minio` -- start infrastructure
5. Create MinIO bucket: `mc mb local/omnicollect`
6. Run migration: `go run . --migrate --source ~/.omnicollect/collection.db --tenant default`
7. Verify migration reports correct item/module counts
8. Start app with cloud config: `DATABASE_URL=... S3_ENDPOINT=... go run . --serve`
9. `curl http://localhost:8080/api/v1/health` -- verify database + storage connected
10. `curl http://localhost:8080/api/v1/modules` -- verify modules loaded from DB
11. `curl http://localhost:8080/api/v1/items` -- verify items queryable
12. Open browser, create new item, upload image -- verify stored in S3
13. Verify thumbnails load in grid view from S3 (via proxy)
14. Verify text search works (tsvector/tsquery)
15. Verify attribute filtering works (JSONB operators)

### Multi-Tenant Isolation
16. Create a second tenant schema
17. Insert items under tenant B
18. Query as tenant A -- verify tenant B items are invisible

### Container
19. `docker build -t omnicollect .` -- verify image builds
20. `docker-compose up` -- verify full stack starts
21. Open browser at localhost:8080 -- verify app works end-to-end through container
22. Verify container is stateless: restart it, verify data persists (in PG + S3)
