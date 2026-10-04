// ABOUTME: Bounded browsing, page/query ownership and palette limits.
// ABOUTME: Deferred API mocks deliberately ignore cancellation to exercise generation guards.
import {beforeEach, afterEach, expect, it, vi} from 'vitest'
import {createPinia, setActivePinia, disposePinia, type Pinia} from 'pinia'
import {useCollectionStore} from './collectionStore'
import {useSelectionStore} from './selectionStore'
import * as api from '../api/client'
vi.mock('../api/client', () => ({get: vi.fn(), post: vi.fn(), del: vi.fn()}))
let pinia: Pinia
const page = (offset = 0, count = 100, hasMore = true, limit = 100) => ({items: Array.from({length:count}, (_, i) => ({id: String(offset+i), title: 'Item '+(offset+i)})), offset, limit, hasMore})
beforeEach(() => {vi.resetAllMocks(); pinia = createPinia(); setActivePinia(pinia)})
afterEach(() => disposePinia(pinia))
it('replaces rather than appends pages, clears page selections, and preserves query constraints', async () => {
  vi.mocked(api.get).mockImplementation(async path => page(Number(new URL(path, 'http://test').searchParams.get('offset'))))
  const store = useCollectionStore(), selection = useSelectionStore()
  store.activeModuleId = 'books'; store.searchQuery = 'needle'; store.activeTags = ['tag']
  store.activeFilters = {price:[{field:'purchasePrice',op:'gte',value:10}]}
  await store.fetchItems(); selection.toggle('0', 0)
  expect(await store.goToPage(100)).toBe(true)
  expect(store.items).toHaveLength(100); expect(store.items[0].id).toBe('100')
  expect(store.offset).toBe(100); expect(selection.count).toBe(0)
  const params = new URL(vi.mocked(api.get).mock.calls.at(-1)![0], 'http://test').searchParams
  expect(params.get('moduleId')).toBe('books'); expect(params.get('query')).toBe('needle')
  expect(params.get('tags')).toBe('["tag"]'); expect(params.get('filters')).toContain('purchasePrice')
  await store.setTags(['changed']); expect(store.offset).toBe(0)
  await store.goToPage(100); await store.fetchItems(); expect(store.offset).toBe(0)
  const calls = vi.mocked(api.get).mock.calls.length
  for (const offset of [-100, 1, 1000001, Infinity]) expect(await store.goToPage(offset)).toBe(false)
  expect(api.get).toHaveBeenCalledTimes(calls)
})
it('rejects a stale page after the query changes even when abort is ignored', async () => {
  vi.mocked(api.get).mockResolvedValueOnce(page())
  const store = useCollectionStore(); await store.fetchItems()
  const pending: Array<(value:any)=>void> = []
  vi.mocked(api.get).mockImplementation(() => new Promise(resolve => pending.push(resolve)))
  const old = store.goToPage(100)
  const signal = vi.mocked(api.get).mock.calls.at(-1)![1]!
  const current = store.setSearch('new')
  expect(signal.aborted).toBe(true)
  pending[0](page(100)); expect(await old).toBe(false)
  expect(store.loading).toBe(true); expect(store.offset).toBe(0)
  pending[1](page(0, 2, false)); expect(await current).toBe(true)
  expect(store.offset).toBe(0); expect(store.items).toHaveLength(2); expect(store.hasMore).toBe(false)
})
it('rejects malformed/unbounded pages and resets pagination on disposal', async () => {
  const store = useCollectionStore()
  for (const invalid of [page(0,101), page(100), page(0,1,false,20), {items:[]}]) {
    vi.mocked(api.get).mockResolvedValue(invalid)
    expect(await store.fetchItems()).toBe(false); expect(store.error).toBe('Invalid item page response')
  }
  vi.mocked(api.get).mockResolvedValue(page(100))
  await store.goToPage(100); expect(store.offset).toBe(100)
  store.$dispose(); expect(store.offset).toBe(0); expect(store.hasMore).toBe(false); expect(store.items).toEqual([])
})
it('limits palette search independently of collection scope and refuses oversized results', async () => {
  const store = useCollectionStore(); store.activeModuleId = 'private-filter'
  vi.mocked(api.get).mockResolvedValue(page(0,20,true,20))
  expect(await store.searchAllItems('needle')).toHaveLength(20)
  expect(api.get).toHaveBeenCalledWith('/api/v1/items/page?query=needle&limit=20&offset=0', undefined)
  vi.mocked(api.get).mockResolvedValue(page(0,21,true,20))
  await expect(store.searchAllItems('needle')).rejects.toThrow('Invalid item page')
})
