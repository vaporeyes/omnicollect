import {defineStore} from 'pinia'
import {ref} from 'vue'
import * as api from '../api/client'
import type {Item} from '../api/types'

export interface AttributeFilter {
  field: string
  op: 'in' | 'eq' | 'gte' | 'lte'
  value?: any
  values?: string[]
}

export const useCollectionStore = defineStore('collection', () => {
  const items = ref<Item[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeModuleId = ref('')
  const searchQuery = ref('')
  const activeFilters = ref<Record<string, AttributeFilter[]>>({})

  function serializeFilters(): string {
    const all: AttributeFilter[] = []
    for (const filters of Object.values(activeFilters.value)) {
      all.push(...filters)
    }
    return all.length > 0 ? JSON.stringify(all) : ''
  }

  function buildQueryString(): string {
    const params = new URLSearchParams()
    if (searchQuery.value) params.set('query', searchQuery.value)
    if (activeModuleId.value) params.set('moduleId', activeModuleId.value)
    const filters = serializeFilters()
    if (filters) params.set('filters', filters)
    const qs = params.toString()
    return '/api/v1/items' + (qs ? '?' + qs : '')
  }

  async function fetchItems() {
    loading.value = true
    error.value = null
    try {
      items.value = await api.get<Item[]>(buildQueryString())
    } catch (e: any) {
      error.value = e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  async function saveItem(item: Item) {
    error.value = null
    try {
      const saved = await api.post<Item>('/api/v1/items', item)
      await fetchItems()
      return saved
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function deleteItem(id: string) {
    error.value = null
    try {
      await api.del('/api/v1/items/' + id)
      await fetchItems()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function searchAllItems(query: string): Promise<Item[]> {
    if (!query) return []
    return await api.get<Item[]>('/api/v1/items?query=' + encodeURIComponent(query))
  }

  function setFilter(moduleId: string) {
    activeModuleId.value = moduleId
    activeFilters.value = {}
    fetchItems()
  }

  function setSearch(query: string) {
    searchQuery.value = query
    fetchItems()
  }

  function setActiveFilters(filters: Record<string, AttributeFilter[]>) {
    activeFilters.value = filters
    fetchItems()
  }

  function clearFilters() {
    activeFilters.value = {}
    fetchItems()
  }

  return {
    items, loading, error, activeModuleId, searchQuery, activeFilters,
    fetchItems, saveItem, deleteItem, searchAllItems,
    setFilter, setSearch, setActiveFilters, clearFilters,
  }
})
