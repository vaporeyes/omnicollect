// ABOUTME: Unit tests for the collection Pinia store.
// ABOUTME: Verifies fetchItems, saveItem, deleteItem, and filter operations with mocked fetch.
import {flushPromises} from '@vue/test-utils'
import {describe, it, expect, vi, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useCollectionStore} from './collectionStore'
import {useSelectionStore} from './selectionStore'

const page = (items: any[], limit = 100, offset = 0, hasMore = false) => ({items, limit, offset, hasMore})
function mockFetchSuccess(data: any) {
  global.fetch = vi.fn().mockImplementation((path: string) => {
    const params = new URL(path, 'http://localhost').searchParams
    const result = Array.isArray(data) ? page(data, Number(params.get('limit') || 100), Number(params.get('offset') || 0)) : data
    return Promise.resolve(new Response(JSON.stringify(result)))
  })
}

function mockFetchError(msg: string, status = 500) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: false,
    status,
    json: () => Promise.resolve({error: msg}),
      text: () => Promise.resolve(JSON.stringify({error: msg})),
  } as unknown as Response)
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.restoreAllMocks()
})

describe('fetchItems', () => {
  it('sets items on success', async () => {
    const items = [{id: '1', moduleId: 'm1', title: 'Item 1', purchasePrice: null, images: [], attributes: {}, createdAt: '', updatedAt: ''}]
    mockFetchSuccess(items)

    const store = useCollectionStore()
    await store.fetchItems()

    expect(store.items).toEqual(items)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('sets error on failure', async () => {
    mockFetchError('database error')

    const store = useCollectionStore()
    await store.fetchItems()

    expect(store.items).toEqual([])
    expect(store.error).toBe('database error')
  })
})

describe('search and selection coherence', () => {
  it('prunes removed IDs on refresh and clears selection when filters change', async () => {
    const store = useCollectionStore(), selection = useSelectionStore()
    selection.toggle('keep', 0); selection.toggle('removed', 1)
    mockFetchSuccess([{id: 'keep'}])
    await store.fetchItems()
    expect(selection.selectedIdArray()).toEqual(['keep'])
    await store.setFilter('other')
    expect(selection.count).toBe(0)
  })

  it('updates canonical search immediately and resolves superseded debounce work', async () => {
    vi.useFakeTimers()
    try {
      mockFetchSuccess([])
      const store = useCollectionStore()
      const older = store.setSearch('older', 300)
      const newer = store.setSearch('latest', 300)
      expect(store.searchQuery).toBe('latest')
      expect(await older).toBe(false)
      expect(fetch).not.toHaveBeenCalled()
      await vi.advanceTimersByTimeAsync(300)
      expect(await newer).toBe(true)
      expect(fetch).toHaveBeenCalledTimes(1)
      expect((fetch as any).mock.calls[0][0]).toBe('/api/v1/items/page?query=latest&limit=100&offset=0')
    } finally { vi.useRealTimers() }
  })

  it('cancels a scheduled search when a new view loads directly', async () => {
    vi.useFakeTimers()
    try {
      mockFetchSuccess([])
      const store = useCollectionStore()
      const scheduled = store.setSearch('shared', 300)
      await store.setFilter('books')
      expect(await scheduled).toBe(false)
      await vi.advanceTimersByTimeAsync(300)
      expect(fetch).toHaveBeenCalledTimes(1)
      expect(store.searchQuery).toBe('shared')
    } finally { vi.useRealTimers() }
  })
})

describe('saveItem', () => {
  it('calls POST and re-fetches items', async () => {
    const savedItem = {id: '1', moduleId: 'm1', title: 'Saved', purchasePrice: null, images: [], attributes: {}, createdAt: '', updatedAt: ''}
    // First call: POST save, second call: GET re-fetch
    let callCount = 0
    global.fetch = vi.fn().mockImplementation(() => {
      callCount++
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(callCount === 1 ? savedItem : page([savedItem])),
      text: () => Promise.resolve(JSON.stringify(callCount === 1 ? savedItem : page([savedItem]))),
      } as unknown as Response)
    })

    const store = useCollectionStore()
    const result = await store.saveItem(savedItem as any)

    expect(result.id).toBe('1')
    expect(fetch).toHaveBeenCalledTimes(2) // save + re-fetch
  })
})

