# Research: Cross-Collection Tags

**Branch**: `014-cross-collection-tags` | **Date**: 2026-04-08

## R1: Tag Storage Strategy

**Decision**: Store tags as a JSON array of strings directly on the item. In SQLite: `tags TEXT NOT NULL DEFAULT '[]'`. In PostgreSQL: `tags JSONB NOT NULL DEFAULT '[]'`.

**Rationale**: Aligns with Constitution Principle III (flat data, no JOINs). Tags are a simple list of strings with no metadata (no creation date, no color, no hierarchy). A JSON array is the most compact representation. No junction table means no additional queries or JOINs for basic item operations.

**Alternatives considered**:
- Junction table (item_tags): More normalized but violates Constitution III. Requires JOINs for every item query. Rejected.
- Comma-separated string: Harder to query (LIKE patterns). Rejected.
- Separate tags column per tag: Absurd. Rejected.

## R2: Tag Querying (SQLite)

**Decision**: Use `json_each()` to check if an item's tags array contains a specific value. For filtering by tag: `WHERE EXISTS (SELECT 1 FROM json_each(items.tags) WHERE value = ?)`.

**Rationale**: SQLite's `json_each()` is a virtual table function that iterates over JSON array elements. Combined with EXISTS, it efficiently checks membership without materializing the full array.

**For multiple tags (OR)**: `WHERE EXISTS (SELECT 1 FROM json_each(items.tags) WHERE value IN (?, ?, ...))`.

## R3: Tag Querying (PostgreSQL)

**Decision**: Use the `?|` operator for "contains any" check: `WHERE tags ?| array['gift', 'rare']`. This checks if the JSONB array contains any of the specified values.

**Rationale**: PostgreSQL's JSONB operators are purpose-built for this. The `?|` operator is efficient with GIN indexes.

**Index**: `CREATE INDEX idx_items_tags ON items USING GIN(tags)` for fast tag-based queries.

## R4: Tags in Full-Text Search

**Decision**: Include tag values in the search index so that searching for "gift" also finds items tagged "gift".

**SQLite**: Update the FTS5 insert/update triggers to concatenate tag values into the `attrs_text` field (or add a separate `tags_text` column to the FTS table).

**PostgreSQL**: Update the tsvector trigger to include tag values: `search_vector = to_tsvector(coalesce(title,'') || ' ' || coalesce(tags_as_text,'') || ...)`.

**Rationale**: Users expect search to find tagged items. Without this, a user searching for "gift" wouldn't find items only tagged (not titled) "gift".

## R5: Tag Autocomplete Strategy

**Decision**: Add a `GetAllTags()` method to the Store interface that returns a distinct list of all tag values used by the current tenant, with item counts.

**SQLite**: `SELECT DISTINCT value, COUNT(*) FROM items, json_each(items.tags) GROUP BY value ORDER BY value`.

**PostgreSQL**: `SELECT tag, COUNT(*) FROM items, jsonb_array_elements_text(tags) AS tag GROUP BY tag ORDER BY tag`.

**Rationale**: A single query returns all tags with counts. The frontend can filter locally for autocomplete (the total tag vocabulary is small -- typically under 100 distinct tags). No need for a server-side search endpoint.

## R6: Tag Rename/Delete Strategy

**Decision**: Tag rename and delete operate directly on the items table, updating the JSON array in all matching items.

**SQLite rename**: Update items where tags contain the old name, replacing it in the array. Use a combination of `json_remove` + `json_insert` or rebuild the array.

**PostgreSQL rename**: `UPDATE items SET tags = (tags - 'old_name') || '"new_name"' WHERE tags ? 'old_name'`.

**Delete**: `UPDATE items SET tags = tags - 'old_name' WHERE tags ? 'old_name'` (PostgreSQL) or equivalent SQLite.

**Rationale**: Direct array manipulation avoids the need for a separate tags table. Operations are atomic per-row and work within a single transaction.

## R7: Tag Filter UI Placement

**Decision**: The tag filter is a separate row below the existing filter bar (or above the collection views), consisting of clickable tag chips. It's distinct from the schema-driven faceted filter bar.

**Rationale**: Tags are universal (not schema-specific). The faceted filter bar only appears when a specific module is selected and is driven by that module's schema. Tags should be visible and filterable even when "All Types" is selected.
