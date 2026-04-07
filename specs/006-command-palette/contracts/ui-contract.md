# UI Contract: CommandPalette Component

**Branch**: `006-command-palette` | **Date**: 2026-04-07

## Component Interface

### Props

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| visible | boolean | yes | Controls palette visibility |

### Emits

| Event | Payload | Description |
|-------|---------|-------------|
| close | none | User dismissed the palette (Escape, click-outside, or Cmd/Ctrl+K toggle) |
| selectItem | Item | User selected a collection item result |
| action | string | User selected a quick action (action identifier) |

### Quick Action Identifiers

| Identifier | Triggered By Keywords | App Behavior |
|------------|----------------------|--------------|
| "newItem" | "new", "add", "create" | Open new item form for active module |
| "newSchema" | "new", "schema", "create" | Open schema builder |
| "openSettings" | "settings", "preferences" | Open settings page |
| "exportBackup" | "backup", "export" | Trigger backup export |

## Keyboard Contract

| Key | Palette Closed | Palette Open |
|-----|---------------|--------------|
| Cmd/Ctrl+K | Open palette | Close palette |
| Escape | (no effect) | Close palette |
| Down Arrow | (no effect) | Highlight next result |
| Up Arrow | (no effect) | Highlight previous result |
| Enter | (no effect) | Select highlighted result |
| Any character | (no effect) | Append to search query |

## Visual Contract

- Overlay: full-screen backdrop with blur and semi-transparent background
- Dialog: centered, max-width ~600px, frosted glass background
- Input: large font, prominent, auto-focused
- Results: scrollable list below input, max ~25 items
- Quick actions: grouped above item results with distinct styling
- Highlighted result: visually distinct background color
- Thumbnails: 40x40px from `/thumbnails/` path; placeholder SVG for items without images
