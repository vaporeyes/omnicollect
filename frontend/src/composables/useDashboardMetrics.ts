// ABOUTME: Computed dashboard metrics derived from collection store items.
// ABOUTME: Provides total value, item count, most valuable item, module breakdown, and acquisition timeline.

import {computed, type Ref} from 'vue'
import type {Item, ModuleSchema} from '../api/types'

export interface ModuleSegment {
  moduleId: string
  moduleName: string
  totalValue: number
  itemCount: number
  percentage: number
}

export interface TimeBucket {
  label: string
  key: string
  count: number
}

export interface DashboardMetrics {
  pricedItems: number
  invalidPrices: number
  valueAvailable: boolean
  totalValue: number
  totalItems: number
  mostValuableItem: {id: string; title: string; price: number} | null
  moduleBreakdown: ModuleSegment[]
  acquisitionTimeline: TimeBucket[]
}

const MAX_SEGMENTS = 6

export function computeDashboardMetrics(
  items: Item[],
  getModuleName: (id: string) => string
): DashboardMetrics {
  const totalItems = items.length
  let totalValue = 0
  let pricedItems = 0
  let invalidPrices = 0
  let best: {id: string; title: string; price: number} | null = null

  // Per-module accumulators
  const moduleMap = new Map<string, {value: number; count: number}>()
  // Per-month accumulators
  const monthMap = new Map<string, number>()

  for (const item of items) {
    const validPrice = typeof item.purchasePrice === 'number' && Number.isFinite(item.purchasePrice) && item.purchasePrice >= 0
    const price = validPrice ? item.purchasePrice! : 0
    if (item.purchasePrice != null && !validPrice) invalidPrices++
    if (validPrice) {
      pricedItems++
      totalValue += price
      if (!best || price > best.price) {
        best = {id: item.id, title: item.title, price}
      }
    }

    // Module grouping
    const entry = moduleMap.get(item.moduleId)
    if (entry) {
      entry.value += price
      entry.count++
    } else {
      moduleMap.set(item.moduleId, {value: price, count: 1})
    }

    // Month grouping from createdAt (ISO string)
    const monthKey = /^\d{4}-(0[1-9]|1[0-2])(?:-|$)/.test(item.createdAt || '') ? item.createdAt.substring(0, 7) : 'unknown'
    monthMap.set(monthKey, (monthMap.get(monthKey) ?? 0) + 1)
  }

  const valueAvailable = Number.isFinite(totalValue)
  // Build module breakdown by value, or count when no usable total exists.
  let segments: ModuleSegment[] = Array.from(moduleMap.entries())
    .map(([moduleId, data]) => ({
      moduleId,
      moduleName: getModuleName(moduleId),
      totalValue: data.value,
      itemCount: data.count,
      percentage: valueAvailable && totalValue > 0 ? (data.value / totalValue) * 100 : 0,
    }))
    .sort((a, b) => (valueAvailable && totalValue > 0 ? b.totalValue - a.totalValue : b.itemCount - a.itemCount) || a.moduleId.localeCompare(b.moduleId))

  // Group smallest into "Other" when more than MAX_SEGMENTS modules
  if (segments.length > MAX_SEGMENTS) {
    const top = segments.slice(0, MAX_SEGMENTS - 1)
    const rest = segments.slice(MAX_SEGMENTS - 1)
    const otherValue = rest.reduce((s, r) => s + r.totalValue, 0)
    const otherCount = rest.reduce((s, r) => s + r.itemCount, 0)
    top.push({
      moduleId: '',
      moduleName: 'Other',
      totalValue: otherValue,
      itemCount: otherCount,
      percentage: valueAvailable && totalValue > 0 ? (otherValue / totalValue) * 100 : 0,
    })
    segments = top
  }

  // Build acquisition timeline sorted chronologically
  const timeline: TimeBucket[] = Array.from(monthMap.entries())
    .filter(([key]) => key !== 'unknown')
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, count]) => ({
      key,
      label: formatMonthLabel(key),
      count,
    }))

  return {totalValue, pricedItems, invalidPrices, valueAvailable, totalItems, mostValuableItem: best, moduleBreakdown: segments, acquisitionTimeline: timeline}
}

function formatMonthLabel(key: string): string {
  // key is "YYYY-MM"
  const [year, month] = key.split('-')
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const idx = parseInt(month, 10) - 1
  return `${months[idx] ?? month} ${year}`
}

export function useDashboardMetrics(
  items: Ref<Item[]>,
  modules: Ref<ModuleSchema[]>
) {
  function getModuleName(id: string): string {
    return modules.value.find(m => m.id === id)?.displayName ?? id
  }

  const metrics = computed(() => computeDashboardMetrics(items.value, getModuleName))

  return {metrics}
}
