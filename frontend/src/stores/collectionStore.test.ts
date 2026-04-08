// ABOUTME: Unit tests for the collection Pinia store.
// ABOUTME: Verifies fetchItems, saveItem, deleteItem, and filter operations with mocked fetch.
import {describe, it, expect, vi, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useCollectionStore} from './collectionStore'

function mockFetchSuccess(data: any) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve(data),
  } as unknown as Response)
}

function mockFetchError(msg: string, status = 500) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: false,
    status,
    json: () => Promise.resolve({error: msg}),
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
        json: () => Promise.resolve(callCount === 1 ? savedItem : [savedItem]),
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
        json: () => Promise.resolve([]),
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

    store.setActiveFilters(filters)

    expect(store.activeFilters).toEqual(filters)
    expect(fetch).toHaveBeenCalled()
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
