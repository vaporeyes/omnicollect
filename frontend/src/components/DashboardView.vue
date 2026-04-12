<script lang="ts" setup>
// ABOUTME: Insights dashboard showing financial summary cards, value-by-module doughnut chart,
// ABOUTME: and acquisitions-over-time bar chart. Displayed when "All Types" is selected.

import {computed, watch, ref} from 'vue'
import type {Item, ModuleSchema} from '../api/types'
import {useDashboardMetrics} from '../composables/useDashboardMetrics'
import DashboardMetricCard from './DashboardMetricCard.vue'
import {Chart as ChartJS, ArcElement, DoughnutController, BarElement, BarController, CategoryScale, LinearScale, Tooltip, Legend} from 'chart.js'
import {Doughnut, Bar} from 'vue-chartjs'

ChartJS.register(ArcElement, DoughnutController, BarElement, BarController, CategoryScale, LinearScale, Tooltip, Legend)

const props = defineProps<{
  items: Item[]
  modules: ModuleSchema[]
  dark: boolean
}>()

const emit = defineEmits<{
  (e: 'selectItem', id: string): void
}>()

const itemsRef = computed(() => props.items)
const modulesRef = computed(() => props.modules)
const {metrics} = useDashboardMetrics(itemsRef, modulesRef)

function formatCurrency(value: number): string {
  return '$' + value.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})
}

// Chart color palette derived from CSS variables
const chartColors = ref<string[]>([])
const textColor = ref('')
const gridColor = ref('')
const tooltipBg = ref('')

function readThemeColors() {
  const style = getComputedStyle(document.documentElement)
  textColor.value = style.getPropertyValue('--text-secondary').trim() || '#888'
  gridColor.value = style.getPropertyValue('--border-primary').trim() || '#e0e0e0'
  tooltipBg.value = style.getPropertyValue('--bg-tertiary').trim() || '#333'

  // Build a palette from the accent + some derived hues
  const accent = style.getPropertyValue('--accent-blue').trim() || '#3b82f6'
  chartColors.value = [
    accent,
    'hsl(280, 55%, 55%)',
    'hsl(160, 50%, 45%)',
    'hsl(35, 75%, 55%)',
    'hsl(350, 60%, 55%)',
    'hsl(200, 50%, 50%)',
  ]
}

// Read colors on mount and when theme changes
readThemeColors()
watch(() => props.dark, () => {
  // Slight delay to allow CSS variables to update
  requestAnimationFrame(readThemeColors)
})

// Doughnut chart data: show value breakdown when prices exist, item count otherwise
const useCountFallback = computed(() => metrics.value.totalValue === 0)

const doughnutData = computed(() => {
  const breakdown = metrics.value.moduleBreakdown
  return {
    labels: breakdown.map(s => s.moduleName),
    datasets: [{
      data: breakdown.map(s => useCountFallback.value ? s.itemCount : s.totalValue),
      backgroundColor: breakdown.map((_, i) => chartColors.value[i % chartColors.value.length] || '#888'),
      borderWidth: 0,
      hoverOffset: 6,
    }],
  }
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '60%',
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: textColor.value,
        padding: 16,
        usePointStyle: true,
        pointStyleWidth: 10,
        font: {size: 12},
      },
    },
    tooltip: {
      backgroundColor: tooltipBg.value,
      titleColor: textColor.value,
      bodyColor: textColor.value,
      padding: 10,
      cornerRadius: 8,
      callbacks: {
        label(ctx: any) {
          const segment = metrics.value.moduleBreakdown[ctx.dataIndex]
          if (!segment) return ''
          if (useCountFallback.value) {
            return `${segment.moduleName}: ${segment.itemCount} items`
          }
          const pct = segment.percentage.toFixed(1)
          return `${segment.moduleName}: $${segment.totalValue.toLocaleString()} (${segment.itemCount} items, ${pct}%)`
        },
      },
    },
  },
}))

// Bar chart data
const barData = computed(() => {
  const timeline = metrics.value.acquisitionTimeline
  return {
    labels: timeline.map(t => t.label),
    datasets: [{
      label: 'Items Added',
      data: timeline.map(t => t.count),
      backgroundColor: chartColors.value[0] || '#3b82f6',
      borderRadius: 6,
      barPercentage: 0.7,
    }],
  }
})

const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {display: false},
    tooltip: {
      backgroundColor: tooltipBg.value,
      titleColor: textColor.value,
      bodyColor: textColor.value,
      padding: 10,
      cornerRadius: 8,
    },
  },
  scales: {
    x: {
      grid: {display: false},
      ticks: {color: textColor.value, font: {size: 11}},
    },
    y: {
      beginAtZero: true,
      grid: {color: gridColor.value},
      ticks: {
        color: textColor.value,
        font: {size: 11},
        precision: 0,
      },
    },
  },
}))

const hasItems = computed(() => props.items.length > 0)
const hasBreakdown = computed(() => metrics.value.moduleBreakdown.length > 0)
const hasTimeline = computed(() => metrics.value.acquisitionTimeline.length > 0)
</script>

<template>
  <div class="dashboard">
    <div class="dashboard-cards">
      <DashboardMetricCard
        label="Total Collection Value"
        :value="formatCurrency(metrics.totalValue)"
        :subtitle="hasItems ? `across ${metrics.totalItems} items` : undefined"
      />
      <DashboardMetricCard
        label="Total Items"
        :value="metrics.totalItems.toLocaleString()"
      />
      <DashboardMetricCard
        v-if="metrics.mostValuableItem"
        label="Most Valuable Item"
        :value="formatCurrency(metrics.mostValuableItem.price)"
        :subtitle="metrics.mostValuableItem.title"
        :clickable="true"
        @click="emit('selectItem', metrics.mostValuableItem!.id)"
      />
      <DashboardMetricCard
        v-else
        label="Most Valuable Item"
        value="--"
        subtitle="No prices recorded"
      />
    </div>

    <div v-if="hasItems" class="dashboard-charts">
      <div class="chart-card">
        <h3 class="chart-title">{{ useCountFallback ? 'Items by Type' : 'Value by Type' }}</h3>
        <div v-if="hasBreakdown" class="chart-container doughnut-container">
          <Doughnut :data="doughnutData" :options="doughnutOptions" />
        </div>
        <div v-else class="chart-empty">
          Add items to see your collection breakdown
        </div>
      </div>

      <div class="chart-card">
        <h3 class="chart-title">Acquisitions Over Time</h3>
        <div v-if="hasTimeline" class="chart-container bar-container">
          <Bar :data="barData" :options="barOptions" />
        </div>
        <div v-else class="chart-empty">
          Add items to see your acquisition timeline
        </div>
      </div>
    </div>

    <div v-else class="dashboard-empty">
      <p class="dashboard-empty-text">Add items to see your collection insights</p>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 960px;
}

.dashboard-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.dashboard-charts {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
}

.chart-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
}

.chart-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
  margin: 0 0 var(--space-md) 0;
}

.chart-container {
  position: relative;
}

.doughnut-container {
  height: 280px;
}

.bar-container {
  height: 280px;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-muted);
  font-size: 13px;
}

.dashboard-empty {
  text-align: center;
  padding: var(--space-xl) var(--space-md);
}

.dashboard-empty-text {
  color: var(--text-muted);
  font-size: 14px;
}
</style>
