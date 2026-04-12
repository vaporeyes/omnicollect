# Feature Specification: Masonry Grid & Item Comparison

**Feature Branch**: `020-masonry-compare`  
**Created**: 2026-04-11  
**Status**: Draft  
**Input**: User description: "Upgrade media presentation for serious collectors: adaptive masonry grid layout removing fixed aspect-ratio, item comparison mode for two selected items with synchronized galleries and attribute diffing."

## Clarifications

### Session 2026-04-11

- Q: Do thumbnails preserve original aspect ratio, or are they square-cropped? (Masonry depends on natural proportions.) → A: Thumbnails preserve original aspect ratio; masonry works with existing thumbnails. No thumbnail generation changes needed.
- Q: Should the comparison diff include core item fields (title, price, tags) or only module schema attributes? → A: Core fields (title, price, tags) are included in the diff table alongside schema attributes.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Masonry Grid Layout (Priority: P1)

A collector browses their collection and sees items displayed in a masonry grid where each card's height adapts to the original image's aspect ratio. Tall portrait photos of figurines fill vertical space naturally, while wide landscape photos of display shelves stay short. The frosted glass caption overlay and smooth hover effects remain intact. The overall visual impression is a polished, Pinterest-style gallery that showcases each item at its best.

**Why this priority**: The masonry layout is the foundational visual upgrade that benefits every user on every page load. It is a standalone improvement with no dependency on comparison mode.

**Independent Test**: Can be fully tested by loading a collection with mixed portrait, landscape, and square images. Each card should size naturally based on image dimensions, with no forced cropping. Hover states and captions must render identically to the current grid.

**Acceptance Scenarios**:

1. **Given** a collection with images of varying aspect ratios, **When** the user views the grid, **Then** each card's height reflects the original image proportions without forced cropping.
2. **Given** a collection item with a tall portrait image (e.g., 3:4), **When** displayed in the grid, **Then** the card is taller than a card with a landscape image (e.g., 16:9) in the same grid.
3. **Given** any grid card, **When** the user hovers over it, **Then** the card scales up slightly with an elevated shadow, matching the existing hover behavior.
4. **Given** any grid card, **When** displayed, **Then** a frosted glass caption at the bottom shows the item title, module name, and date.
5. **Given** a collection item with no images, **When** displayed in the grid, **Then** a placeholder card renders at a reasonable default size.
6. **Given** the grid is resized (window resize or sidebar toggle), **When** columns reflow, **Then** cards redistribute cleanly with no large vertical gaps.

---

### User Story 2 - Compare Button Activation (Priority: P2)

A collector multi-selects exactly two items in the grid. The floating bulk action bar shows a "Compare" button alongside the existing actions. The button only appears when exactly two items are selected and disappears if the selection changes to any other count.

**Why this priority**: This is the entry point to comparison mode. Without it, comparison mode is unreachable. It is a small, self-contained change to the existing bulk action bar.

**Independent Test**: Can be tested by selecting 0, 1, 2, and 3+ items and verifying the Compare button appears only at exactly 2.

**Acceptance Scenarios**:

1. **Given** zero or one item is selected, **When** the user views the bulk action bar, **Then** no Compare button is visible.
2. **Given** exactly two items are selected, **When** the user views the bulk action bar, **Then** a Compare button appears alongside Delete, Export, Edit Module, and Deselect All.
3. **Given** exactly two items are selected, **When** the user selects a third item, **Then** the Compare button disappears.
4. **Given** exactly two items are selected, **When** the user clicks Compare, **Then** the main content area transitions to the comparison view.

---

### User Story 3 - Side-by-Side Comparison View (Priority: P2)

A collector enters comparison mode and sees the two selected items displayed side-by-side, each taking half the content area. Each side shows the item's image gallery at the top and its attributes table below. The user can browse through each item's images. The attributes table highlights any differences between the two items (e.g., different Condition values, different Mint Mark fields) so the collector can quickly spot what distinguishes one item from the other.

**Why this priority**: This is the core value of comparison mode. Tied with US2 since they are both required for the feature to function.

**Independent Test**: Can be tested by entering comparison mode with two items that share the same module schema and have overlapping but non-identical attribute values. Differences should be visually highlighted.

**Acceptance Scenarios**:

