# Tasks: Markdown Textarea Support

**Input**: Design documents from `/specs/008-markdown-textarea/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ui-contract.md, quickstart.md

**Tests**: No automated test framework in this project. Tests not requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Install dependencies needed for Markdown editing and rendering

- [x] T001 Install npm dependencies: run `npm install @codemirror/lang-markdown @codemirror/language marked dompurify` and `npm install -D @types/dompurify` in `frontend/`

**Checkpoint**: New packages available in node_modules; package.json updated

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational/blocking tasks needed. Both user stories are component-level work that depends only on the npm packages from Phase 1.

---

## Phase 3: User Story 1 - Write Rich Text in Collection Item Forms (Priority: P1) MVP

**Goal**: Schema attributes with `widget: "textarea"` render a Markdown-aware CodeMirror editor with formatting toolbar instead of a plain textarea.

**Independent Test**: Create/edit an item with a textarea widget field, verify the Markdown editor appears with toolbar, type Markdown, save, confirm raw Markdown string is persisted.

### Implementation for User Story 1

- [x] T002 [US1] Create `frontend/src/components/MarkdownEditor.vue` scaffold: props (`modelValue: string`), emits (`update:modelValue`), CodeMirror editor setup using `vue-codemirror` with `@codemirror/lang-markdown` extension, basic theme matching the app's design (dark/light via CSS variables)
- [x] T003 [US1] Add formatting toolbar to `MarkdownEditor.vue`: a row of icon buttons above the editor for Bold (`**`), Italic (`*`), Heading (`## `), Bullet List (`- `), Numbered List (`1. `), and Link (`[text](url)`); each button inserts corresponding Markdown syntax at cursor position using CodeMirror's dispatch API
- [x] T004 [US1] Style the toolbar and editor in `MarkdownEditor.vue`: toolbar uses `font-family: 'Outfit'`, buttons match app's `var(--bg-hover)` and `var(--border-primary)` theme variables, editor height auto-grows with content up to a max height with scroll
- [x] T005 [US1] Modify `frontend/src/components/FormField.vue`: when `attribute.display?.widget === 'textarea'`, render `<MarkdownEditor v-model="value">` instead of the existing `<textarea>` element; import MarkdownEditor component

**Checkpoint**: Textarea widget fields show the Markdown editor with toolbar; content saves as raw Markdown string in item attributes

---

## Phase 4: User Story 2 - View Rendered Markdown in Item Details (Priority: P2)

**Goal**: Textarea attributes in the item detail view render as beautifully formatted, sanitized HTML instead of raw Markdown text.

**Independent Test**: Save an item with Markdown content, open the detail view, verify headings, bold, italic, lists, and links render correctly as formatted HTML.

### Implementation for User Story 2

- [x] T006 [P] [US2] Create `frontend/src/components/MarkdownRenderer.vue`: props (`content: string`), uses `marked` to convert Markdown to HTML, pipes through `DOMPurify.sanitize()`, renders via `v-html` inside a `<div class="prose">` container; adds `target="_blank"` and `rel="noopener noreferrer"` to all links via marked renderer override
- [x] T007 [US2] Modify `frontend/src/components/ItemDetail.vue`: for attributes where the schema defines `widget: "textarea"`, render `<MarkdownRenderer :content="value">` instead of the plain text `{{ value }}` display; import MarkdownRenderer component; pass schema to determine which attributes are textarea widgets

**Checkpoint**: Detail view shows rendered Markdown with formatting, clickable links, and no XSS vectors

---

## Phase 5: User Story 3 - Prose Typography Styling (Priority: P3)

**Goal**: A reusable `.prose` CSS class ensures all rendered Markdown looks polished with the app's Instrument Serif / Outfit font pairing.

**Independent Test**: Apply the prose class to rendered Markdown, verify headings use serif font, body uses sans-serif, code uses monospace, lists/blockquotes/links are styled consistently.

### Implementation for User Story 3

