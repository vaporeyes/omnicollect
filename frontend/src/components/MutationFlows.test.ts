// ABOUTME: Mounted application mutation flows preserve fixed targets and disclose tag merges.
// ABOUTME: Stubs unrelated heavy views and networking, not confirmation logic.
import {afterEach,beforeEach,expect,it,vi} from 'vitest'
import {mount,flushPromises,type VueWrapper} from '@vue/test-utils'
import {createPinia,disposePinia,type Pinia} from 'pinia'
import App from '../App.vue'
import AppSidebar from './AppSidebar.vue'
import BulkActionBar from './BulkActionBar.vue'
import {useCollectionStore} from '../stores/collectionStore'
import {useSelectionStore} from '../stores/selectionStore'
import {useSmartFolderStore} from '../stores/smartFolderStore'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
import * as api from '../api/client'
vi.mock('../api/client',()=>({get:vi.fn(),getAllTags:vi.fn(),post:vi.fn(),renameTag:vi.fn(),deleteTag:vi.fn()}))
vi.mock('../auth/plugin',()=>({isAuthConfigured:false}))
let wrapper:VueWrapper|undefined,pinia:Pinia,cleanup:()=>void
beforeEach(()=>{
 vi.clearAllMocks();cleanup=stubNativeDialogs();pinia=createPinia()
 vi.stubGlobal('matchMedia',vi.fn(()=>({matches:false,addEventListener:vi.fn(),removeEventListener:vi.fn()})))
 vi.mocked(api.get).mockImplementation(async(path)=>path==='/api/v1/modules' ? [{id:'a',displayName:'Original',attributes:[]},{id:'b',displayName:'Destination',attributes:[]}] : path.startsWith('/api/v1/items/page') ? {items:[],limit:100,offset:0,hasMore:false} : path === '/api/v1/items/summary' ? {items:2,pricedItems:0,invalidPrices:0,purchaseTotal:null,valueAvailable:true,modules:[{moduleId:'a',items:1},{moduleId:'b',items:1}],modulesTruncated:false} : {})
 vi.mocked(api.getAllTags).mockResolvedValue([{name:'old',count:2},{name:'existing',count:1}])
})
afterEach(()=>{wrapper?.unmount();disposePinia(pinia);cleanup();vi.unstubAllGlobals()})
async function start(){wrapper=mount(App,{global:{plugins:[pinia],stubs:{AppSidebar:true,DashboardView:true,ImageLightbox:true,CommandPalette:true,ImportDialog:true,ToastProvider:true}}});await flushPromises();return wrapper}
it('moves the original selection and destination despite subsequent selection changes',async()=>{
 const view=await start();const items=useCollectionStore(pinia),selection=useSelectionStore(pinia)
 items.items=[{id:'one',title:'First',moduleId:'a'} as any,{id:'two',title:'Second',moduleId:'a'} as any]
 selection.toggle('one',0);await flushPromises()
 view.getComponent(BulkActionBar).vm.$emit('editModule');await flushPromises()
 selection.clear();selection.toggle('two',1)
 await view.get('select[aria-label="Destination module"]').setValue('b')
 const review=view.findAll('button').find(b=>b.text()==='Review move')!;await review.trigger('click')
 expect(view.text()).toContain('First');expect(view.get('dialog').text()).toContain('Destination')
 vi.mocked(api.post).mockResolvedValue({updated:1})
 await view.get('.confirm-action').trigger('click');await flushPromises()
 expect(api.post).toHaveBeenCalledWith('/api/v1/items/batch-update-module',{ids:['one'],newModuleId:'b'})
 expect(vi.mocked(api.get).mock.calls.filter(([path])=>path==='/api/v1/items/summary')).toHaveLength(2)
})
it('requires merge consent and preserves the rename draft when cancelled',async()=>{
 const view=await start();view.getComponent(AppSidebar).vm.$emit('openTags');await flushPromises()
 const rename=view.findAll('.tag-row').find(row=>row.text().includes('old'))!
 await rename.get('.rename-btn').trigger('click');await rename.get('input').setValue('existing');await rename.get('.save-btn').trigger('click')
 expect(view.get('dialog').text()).toContain('both tags will be merged')
 expect(api.renameTag).not.toHaveBeenCalled()
 view.get('dialog').element.dispatchEvent(new Event('cancel',{cancelable:true}));await flushPromises()
 expect((view.get('.tag-edit-input').element as HTMLInputElement).value).toBe('existing')
})
it('keeps saved-view edits blocked after settings failure and retries the baseline',async()=>{
 vi.mocked(api.get).mockImplementation(async(path)=>{if(path==='/api/v1/settings')throw new Error('offline');return path.startsWith('/api/v1/items/page')?{items:[],limit:100,offset:0,hasMore:false}:[]})
 const view=await start(),store=useSmartFolderStore(pinia)
 expect(view.text()).toContain('Settings could not be loaded')
 expect(store.create('overwrite','','',{},[])).toBeNull()
 vi.mocked(api.get).mockImplementation(async(path)=>path==='/api/v1/settings'?{smartFolders:[{id:'preserved',name:'Server view'}]}:path.startsWith('/api/v1/items/page')?{items:[],limit:100,offset:0,hasMore:false}:[])
 await view.findAll('button').find(button=>button.text()==='Retry settings')!.trigger('click');await flushPromises()
 expect(store.baselineReady).toBe(true);expect(store.folders[0].id).toBe('preserved')
 expect(view.text()).not.toContain('Settings could not be loaded')
})
it('does not treat an empty later page as an empty collection',async()=>{
 const view=await start(),store=useCollectionStore(pinia)
 store.offset=100;store.items=[];store.loaded=true
 await flushPromises()
 expect(view.text()).toContain('collection may still contain items')
 expect(view.find('dashboard-view-stub').exists()).toBe(false)
 expect(view.text()).not.toContain('No items match the current search or filters.')
})
it('does not render dashboard totals after an initial collection failure',async()=>{
 vi.mocked(api.get).mockImplementation(async(path)=>{if(path.startsWith('/api/v1/items'))throw new Error('database unavailable');return path==='/api/v1/settings'?{}:[]})
 const view=await start()
 expect(view.text()).toContain('Collection results unavailable')
 expect(view.find('dashboard-view-stub').exists()).toBe(false)
 expect(view.text()).not.toContain('Loading...')
 vi.mocked(api.get).mockResolvedValue({items:[],limit:100,offset:0,hasMore:false})
 await view.findAll('button').find(button=>button.text()==='Retry items')!.trigger('click');await flushPromises()
 expect(view.find('dashboard-view-stub').exists()).toBe(true)
})
