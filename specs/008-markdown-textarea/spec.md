# Feature Specification: Markdown Textarea Support

**Feature Branch**: `008-markdown-textarea`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Upgrade textarea widget to support Markdown editing and rendering. Integrate a Markdown editor for form input and safe Markdown rendering for detail views. Add prose typography styling."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Write Rich Text in Collection Item Forms (Priority: P1)

A collector creating or editing an item has a schema with a `widget: "textarea"` field (e.g., "Notes", "Description", "History"). Instead of a plain textarea, they see a Markdown-aware editor with a minimal toolbar for common formatting (bold, italic, headings, lists, links). They write content using Markdown syntax or the toolbar buttons. The raw Markdown string is saved as the attribute value.

**Why this priority**: The editing experience is the core of this feature. Without a Markdown editor, there is nothing to render. This story alone upgrades the plain textarea to a useful rich-text input.

**Independent Test**: Create/edit an item with a textarea widget field, verify the editor appears with formatting toolbar, type Markdown, save the item, confirm the raw Markdown string is persisted in the item's attributes.

**Acceptance Scenarios**:

1. **Given** a schema attribute with `widget: "textarea"`, **When** the form renders, **Then** a Markdown editor with formatting toolbar replaces the plain textarea.
2. **Given** the Markdown editor is displayed, **When** the user types `**bold**` or clicks the bold toolbar button, **Then** the editor shows the formatted text or inserts the Markdown syntax.
3. **Given** the user has entered Markdown content, **When** they save the item, **Then** the raw Markdown string is persisted as the attribute value (no HTML stored).
4. **Given** an existing item with Markdown content, **When** the user edits it, **Then** the editor loads with the existing Markdown content pre-filled.
5. **Given** a schema attribute without `widget: "textarea"` (e.g., plain text, number), **Then** the existing input controls remain unchanged.

---

### User Story 2 - View Rendered Markdown in Item Details (Priority: P2)

A collector opens the detail view for an item that has Markdown content in a textarea field. Instead of seeing raw Markdown syntax, they see beautifully rendered text with headings, bold, italics, bullet lists, numbered lists, and clickable links. The rendered content uses the application's typography (serif headings, sans-serif body) and is safe from script injection.

**Why this priority**: Rendering makes the Markdown useful to read. Without it, users see raw syntax in the detail view, which defeats the purpose of rich text.

**Independent Test**: Save an item with Markdown content in a textarea field, open the item detail view, verify the content renders as formatted HTML with proper typography.

**Acceptance Scenarios**:

1. **Given** an item with Markdown content in a textarea attribute, **When** the detail view renders, **Then** the Markdown is displayed as formatted HTML (headings, bold, italic, lists, links).
2. **Given** Markdown content with links, **When** rendered, **Then** links are clickable and open in the system browser.
3. **Given** Markdown content containing script tags or HTML injection attempts, **When** rendered, **Then** the dangerous content is stripped and not executed (sanitized output only).
4. **Given** rendered Markdown content, **Then** headings use the serif font, body text uses the sans-serif font, and the overall typography matches the application's design language.
5. **Given** an attribute with no content or plain text (no Markdown syntax), **When** the detail view renders, **Then** it displays as plain text without errors.

---

### User Story 3 - Prose Typography Styling (Priority: P3)

The application provides a reusable prose styling class that ensures all rendered Markdown content throughout the app looks polished. Headings, paragraphs, lists, blockquotes, code blocks, and links all have consistent spacing, font sizing, and color treatment that harmonizes with the existing Outfit (body) and Instrument Serif (headings) font pairing.

**Why this priority**: Typography styling is a polish layer. The feature works without it (US1+US2), but it won't look premium. This story ensures visual consistency across all rendered Markdown.

**Independent Test**: Apply the prose class to rendered Markdown content, verify that headings, paragraphs, lists, blockquotes, code, and links all render with appropriate fonts, sizes, spacing, and colors.

