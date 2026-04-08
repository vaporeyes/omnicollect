# Tags Contract: REST API + UI Components

**Branch**: `014-cross-collection-tags` | **Date**: 2026-04-08

## REST API

### List All Tags
`GET /api/v1/tags`

**Response**: `200 OK` -- `TagCount[]`
```json
[
  {"name": "gift", "count": 12},
  {"name": "rare", "count": 5}
]
```

### Rename Tag
`POST /api/v1/tags/rename`

**Request Body**: `{"oldName": "gift", "newName": "presents"}`

**Response**: `200 OK` -- `{"updated": 12}`

### Delete Tag
`DELETE /api/v1/tags/{name}`

**Response**: `200 OK` -- `{"updated": 5}`

### Modified: List Items with Tag Filter
`GET /api/v1/items?tags=["gift","rare"]`

The `tags` query param is a JSON array of tag names. Items matching ANY of the specified tags are returned (OR logic). Combines with existing `query`, `moduleId`, `filters` params via AND.

## UI Component Contracts

### TagInput Component

**Props**:
| Prop | Type | Description |
|------|------|-------------|
| modelValue | string[] | Current tags (v-model) |
| allTags | TagCount[] | Available tags for autocomplete |

**Emits**:
| Event | Payload | Description |
|-------|---------|-------------|
| update:modelValue | string[] | Tags changed (add/remove) |

### TagFilter Component

**Props**:
| Prop | Type | Description |
|------|------|-------------|
| allTags | TagCount[] | Available tags to filter by |
| selectedTags | string[] | Currently active tag filters |

**Emits**:
| Event | Payload | Description |
|-------|---------|-------------|
| update | string[] | Selected tags changed |

### TagManager Component

**Props**:
| Prop | Type | Description |
|------|------|-------------|
| tags | TagCount[] | All tags with counts |

**Emits**:
| Event | Payload | Description |
|-------|---------|-------------|
| rename | {oldName: string, newName: string} | User renamed a tag |
| delete | string | User deleted a tag |
