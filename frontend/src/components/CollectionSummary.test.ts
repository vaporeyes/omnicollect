// ABOUTME: Aggregate rendering distinguishes global scope, missing prices, zero and failures.
// ABOUTME: Exercises rendered DOM only, not browser visuals or assistive technology.
import {afterEach,beforeEach,it,expect,vi} from 'vitest'
import {mount,flushPromises,type VueWrapper} from '@vue/test-utils'
import {createPinia,disposePinia,type Pinia} from 'pinia'
import CollectionSummary from './CollectionSummary.vue'
import {useSummaryStore} from '../stores/summaryStore'
import * as api from '../api/client'
vi.mock('../api/client',()=>({get:vi.fn()}))
let pinia:Pinia,wrapper:VueWrapper
beforeEach(()=>{vi.resetAllMocks();pinia=createPinia();wrapper=mount(CollectionSummary,{global:{plugins:[pinia]}})})
afterEach(()=>{wrapper.unmount();disposePinia(pinia)})
it('shows missing and zero prices distinctly and discloses truncated groups',async()=>{
 const store=useSummaryStore(pinia)
 store.summary={items:200,pricedItems:0,invalidPrices:2,purchaseTotal:null,valueAvailable:true,modules:[],modulesTruncated:true}
 await flushPromises()
 expect(wrapper.text()).toContain('No recorded prices');expect(wrapper.text()).toContain('198 missing prices')
 expect(wrapper.text()).toContain('Overall totals remain complete')
 expect(wrapper.text()).toContain('independent of search, filters and page')
 store.summary={...store.summary,pricedItems:1,purchaseTotal:0};await flushPromises()
 expect(wrapper.text()).not.toContain('No recorded prices');expect(wrapper.findAll('dd')[1].text()).toMatch(/0[.,]00/)
 store.summary={...store.summary,valueAvailable:false,purchaseTotal:null};await flushPromises()
 expect(wrapper.text()).toContain('Unavailable (overflow)')
})
it('hides stale numbers after refresh failure and permits independent retry',async()=>{
 const store=useSummaryStore(pinia)
 store.summary={items:12345,pricedItems:0,invalidPrices:0,purchaseTotal:null,valueAvailable:true,modules:[{moduleId:'m',items:12345}],modulesTruncated:false}
 vi.mocked(api.get).mockRejectedValue(new Error('offline'))
 await wrapper.get('button').trigger('click');await flushPromises()
 expect(wrapper.findAll('dd')).toHaveLength(0);expect(wrapper.get('[role="alert"]').text()).toContain('offline')
 vi.mocked(api.get).mockResolvedValue({items:0,pricedItems:0,invalidPrices:0,purchaseTotal:null,valueAvailable:true,modules:[],modulesTruncated:false})
 await wrapper.get('button').trigger('click');await flushPromises()
 expect(wrapper.findAll('dd')[0].text()).toBe('0')
})
