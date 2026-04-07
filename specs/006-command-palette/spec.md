# Feature Specification: Spotlight-style Command Palette

**Feature Branch**: `006-command-palette`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Implement a Spotlight-style Command Palette component overlaid on the app, triggered globally by Cmd/Ctrl+K with a blurred glass input field, querying collectionStore for items across all modules with rich results, Quick Actions for keywords like New and Settings, and instant navigation to ItemDetail view."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search and Navigate to Any Item (Priority: P1)

A power user with hundreds of collection items across multiple modules wants to jump directly to a specific item without scrolling through lists or switching module filters. They press Cmd/Ctrl+K, type a few characters of the item name, see matching results with thumbnails and module badges, and select one to immediately open its detail view.

**Why this priority**: This is the core value proposition of the command palette. Without cross-module search and instant navigation, the feature has no reason to exist.

**Independent Test**: Can be fully tested by opening the palette, typing a search query, and verifying that selecting a result navigates to the correct item detail view.

**Acceptance Scenarios**:

1. **Given** the user is anywhere in the application, **When** they press Cmd/Ctrl+K, **Then** a centered overlay appears with a focused text input field and a blurred glass backdrop.
2. **Given** the palette is open, **When** the user types characters, **Then** matching items from all modules appear as results showing thumbnail, title, and module badge.
3. **Given** results are displayed, **When** the user clicks a result or presses Enter on a highlighted result, **Then** the palette closes and the item's detail view opens.
4. **Given** the palette is open, **When** the user presses Escape or clicks outside the palette, **Then** it closes without navigating.
5. **Given** the palette is open, **When** the user types a query with no matches, **Then** a "No results" message is displayed.

---

### User Story 2 - Keyboard Navigation of Results (Priority: P2)

A keyboard-driven user wants to browse results entirely with arrow keys. After typing a query, they use Up/Down arrows to move the highlight through results and press Enter to select, never touching the mouse.

**Why this priority**: Keyboard navigation is essential for a power-user tool. Without it, the palette is just a search box with extra steps.

**Independent Test**: Can be tested by opening the palette, typing a query, using arrow keys to move selection, and pressing Enter to confirm navigation.

**Acceptance Scenarios**:

1. **Given** results are displayed, **When** the user presses the Down arrow, **Then** the next result is highlighted and scrolled into view if necessary.
2. **Given** the second result is highlighted, **When** the user presses the Up arrow, **Then** the first result is highlighted.
3. **Given** a result is highlighted, **When** the user presses Enter, **Then** the palette closes and navigates to that item.
4. **Given** the highlight is on the first result, **When** the user presses Up, **Then** the highlight stays on the first result (no wrap-around).

---

### User Story 3 - Quick Actions via Keywords (Priority: P3)

A user wants to access common application actions without leaving the keyboard. Typing contextual keywords like "new", "settings", or "backup" surfaces relevant quick actions at the top of results, allowing one-keystroke access to create items, open settings, or trigger exports.

**Why this priority**: Quick actions add breadth to the palette beyond item search. They are valuable but the palette is still useful without them.

**Independent Test**: Can be tested by opening the palette, typing a keyword like "new", verifying quick action entries appear above item results, and selecting one to trigger the expected application action.

**Acceptance Scenarios**:

1. **Given** the palette is open, **When** the user types "new", **Then** quick actions "Create New Schema" and "Add New Item" appear above any item search results.
2. **Given** the palette is open, **When** the user types "settings", **Then** a quick action "Open Settings" appears.
3. **Given** a quick action is displayed, **When** the user selects it, **Then** the palette closes and the corresponding action is triggered (e.g., settings page opens, new item form opens).
4. **Given** the palette is open, **When** the user types "backup" or "export", **Then** a quick action "Export Backup" appears.

---

### Edge Cases

- What happens when the palette is opened while another overlay (lightbox, form, builder) is active? The palette should overlay on top of everything; closing it returns to the previous state.
- What happens when the user types very quickly? Results should debounce input to avoid excessive queries.
- What happens when the collection has zero items? Only quick actions should appear (if keyword matches), with an appropriate empty state otherwise.
- What happens if the palette is already open and the user presses Cmd/Ctrl+K again? It should close (toggle behavior).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a command palette overlay when the user presses Cmd/Ctrl+K from any application state.
- **FR-002**: System MUST close the palette when the user presses Escape, clicks outside the overlay, or selects a result.
- **FR-003**: System MUST toggle the palette closed if the user presses Cmd/Ctrl+K while it is already open.
- **FR-004**: The palette input field MUST be auto-focused when the palette opens.
- **FR-005**: System MUST search items across all modules (ignoring the current module filter) as the user types.
- **FR-006**: Search results MUST display item thumbnail (or a placeholder), item title, and module name badge.
- **FR-007**: System MUST debounce search input to avoid querying on every keystroke.
- **FR-008**: System MUST support keyboard navigation of results using Up/Down arrow keys and selection via Enter.
- **FR-009**: The currently highlighted result MUST be visually distinct and scrolled into view.
- **FR-010**: When the user selects an item result, the system MUST navigate to that item's detail view.
- **FR-011**: System MUST display contextual quick actions above search results when the query matches predefined keywords ("new", "settings", "backup"/"export").
- **FR-012**: Quick actions MUST trigger the corresponding application action when selected (open new item form, open settings, trigger backup export).
- **FR-013**: The palette MUST display a "No results" message when neither items nor quick actions match the query.
- **FR-014**: The palette overlay MUST use a frosted glass / backdrop blur visual style consistent with the application's design language.

### Key Entities

- **Search Result**: An item from the collection displayed as a selectable row with thumbnail, title, and module badge.
- **Quick Action**: A predefined application action surfaced when the user's query matches a keyword. Has a label, an icon or indicator, and a trigger action.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can open the command palette and navigate to any item in their collection within 3 seconds (open palette, type, select).
- **SC-002**: Search results appear within 300ms of the user pausing their typing.
- **SC-003**: 100% of quick actions ("new", "settings", "backup") are discoverable by typing the corresponding keyword.
- **SC-004**: The palette is fully operable via keyboard alone (open, search, navigate results, select, close).
- **SC-005**: The palette does not interfere with other application overlays; closing it returns to the previous state.

## Assumptions

- The existing item search capability (full-text search via the backend) is sufficient for command palette queries; no new search infrastructure is needed.
- The palette searches across all modules regardless of the currently active sidebar filter. This is intentional to provide universal access.
- Quick action keywords are matched case-insensitively against a hardcoded set ("new", "settings", "backup", "export"). No fuzzy matching is required for quick actions.
- The "Add New Item" quick action will use the currently active module filter, or the first available module if no filter is active (same behavior as Cmd+N).
- The palette appears on top of all content with a high z-index. It does not dismiss other overlays; it layers above them.
- Result count is capped at a reasonable limit (e.g., 20-50 items) to keep the dropdown performant and scannable.