- [x] T008 [US3] Add `.prose` CSS class to `frontend/src/style.css`: style `h1`-`h4` with `font-family: 'Instrument Serif'` and progressive sizes (24px, 20px, 17px, 15px) with `var(--leading-tight)` line-height; style `p`, `ul`, `ol`, `li` with `font-family: 'Outfit'` at 14px; style `blockquote` with left border using `var(--accent-blue)`, italic, indented; style `code` and `pre` with monospace font and `var(--bg-secondary)` background; style `a` with `var(--accent-blue)` color and underline on hover; add appropriate margin/spacing between elements
- [x] T009 [US3] Verify `.prose` class is applied in `MarkdownRenderer.vue` container div (should already be set from T006); adjust any component-level overrides if needed

**Checkpoint**: All rendered Markdown uses consistent, polished typography matching the app's design language

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and edge case handling

- [x] T010 Handle edge case in `MarkdownEditor.vue`: when content is pasted as HTML, strip HTML tags and insert as plain text (intercept paste event)
- [x] T011 [P] Update `CLAUDE.md` to document MarkdownEditor and MarkdownRenderer components, `widget: "textarea"` Markdown upgrade, new npm dependencies (`@codemirror/lang-markdown`, `marked`, `dompurify`)
- [x] T012 [P] Update project `README.md` to document Markdown support in textarea fields, supported formatting (bold, italic, headings, lists, links), and iteration history entry
- [x] T013 Run quickstart.md acceptance test flow (all 12 steps) and fix any issues found

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **User Story 1 (Phase 3)**: Depends on Phase 1 (npm packages)
- **User Story 2 (Phase 4)**: Depends on Phase 1 (npm packages); independent of US1 (different components)
- **User Story 3 (Phase 5)**: Depends on Phase 4 (prose class applies to MarkdownRenderer output)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends only on Setup. This is the MVP (editor works, saves raw Markdown).
- **US2 (P2)**: Depends only on Setup. Can run in parallel with US1 (different files: MarkdownRenderer vs MarkdownEditor, ItemDetail vs FormField).
- **US3 (P3)**: Depends on US2 (prose class styles the output of MarkdownRenderer).

### Parallel Opportunities

- T002-T004 (MarkdownEditor) and T006 (MarkdownRenderer) target different files -- can run in parallel
- T005 (FormField mod) and T007 (ItemDetail mod) target different files -- can run in parallel
- T011 and T012 (documentation) target different files -- can run in parallel

---

## Parallel Example: US1 + US2 Components

```bash
# These target different files and can run simultaneously:
Task: "Create MarkdownEditor.vue" (T002-T004)
Task: "Create MarkdownRenderer.vue" (T006)
```

## Parallel Example: Integration

```bash
# These modify different files:
Task: "Modify FormField.vue for MarkdownEditor" (T005)
Task: "Modify ItemDetail.vue for MarkdownRenderer" (T007)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001) -- install npm packages
2. Complete Phase 3: User Story 1 (T002-T005) -- editor with toolbar in forms
3. **STOP and VALIDATE**: Create item with Markdown, save, verify raw string persisted
4. This alone upgrades the plain textarea to a usable Markdown editor

### Incremental Delivery

1. Phase 1 (Setup) + Phase 3 (US1) = MVP: Markdown editing in forms
2. Add Phase 4 (US2) = rendered Markdown in detail views
3. Add Phase 5 (US3) = polished prose typography
4. Phase 6 (Polish) = edge cases, documentation, final validation
5. Each phase adds rendering/styling quality without breaking previous phases

### Parallel Team Strategy

With two developers:
1. Both complete Phase 1 (shared)
2. Developer A: US1 (MarkdownEditor + FormField)
3. Developer B: US2 (MarkdownRenderer + ItemDetail)
4. Then: US3 (prose CSS) + Polish

---

## Notes

- No backend changes needed -- raw Markdown stored as plain string in existing attributes JSON
- CodeMirror is already a project dependency (vue-codemirror); only `@codemirror/lang-markdown` is new
- `marked` converts Markdown to HTML; `DOMPurify` sanitizes output -- standard secure pattern
- The `.prose` class is global in style.css, reusable anywhere Markdown is rendered
- Existing items with plain text in textarea fields display correctly without migration