describe('deleteItem', () => {
  it('calls DELETE and re-fetches items', async () => {
    let callCount = 0
    global.fetch = vi.fn().mockImplementation(() => {
      callCount++
      return Promise.resolve({
        ok: true,
        status: callCount === 1 ? 204 : 200,
        json: () => Promise.resolve(page([])),
      text: () => Promise.resolve(JSON.stringify(page([]))),
      } as unknown as Response)
    })

    const store = useCollectionStore()
    await store.deleteItem('item-1')

    expect(fetch).toHaveBeenCalledTimes(2) // delete + re-fetch
  })
})

describe('searchAllItems', () => {
  it('returns items matching the query', async () => {
    const results = [{id: '2', title: 'Found'}]
    mockFetchSuccess(results)

    const store = useCollectionStore()
    const items = await store.searchAllItems('Found')

    expect(items).toEqual(results)
  })

  it('returns empty array for empty query', async () => {
    const store = useCollectionStore()
    const items = await store.searchAllItems('')
    expect(items).toEqual([])
  })
})

describe('setFilter', () => {
  it('clears active filters and sets module', async () => {
    mockFetchSuccess([])
    const store = useCollectionStore()
    store.activeFilters = {field1: [{field: 'f', op: 'eq', value: 'x'}]}

    store.setFilter('new-module')

    expect(store.activeModuleId).toBe('new-module')
    expect(store.activeFilters).toEqual({})
  })
})

describe('setActiveFilters', () => {
  it('sets filters and triggers re-fetch', async () => {
    mockFetchSuccess([])
    const store = useCollectionStore()
    const filters = {condition: [{field: 'condition', op: 'in' as const, values: ['Mint']}]}

    await store.setActiveFilters(filters)

    expect(store.activeFilters).toEqual(filters)
    expect(fetch).toHaveBeenCalled()
  })
})

describe('request ordering', () => {
  it('ignores stale responses and keeps loading until the current request finishes', async () => {
    const pending: Array<(r: Response) => void> = []
    global.fetch = vi.fn().mockImplementation(() => new Promise<Response>(resolve => pending.push(resolve)))
    const store = useCollectionStore()
    const first = store.setSearch('old')
    await flushPromises()
    const second = store.setSearch('new')
    await flushPromises()
    pending[0](new Response(JSON.stringify(page([{id: 'old'}]))))
    await first
    expect(store.loading).toBe(true)
    expect(store.items).toEqual([])
    pending[1](new Response(JSON.stringify(page([{id: 'new'}]))))
    await second
    expect(store.loading).toBe(false)
    expect(store.items[0].id).toBe('new')
  })

  it('does not let an older late response overwrite the current result', async () => {
    const pending: Array<(r: Response) => void> = []
    global.fetch = vi.fn().mockImplementation(() => new Promise<Response>(resolve => pending.push(resolve)))
    const store = useCollectionStore()
    const first = store.setSearch('old')
    await flushPromises()
    const second = store.setSearch('new')
    await flushPromises()
    pending[1](new Response(JSON.stringify(page([{id: 'new'}]))))
    await second
    pending[0](new Response(JSON.stringify(page([{id: 'old'}]))))
    await first
    expect(store.items[0].id).toBe('new')
  })
})

describe('clearFilters', () => {
  it('empties active filters', async () => {
    mockFetchSuccess([])
    const store = useCollectionStore()
    store.activeFilters = {field1: [{field: 'f', op: 'eq', value: 'x'}]}

    store.clearFilters()

    expect(store.activeFilters).toEqual({})
  })
})
