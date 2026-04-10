# Data Model: Insights Dashboard

**Branch**: `018-insights-dashboard` | **Date**: 2026-04-10

## Overview

No new database entities or backend changes. All data models are computed frontend types derived from the existing `Item[]` array in the collection store.

## Computed Types

### DashboardMetrics

Aggregated summary computed from all items in the store.

| Field | Type | Description |
|-------|------|-------------|
| totalValue | number | Sum of purchasePrice across all items (null prices treated as 0) |
| totalItems | number | Count of all items |
| mostValuableItem | { id: string, title: string, price: number } or null | Item with highest purchasePrice; null if no items have prices |
| moduleBreakdown | ModuleSegment[] | Per-module value distribution for doughnut chart |
| acquisitionTimeline | TimeBucket[] | Monthly item counts for bar chart |

### ModuleSegment

One slice of the doughnut chart representing a module (or "Other").

| Field | Type | Description |
|-------|------|-------------|
| moduleId | string | Module identifier (empty string for "Other" group) |
| moduleName | string | Display name from module schema (or "Other") |
| totalValue | number | Sum of purchasePrice for items in this module |
| itemCount | number | Count of items in this module |
| percentage | number | Percentage of total collection value (0-100) |

### TimeBucket

One bar in the acquisitions-over-time chart.

| Field | Type | Description |
|-------|------|-------------|
| label | string | Month label (e.g., "Jan 2026") |
| key | string | Sort key (e.g., "2026-01") for ordering |
| count | number | Number of items created in this month |

## Computation Rules

1. **totalValue**: `items.filter(i => i.purchasePrice != null).reduce((sum, i) => sum + i.purchasePrice, 0)`
2. **mostValuableItem**: Item with max purchasePrice; ties broken by first occurrence
3. **moduleBreakdown**: Group items by moduleId, sum purchasePrice per group. If more than 6 groups, keep top 5 by value, merge rest into "Other". Percentages computed from totalValue.
4. **acquisitionTimeline**: Parse createdAt to YYYY-MM, count items per month, sort chronologically. If span exceeds 24 months, group by quarter instead.

## Relationships to Existing Types

- **Item** (existing): Source data. Fields used: `id`, `title`, `purchasePrice`, `moduleId`, `createdAt`
- **ModuleSchema** (existing): Used for `displayName` lookup via `moduleStore.getModuleById()`
- No new database tables, columns, or migrations required
