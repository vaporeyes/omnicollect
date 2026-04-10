# Feature Specification: Smart Folders (Saved Views)

**Feature Branch**: `019-smart-folders`  
**Created**: 2026-04-10  
**Status**: Draft  
**Input**: User description: "Add Smart Folders (Saved Views) to sidebar navigation with serialized search/filter state, save/rename/delete support."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Save Current View as a Smart Folder (Priority: P1)

A collector has navigated to a specific module, typed a search query, and applied several faceted filters to narrow down their collection (e.g., "Coins" module, search "Liberty", condition filter "Mint State"). They want to save this exact view so they can return to it instantly later. They click a "Save Current View" button, give it a name (e.g., "Mint Liberty Coins"), and it appears in the sidebar as a Smart Folder.

**Why this priority**: This is the core value proposition -- without the ability to save a view, no other Smart Folder features are useful. It must work end-to-end as an MVP.

**Independent Test**: Can be fully tested by applying a module filter, search query, and attribute filters, then saving the view, refreshing the app, and verifying the Smart Folder persists in the sidebar.

**Acceptance Scenarios**:

1. **Given** a user has an active module, search query, and/or attribute filters applied, **When** they click "Save Current View", **Then** an inline text field appears in the Smart Folders sidebar section. **When** they type a name and press Enter, **Then** a new Smart Folder appears in the sidebar with the given name and a distinct visual indicator.
2. **Given** a user has no search query or filters active (just a module selected), **When** they click "Save Current View", **Then** the system still allows saving (a Smart Folder can represent just a module selection).
3. **Given** a user has "All Types" selected with no filters, **When** they click "Save Current View", **Then** the system still allows saving (representing an unfiltered "everything" view).
4. **Given** a user tries to save a Smart Folder with an empty name, **When** they confirm, **Then** the system prevents saving and shows a validation message.
5. **Given** a user saves a Smart Folder, **When** they close and reopen the application, **Then** the Smart Folder is still present in the sidebar.

---

### User Story 2 - Apply a Smart Folder (Priority: P2)

A collector sees their saved "Mint Liberty Coins" Smart Folder in the sidebar. They click it, and the application instantly restores the saved module, search query, and filter state -- showing the same filtered view they originally saved.

**Why this priority**: Applying saved views is the primary consumption action. Without it, saving is pointless. It depends on P1 existing but is independently testable once Smart Folders exist.

**Independent Test**: Can be tested by creating a Smart Folder with known state (module + search + filters), clicking it, and verifying the collection view matches the expected filtered results.

**Acceptance Scenarios**:

1. **Given** a Smart Folder exists with a saved module, search query, and filters, **When** the user clicks it in the sidebar, **Then** the application switches to the saved module, populates the search field, applies the saved filters, and displays the matching items.
2. **Given** a Smart Folder was saved with a module that has since been deleted, **When** the user clicks it, **Then** the system shows all items (falls back to "All Types") and displays an informational message that the original module no longer exists.
3. **Given** the user is currently viewing a different module with active filters, **When** they click a Smart Folder, **Then** the previous view state is fully replaced by the Smart Folder's saved state.
4. **Given** a Smart Folder is active, **When** the user manually changes the search or filters, **Then** the Smart Folder is visually deselected in the sidebar (the view is no longer the saved state).

---

### User Story 3 - Manage Smart Folders (Rename and Delete) (Priority: P3)

A collector wants to rename an existing Smart Folder to something more descriptive, or delete one they no longer need. They right-click (or long-press) on a Smart Folder in the sidebar to access Rename and Delete options via a context menu.

**Why this priority**: Management operations are secondary to creation and usage. Users can live without rename/delete initially, but need them for long-term maintenance of their saved views.

**Independent Test**: Can be tested by creating a Smart Folder, right-clicking it, selecting Rename, entering a new name, and verifying it updates. Separately, right-clicking and selecting Delete, confirming, and verifying removal.

**Acceptance Scenarios**:

