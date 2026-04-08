# Research: Cloud Infrastructure Migration

**Branch**: `011-cloud-infrastructure` | **Date**: 2026-04-08

## R1: PostgreSQL Schema-Per-Tenant Pattern

**Decision**: Each tenant gets a PostgreSQL schema (e.g., `tenant_abc123`). On each request, the connection sets `SET search_path TO tenant_{id}, public` before executing queries. Tables within each schema mirror the current SQLite structure.

**Rationale**: Per clarification session, schema-per-tenant was chosen. Benefits: strong isolation (no tenant_id columns needed), queries remain nearly identical to SQLite, easy to drop a tenant (DROP SCHEMA CASCADE), compatible with connection pooling (pgbouncer can set search_path per transaction).

**Tenant provisioning**: A `CREATE SCHEMA tenant_{id}` followed by running the DDL statements (same CREATE TABLE/INDEX as current SQLite, adapted for PostgreSQL syntax) within that schema.

**Alternatives considered**:
- Row-level tenant_id: Rejected -- requires modifying every query; risk of data leaks if WHERE clause omitted.
- Turso per-user databases: Rejected -- newer technology, less ecosystem support.

## R2: PostgreSQL Full-Text Search (Replacing FTS5)

**Decision**: Per clarification, use PostgreSQL built-in tsvector/tsquery with GIN index.

**Implementation**: Add a `search_vector tsvector` column to the `items` table. Create a GIN index on it. Use a trigger (or computed column) that updates the vector on INSERT/UPDATE by concatenating `title` and extracted attribute text values. Queries use `WHERE search_vector @@ plainto_tsquery(?)` with `ts_rank` for ordering.

**Rationale**: Maps directly to the current FTS5 pattern (separate search index updated by triggers). PostgreSQL's full-text search supports prefix matching, ranking, and language-aware stemming.

**Key differences from SQLite FTS5**:
- No separate virtual table; tsvector is a column on the items table
- Triggers update the column instead of a separate table
- `MATCH` syntax becomes `@@` operator
- `rank` column becomes `ts_rank()` function

## R3: PostgreSQL JSONB for Attributes

**Decision**: Use `JSONB` column type for the `attributes` field. Replace SQLite `json_extract()` calls with PostgreSQL `->>` and `->` operators.

**Rationale**: JSONB is PostgreSQL's native binary JSON type. It supports indexing (GIN), efficient extraction, and the same flat-data pattern as SQLite's JSON functions.

**Query translation**:
- SQLite: `json_extract(attributes, '$.condition')` -> PostgreSQL: `attributes->>'condition'`
- SQLite: `json_extract(attributes, '$.year') >= ?` -> PostgreSQL: `(attributes->>'year')::numeric >= ?`

## R4: S3-Compatible Object Storage

**Decision**: Use the AWS SDK v2 for Go with configurable endpoint URL. This works with AWS S3, Cloudflare R2, Google Cloud Storage (via S3-compatible API), and MinIO for local development.

**Image flow**:
1. Upload handler receives multipart file
2. Process image in memory (validate, generate thumbnail)
3. Upload original to `s3://{bucket}/originals/{filename}`
4. Upload thumbnail to `s3://{bucket}/thumbnails/{filename}`
5. Return the filename (frontend constructs URL from configured base)

**Frontend image URLs**: In cloud mode, images are served either via presigned S3 URLs or via a proxy endpoint on the backend (`/thumbnails/{filename}` proxies to S3). The proxy approach keeps URLs simple and avoids CORS/presigning complexity.

**Alternatives considered**:
- Direct S3 presigned URLs: More complex (frontend needs to know bucket URL, presigning logic). Rejected for v1.
- Backend proxy only: Simpler but adds latency. Acceptable for v1; can optimize to direct URLs later.

## R5: Docker Multi-Stage Build

**Decision**: Three-stage Dockerfile:
1. **Go builder**: `golang:1.25-alpine` -- builds the Go binary
2. **Node builder**: `node:18-alpine` -- builds the Vue frontend (`npm run build`)
3. **Runtime**: `alpine:3.19` -- copies binary + dist, runs the server

**Final image size**: ~30-50MB (Go binary + frontend assets + alpine base).

**Rationale**: Multi-stage keeps the final image minimal. Alpine base is standard for Go services. No development tools or source code in the runtime image.

## R6: Configuration Strategy

**Decision**: 12-factor environment variables with sensible defaults for local mode.

| Variable | Default | Cloud Example |
|----------|---------|---------------|
| DATABASE_URL | (empty = use local SQLite) | postgres://user:pass@host:5432/omnicollect |
| S3_ENDPOINT | (empty = use local filesystem) | https://s3.amazonaws.com |
| S3_BUCKET | (empty) | omnicollect-media |
| S3_ACCESS_KEY | (empty) | AKIA... |
| S3_SECRET_KEY | (empty) | secret... |
| S3_REGION | us-east-1 | us-east-1 |
| PORT | 8080 | 8080 |
| TENANT_ID | default | (set per-request by auth middleware in future) |

**Rationale**: Empty DATABASE_URL/S3_ENDPOINT triggers local mode. This preserves backward compatibility. Cloud mode requires both to be set.

## R7: Module Schema Storage in PostgreSQL

**Decision**: Per clarification, module schemas are stored in a `modules` table within each tenant's schema.

**Table schema**:
```sql
CREATE TABLE modules (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  description TEXT,
  schema_json JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

The `schema_json` column stores the full module definition (including attributes array). `GetActiveModules` queries this table. `SaveCustomModule` upserts into it.

**Rationale**: Same data, different storage location. The Schema Builder UI and all module-related APIs work unchanged -- only the storage backend switches.

## R8: Data Migration Tool

**Decision**: A CLI command (`go run . --migrate --source /path/to/collection.db --tenant tenant_id`) that reads the SQLite database and writes all items, modules, and settings into the PostgreSQL tenant schema.

**Steps**:
1. Open source SQLite database
2. Connect to PostgreSQL (from DATABASE_URL)
3. Create tenant schema if not exists
4. Run DDL within tenant schema
5. Copy all items (adapting JSON to JSONB)
6. Copy all module schemas from `~/.omnicollect/modules/*.json` files into the modules table
7. Report counts (items migrated, modules migrated, errors)

**Rationale**: One-time migration tool. Not a runtime feature -- used during initial cloud deployment setup.
