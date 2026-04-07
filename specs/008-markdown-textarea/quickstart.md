# Quickstart: Markdown Textarea Support

**Branch**: `008-markdown-textarea` | **Date**: 2026-04-07

## Prerequisites

- Wails v2 development environment (`wails dev` works)
- Existing codebase on the `008-markdown-textarea` branch
- A module schema with at least one `widget: "textarea"` attribute

## New Dependencies to Install

```bash
cd frontend
npm install @codemirror/lang-markdown @codemirror/language marked dompurify
npm install -D @types/dompurify
```

## Files to Create

1. **`frontend/src/components/MarkdownEditor.vue`** -- CodeMirror-based Markdown editor with formatting toolbar (bold, italic, heading, list, link)
2. **`frontend/src/components/MarkdownRenderer.vue`** -- Safe Markdown-to-HTML renderer using marked + DOMPurify

## Files to Modify

1. **`frontend/src/components/FormField.vue`** -- Replace `<textarea>` with `<MarkdownEditor>` when `widget === "textarea"`
2. **`frontend/src/components/ItemDetail.vue`** -- Use `<MarkdownRenderer>` for textarea attributes in the detail view
3. **`frontend/src/style.css`** -- Add `.prose` class with typography for rendered Markdown

## Implementation Order

1. Install npm dependencies
2. Create MarkdownEditor.vue (CodeMirror + toolbar)
3. Create MarkdownRenderer.vue (marked + DOMPurify)
4. Add `.prose` CSS class to style.css
5. Modify FormField.vue to use MarkdownEditor
6. Modify ItemDetail.vue to use MarkdownRenderer
7. Update CLAUDE.md and README

## Acceptance Test Flow

1. Open a module schema that has a `widget: "textarea"` attribute (e.g., "Notes")
2. Click to create a new item -- verify the Markdown editor appears with toolbar instead of plain textarea
3. Type `**bold**` -- verify syntax highlighting in the editor
4. Click the Bold toolbar button -- verify `**` markers are inserted
5. Click Italic, Heading, List, Link toolbar buttons -- verify each inserts correct syntax
6. Save the item
7. Open the item in detail view -- verify Markdown is rendered as formatted HTML
8. Verify bold text is bold, lists are bulleted/numbered, links are clickable
9. Verify headings use the serif font, body text uses sans-serif
10. Edit the item -- verify the Markdown editor loads with existing content pre-filled
11. Enter `<script>alert('xss')</script>` in the editor, save, view detail -- verify the script tag is stripped
12. Create an item with plain text (no Markdown syntax) -- verify it displays correctly as plain paragraphs
