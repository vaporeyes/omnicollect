# Research: Markdown Textarea Support

**Branch**: `008-markdown-textarea` | **Date**: 2026-04-07

## R1: Markdown Editor Library Choice

**Decision**: Use CodeMirror 6 with `@codemirror/lang-markdown` extension, wrapped in the existing `vue-codemirror` package.

**Rationale**: CodeMirror 6 and vue-codemirror are already project dependencies (used by SchemaCodeEditor for JSON editing). Adding `@codemirror/lang-markdown` leverages the same infrastructure with minimal new dependency weight. The editor gets syntax highlighting, and we build a thin toolbar component on top.

**Alternatives considered**:
- Milkdown: Full WYSIWYG Markdown editor. Rejected -- heavy dependency (~200KB+), not already in the project, and WYSIWYG can introduce HTML artifacts. CodeMirror keeps content as pure text.
- TipTap/ProseMirror: Rich text editor. Rejected -- stores content as structured JSON/HTML, violating the requirement to persist raw Markdown strings.
- Plain textarea with toolbar: Simpler but no syntax highlighting or bracket matching. CodeMirror is already available and provides a much better editing experience.

## R2: Markdown Rendering Library Choice

**Decision**: Use `marked` for Markdown-to-HTML conversion with `DOMPurify` for sanitization.

**Rationale**: `marked` is lightweight (~40KB), fast, and widely used. It converts Markdown to HTML strings which Vue can render via `v-html`. `DOMPurify` ensures all rendered HTML is safe from XSS by stripping script tags and event handlers. The two-library approach (parser + sanitizer) is the standard secure pattern.

**Alternatives considered**:
- markdown-it: Equally capable but slightly larger. No meaningful advantage for our use case.
- remark/rehype: AST-based pipeline. More powerful but overkill for simple rendering. Adds multiple dependencies.
- Rendering without sanitization: Rejected -- Constitution doesn't require it but security best practice demands sanitization of any user-authored HTML.

## R3: Toolbar Design

**Decision**: Build a minimal custom toolbar as a row of icon buttons above the CodeMirror editor. Buttons: Bold, Italic, Heading, Bullet List, Numbered List, Link. Each button inserts the corresponding Markdown syntax at the cursor position.

**Rationale**: A custom toolbar keeps the UI consistent with the app's design language (Outfit font, theme variables). CodeMirror's API provides `replaceSelection` and cursor manipulation for inserting syntax. Five to six buttons is the right density -- enough to be useful, not enough to overwhelm.

**Alternatives considered**:
- No toolbar (syntax-only editing): Rejected -- reduces discoverability for users unfamiliar with Markdown.
- Full formatting toolbar (tables, code blocks, images, etc.): Rejected per spec -- v1 scope is limited to essential formatting.

## R4: Link Handling in Rendered Markdown

**Decision**: Add `target="_blank"` and `rel="noopener noreferrer"` to all rendered links. In the Wails webview, external links automatically open in the system browser.

**Rationale**: Wails webview intercepts navigation by default. Setting target="_blank" triggers the system browser for external URLs, which is the expected behavior for a desktop app. The `rel` attributes prevent the target page from accessing the source window.

**Alternatives considered**:
- No special handling: Rejected -- clicking a link would navigate the entire app away from OmniCollect.

## R5: Prose CSS Class Scope

**Decision**: Define `.prose` as a global class in `style.css` that styles all Markdown HTML elements (h1-h4, p, ul, ol, li, blockquote, code, pre, a, strong, em) with the app's typography. Headings use Instrument Serif; body/lists use Outfit; code uses system monospace.

**Rationale**: A global class is reusable anywhere Markdown is rendered (detail view, potentially future preview panes). Scoping to `.prose` prevents styles from leaking into non-Markdown UI elements. This is the same pattern used by Tailwind's `@tailwindcss/typography` plugin.

**Alternatives considered**:
- Scoped styles per component: Rejected -- would duplicate the same styles in every component that renders Markdown.
- Inline styles on each element: Rejected -- unmaintainable and inconsistent.
