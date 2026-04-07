# Feature Specification: Schema-Driven Faceted Filtering

**Feature Branch**: `007-faceted-filtering`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Upgrade ItemList and CollectionGrid with faceted filtering. Dynamically generate filter bar from module schema for enum, boolean, and number fields. Update backend query to accept attribute filters combined with FTS5 search."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Filter Items by Enum Attributes (Priority: P1)

A collector with hundreds of coins selects the "Coins" collection type in the sidebar. A filter bar appears showing interactive pills for each enum attribute (e.g., Condition). Clicking "Mint" shows only items where Condition equals "Mint". Clicking a second value (e.g., "Poor") adds it as an OR filter, showing items matching either value. Clicking an active pill deselects it. The filter combines with any active text search.

**Why this priority**: Enum filtering is the most common and highest-value faceted search use case. Most collection schemas define enum fields (condition, category, rarity) that users frequently want to narrow by.

**Independent Test**: Select a module with enum attributes, verify filter pills appear, click a pill, confirm only matching items display. Click again to deselect, confirm all items return.

**Acceptance Scenarios**:

1. **Given** a module with enum attributes is selected, **When** the filter bar renders, **Then** each enum attribute appears as a group of selectable pills, one per option value.
2. **Given** no pills are selected, **When** the user clicks an enum pill (e.g., "Condition: Mint"), **Then** the item list updates to show only items where that attribute matches the selected value.
3. **Given** one enum pill is active, **When** the user clicks a second pill in the same attribute group, **Then** both values are included (OR logic within the same attribute).
4. **Given** an active pill, **When** the user clicks it again, **Then** it deselects and the filter removes that value.
5. **Given** a text search is active, **When** the user also selects a filter pill, **Then** both the text search and attribute filter apply simultaneously (AND logic between text search and attribute filters).
6. **Given** the user switches to a different module, **When** the filter bar re-renders, **Then** all previous filter selections are cleared and new pills reflect the new module's schema.

---

### User Story 2 - Filter Items by Boolean Attributes (Priority: P2)

A collector wants to filter items by boolean fields (e.g., "Is Graded", "For Sale"). The filter bar shows toggle pills for each boolean attribute. Clicking a boolean pill filters to items where that attribute is true. Clicking again removes the filter.

**Why this priority**: Boolean filters are simple and common. They complement enum filters to cover the majority of schema attribute types.

**Independent Test**: Select a module with boolean attributes, verify toggle pills appear, click one, confirm only items with that attribute set to true are shown.

**Acceptance Scenarios**:

1. **Given** a module with boolean attributes is selected, **When** the filter bar renders, **Then** each boolean attribute appears as a single toggle pill (e.g., "Is Graded").
2. **Given** no boolean pills are active, **When** the user clicks a boolean pill once, **Then** only items where that attribute is true are shown (pill shows "Yes" state).
3. **Given** a boolean pill is in "Yes" state, **When** the user clicks it again, **Then** only items where that attribute is false are shown (pill shows "No" state).
4. **Given** a boolean pill is in "No" state, **When** the user clicks it again, **Then** the filter is removed and all items are shown (pill returns to off state).
5. **Given** both an enum filter and a boolean filter are active, **Then** items must match both filters (AND logic between different attributes).

---

### User Story 3 - Filter Items by Number Range (Priority: P3)

A collector wants to filter coins by mint year (e.g., 1900-1950) or by purchase price range. For number-type attributes, the filter bar shows a range control (min/max inputs). Adjusting the range filters items to those within the specified bounds.

**Why this priority**: Number range filtering is powerful but more complex to implement (requires range inputs and backend range query logic). Still valuable for attributes like year, price, and quantity.

**Independent Test**: Select a module with number attributes, verify range controls appear, set a min/max range, confirm only items within that range are shown.

**Acceptance Scenarios**:

1. **Given** a module with number attributes is selected, **When** the filter bar renders, **Then** each number attribute shows a control with min and max inputs.
2. **Given** the user enters a minimum value for a number attribute, **Then** only items where that attribute is greater than or equal to the minimum are shown.
3. **Given** the user enters a maximum value, **Then** only items where that attribute is less than or equal to the maximum are shown.
4. **Given** both min and max are set, **Then** only items within the range (inclusive) are shown.
5. **Given** a number range filter is active along with enum/boolean filters, **Then** all filters combine with AND logic.
6. **Given** the user clears both min and max inputs, **Then** the number filter is removed.

---

### Edge Cases

