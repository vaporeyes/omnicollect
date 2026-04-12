# Research: Masonry Grid & Item Comparison

**Branch**: `020-masonry-compare` | **Date**: 2026-04-11

## R1: CSS Masonry Layout Approach

**Decision**: Use CSS `column-count` (multi-column layout) for the masonry effect.

**Rationale**: CSS Grid `masonry` value (from CSS Grid Level 3) is still experimental and not supported in all browsers. CSS `column-count` is universally supported, produces a true masonry layout with variable-height items, and requires no JavaScript layout library. It naturally fills columns top-to-bottom, left-to-right, which matches the Pinterest-style visual the spec calls for.

**Alternatives considered**:
- CSS Grid `grid-template-rows: masonry` -- Not yet supported in Chrome/Edge (Firefox only behind flag). Would be ideal when standardized but too risky for production today.
- JavaScript masonry libraries (Masonry.js, Isotope) -- Adds a dependency, requires resize observers, and fights Vue's reactivity. Overkill for this use case.
- CSS Grid with `grid-auto-rows: 1px` and `span` calculation -- Requires JavaScript to measure image heights and set `grid-row-end` spans. Works but complex and fragile.

**Implementation notes**:
- `column-count` with `column-gap` replaces `display: grid` and `grid-template-columns`.
- Cards use `break-inside: avoid` to prevent splitting across columns.
- Column count is responsive via media queries or `column-width` (e.g., `column-width: 240px` lets the browser pick column count).
- Card order is column-first (top-to-bottom per column) rather than row-first. This is the standard masonry behavior and matches user expectations.
- The staggered entrance animation (per-card delay) still works since cards render in DOM order.

## R2: Image Aspect Ratio in Cards

**Decision**: Remove `aspect-ratio: 1` from `.card-image`. Let images render at natural dimensions with `width: 100%` and `height: auto`.

**Rationale**: Thumbnails already preserve the original aspect ratio (confirmed in clarification). Removing the forced square crop and letting images fill their natural height is the minimal change needed.

**Implementation notes**:
- Replace `aspect-ratio: 1` with `display: block; width: 100%; height: auto`.
- Keep `object-fit: cover` as a fallback only if a max-height is desired (spec says no -- use natural proportions).
- The frosted glass caption overlay is positioned absolutely at the bottom of `.card-image`, so it naturally stays at the image's bottom edge regardless of height.

## R3: Comparison View Layout

**Decision**: Two-column CSS grid layout (`grid-template-columns: 1fr 1fr`) with a shared navigation control bar between the galleries and a unified diff table below.

**Rationale**: A simple two-column grid is the clearest side-by-side layout. The shared navigation bar (centered between both galleries) reinforces the synchronized browsing concept. The diff table as a single table with three columns (Label | Item A | Item B) is more scannable than two separate attribute lists.

**Alternatives considered**:
- Two independent ItemDetail components side by side -- Too heavy; ItemDetail includes edit/delete UI and navigation that are irrelevant in comparison mode. Better to build a focused comparison component.
- Tab-based switching between items -- Defeats the side-by-side comparison purpose.

## R4: Attribute Diff Highlighting

**Decision**: Use a subtle background color on rows where values differ. Same color in light and dark mode (using CSS variables for theme adaptation).

**Rationale**: Background highlighting is the most accessible approach -- it works for colorblind users when combined with the inherent text difference, and it is immediately visible without requiring icon interpretation.

**Implementation notes**:
- Diff rows get a CSS class (e.g., `.diff-row`) with a light yellow/amber background in light mode, desaturated amber in dark mode.
- Tags comparison: compare as sorted comma-joined strings; highlight if different.
- Price comparison: highlight if values differ (handle null vs 0 as different).
- Schema attributes: simple string/number equality comparison per field.
- Missing attributes (cross-module comparison): always highlighted since one side is empty.

## R5: Responsive Behavior for Comparison View

**Decision**: Stack vertically (single column) when viewport width drops below 768px.

**Rationale**: Side-by-side comparison becomes unreadable below ~700px. Stacking vertically preserves usability on narrow windows while keeping the diff table scannable.

**Implementation notes**:
- Media query `@media (max-width: 768px)` switches from `grid-template-columns: 1fr 1fr` to `1fr`.
- Gallery sync still works in stacked mode; the user scrolls vertically between the two items.
