# Feature Specification: Insights Dashboard

**Feature Branch**: `018-insights-dashboard`  
**Created**: 2026-04-10  
**Status**: Draft  
**Input**: User description: "Build a premium Insights Dashboard with financial metric cards, collection breakdown doughnut chart, acquisitions over time chart, and theme-reactive data visualization."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Financial Summary at a Glance (Priority: P1)

A collector opens the app and lands on the "All Types" view. Instead of jumping straight into a grid of items, they see a dashboard with summary cards showing their total collection value, total item count, and their most valuable item. This gives them an immediate sense of their collection's scope and worth without scrolling through individual items.

**Why this priority**: The summary cards are the core value proposition -- they transform raw data into meaningful insights with minimal complexity. They require no new dependencies and deliver immediate user value.

**Independent Test**: Can be fully tested by creating several items with purchase prices across multiple modules, then selecting "All Types" and verifying the summary cards display correct aggregated values.

**Acceptance Scenarios**:

1. **Given** a user has items with purchase prices across multiple modules, **When** they select "All Types", **Then** they see a dashboard with three summary cards: Total Collection Value, Total Items, and Most Valuable Item.
2. **Given** a user has items where some have no purchase price, **When** they view the dashboard, **Then** Total Collection Value sums only items with a price, and Total Items counts all items regardless of price.
3. **Given** a user has no items, **When** they view the dashboard, **Then** the cards show $0.00 for value, 0 for items, and a placeholder message for most valuable item.
4. **Given** a user is viewing the dashboard, **When** they click the Most Valuable Item card, **Then** they navigate to that item's detail view.

---

### User Story 2 - Understand Collection Composition by Module (Priority: P2)

A collector wants to understand the value breakdown of their collection. A doughnut chart on the dashboard shows how collection value is distributed across modules (e.g., 60% Coins by value, 30% Vinyl, 10% Stamps). Hovering or tapping a segment reveals the exact value, item count, and percentage.

**Why this priority**: The composition chart provides the most requested visual insight -- understanding what makes up the collection. It requires a charting library but is a single, self-contained visualization.

**Independent Test**: Can be tested by creating items across 3+ modules, selecting "All Types", and verifying the doughnut chart displays correct proportions with interactive tooltips showing counts and percentages.

**Acceptance Scenarios**:

1. **Given** a user has items with prices in 3 modules, **When** they view the dashboard, **Then** a doughnut chart shows segments proportional to each module's total value with correct labels.
2. **Given** a user interacts with a chart segment, **When** they hover (desktop) or tap (touch), **Then** a tooltip shows the module name, total value, item count, and percentage of total collection value.
3. **Given** a user has items in only one module, **When** they view the dashboard, **Then** the chart shows a single full segment for that module.
4. **Given** the user switches between light and dark mode, **When** the theme changes, **Then** the chart colors, labels, and tooltip styling update immediately to match the active theme.

---

### User Story 3 - Track Acquisitions Over Time (Priority: P3)

A collector is curious about their collecting habits. A bar chart on the dashboard shows how many items they have added over time, grouped by month. This helps them see periods of active collecting versus quiet stretches.

**Why this priority**: The timeline chart adds depth to the dashboard but depends on the same charting infrastructure as P2. It provides secondary insight (collecting patterns over time) that is valuable but not essential for a first release.

**Independent Test**: Can be tested by creating items with different creation dates spanning multiple months, selecting "All Types", and verifying the bar chart shows correct monthly counts.

**Acceptance Scenarios**:

1. **Given** a user has items created across several months, **When** they view the dashboard, **Then** a bar chart shows item counts grouped by month with readable date labels.
2. **Given** a user hovers over a bar, **When** the tooltip appears, **Then** it shows the month label and exact item count.
3. **Given** all items were created in the same month, **When** the chart renders, **Then** it shows a single bar rather than an empty or broken chart.
4. **Given** the user switches between light and dark mode, **When** the theme changes, **Then** the chart colors, gridlines, and labels update to match the active theme.

---

### Edge Cases

