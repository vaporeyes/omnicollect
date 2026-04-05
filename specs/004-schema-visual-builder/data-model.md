# Data Model: Schema Visual Builder

**Date**: 2026-04-05
**Feature**: 004-schema-visual-builder

## Entities

### Draft Schema (in-memory reactive state)

The central reactive state shared between the visual builder and code
editor panes. Not persisted until the user explicitly saves.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Module identifier (auto-slugified from displayName) |
| displayName | string | Human-readable collection type name |
| description | string | Optional description |
| attributes | DraftAttribute[] | Ordered list of field definitions |

### DraftAttribute (nested in Draft Schema)

| Field | Type | Description |
|-------|------|-------------|
| name | string | Attribute key (unique within schema) |
| type | string | "string", "number", "boolean", "date", "enum" |
| required | boolean | Whether the field is mandatory |
| options | string[] | Enum options (only for type "enum") |
| display | DraftDisplayHints | Optional rendering hints |

### DraftDisplayHints (nested in DraftAttribute)

| Field | Type | Description |
|-------|------|-------------|
| label | string | Override display label |
| placeholder | string | Input placeholder text |
| widget | string | Widget hint: "text", "textarea", "dropdown" |
| group | string | Group name for form sectioning |
| order | number | Sort priority |

## State Transitions

```
[Empty] ---(new)--> [Draft] ---(edit)--> [Draft]
                        |                    |
                    (save)               (save)
                        |                    |
                        v                    v
                   [Saved to disk]    [Updated on disk]
                        |
                  (reload modules)
                        |
                        v
                [Available in app]
```

## Bidirectional Sync

The draft schema exists simultaneously as:
1. A structured object (drives visual builder)
2. A JSON string (drives code editor)

**Object-to-JSON**: On any visual builder change, serialize the draft
object to formatted JSON and update the code editor content.

**JSON-to-Object**: On any code editor change, attempt `JSON.parse()`.
If successful, update the draft object. If parsing fails (mid-typing),
retain the last valid object state and show an error indicator.

## Validation Rules (pre-save)

- `displayName` MUST be non-empty.
- `id` MUST be non-empty (auto-generated or user-provided).
- `id` MUST NOT conflict with a different existing schema file
  (same ID in a different file).
- Each attribute `name` MUST be non-empty and unique within the schema.
- Each attribute `type` MUST be one of: "string", "number", "boolean",
  "date", "enum".
- Enum attributes MUST have at least one option.
- The serialized JSON MUST be valid (parseable).

## Slug Generation

Auto-generate `id` from `displayName`:
- Lowercase
- Replace spaces/special characters with hyphens
- Remove consecutive hyphens
- Example: "Vinyl Records" -> "vinyl-records"
