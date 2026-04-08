# Feature Specification: Cloud Infrastructure Migration

**Feature Branch**: `011-cloud-infrastructure`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Migrate from local SQLite and filesystem to cloud infrastructure: database pivot (PostgreSQL or Turso/LibSQL), S3-compatible object storage for images, Docker containerization."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Database Migrates to Cloud with Multi-Tenant Support (Priority: P1)

The application's data storage moves from a local SQLite file to a cloud-hosted database that supports multiple users. Each user's collections, items, and module schemas are isolated -- a user can only query and modify their own data. All existing queries (text search, attribute filtering, JSON attribute access) continue to work against the new database. Existing single-user data can be migrated to the multi-tenant schema.

**Why this priority**: The database is the foundation of all operations. Without cloud database support, no other cloud feature (object storage, containerization) can serve multiple users. This is the highest-risk, highest-value change.

**Independent Test**: Deploy the application with the cloud database. Create two user accounts. Verify that items created by user A are invisible to user B. Verify text search, attribute filtering, and all CRUD operations work identically to the current SQLite-based app.

**Acceptance Scenarios**:

1. **Given** the application is configured for the cloud database, **When** a user creates an item, **Then** the item is stored with the user's tenant identifier and is only visible to that user.
2. **Given** existing single-user data, **When** a migration script is run, **Then** all items, modules, and settings are imported into the cloud database under a designated tenant.
3. **Given** two users with separate collections, **When** user A searches for items, **Then** only user A's items are returned -- never user B's.
4. **Given** items with JSON attributes, **When** a user applies faceted filters, **Then** attribute filtering works identically to the current implementation.
5. **Given** the cloud database is temporarily unreachable, **When** the application attempts a query, **Then** it returns a clear error message rather than silently failing.

---

### User Story 2 - Images Stored in Cloud Object Storage (Priority: P2)

Uploaded images (originals and thumbnails) are stored in a cloud object store instead of the local filesystem. The image upload flow generates a thumbnail, uploads both original and thumbnail to the object store, and returns a URL. The frontend loads images from the object store URL instead of the local `/thumbnails/` and `/originals/` paths.

**Why this priority**: Image storage is the second pillar of the cloud migration. Without it, the application cannot scale beyond a single server's filesystem. This is independent of the database migration and can be tested separately.

**Independent Test**: Upload an image through the application. Verify the original and thumbnail are stored in the cloud object store. Verify the image displays correctly in the grid, detail, and lightbox views via the object store URL.

**Acceptance Scenarios**:

1. **Given** the application is configured for cloud object storage, **When** a user uploads an image, **Then** the original and thumbnail are stored in the object store (not the local filesystem).
2. **Given** images in the object store, **When** the grid view loads, **Then** thumbnail images load from the object store URL and display correctly.
3. **Given** an item with images, **When** the user opens the detail view or lightbox, **Then** the original image loads from the object store URL.
4. **Given** the object store is temporarily unreachable, **When** an image fails to load, **Then** the existing placeholder is shown (graceful degradation).
5. **Given** a user deletes an item with images, **Then** the images remain in the object store (deletion is not required in v1; storage cleanup can be added later).

---

### User Story 3 - Application Packaged as a Container (Priority: P3)

The Go backend and compiled Vue frontend are packaged into a single container image. The container can be started with environment variables for database connection, object store credentials, and server port. A developer or operator can pull the image and run it with `docker run` plus the required configuration.

**Why this priority**: Containerization is the deployment wrapper. It depends on the database and storage migrations being complete (the container needs to connect to external services). It is the lowest-risk change -- it packages what already works.

**Independent Test**: Build the container image. Run it with the required environment variables pointing to a cloud database and object store. Verify the application starts, the frontend loads, and all features work through the container.

**Acceptance Scenarios**:

1. **Given** the container image is built, **When** it is started with valid database and object store configuration, **Then** the application starts and serves the frontend at the configured port.
2. **Given** a running container, **When** a user performs any operation (CRUD, search, filter, image upload, export), **Then** it works identically to the non-containerized version.
3. **Given** missing or invalid configuration (e.g., bad database URL), **When** the container starts, **Then** it exits with a clear error message indicating which configuration is missing.
4. **Given** the container, **Then** it exposes a single HTTP port and requires no volume mounts for operation (all state is in the cloud database and object store).

---

### Edge Cases