**Acceptance Scenarios**:

1. **Given** rendered Markdown with h1-h4 headings, **Then** headings use the serif font with progressively smaller sizes and appropriate spacing.
2. **Given** rendered Markdown with bullet and numbered lists, **Then** lists have consistent indentation, markers, and line spacing.
3. **Given** rendered Markdown with inline code and code blocks, **Then** code uses a monospace font with a subtle background color.
4. **Given** rendered Markdown with blockquotes, **Then** blockquotes have a left border accent and italic styling.
5. **Given** rendered Markdown with links, **Then** links use the accent color and show an underline on hover.

---

### Edge Cases

- What happens when a textarea field contains extremely long Markdown content? The editor should scroll; the rendered view should display the full content without layout breakage.
- What happens when a user pastes HTML content into the Markdown editor? The editor should accept it as plain text (strip HTML tags) or convert to Markdown equivalent.
- What happens with existing items that have plain text in textarea fields (pre-Markdown)? Plain text renders fine as Markdown (no syntax = no formatting), so backward compatibility is automatic.
- What happens when the Markdown contains image references? Images are not supported in v1 (items have a separate image system). Image Markdown syntax should render as the alt text or be ignored gracefully.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Schema attributes with `widget: "textarea"` MUST render a Markdown editor with formatting toolbar in create/edit forms.
- **FR-002**: The Markdown editor MUST support at minimum: bold, italic, headings (h1-h3), bullet lists, numbered lists, and links.
- **FR-003**: The editor MUST persist the raw Markdown string as the attribute value (no HTML stored in the database).
- **FR-004**: The editor MUST pre-fill with existing Markdown content when editing an existing item.
- **FR-005**: Schema attributes without `widget: "textarea"` MUST continue using their existing input controls unchanged.
- **FR-006**: The item detail view MUST render Markdown content as formatted HTML for textarea attributes.
- **FR-007**: All rendered HTML MUST be sanitized to prevent script injection (no executable HTML allowed).
- **FR-008**: Rendered Markdown links MUST be clickable and open in the user's system browser.
- **FR-009**: The application MUST provide a reusable prose CSS class that styles rendered Markdown with the app's typography (serif headings, sans-serif body, monospace code).
- **FR-010**: The Markdown editor toolbar MUST be minimal and non-intrusive, consistent with the application's visual design language.
- **FR-011**: Plain text content (no Markdown syntax) in textarea fields MUST render correctly without errors in both editor and detail view.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can format text using Markdown in textarea fields and see the formatting applied in the detail view within a single create-save-view cycle.
- **SC-002**: 100% of textarea widget fields across all schemas use the Markdown editor.
- **SC-003**: Zero script injection vectors in rendered Markdown content (all HTML sanitized).
- **SC-004**: Existing items with plain text in textarea fields display correctly without any user intervention or data migration.
- **SC-005**: Rendered Markdown content is visually consistent across all item detail views, using the application's font pairing.

## Assumptions

- The data format remains unchanged: Markdown is stored as a plain string in the JSON `attributes` blob. No new database columns or schema changes are needed.
- The Markdown editor is a lightweight component that loads quickly (under 200ms); it does not require a heavy WYSIWYG framework.
- Markdown image syntax (`![alt](url)`) is not rendered as images in v1. The alt text is displayed instead. Items already have a dedicated image attachment system.
- The toolbar includes only essential formatting actions (bold, italic, heading, list, link). Advanced features like tables, footnotes, or math are out of scope for v1.
- Links in rendered Markdown open in the system's default browser (standard behavior for desktop apps using webview).
- The prose CSS class is global (defined in the app's main stylesheet) and reusable anywhere rendered Markdown appears.
- CodeMirror is already a project dependency (used by the schema code editor). The Markdown editor can leverage the same CodeMirror infrastructure if desired, keeping the dependency footprint small.
