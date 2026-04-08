# Feature Specification: Cross-Collection Tags

**Feature Branch**: `014-cross-collection-tags`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Add a lightweight tags system that lets users slice across collections -- 'show me everything tagged gift' regardless of whether it's a coin or a book."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add and Remove Tags on Items (Priority: P1)

A collector viewing or editing an item wants to assign free-form tags to it (e.g., "gift", "rare", "for-sale"). In the item detail view and edit form, a tag input area lets them type a tag name and press Enter to add it. Existing tags display as removable chips. Tags are saved as part of the item and persist across sessions. An item can have multiple tags, and the same tag label can be used across items in different modules.

**Why this priority**: Tag assignment is the foundation. Without the ability to add tags to items, there is nothing to filter by. This story delivers the core data model and UI for tagging.

**Independent Test**: Open an item, add tags "gift" and "rare", save. Reopen the item, verify both tags are present. Remove "rare", save, verify only "gift" remains.

**Acceptance Scenarios**:

1. **Given** an item in the detail or edit view, **When** the user types a tag name and presses Enter, **Then** the tag is added and displayed as a chip on the item.
2. **Given** an item with tags, **When** the user clicks the remove button on a tag chip, **Then** the tag is removed from the item.
3. **Given** an item with tags, **When** the item is saved, **Then** the tags persist and are visible when the item is reopened.
4. **Given** two items in different modules, **When** the user adds the tag "gift" to both, **Then** both items have the "gift" tag independently.
5. **Given** the tag input, **When** the user types a tag that already exists on the item, **Then** the duplicate is ignored (no duplicate tags per item).

---

### User Story 2 - Filter Items by Tag Across All Modules (Priority: P2)

A collector wants to find all items tagged "gift" regardless of which collection type they belong to. In the collection views (list and grid), a tag filter control lets the user select one or more tags. When a tag filter is active, only items with the matching tag(s) are shown -- spanning across all modules. Tag filtering combines with existing text search and module filters.

**Why this priority**: Cross-collection filtering is the unique value of tags. Without it, tags are just metadata with no discoverability. This story delivers the "show me everything tagged X" capability.

**Independent Test**: Tag items from two different modules with "gift". Activate the "gift" tag filter in the collection view (with "All Types" selected). Verify items from both modules appear.

**Acceptance Scenarios**:

1. **Given** items in multiple modules tagged with "gift", **When** the user selects the "gift" tag filter with "All Types" active, **Then** all items with that tag are shown regardless of module.
2. **Given** a tag filter is active, **When** the user also applies a text search, **Then** both filters combine (AND logic -- items must match both the tag and the search query).
3. **Given** a tag filter is active along with a module filter, **Then** only items in that module with the matching tag are shown.
4. **Given** no items have the selected tag, **Then** a "No items match" message is displayed.
5. **Given** the user selects multiple tags, **Then** items matching any of the selected tags are shown (OR logic within tag filter).
6. **Given** a tag filter is active, **When** the user clears it, **Then** all items are shown again.

---

### User Story 3 - Tag Autocomplete and Management (Priority: P3)

A collector wants to reuse existing tags rather than retyping them. The tag input offers autocomplete suggestions based on tags already used across their collection. A tag management view (or section in settings) shows all existing tags with item counts, allowing the user to rename or delete tags in bulk.

**Why this priority**: Autocomplete and management prevent tag sprawl (e.g., "gift" vs "Gift" vs "gifts"). This is a quality-of-life feature that becomes important as the tag vocabulary grows.

**Independent Test**: Type "gi" in the tag input, verify "gift" appears as a suggestion. Select it. Open tag management, rename "gift" to "presents", verify all items previously tagged "gift" now show "presents".

**Acceptance Scenarios**:

