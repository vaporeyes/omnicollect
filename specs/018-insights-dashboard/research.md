# Research: Insights Dashboard

**Branch**: `018-insights-dashboard` | **Date**: 2026-04-10

## R1: Charting Library Selection

**Decision**: Chart.js 4.x with vue-chartjs 5.x wrapper

**Rationale**:
- Chart.js 4.x supports tree-shaking via named imports (register only Doughnut + Bar controllers, tooltip plugin, scales). Tree-shaken bundle is approximately 40-45KB gzipped, under the 50KB target.
- vue-chartjs provides reactive Vue 3 components (`<Doughnut>`, `<Bar>`) that re-render on prop changes. Integrates naturally with Vue's reactivity system.
- Chart.js reads colors from JavaScript config objects, so theme reactivity is achieved by watching CSS variable changes and updating chart options/data objects -- Chart.js re-renders on prop changes via vue-chartjs.
- Mature ecosystem with extensive documentation; most commonly used charting library in Vue projects.

**Alternatives considered**:
- **ECharts** (via vue-echarts): More powerful but significantly larger (~200KB min gzipped even with tree-shaking). Overkill for two charts.
- **Lightweight Charts** (TradingView): Optimized for financial time-series, not general-purpose doughnut/bar charts. Wrong tool.
- **D3.js**: Maximum flexibility but requires building chart primitives from scratch. Too much effort for two standard chart types.
- **unovis**: Newer, smaller, but immature Vue integration and sparse documentation.

## R2: Theme Reactivity Strategy

**Decision**: Read CSS variables via `getComputedStyle()` on theme change, rebuild Chart.js color config

**Rationale**:
- Poline injects theme colors as CSS variables on `:root` via `applyPolineTheme()`. Chart.js does not natively read CSS variables for colors.
- The composable will expose a reactive `chartColors` object computed from CSS variables. A MutationObserver on `document.documentElement.style` (or a watcher on the existing `systemDark` ref pattern from App.vue) triggers recomputation.
- Simpler approach: since App.vue already manages theme state reactively, the dashboard can accept a `dark` boolean prop (or read a shared reactive ref) and recompute chart colors when it changes. This avoids MutationObserver complexity.
- Chart.js tooltip styling can use CSS classes with the `external` tooltip handler, inheriting theme variables directly.

**Alternatives considered**:
- **CSS-only theming**: Chart.js renders to `<canvas>`, which cannot use CSS variables for fill colors. Requires JS-side color injection.
- **MutationObserver on :root style**: Works but adds complexity. Prop-based reactivity via Vue is simpler and more predictable.

## R3: Dashboard/Grid View Toggle Approach

**Decision**: Add a `showDashboard` ref to the view state in App.vue (session-only, not persisted)

**Rationale**:
- The toggle is session-only per spec clarification. A simple `ref(true)` initialized on mount (defaults to dashboard) is sufficient.
- When `activeModuleId` is empty and `showDashboard` is true, render DashboardView. When false, render existing grid/list. When a specific module is selected, always show grid/list.
- Resetting `showDashboard = true` when user returns to "All Types" ensures dashboard is always the default.

**Alternatives considered**:
- **Separate route/page**: OmniCollect is a single-page app without a router. Adding routing for one view toggle is unnecessary.
- **Store-level state**: Adds complexity vs a local ref in App.vue. Since this is session-only and view-layer state, it belongs in the view component.

## R4: "Other" Category Grouping for Doughnut Chart

**Decision**: Sort modules by value descending, keep top 5, merge rest into "Other"

**Rationale**:
- FR-010 specifies grouping when more than 6 modules exist. Top 5 by value + "Other" ensures the most valuable segments are always visible.
- The "Other" segment sums the values and counts of all remaining modules.
- If fewer than 7 modules exist, no grouping is applied.

## R5: Time Axis Granularity for Bar Chart

**Decision**: Default monthly grouping with adaptive axis when data spans over 24 months

**Rationale**:
- Per spec, monthly granularity is the default. For collections spanning over 2 years, grouping by quarter provides better readability.
- Chart.js time scale with `unit: 'month'` handles this natively when using the date adapter (chartjs-adapter-date-fns or similar lightweight adapter).
- For simplicity in v1, use manual month bucketing (parse createdAt, group by YYYY-MM) rather than Chart.js time scale. This avoids the date adapter dependency and keeps the bundle smaller.
