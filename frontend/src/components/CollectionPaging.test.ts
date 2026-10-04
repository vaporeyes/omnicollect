// ABOUTME: Native paging controls expose bounded, page-local scope and honest empty/error states.
// ABOUTME: Uses the real Pinia store with deferred reads; no browser verification claimed.
import {beforeEach, afterEach, expect, it, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {createPinia, disposePinia, type Pinia} from 'pinia'
import CollectionPaging from './CollectionPaging.vue'
import {useCollectionStore} from '../stores/collectionStore'
import * as api from '../api/client'
vi.mock('../api/client', () => ({get: vi.fn()}))
let pinia: Pinia, wrapper: VueWrapper
beforeEach(() => {vi.resetAllMocks(); pinia=createPinia(); wrapper=mount(CollectionPaging,{global:{plugins:[pinia]}})})
afterEach(() => {wrapper.unmount(); disposePinia(pinia)})
function button(text:string) {return wrapper.findAll('button').find(b => b.text()===text)!}
it('blocks duplicate page clicks and reports empty shifted pages without claiming an empty collection', async () => {
  const store=useCollectionStore(pinia); store.loaded=true; store.hasMore=true
  await flushPromises()
  expect(wrapper.text()).toContain('current page only')
  expect(button('Previous page').attributes('disabled')).toBeDefined()
  let resolve!:(value:any)=>void
  vi.mocked(api.get).mockImplementation(()=>new Promise(done=>{resolve=done}))
  await button('Next page').trigger('click'); await button('Next page').trigger('click')
  expect(api.get).toHaveBeenCalledTimes(1)
  resolve({items:[],limit:100,offset:100,hasMore:false}); await flushPromises()
  expect(wrapper.text()).toContain('Page 2')
  expect(wrapper.text()).toContain('collection may still contain items')
  expect(button('Next page').attributes('disabled')).toBeDefined()
  expect(button('Previous page').attributes('disabled')).toBeUndefined()
})
it('does not report a failed fetch as successful results and permits restart', async () => {
  const store=useCollectionStore(pinia); store.loaded=true; store.hasMore=true; store.offset=100
  vi.mocked(api.get).mockRejectedValue(new Error('offline'))
  await button('Refresh from first page').trigger('click'); await flushPromises()
  expect(wrapper.find('[role="status"]').exists()).toBe(false)
  expect(button('Next page').attributes('disabled')).toBeDefined()
  vi.mocked(api.get).mockResolvedValue({items:[],limit:100,offset:0,hasMore:false})
  await button('Refresh from first page').trigger('click'); await flushPromises()
  expect(wrapper.text()).toContain('Page 1'); expect(store.error).toBeNull()
})