1. **Given** existing tags in the collection, **When** the user starts typing in the tag input, **Then** matching existing tags appear as autocomplete suggestions.
2. **Given** autocomplete suggestions, **When** the user selects one, **Then** the tag is added to the item.
3. **Given** the tag management view, **When** the user renames a tag, **Then** all items with the old tag name are updated to the new name.
4. **Given** the tag management view, **When** the user deletes a tag, **Then** the tag is removed from all items that had it.
5. **Given** the tag management view, **Then** each tag shows a count of how many items use it.

---

### Edge Cases

- What happens when a tag name contains special characters (commas, quotes, spaces)? Tags are trimmed of leading/trailing whitespace. Internal spaces are allowed. Tags are case-insensitive for matching (stored lowercase).
- What happens when an item is deleted that has tags? The tags are removed with the item. If no other items use that tag, it disappears from autocomplete and management.
- What happens with very long tag names? Tag names are limited to 50 characters.
- What happens when tags are used in CSV export? Tags should be included as a comma-separated list in a "tags" column.
- What happens in the command palette? Tags should be searchable in the command palette alongside item titles.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each item MUST support zero or more tags, stored as an ordered list of strings.
- **FR-002**: Tags MUST be free-form text labels, case-insensitive (stored and compared in lowercase), trimmed, and limited to 50 characters.
- **FR-003**: The item edit form and detail view MUST display a tag input area with add/remove capability.
- **FR-004**: Adding a tag that already exists on the item MUST be a no-op (no duplicates per item).
- **FR-005**: The collection views (list and grid) MUST provide a tag filter that shows items matching selected tags across all modules.
- **FR-006**: Multiple selected tags MUST use OR logic (show items with any of the selected tags).
- **FR-007**: Tag filtering MUST combine with text search, module filter, and attribute filters using AND logic.
- **FR-008**: The backend item query MUST support filtering by tags alongside existing query parameters.
- **FR-009**: The tag input MUST offer autocomplete suggestions from all tags used in the user's collection.
- **FR-010**: A tag management capability MUST allow users to see all tags with item counts, rename tags (updating all items), and delete tags (removing from all items).
- **FR-011**: Tags MUST be included in CSV export as a comma-separated "tags" column.
- **FR-012**: Tags MUST be searchable in the command palette alongside item titles.
- **FR-013**: Tag data MUST be stored as part of the item (not in a separate junction table) to maintain the flat data architecture.

### Key Entities

- **Tag**: A free-form text label attached to an item. Case-insensitive, max 50 characters, unique per item but shared across items and modules.
- **Tag Filter**: A user-applied filter in the collection view that shows only items containing the selected tag(s). Works across all modules.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can add a tag to an item and find it via tag filter in under 5 seconds.
- **SC-002**: Tag filtering across all modules returns correct results -- zero items are missed and zero false positives appear.
- **SC-003**: Autocomplete suggestions appear within 200ms of typing in the tag input.
- **SC-004**: Tag rename in management updates 100% of affected items atomically.
- **SC-005**: Tags are preserved across all item operations (create, update, delete, export, backup) with zero data loss.

## Assumptions

- Tags are stored as a JSON array of strings on the item itself (e.g., `"tags": ["gift", "rare"]`). This aligns with Constitution Principle III (flat data architecture, no junction tables). In SQLite, the `tags` field is a TEXT column containing a JSON array. In PostgreSQL, it's a JSONB array.
- Tag names are normalized to lowercase on save. Display can show original casing, but matching is always case-insensitive.
- The tag filter in the collection view is a separate control from the faceted filter bar (which is schema-driven). Tags are universal across all modules, not tied to a schema.
- Tag autocomplete queries the backend for a distinct list of all tags used by the current user/tenant.
- Tag management is a lightweight section (possibly in settings or a dedicated view), not a full CRUD page. Rename and delete are the only management operations.
- The existing `GetItems` query is extended with an optional `tags` parameter (JSON array of tags to filter by). The backend filters items whose `tags` array contains any of the specified tags.
- Tags are included in the FTS5/tsvector search index so that searching for "gift" also matches items tagged "gift" even if the word doesn't appear in the title or attributes.