1. **Given** the user enters comparison mode with two items, **When** the view renders, **Then** both items appear side-by-side with equal width.
2. **Given** both items have image galleries, **When** the user navigates images on one side, **Then** the other side also advances to the same image index (synchronized navigation).
3. **Given** both items have image galleries of different lengths, **When** synchronized navigation reaches the end of the shorter gallery, **Then** the shorter side stops at its last image while the longer side continues.
4. **Given** both items share a module schema with attribute "Condition", **When** Item A has "Mint" and Item B has "Good", **Then** the Condition row is visually highlighted as different.
5. **Given** two items with different purchase prices, **When** compared, **Then** the price row in the diff table is visually highlighted as different.
6. **Given** both items share a module schema, **When** all core fields and attribute values are identical, **Then** no difference highlighting appears.
7. **Given** the user is in comparison mode, **When** the user clicks a close/back button, **Then** the view returns to the collection grid with the previous selection preserved.
8. **Given** two items from different modules, **When** compared, **Then** the diff table shows core fields for both plus the union of both schemas with missing fields marked as empty on the respective side.

---

### Edge Cases

- What happens when an item has no images? Display a placeholder on that side of the comparison; synchronized navigation controls the side that has images.
- What happens when both items have zero images? Both sides show placeholders; gallery navigation controls are hidden.
- What happens when the user resizes the window while in comparison mode? The side-by-side layout remains responsive, stacking vertically if the viewport becomes too narrow.
- What happens when one item is deleted externally while comparison mode is open? The comparison view closes gracefully and returns to the grid.
- What happens when the masonry grid contains only one item? It renders as a single card with natural dimensions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The collection grid MUST arrange cards in a masonry (variable-height column) layout where each card's height is determined by its primary image's aspect ratio.
- **FR-002**: Grid cards MUST NOT apply a fixed aspect ratio to images; images MUST display without cropping, using their original proportions.
- **FR-003**: Grid cards MUST retain the existing frosted glass caption overlay (blurred background, item title, module name, date).
- **FR-004**: Grid cards MUST retain the existing hover effect (slight scale increase and shadow elevation).
- **FR-005**: Grid cards with no images MUST render with a placeholder at a default card height.
- **FR-006**: The masonry grid MUST reflow correctly when the viewport width changes or columns are added/removed.
- **FR-007**: The bulk action bar MUST display a "Compare" button when exactly two items are selected.
- **FR-008**: The Compare button MUST NOT appear when fewer or more than two items are selected.
- **FR-009**: Clicking Compare MUST open a comparison view that replaces the main content area.
- **FR-010**: The comparison view MUST display both items side-by-side, each occupying half the available width.
- **FR-011**: Each side of the comparison view MUST display the item's image gallery with navigation controls (previous/next).
- **FR-012**: Image gallery navigation MUST be synchronized: advancing one side advances the other.
- **FR-013**: Synchronized navigation MUST handle galleries of unequal length gracefully (shorter gallery stops at its last image).
- **FR-014**: The comparison view MUST display a combined diff table below the images that includes core item fields (title, purchase price, tags) and all module schema attributes, using the schema for field labels.
- **FR-015**: Rows where the two items have different values MUST be visually highlighted, including differences in core fields (title, price, tags) and schema attributes.
- **FR-016**: The comparison view MUST provide a way to exit and return to the collection grid.
- **FR-017**: Items from different modules MUST be comparable; the attributes table shows the union of both schemas with empty values where an attribute does not exist on one side.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Collection grid displays all items at their original image aspect ratios with no visible cropping in the default grid view.
- **SC-002**: Users can identify the Compare button and enter comparison mode within 5 seconds of selecting two items.
- **SC-003**: Attribute differences between two compared items are visually distinguishable without reading each value individually (color or highlight is sufficient).
- **SC-004**: Synchronized gallery navigation keeps both sides in lockstep: clicking next on either side advances both to the same index.
- **SC-005**: The masonry grid reflows within 200ms of a viewport resize with no visible layout jank.
- **SC-006**: Existing grid interactions (hover states, selection checkboxes, context menus, keyboard shortcuts) continue to function identically after the masonry change.

## Assumptions

- The masonry layout uses CSS-only techniques (CSS columns or CSS grid with masonry-like behavior) rather than requiring a JavaScript layout library. If native CSS masonry proves insufficient, a lightweight JS approach may be used.
- Thumbnails preserve the original image's aspect ratio (scaled down, not square-cropped). Image aspect ratios are determined from the loaded thumbnail dimensions at render time. No server-side dimension metadata is stored or required.
- Comparison mode is a client-side-only view with no new backend endpoints or data persistence.
- The comparison view reuses existing item data already loaded in the collection store; no additional API calls are needed to enter comparison mode.
- Multi-select and shift-click behavior in the grid remain unchanged.
- The staggered card entrance animation from the current grid is preserved in the masonry layout.