1. **Given** a Smart Folder exists, **When** the user right-clicks it, **Then** a context menu appears with "Rename" and "Delete" options.
2. **Given** the user selects "Rename" from the context menu, **When** they enter a new name and confirm, **Then** the Smart Folder's name updates in the sidebar and persists across sessions.
3. **Given** the user selects "Delete" from the context menu, **When** they confirm the deletion, **Then** the Smart Folder is removed from the sidebar and from persistent storage.
4. **Given** a Smart Folder is currently active (its view is displayed), **When** the user deletes it, **Then** the view remains as-is but the Smart Folder no longer appears in the sidebar.

---

### Edge Cases

- What happens when a user saves a Smart Folder with filter values that reference enum options that have since been removed from the module schema? The system applies the filters it can and ignores invalid filter values silently.
- What happens when the user has many Smart Folders (20+)? The sidebar section scrolls independently so Smart Folders do not push module list off-screen.
- What happens when two Smart Folders have the same name? The system allows duplicate names (they are distinguished by their saved state content).
- What happens when Smart Folder state includes active tags? Tags are included in the saved state alongside module, search, and attribute filters.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to save the current view state (active module, search query, attribute filters, and active tags) as a named Smart Folder. The naming input MUST be an inline text field in the sidebar (not a modal), confirmed by pressing Enter and dismissable by pressing Escape.
- **FR-002**: System MUST display saved Smart Folders in a dedicated section of the sidebar, visually distinct from the module list (using a bookmark or similar icon).
- **FR-003**: System MUST persist Smart Folders across application sessions (survives close and reopen).
- **FR-004**: Clicking a Smart Folder MUST restore the full saved view state: module selection, search query, attribute filters, and tag filters.
- **FR-005**: System MUST visually highlight the active Smart Folder in the sidebar when its saved state matches the current view.
- **FR-006**: System MUST deselect the active Smart Folder highlight when the user manually changes the module, search, or filters.
- **FR-007**: System MUST provide a context menu on right-click (or equivalent gesture) for each Smart Folder with "Rename" and "Delete" actions.
- **FR-008**: Renaming a Smart Folder MUST update the display name and persist the change.
- **FR-009**: Deleting a Smart Folder MUST remove it from the sidebar and from persistent storage, with a confirmation step.
- **FR-010**: System MUST validate that Smart Folder names are non-empty before saving.
- **FR-011**: System MUST handle gracefully the case where a Smart Folder references a module that no longer exists (fall back to "All Types" with an informational message).
- **FR-012**: Smart Folder ordering in the sidebar MUST reflect the order in which they were created (newest last).

### Key Entities

- **Smart Folder**: A named, persisted view configuration containing: unique identifier, user-given display name, optional module ID, optional search query string, optional attribute filter set, optional tag filter list, creation timestamp.
- **View State Snapshot**: The serialized combination of module selection, search query, attribute filters, and tag filters that represents a specific collection view at the time of saving. View mode (dashboard/list/grid) is explicitly excluded; the user's current view mode is preserved when applying a Smart Folder.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can save the current view as a Smart Folder in under 5 seconds (click button, type name, confirm).
- **SC-002**: Clicking a Smart Folder restores the saved view within 1 second.
- **SC-003**: Smart Folders persist across 100% of application restarts with no data loss.
- **SC-004**: Users can rename or delete a Smart Folder in under 3 seconds via context menu.
- **SC-005**: The sidebar remains usable with up to 50 Smart Folders without performance degradation.

## Clarifications

### Session 2026-04-10

- Q: How does the user enter a Smart Folder name when saving? → A: Inline text field in the sidebar (type name, press Enter). No modal dialog.
- Q: Should Smart Folder state include view mode (dashboard/list/grid)? → A: No. Smart Folders only restore data filters; user's current view mode is preserved.

## Assumptions

- Smart Folders are stored locally alongside application settings (same persistence mechanism as theme preferences). No new backend API endpoints are required for local/desktop mode.
- For cloud mode, Smart Folders are stored as part of the user's settings in the existing settings storage mechanism.
- The existing context menu component is reusable for Smart Folder right-click actions.
- Smart Folder names have no maximum length enforced beyond the practical display width in the sidebar.
- There is no limit on the number of Smart Folders a user can create.
- The "Save Current View" button is always visible in the sidebar (not hidden behind a menu), positioned between the module list and the bottom action buttons.
