# Implementation Plan: Markdown Textarea Support

**Branch**: `008-markdown-textarea` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-markdown-textarea/spec.md`

## Summary

Upgrade the `widget: "textarea"` schema hint to render a Markdown-aware CodeMirror editor with a minimal formatting toolbar. In the item detail view, render Markdown content as sanitized HTML using marked + DOMPurify. Add a global `.prose` CSS class for consistent Markdown typography using the Outfit/Instrument Serif font pairing.

## Technical Context

**Language/Version**: Go 1.25+ (backend, no changes needed), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Wails v2, Pinia, vue-codemirror (existing), CodeMirror markdown extensions (new), marked (new), DOMPurify (new)
**Storage**: SQLite (no changes -- raw Markdown stored as string in existing JSON attributes)
**Testing**: Manual acceptance testing
**Target Platform**: macOS desktop (Wails v2)
**Project Type**: Desktop application (Go + Vue via Wails)
**Performance Goals**: Editor load under 200ms; Markdown render under 50ms for typical content
**Constraints**: Offline-capable; no HTML stored in database; all rendered HTML sanitized
**Scale/Scope**: Single-user desktop app; textarea fields contain typically 50-500 words

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS | No network dependencies; Markdown parsing/rendering is client-side |
| II. Schema-Driven UI | PASS | Editor is triggered by the existing `widget: "textarea"` hint; no type-specific templates |
| III. Flat Data Architecture | PASS | Raw Markdown stored as string in existing JSON attributes column; no schema changes |
| IV. Performance & Memory | PASS | No media changes; editor and renderer are lightweight |
| V. Type-Safe IPC | PASS | No backend changes; no new Wails bindings |
| VI. Documentation | PASS | Spec artifacts produced; CLAUDE.md and README update required at completion |

All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/008-markdown-textarea/
  plan.md              # This file
  research.md          # Phase 0 output
  data-model.md        # Phase 1 output
  quickstart.md        # Phase 1 output
  contracts/           # Phase 1 output
  spec.md              # Feature specification
  checklists/          # Quality checklists
```

### Source Code (repository root)

```text
# Frontend only -- no backend changes needed
frontend/src/
  components/
    MarkdownEditor.vue     # New: CodeMirror markdown editor with toolbar
    MarkdownRenderer.vue   # New: Safe Markdown-to-HTML renderer
    FormField.vue          # Modified: use MarkdownEditor when widget is "textarea"
    ItemDetail.vue         # Modified: use MarkdownRenderer for textarea attributes
  style.css                # Modified: add .prose typography class
```

**Structure Decision**: Two new components (editor + renderer) plus modifications to existing FormField and ItemDetail. No backend changes -- Markdown is stored as a raw string in the existing attributes JSON, same as plain textarea text was before.

## Complexity Tracking

No constitution violations. No complexity justification needed.
