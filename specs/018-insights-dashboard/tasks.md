# Tasks: Insights Dashboard

**Input**: Design documents from `/specs/018-insights-dashboard/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Install dependencies and create project structure

- [x] T001 Install chart.js and vue-chartjs dependencies in frontend/package.json
- [x] T002 Create frontend/src/composables/ directory structure

---

## Phase 2: Foundational (Shared Composable)

**Purpose**: Core metric computation logic that all user stories depend on

- [x] T003 Implement useDashboardMetrics composable in frontend/src/composables/useDashboardMetrics.ts -- exports computed DashboardMetrics (totalValue, totalItems, mostValuableItem, moduleBreakdown, acquisitionTimeline) from collectionStore items and moduleStore modules
- [x] T004 Write Vitest tests for useDashboardMetrics in frontend/src/composables/__tests__/useDashboardMetrics.test.ts -- cover: empty items, mixed null prices, multiple modules, "Other" grouping at 7+ modules, monthly bucketing, single-month edge case

**Checkpoint**: Composable tested and passing before UI work begins

---

## Phase 3: User Story 1 - Financial Summary Cards (Priority: P1)

**Goal**: Three glassmorphism summary cards showing Total Collection Value, Total Items, and Most Valuable Item

**Independent Test**: Select "All Types", verify cards show correct aggregated values

### Implementation for User Story 1

- [x] T005 [P] [US1] Create DashboardMetricCard.vue in frontend/src/components/DashboardMetricCard.vue -- reusable glassmorphism card component with props: label, value, subtitle, clickable; uses --bg-secondary, --glass-blur, --border-primary, --shadow-sm tokens
- [x] T006 [P] [US1] Create DashboardView.vue in frontend/src/components/DashboardView.vue -- layout with three summary cards row (Total Collection Value, Total Items, Most Valuable Item) using DashboardMetricCard; empty state when no items; click handler on Most Valuable Item card emits item selection
- [x] T007 [US1] Wire DashboardView into App.vue -- add showDashboard ref (defaults true), render DashboardView when activeModuleId is empty and showDashboard is true; add dashboard/grid-list toggle button; reset showDashboard to true when setFilter changes to "All Types"

**Checkpoint**: Summary cards visible and correct when "All Types" selected; toggle switches to grid/list

---

## Phase 4: User Story 2 - Collection Breakdown Doughnut Chart (Priority: P2)

**Goal**: Doughnut chart showing collection value distribution by module with interactive tooltips

**Independent Test**: Create items across 3+ modules with prices, verify chart shows correct value proportions

### Implementation for User Story 2

- [x] T008 [US2] Register Chart.js tree-shaken modules in frontend/src/components/DashboardView.vue -- import and register only: ArcElement, DoughnutController, Tooltip, Legend from chart.js
- [x] T009 [US2] Add doughnut chart to DashboardView.vue -- vue-chartjs Doughnut component using moduleBreakdown from useDashboardMetrics; chart colors derived from CSS variables via getComputedStyle; tooltips show module name, value, item count, percentage; empty state placeholder when no data
- [x] T010 [US2] Implement theme reactivity for doughnut chart -- watch for theme changes (dark mode toggle), recompute chart colors from CSS variables, update chart options reactively

**Checkpoint**: Doughnut chart renders with correct proportions and updates on theme change

---

## Phase 5: User Story 3 - Acquisitions Over Time Bar Chart (Priority: P3)

**Goal**: Bar chart showing monthly item acquisition counts

**Independent Test**: Create items with different creation dates, verify bar chart shows correct monthly counts

### Implementation for User Story 3

- [x] T011 [US3] Register Bar chart modules in DashboardView.vue -- add BarElement, BarController, CategoryScale, LinearScale to Chart.js registration
- [x] T012 [US3] Add bar chart to DashboardView.vue -- vue-chartjs Bar component using acquisitionTimeline from useDashboardMetrics; month labels on x-axis, item counts on y-axis; tooltips show month and count; empty state placeholder; theme-reactive colors

**Checkpoint**: Bar chart renders with correct monthly counts and updates on theme change

---

## Phase 6: Polish & Documentation

**Purpose**: Final validation and documentation updates

- [x] T013 [P] Update CLAUDE.md -- add DashboardView.vue, DashboardMetricCard.vue, useDashboardMetrics.ts to component list; add chart.js, vue-chartjs to dependency list; document dashboard view behavior
- [x] T014 [P] Run go vet ./... and cd frontend && npm test to verify no regressions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1
- **Phase 3 (US1)**: Depends on Phase 2; T005 and T006 can run in parallel, T007 depends on both
- **Phase 4 (US2)**: Depends on Phase 3 (DashboardView.vue must exist); T008 before T009 before T010
- **Phase 5 (US3)**: Depends on Phase 4 (chart registration pattern established); T011 before T012
- **Phase 6 (Polish)**: Depends on all story phases

### Parallel Opportunities

- T005 and T006 can run in parallel (different files)
- T013 and T014 can run in parallel (different concerns)

---

## Notes

- No backend changes required (FR-011)
- Chart.js must be tree-shaken -- import only needed controllers/elements
- All chart colors must read from CSS variables for theme reactivity
- Dashboard toggle is session-only (ref in App.vue, not persisted)
