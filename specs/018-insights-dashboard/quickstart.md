# Quickstart: Insights Dashboard

**Branch**: `018-insights-dashboard` | **Date**: 2026-04-10

## Prerequisites

- Node.js (for frontend dev)
- Go 1.25+ (for backend, though no backend changes in this feature)
- Wails v2 CLI (for desktop dev mode)

## Setup

```bash
# Install new frontend dependencies
cd frontend
npm install chart.js vue-chartjs

# Verify dev server starts
cd ..
wails dev
```

## Development Workflow

1. **Composable first**: Implement `useDashboardMetrics.ts` with pure computation logic. Write Vitest tests against it using mock Item arrays.

2. **Cards next**: Build `DashboardMetricCard.vue` using existing glassmorphism tokens (`--bg-secondary`, `--glass-blur`, `--border-primary`, `--shadow-sm`).

3. **Charts**: Build chart components using vue-chartjs `<Doughnut>` and `<Bar>`. Register only needed Chart.js modules (tree-shaking).

4. **Integration**: Wire `DashboardView.vue` into `App.vue` conditional rendering alongside existing grid/list views.

## Testing

```bash
# Run frontend unit tests
cd frontend && npm test

# Test specific dashboard composable
cd frontend && npx vitest run --reporter=verbose useDashboardMetrics
```

## Key Files to Modify

| File | Change |
|------|--------|
| `frontend/src/App.vue` | Add showDashboard ref, conditional DashboardView rendering when activeModuleId is empty |
| `frontend/src/components/DashboardView.vue` | New: main dashboard layout |
| `frontend/src/components/DashboardMetricCard.vue` | New: reusable glassmorphism card |
| `frontend/src/composables/useDashboardMetrics.ts` | New: computed metrics from items |
| `frontend/package.json` | Add chart.js, vue-chartjs dependencies |

## Verification

1. Select "All Types" in module selector -- dashboard should appear
2. Toggle view icon -- should switch between dashboard and grid/list
3. Select a specific module -- should show grid/list (no dashboard)
4. Toggle light/dark theme -- chart colors should update immediately
5. Empty collection -- should show placeholder cards and empty chart messages