- What happens when the database migration encounters items with malformed JSON attributes? The migration script should log a warning and skip the malformed item, reporting a count of skipped items at the end.
- What happens when object store upload fails mid-request (e.g., network timeout)? The image upload endpoint should return an error; the item should not reference a non-existent image.
- What happens to the desktop (Wails) mode after this migration? The desktop mode should continue working with local SQLite and local filesystem as a fallback when cloud configuration is not provided. Cloud infrastructure is for the server deployment path.
- What happens to existing backup export functionality? Backup export should work against whatever database is configured (cloud or local). The backup ZIP should still include all data.
- What happens when the container is restarted? Since all state is external (database + object store), the container is stateless and can restart without data loss.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST support connecting to a cloud-hosted database as an alternative to local SQLite, configurable via environment variables.
- **FR-002**: The cloud database MUST use a schema-per-tenant approach where each user gets their own isolated database schema. Queries operate within a tenant's schema without needing a tenant_id column in every WHERE clause. Tenant provisioning creates a new schema with the standard table structure.
- **FR-003**: All existing queries (text search, attribute filtering, CRUD, batch operations) MUST work against the cloud database with identical behavior. Text search MUST use the database's native full-text search capability (replacing the current FTS5 virtual table) with an indexed search column for performant queries.
- **FR-004**: A data migration tool MUST exist to import existing single-user SQLite data into the cloud database under a specified tenant.
- **FR-005**: The application MUST support storing images in an S3-compatible object store as an alternative to the local filesystem, configurable via environment variables.
- **FR-006**: Image upload MUST generate thumbnails locally (in memory or temp file), then upload both original and thumbnail to the object store.
- **FR-007**: The frontend MUST load images from the object store URLs when cloud storage is configured, and from local paths when it is not.
- **FR-008**: The application MUST be packageable as a single container image containing the Go backend and compiled Vue frontend.
- **FR-009**: The container MUST accept all external service configuration (database URL, object store credentials, server port) via environment variables.
- **FR-010**: The container MUST exit with a clear error if required configuration is missing at startup.
- **FR-011**: The application MUST continue to work with local SQLite and local filesystem when cloud configuration is not provided (backward compatibility for desktop mode).
- **FR-012**: The container MUST be stateless -- no local volume mounts required for operation.

### Key Entities

- **Tenant**: An isolated user context in the multi-tenant database. All items, modules, and settings belong to exactly one tenant.
- **Storage Backend**: An abstraction over local filesystem and cloud object store. Provides upload, URL generation, and health check operations.
- **Container Image**: A self-contained deployment artifact with the application binary, frontend assets, and a health check endpoint.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The application operates against the cloud database with zero behavioral regressions compared to local SQLite (all CRUD, search, filter, export operations work identically).
- **SC-002**: Images upload to and load from the cloud object store with no user-visible latency increase greater than 500ms compared to local filesystem.
- **SC-003**: The container image builds in under 5 minutes and starts serving requests in under 10 seconds.
- **SC-004**: Two distinct users can operate simultaneously against the same deployment without seeing each other's data.
- **SC-005**: Existing single-user data can be migrated to the cloud database with zero data loss (item count matches before and after migration).
- **SC-006**: The application still works in desktop/local mode when cloud configuration is absent.

## Clarifications

### Session 2026-04-08

- Q: How should full-text search work in PostgreSQL (replacing SQLite FTS5)? -> A: Use PostgreSQL built-in full-text search with tsvector/tsquery and GIN index.
- Q: Where should module schemas be stored in cloud mode (no local filesystem)? -> A: Database table -- store as rows in a modules table within each tenant's schema.

## Assumptions

- The database uses PostgreSQL with a schema-per-tenant isolation model. Each tenant gets their own PostgreSQL schema containing the standard tables (items, items_fts equivalent). Queries are scoped to the tenant's schema via `SET search_path` at connection time. This keeps queries close to the current SQLite form (no tenant_id columns needed in WHERE clauses) while providing strong isolation.
- S3-compatible object storage is used (works with AWS S3, Cloudflare R2, Google Cloud Storage, and MinIO for local development). The specific provider is configurable via endpoint URL + credentials.
- Authentication and user management are out of scope for this feature. The tenant identifier is provided as a configuration value or header. A future auth feature will handle user registration, login, and tenant provisioning.
- The Dockerfile uses a multi-stage build: Go build stage produces the binary, Node stage builds the frontend, final stage is a minimal runtime image.
- Environment variables for configuration follow the standard 12-factor app pattern (DATABASE_URL, S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, PORT).
- Local development can use MinIO (local S3-compatible storage) and local SQLite to avoid cloud dependencies during development.
- Module schemas in the cloud deployment are stored as rows in a `modules` table within each tenant's database schema (not as JSON files on disk), since the container is stateless. The `SaveCustomModule` and `GetActiveModules` operations read/write this table instead of the filesystem.
