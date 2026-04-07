# Data Model: Command Palette

**Branch**: `006-command-palette` | **Date**: 2026-04-07

## Existing Entities (no changes)

### Item

The command palette reads existing `Item` entities. No schema changes are needed.

| Field | Type | Notes |
|-------|------|-------|
| id | string (UUID) | Primary key |
| moduleId | string | Links to ModuleSchema |
| title | string | Displayed in palette results; searched via FTS5 |
| purchasePrice | number (nullable) | Not displayed in palette |
| images | string[] | First entry used for thumbnail in results |
| attributes | Record<string, any> | Searched via FTS5 (attrs_text) |
| createdAt | string (RFC3339) | Not displayed in palette |
| updatedAt | string (RFC3339) | Not displayed in palette |

### ModuleSchema

Used to resolve the module display name for the badge in palette results.

| Field | Type | Notes |
|-------|------|-------|
| id | string | Module identifier |
| displayName | string | Shown as badge text in palette results |

## New Frontend-Only Types

### QuickAction

A predefined application action surfaced by keyword matching. Exists only in the Vue component, not persisted.

| Field | Type | Notes |
|-------|------|-------|
| label | string | Display text (e.g., "Add New Item") |
| keywords | string[] | Trigger words (e.g., ["new", "add", "create"]) |
| action | string | Action identifier dispatched on selection |

### PaletteResult

A union type representing a single row in the palette results list.

| Variant | Fields | Notes |
|---------|--------|-------|
| QuickAction | label, action | Rendered with an action icon, no thumbnail |
| Item | item (Item), moduleName (string) | Rendered with thumbnail + title + module badge |

## State Management

The palette manages its own local state (open/closed, query, highlighted index, results). It does not add persistent state to any Pinia store. The only store interaction is calling `GetItems` for cross-module search.

## No Database Changes

The command palette is read-only. It queries existing data via the existing `GetItems` backend binding. No new tables, columns, indexes, or migrations are needed.