- What happens when a module has no filterable attributes (no enum, boolean, or number fields)? The filter bar should not appear.
- What happens when all items are filtered out? A "No items match filters" message should display, distinct from the "collection is empty" state.
- What happens when the user switches from a filtered module to "All Types"? Faceted filters should be hidden since there is no single schema to derive facets from.
- What happens when an item's attribute is null/missing for a filtered field? Items with null/missing values should be excluded when any filter is active for that attribute.
- What happens when enum options in the schema change after items are already saved? Only options defined in the current schema appear as pills; items with legacy values still match if selected via other means.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST dynamically generate a filter bar from the active module's schema when a specific collection type is selected.
- **FR-002**: System MUST create selectable pills for each value of enum-type attributes.
- **FR-003**: System MUST create tri-state toggle pills for boolean-type attributes, cycling through off (no filter), true-only, and false-only on repeated clicks.
- **FR-004**: System MUST create inline min/max number inputs for number-type attributes directly in the filter bar (no popover).
- **FR-005**: Selecting multiple values within a single enum attribute MUST use OR logic (item matches if any selected value matches).
- **FR-006**: Filters across different attributes MUST combine with AND logic (item must match all active attribute filters).
- **FR-007**: Attribute filters MUST combine with text search using AND logic (item must match both text search and all attribute filters).
- **FR-008**: The filter bar MUST NOT appear when no module is selected (showing "All Types") or when the active module has no filterable attributes.
- **FR-015**: The filter bar MUST be collapsible: by default it shows a single compact row with a summary of active filters and an expand/collapse toggle. Expanding reveals all filter controls.
- **FR-009**: Switching modules MUST clear all active filters and regenerate the filter bar for the new module's schema.
- **FR-010**: The backend item query MUST accept an optional set of attribute filters and apply them alongside existing text search and module filtering.
- **FR-011**: Items with null or missing values for a filtered attribute MUST be excluded when any filter is active for that attribute.
- **FR-012**: Filtered results MUST update within 500ms of a filter change.
- **FR-013**: Both list view and grid view MUST respect the same active filters.
- **FR-014**: A "Clear all filters" action MUST be available when any filters are active.

### Key Entities

- **Attribute Filter**: A user-applied constraint on a specific schema attribute. Has a field name, filter type (enum/boolean/range), and selected values or range bounds.
- **Filter Bar**: A dynamically generated UI strip derived from the active module's schema, containing filter controls for each filterable attribute.
- **Filter Payload**: A structured set of active filters sent to the backend alongside the existing query and module ID parameters.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can narrow a 500-item collection to a specific subset in under 3 seconds using faceted filters.
- **SC-002**: 100% of enum and boolean attributes in any schema are surfaced as filter controls when their module is active.
- **SC-003**: Number range filters correctly constrain results to the specified bounds with zero false positives.
- **SC-004**: Filters and text search combine correctly: no items appear that fail any active filter or text query.
- **SC-005**: Switching modules clears filters and regenerates the bar in under 500ms.
- **SC-006**: The filter bar adds zero visual clutter when no module is selected or when a module has no filterable attributes.

## Clarifications

### Session 2026-04-07

- Q: Should boolean filters only filter to true, or support filtering to false as well? -> A: Tri-state toggle cycling off -> true -> false -> off on repeated clicks.
- Q: How should the filter bar handle visual overflow with many filterable attributes? -> A: Collapsible bar -- compact row with active filter summary by default, expand to reveal all controls.
- Q: Should number range controls use inline inputs, a popover, or sliders? -> A: Inline min/max number inputs directly in the filter bar (no popover or slider).

## Assumptions

- Only `enum`, `boolean`, and `number` attribute types are filterable. String, date, and other types are not included in v1 (users can still search strings via text search).
- The `purchasePrice` base field on items is treated as a filterable number field when any module is selected (it exists on all items regardless of schema).
- The filter bar appears in the same location for both list and grid views (above the content, below existing controls).
- Number range controls use inline min/max number inputs directly in the filter bar (no popover or slider). This keeps interaction fast and avoids popover dismissal issues with unknown value ranges.
- Multi-select within a single enum attribute uses OR logic (show items matching any selected value). Cross-attribute filtering uses AND logic. This is the standard faceted search convention.
- Filter state is ephemeral (not persisted across sessions or module switches). Each module switch resets filters.
- The backend query filters by extracting attribute values from the stored JSON. Performance is acceptable for collections up to a few thousand items.