- What happens when a user has zero items? Dashboard shows empty-state cards with zeros and placeholder text for charts (e.g., "Add items to see your collection breakdown").
- What happens when no items have a purchase price? Total Collection Value shows $0.00; Most Valuable Item card shows "No prices recorded".
- What happens with a very large number of modules (10+)? The doughnut chart groups the smallest modules into an "Other" segment to keep the chart readable.
- What happens when the createdAt timestamps span years? The bar chart adapts its time axis granularity (months for under 2 years, quarters or years for longer ranges).
- How does the dashboard coexist with the existing grid/list view? The dashboard appears as the default view when "All Types" is selected, with a toggle to switch to the traditional grid/list view.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a dashboard view as the default landing page when "All Types" is selected in the module selector.
- **FR-002**: System MUST show a "Total Collection Value" summary card that sums `purchasePrice` across all items, displaying the result in a currency format.
- **FR-003**: System MUST show a "Total Items" summary card with the count of all items across all modules.
- **FR-004**: System MUST show a "Most Valuable Item" summary card displaying the title and price of the item with the highest `purchasePrice`, linking to its detail view.
- **FR-005**: System MUST render a doughnut chart showing collection value distribution by module, where each segment represents one module's proportion of total collection value (sum of purchasePrice). Modules with no priced items show as zero-value segments.
- **FR-006**: System MUST render a bar chart showing item acquisitions over time, grouped by month using item creation timestamps.
- **FR-007**: Both charts MUST display interactive tooltips on hover/tap showing exact values (count, percentage, month label).
- **FR-008**: All dashboard visualizations MUST reactively update their colors, text, and styling when the user switches between light and dark themes, using the existing theme variable system.
- **FR-009**: System MUST provide a toggle allowing users to switch between dashboard view and the existing grid/list view while in "All Types" mode. The toggle is session-only; the dashboard is always the default view on each visit.
- **FR-010**: The doughnut chart MUST group modules into an "Other" category when more than 6 modules exist, consolidating the smallest segments.
- **FR-011**: Dashboard data MUST be computed from items already loaded in the frontend, requiring no new backend API endpoints.
- **FR-012**: Summary cards MUST use the existing glassmorphism design tokens (semi-transparent backgrounds, backdrop blur, border styling) for visual consistency.

### Key Entities

- **Dashboard Metrics**: Computed aggregation of item data -- total value (sum of purchasePrice), total count, most valuable item reference, per-module item counts, and monthly acquisition counts.
- **Chart Segment**: A visual slice of the doughnut chart representing one module (or "Other"), containing module name, total value, item count, and percentage of total collection value.
- **Time Bucket**: A monthly grouping of items by creation timestamp for the acquisitions bar chart, containing month label and item count.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can see total collection value, item count, and most valuable item within 1 second of selecting "All Types".
- **SC-002**: Collection breakdown chart accurately reflects item distribution across all modules to within rounding precision.
- **SC-003**: Acquisitions timeline correctly groups all items by their creation month with no items omitted.
- **SC-004**: Theme changes (light/dark toggle) update all chart colors and card styling within 0.5 seconds with no page reload.
- **SC-005**: Dashboard remains usable and readable with collections ranging from 0 to 10,000+ items.
- **SC-006**: Users can switch between dashboard and grid/list view without losing their selection state.

## Clarifications

### Session 2026-04-10

- Q: Should the doughnut chart show proportion by item count or by collection value? → A: By total value per module (sum of purchasePrice).
- Q: Should the dashboard/grid view toggle persist across sessions? → A: Session-only; always defaults to dashboard on each visit.

## Assumptions

- Currency formatting uses a simple fixed format (e.g., $1,234.56) since there is no user locale/currency setting in the application today. A future iteration could add currency preferences.
- The charting library will be lightweight (under 50KB gzipped) to maintain fast load times consistent with the desktop app's performance profile.
- All dashboard computations happen client-side from items already fetched by the collection store; no new backend endpoints or database queries are needed.
- The "All Types" module selector state (empty activeModuleId) is the sole trigger for showing the dashboard; selecting a specific module returns to the standard grid/list view.
- The existing Poline theme system exposes sufficient CSS variables for the charting library to read and react to theme changes.
- Monthly granularity is the default for the acquisitions chart; finer granularity (weekly/daily) is out of scope for v1.
