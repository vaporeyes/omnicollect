import {defineStore} from 'pinia'
import {ref, onScopeDispose} from 'vue'
import * as api from '../api/client'
import type {Item, ItemPage} from '../api/types'

export const COLLECTION_PAGE_SIZE = 100
export const PALETTE_RESULT_LIMIT = 20
const MAX_OFFSET = 1000000
function pageItems(page: ItemPage, limit: number, offset: number): Item[] {
  if (!page || !Array.isArray(page.items) || page.items.length > limit || page.limit !== limit || page.offset !== offset || typeof page.hasMore !== 'boolean') throw new Error('Invalid item page response')
  return page.items
}
import {useSelectionStore} from './selectionStore'

export interface AttributeFilter {
  field: string
  op: 'in' | 'eq' | 'gte' | 'lte'
  value?: any
  values?: string[]
}

export const useCollectionStore = defineStore('collection', () => {
  const selection = useSelectionStore()
  const items = ref<Item[]>([])
  const loading = ref(false)
  const loaded = ref(false)
  const offset = ref(0)
  const hasMore = ref(false)
  const error = ref<string | null>(null)
  const activeModuleId = ref('')
  const searchQuery = ref('')
  const activeFilters = ref<Record<string, AttributeFilter[]>>({})
  const activeTags = ref<string[]>([])

  function serializeFilters(): string {
    const all: AttributeFilter[] = []
    for (const filters of Object.values(activeFilters.value)) {
      all.push(...filters)
    }
    return all.length > 0 ? JSON.stringify(all) : ''
  }

  function buildQueryString(pageOffset = 0): string {
    const params = new URLSearchParams()
    if (searchQuery.value) params.set('query', searchQuery.value)
    if (activeModuleId.value) params.set('moduleId', activeModuleId.value)
    const filters = serializeFilters()
    if (filters) params.set('filters', filters)
    if (activeTags.value.length > 0) params.set('tags', JSON.stringify(activeTags.value))
    params.set('limit', String(COLLECTION_PAGE_SIZE))
    params.set('offset', String(pageOffset))
    return '/api/v1/items/page?' + params.toString()
  }

  let selectionScope = buildQueryString()
  let scheduledSearch: {timer: ReturnType<typeof setTimeout>; resolve: (value: boolean) => void} | null = null
  function cancelScheduledSearch() {
    if (!scheduledSearch) return
    clearTimeout(scheduledSearch.timer)
    scheduledSearch.resolve(false)
    scheduledSearch = null
  }
  let requestVersion = 0
  let activeRequest: AbortController | null = null

  let disposed = false
  onScopeDispose(() => {
    disposed = true
    loaded.value = false
    offset.value = 0
    hasMore.value = false
    requestVersion++
    activeRequest?.abort()
    cancelScheduledSearch()
    items.value = []
    loading.value = false
    error.value = null
    searchQuery.value = ''
    activeModuleId.value = ''
    activeFilters.value = {}
    activeTags.value = []
    selection.clear()
  })
  // Refresh and all scope/mutation changes restart at page one; offsets are not snapshots.
  function fetchItems(): Promise<boolean> {return loadPage(0)}
  function goToPage(pageOffset: number): Promise<boolean> {
    if (loading.value || !Number.isInteger(pageOffset) || pageOffset < 0 || pageOffset > MAX_OFFSET || pageOffset % COLLECTION_PAGE_SIZE !== 0) return Promise.resolve(false)
    return loadPage(pageOffset)
  }
  async function loadPage(pageOffset: number): Promise<boolean> {
    if (disposed) return false
    cancelScheduledSearch()
    const path = buildQueryString(pageOffset)
    if (path !== selectionScope) selection.clear()
    selectionScope = path
    const version = ++requestVersion
    activeRequest?.abort()
    const controller = new AbortController()
    activeRequest = controller
    loading.value = true
    error.value = null
    try {
      const page = await api.get<ItemPage>(path, controller.signal)
      if (version !== requestVersion) return false
      const result = pageItems(page, COLLECTION_PAGE_SIZE, pageOffset)
      items.value = result
      offset.value = pageOffset
      hasMore.value = page.hasMore
      loaded.value = true
      selection.prune(result)
      return true
    } catch (e: any) {
      if (version === requestVersion && e?.name !== 'AbortError') error.value = e?.message ?? String(e)
      return false
    } finally {
      if (version === requestVersion) {
        loading.value = false
        activeRequest = null
      }
    }
  }

  async function saveItem(item: Item) {
    if (disposed) throw new DOMException('Store disposed', 'AbortError')
    error.value = null
    try {
      const saved = await api.post<Item>('/api/v1/items', item)
      if (disposed) throw new DOMException('Store disposed', 'AbortError')
      await fetchItems()
      return saved
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function deleteItem(id: string) {
    if (disposed) throw new DOMException('Store disposed', 'AbortError')
    error.value = null
    try {
      await api.del('/api/v1/items/' + id)
      await fetchItems()
    } catch (e: any) {
      error.value = e?.message ?? String(e)
      throw e
    }
  }

  async function searchAllItems(query: string, signal?: AbortSignal): Promise<Item[]> {
    if (!query) return []
    if (disposed) throw new DOMException('Store disposed', 'AbortError')
    const page = await api.get<ItemPage>('/api/v1/items/page?query=' + encodeURIComponent(query) + '&limit=' + PALETTE_RESULT_LIMIT + '&offset=0', signal)
    if (disposed) throw new DOMException('Store disposed', 'AbortError')
    return pageItems(page, PALETTE_RESULT_LIMIT, 0)
  }

  function setTags(tags: string[]) {
    activeTags.value = tags
    return fetchItems()
  }

  function setFilter(moduleId: string) {
    activeModuleId.value = moduleId
    activeFilters.value = {}
    activeTags.value = []
    return fetchItems()
  }

  function setSearch(query: string, delay = 0): Promise<boolean> {
    if (disposed) return Promise.resolve(false)
    searchQuery.value = query
    error.value = null
    selection.clear()
    cancelScheduledSearch()
    // Invalidate an old request immediately, not after the debounce expires.
    activeRequest?.abort()
    requestVersion++
    if (!delay) return fetchItems()
    loading.value = true
    return new Promise(resolve => {
      scheduledSearch = {resolve, timer: setTimeout(() => {
        scheduledSearch = null
        void fetchItems().then(resolve)
      }, delay)}
    })
  }

  function setActiveFilters(filters: Record<string, AttributeFilter[]>) {
    activeFilters.value = filters
    return fetchItems()
  }

  function clearFilters() {
    activeFilters.value = {}
    return fetchItems()
  }

  return {
    items, loading, loaded, error, offset, hasMore, activeModuleId, searchQuery, activeFilters, activeTags,
    fetchItems, goToPage, saveItem, deleteItem, searchAllItems,
    setFilter, setSearch, setActiveFilters, clearFilters, setTags,
  }
})
