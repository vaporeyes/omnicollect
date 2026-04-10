# Data Model: Public Showcase URLs

**Branch**: `017-public-showcase` | **Date**: 2026-04-10

## New Entity: Showcase

| Field | Type (Go) | SQLite | PostgreSQL | Description |
|-------|-----------|--------|------------|-------------|
| ID | string | TEXT PRIMARY KEY | TEXT PRIMARY KEY | UUID |
| Slug | string | TEXT NOT NULL UNIQUE | TEXT NOT NULL UNIQUE | URL slug (e.g., "rare-coins-a3f7b2c1") |
| TenantID | string | TEXT NOT NULL | (implicit, schema-per-tenant) | Owner tenant |
| ModuleID | string | TEXT NOT NULL | TEXT NOT NULL | Collection module being showcased |
| Enabled | bool | INTEGER NOT NULL DEFAULT 0 | BOOLEAN NOT NULL DEFAULT FALSE | Public visibility toggle |
| CreatedAt | string | TEXT NOT NULL | TIMESTAMPTZ NOT NULL | First time toggled public |
| UpdatedAt | string | TEXT NOT NULL | TIMESTAMPTZ NOT NULL | Last toggle change |

**Indexes**: `idx_showcases_slug ON showcases(slug)` for fast lookup by slug.

**Uniqueness**: One showcase per (tenant_id, module_id) pair. A collection can only have one showcase URL.

## Store Interface Additions

| Method | Signature | Description |
|--------|-----------|-------------|
| GetShowcaseBySlug | `(slug string) (*Showcase, error)` | Look up showcase by URL slug (cross-tenant for public access) |
| GetShowcaseForModule | `(moduleID string) (*Showcase, error)` | Get showcase for a module within current tenant |
| UpsertShowcase | `(showcase Showcase) error` | Create or update showcase record |
| ListShowcases | `() ([]Showcase, error)` | List all showcases for current tenant |

Note: `GetShowcaseBySlug` must query across all tenants (it's used by the public route which has no tenant context). In PostgreSQL, this means querying a shared `public.showcases` table or scanning tenant schemas.

**Decision**: Store showcases in a `public.showcases` table (not per-tenant schema) in PostgreSQL, since the public route needs to look up by slug without knowing the tenant. In SQLite, the showcases table is in the main database (single-tenant anyway).

## REST API Additions

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| /api/v1/showcases | GET | yes | List showcases for current tenant |
| /api/v1/showcases/toggle | POST | yes | Toggle a module public/private |
| /showcase/{slug} | GET | NO | Public gallery page (server-rendered HTML) |

## TypeScript Type

```typescript
interface Showcase {
  id: string
  slug: string
  moduleId: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}
```
