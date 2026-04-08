# Infrastructure Contract: Cloud Deployment

**Branch**: `011-cloud-infrastructure` | **Date**: 2026-04-08

## Environment Variables

| Variable | Required (Cloud) | Default | Description |
|----------|-----------------|---------|-------------|
| DATABASE_URL | yes | (empty = local SQLite) | PostgreSQL connection string |
| S3_ENDPOINT | yes | (empty = local filesystem) | S3-compatible endpoint URL |
| S3_BUCKET | yes | (empty) | Bucket name for media storage |
| S3_ACCESS_KEY | yes | (empty) | S3 access key |
| S3_SECRET_KEY | yes | (empty) | S3 secret key |
| S3_REGION | no | us-east-1 | S3 region |
| PORT | no | 8080 | HTTP server listen port |

**Mode detection**: If `DATABASE_URL` is empty, use local SQLite. If `S3_ENDPOINT` is empty, use local filesystem. Both can be mixed (e.g., cloud DB + local files for testing).

## Docker Image

### Build

```
docker build -t omnicollect .
```

### Run

```
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/omnicollect \
  -e S3_ENDPOINT=https://s3.amazonaws.com \
  -e S3_BUCKET=omnicollect-media \
  -e S3_ACCESS_KEY=AKIA... \
  -e S3_SECRET_KEY=secret... \
  omnicollect
```

### Health Check

`GET /api/v1/health` -- returns `{"status": "ok", "database": "connected", "storage": "connected"}`

## Docker Compose (Development)

```yaml
services:
  app:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://omni:omni@postgres:5432/omnicollect
      S3_ENDPOINT: http://minio:9000
      S3_BUCKET: omnicollect
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
    depends_on: [postgres, minio]

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: omni
      POSTGRES_PASSWORD: omni
      POSTGRES_DB: omnicollect
    ports: ["5432:5432"]

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports: ["9000:9000", "9001:9001"]
```

## Migration CLI

```
go run . --migrate --source ~/.omnicollect/collection.db --tenant default
```

Reads SQLite database and writes to PostgreSQL (from DATABASE_URL). Creates tenant schema if needed. Reports item/module/error counts.

## Tenant Schema Provisioning

When a new tenant is created:
1. `CREATE SCHEMA IF NOT EXISTS tenant_{id}`
2. `SET search_path TO tenant_{id}`
3. Run DDL (CREATE TABLE items, modules, settings + indexes + triggers)

## REST API Changes

No endpoint changes. All existing endpoints work identically. The backend transparently uses the configured database and storage backends.

One new endpoint:
- `GET /api/v1/health` -- returns database and storage connectivity status
