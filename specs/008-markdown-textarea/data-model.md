# Data Model: Markdown Textarea Support

**Branch**: `008-markdown-textarea` | **Date**: 2026-04-07

## No Data Model Changes

This feature does not modify any data structures, database schema, or stored formats.

### Storage Format (unchanged)

Markdown content is stored as a raw string value in the item's `attributes` JSON object, exactly as plain textarea text was stored before. The attribute key, type, and storage mechanism are all unchanged.

**Before**: `{"notes": "Some plain text notes"}`
**After**: `{"notes": "Some **bold** and *italic* notes\n\n- Item 1\n- Item 2"}`

The only difference is that users can now include Markdown syntax in the string value. The backend treats it identically to any other string attribute.

### Schema Hint (unchanged)

The `widget: "textarea"` display hint in module schemas continues to work as before. No new schema properties are added. The presence of `widget: "textarea"` now triggers the Markdown editor instead of a plain `<textarea>`, but this is a frontend rendering change only.

### Backward Compatibility

Existing items with plain text in textarea fields are fully compatible. Plain text without Markdown syntax renders identically as paragraphs. No data migration needed.
