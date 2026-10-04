// ABOUTME: Independent tenant-wide aggregate snapshots, not sums of visible pages.
// ABOUTME: Cancels/discards superseded and disposed reads; failures never expose stale totals.
import {defineStore} from 'pinia'
import {ref, onScopeDispose} from 'vue'
import * as api from '../api/client'
import type {CollectionSummary} from '../api/types'
function validSummary(value: CollectionSummary): boolean {
  const count = (n: number) => Number.isSafeInteger(n) && n >= 0
  if (!value || !count(value.items) || !count(value.pricedItems) || !count(value.invalidPrices) ||
      value.pricedItems + value.invalidPrices > value.items || typeof value.valueAvailable !== 'boolean' ||
      typeof value.modulesTruncated !== 'boolean' || !Array.isArray(value.modules) || value.modules.length > 1000) return false
  if (value.purchaseTotal !== null && (!Number.isFinite(value.purchaseTotal) || value.purchaseTotal < 0)) return false
  if ((!value.valueAvailable || value.pricedItems === 0) && value.purchaseTotal !== null) return false
  if (value.valueAvailable && value.pricedItems > 0 && value.purchaseTotal === null) return false
  const ids = new Set<string>()
  let total = 0
  for (const group of value.modules) {
    if (!group || typeof group.moduleId !== 'string' || !group.moduleId || ids.has(group.moduleId) || !count(group.items)) return false
    ids.add(group.moduleId); total += group.items
  }
  return total <= value.items && (value.modulesTruncated || total === value.items)
}
export const useSummaryStore = defineStore('summary', () => {
  const summary = ref<CollectionSummary | null>(null)
  const loading = ref(false), error = ref(''), updatedAt = ref('')
  let generation = 0, disposed = false, request: AbortController | null = null
  onScopeDispose(() => {disposed = true; generation++; request?.abort(); summary.value = null; loading.value = false; error.value = ''; updatedAt.value = ''})
  async function refresh(): Promise<boolean> {
    if (disposed) return false
    const current = ++generation
    request?.abort(); request = new AbortController()
    loading.value = true; error.value = ''; summary.value = null; updatedAt.value = ''
    try {
      const result = await api.get<CollectionSummary>('/api/v1/items/summary', request.signal)
      if (disposed || current !== generation) return false
      if (!validSummary(result)) throw new Error('Invalid collection summary')
      summary.value = result; updatedAt.value = new Date().toLocaleTimeString()
      return true
    } catch (e: any) {
      if (!disposed && current === generation) error.value = e?.message || 'Summary unavailable'
      return false
    } finally {if (current === generation) {loading.value = false; request = null}}
  }
  return {summary, loading, error, updatedAt, refresh}
})
