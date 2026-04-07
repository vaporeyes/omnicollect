# UI Contract: Markdown Components

**Branch**: `008-markdown-textarea` | **Date**: 2026-04-07

## MarkdownEditor Component

### Props

| Prop       | Type   | Required | Description                         |
|------------|--------|----------|-------------------------------------|
| modelValue | string | yes      | Raw Markdown string (v-model)       |

### Emits

| Event            | Payload | Description                       |
|------------------|---------|-----------------------------------|
| update:modelValue | string  | Emitted on content change (v-model) |

### Toolbar Actions

| Button       | Markdown Inserted      | Behavior                           |
|-------------|------------------------|-------------------------------------|
| Bold        | `**text**`             | Wraps selection or inserts placeholder |
| Italic      | `*text*`               | Wraps selection or inserts placeholder |
| Heading     | `## text`              | Prepends `## ` to line             |
| Bullet List | `- item`               | Prepends `- ` to line              |
| Numbered List | `1. item`            | Prepends `1. ` to line             |
| Link        | `[text](url)`          | Wraps selection as link text        |

## MarkdownRenderer Component

### Props

| Prop    | Type   | Required | Description                      |
|---------|--------|----------|----------------------------------|
| content | string | yes      | Raw Markdown string to render    |

### Behavior

- Converts Markdown to HTML
- Sanitizes HTML to remove all script injection vectors
- Wraps output in a container with the `.prose` CSS class
- Links get `target="_blank"` and `rel="noopener noreferrer"`
- Empty/null content renders nothing (no errors)

## Integration Points

### FormField.vue

When `attribute.display.widget === "textarea"`, render `<MarkdownEditor>` instead of `<textarea>`. Bind to the same reactive attribute value via v-model.

### ItemDetail.vue

For textarea-type attributes in the detail view, render `<MarkdownRenderer :content="value">` instead of displaying the raw string.
