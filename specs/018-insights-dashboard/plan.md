# Implementation Plan: Insights Dashboard

**Branch**: `018-insights-dashboard` | **Date**: 2026-04-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/018-insights-dashboard/spec.md`

## Summary

Add a DashboardView component that appears as the default landing page when "All Types" is selected (empty activeModuleId). Displays three glassmorphism summary cards (Total Collection Value, Total Items, Most Valuable Item) and two interactive charts (doughnut for value-by-module, bar for acquisitions-over-time). All data is computed client-side from the existing collectionStore items array. Charts use Chart.js (tree-shaken) and react to Poline theme changes via CSS variable reads.

## Technical Context

**Language/Version**: TypeScript 4.6+ (frontend), Go 1.25+ (backend -- no changes)
**Primary Dependencies**: Vue 3.2+, Pinia 3.0+, Chart.js 4.x (new), vue-chartjs 5.x (new)
**Storage**: N/A (no backend changes; all computation is client-side from existing store data)
**Testing**: Vitest (frontend unit tests for metric computation logic)
**Target Platform**: Desktop (Wails) + standalone HTTP server (browser)
**Project Type**: Desktop app (Wails v2 shell)
**Performance Goals**: Dashboard renders within 1 second; theme transitions within 0.5 seconds
**Constraints**: Chart.js tree-shaken bundle under 50KB gzipped; no new backend API endpoints
**Scale/Scope**: Collections from 0 to 10,000+ items; up to 20+ modules

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First Mandate | PASS | Dashboard computes from local data already in the store. No network dependency. |
| II. Schema-Driven UI | PASS | Dashboard is a cross-module aggregate view, not a type-specific form. No hardcoded collection templates. |
| III. Flat Data Architecture | PASS | No new database queries. Reads existing item fields (purchasePrice, moduleId, createdAt). |
| IV. Performance & Memory Protection | PASS | Dashboard renders only computed aggregates, not images. No media loaded. |
| V. Type-Safe IPC | PASS | No new IPC/API calls. Frontend-only feature consuming typed Item[] from store. |
| VI. Documentation is Paramount | REQUIRES | CLAUDE.md and README must be updated after implementation. |

## Project Structure

### Documentation (this feature)

```text
specs/018-insights-dashboard/
  plan.md              # This file
  research.md          # Charting library decision
  data-model.md        # Computed entity definitions
  quickstart.md        # Setup and dev guide
  tasks.md             # Implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
frontend/src/
  components/
    DashboardView.vue       # Main dashboard layout: cards + charts + empty states
    DashboardMetricCard.vue  # Reusable glassmorphism summary card
  composables/
    useDashboardMetrics.ts   # Computed metrics from collectionStore items
  stores/
    collectionStore.ts       # Existing (add showDashboard session ref)
```

**Structure Decision**: Two new Vue components in the existing `components/` directory, one new composable in a new `composables/` directory for metric computation logic (testable in isolation). The composable pattern separates data transformation from rendering, keeping the dashboard components focused on display.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
