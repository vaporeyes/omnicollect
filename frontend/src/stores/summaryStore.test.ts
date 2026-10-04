// ABOUTME: Independent summary scope, latest-result ownership and disposal contracts.
// ABOUTME: Covers missing, zero, invalid, overflow and truncated aggregate shapes.
import {beforeEach,afterEach,it,expect,vi} from 'vitest'
import {createPinia,setActivePinia,disposePinia,type Pinia} from 'pinia'
import {useSummaryStore} from './summaryStore'
import {useCollectionStore} from './collectionStore'
import * as api from '../api/client'
vi.mock('../api/client',()=>({get:vi.fn()}))
let pinia:Pinia
const summary=(items=200)=>({items,pricedItems:0,invalidPrices:0,purchaseTotal:null,valueAvailable:true,modules:[{moduleId:'m',items}],modulesTruncated:false})
beforeEach(()=>{vi.resetAllMocks();pinia=createPinia();setActivePinia(pinia)})
afterEach(()=>disposePinia(pinia))
it('uses a filter-independent endpoint rather than the visible page',async()=>{
 const items=useCollectionStore();items.items=[{id:'one'} as any];items.searchQuery='filter';items.activeModuleId='other'
 vi.mocked(api.get).mockResolvedValue(summary())
 const store=useSummaryStore();expect(await store.refresh()).toBe(true)
 expect(api.get).toHaveBeenCalledWith('/api/v1/items/summary',expect.any(AbortSignal))
 expect(store.summary?.items).toBe(200)
})
it('hides old values while loading and ignores superseded or disposed completions',async()=>{
 const store=useSummaryStore();vi.mocked(api.get).mockResolvedValue(summary());await store.refresh()
 const pending:Array<(value:any)=>void>=[];vi.mocked(api.get).mockImplementation(()=>new Promise(resolve=>pending.push(resolve)))
 const first=store.refresh();const signal=vi.mocked(api.get).mock.calls.at(-1)![1]!
 expect(store.summary).toBeNull()
 const second=store.refresh();expect(signal.aborted).toBe(true)
 pending[1](summary(300));expect(await second).toBe(true)
 pending[0](summary(100));expect(await first).toBe(false);expect(store.summary?.items).toBe(300)
 const third=store.refresh();const lastSignal=vi.mocked(api.get).mock.calls.at(-1)![1]!
 store.$dispose();expect(lastSignal.aborted).toBe(true);pending[2](summary(500));expect(await third).toBe(false)
 expect(store.summary).toBeNull();expect(store.updatedAt).toBe('')
})
it('accepts explicit zero/missing/overflow coverage and rejects inconsistent payloads',async()=>{
 const store=useSummaryStore()
 for(const value of [summary(),{...summary(),pricedItems:1,purchaseTotal:0},{...summary(),pricedItems:200,valueAvailable:false},{...summary(),modules:[],modulesTruncated:true}]){
  vi.mocked(api.get).mockResolvedValue(value);expect(await store.refresh()).toBe(true)
 }
 for(const value of [{}, {...summary(),items:-1}, {...summary(),purchaseTotal:Infinity}, {...summary(),pricedItems:300}, {...summary(),modules:[]}, {...summary(),modules:[{moduleId:'m',items:100},{moduleId:'m',items:100}]}]){
  vi.mocked(api.get).mockResolvedValue(value);expect(await store.refresh()).toBe(false);expect(store.summary).toBeNull();expect(store.error).toContain('Invalid')
 }
})
