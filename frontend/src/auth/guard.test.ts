// ABOUTME: Account boundaries remount all application state and dispose old private stores.
// ABOUTME: Simulates account changes and suspended auth without contacting Auth0.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, defineComponent, h, nextTick, ref, Teleport, type App} from 'vue'
import {createPinia, disposePinia, type Pinia} from 'pinia'
import {flushPromises} from '@vue/test-utils'
import {AuthGuard} from './guard'
import {beginSignOut, logoutRequested, sessionFailure, setSessionIdentity} from './session'
import {useCollectionStore} from '../stores/collectionStore'
import {useModuleStore} from '../stores/moduleStore'
import {useSmartFolderStore} from '../stores/smartFolderStore'
import {useSelectionStore} from '../stores/selectionStore'
import {useToastStore} from '../stores/toastStore'
import {useSummaryStore} from '../stores/summaryStore'
import * as api from '../api/client'
const holder=vi.hoisted(()=>({auth:null as any}))
vi.mock('@auth0/auth0-vue',()=>({useAuth0:()=>holder.auth}))
vi.mock('../api/client',()=>({get:vi.fn(),put:vi.fn()}))
let app:App|undefined, pinia:Pinia, host:HTMLElement
let instances:any[]
beforeEach(()=>{
 vi.clearAllMocks();setSessionIdentity('local');logoutRequested.value=false;sessionFailure.value=''
 holder.auth={isLoading:ref(false),isAuthenticated:ref(true),user:ref({sub:'A'}),loginWithRedirect:vi.fn().mockResolvedValue(undefined)}
 vi.mocked(api.get).mockResolvedValue([])
 instances=[];host=document.createElement('div');document.body.append(host);pinia=createPinia()
})
afterEach(()=>{app?.unmount();disposePinia(pinia);host.remove();vi.useRealTimers();setSessionIdentity('local');logoutRequested.value=false})
function mount(){
 const Child=defineComponent({setup(){
  const state={summary:useSummaryStore(),items:useCollectionStore(),modules:useModuleStore(),folders:useSmartFolderStore(),selection:useSelectionStore(),toast:useToastStore()}
  instances.push(state)
  return()=>h('div',[h('p',state.items.items.map(item=>item.title).join(',')),h(Teleport,{to:'body'},h('div',{'data-private-overlay':''},state.folders.folders.map(folder=>folder.name).join(',')))])
 }})
 app=createApp({render:()=>h(AuthGuard,null,{default:()=>h(Child)})});app.use(pinia);app.mount(host)
}
it('does not set up application stores before an account is ready',async()=>{
 holder.auth.isLoading.value=true;mount();expect(instances).toHaveLength(0)
 holder.auth.isLoading.value=false;await nextTick();expect(instances).toHaveLength(1)
})
it('clears private stores/overlays and mounts fresh state when subjects change',async()=>{
 mount();const old=instances[0]
 old.summary.summary={items:12345,pricedItems:0,invalidPrices:0,purchaseTotal:null,valueAvailable:true,modules:[],modulesTruncated:true}
 old.items.items=[{id:'private',title:'Alice item'}];old.items.searchQuery='Alice search'
 old.folders.folders=[{id:'folder',name:'Alice folder'}];old.selection.toggle('private',0);old.toast.show('Alice toast')
 await nextTick();expect(document.body.textContent).toContain('Alice folder')
 holder.auth.user.value={sub:'B'};await nextTick()
 expect(instances).toHaveLength(2)
 const current=instances[1]
 expect(current.items).not.toBe(old.items)
 expect(current.items.items).toEqual([]);expect(current.items.searchQuery).toBe('')
 expect(current.folders.folders).toEqual([]);expect(current.selection.count).toBe(0);expect(current.toast.toasts).toEqual([])
 expect(document.body.textContent).not.toContain('Alice')
 expect(old.items.items).toEqual([])
 expect(old.summary.summary).toBeNull();expect(current.summary.summary).toBeNull()
})
it('drops queued settings writes and late module results after disposal',async()=>{
 mount();const old=instances[0]
 let finishWrite!:()=>void,finishRead!:(value:any)=>void
 vi.mocked(api.put).mockReturnValue(new Promise<void>(resolve=>{finishWrite=resolve}))
 vi.mocked(api.get).mockReturnValue(new Promise(resolve=>{finishRead=resolve}))
 old.folders.loadFromSettings({smartFolders:[{id:'a',name:'private A'}]})
 const first=old.folders.saveToSettings();await flushPromises()
 const second=old.folders.saveToSettings()
 const read=old.modules.fetchModules()
 const search=old.items.setSearch('private query',10000)
 holder.auth.user.value={sub:'B'};await nextTick()
 expect(await search).toBe(false)
 finishWrite();finishRead([{id:'private-module'}]);await Promise.all([first,second,read])
 expect(api.put).toHaveBeenCalledTimes(1)
 expect(old.modules.modules).toEqual([]);expect(instances[1].modules.modules).toEqual([])
 await old.folders.saveToSettings();expect(api.put).toHaveBeenCalledTimes(1)
})
it('clears immediately on sign out without starting a competing login',async()=>{
 mount();beginSignOut();await nextTick();expect(host.textContent).toContain('Signed out')
 holder.auth.isAuthenticated.value=false;await nextTick()
 expect(holder.auth.loginWithRedirect).not.toHaveBeenCalled()
})
it('shows login rejection rather than leaving an unhandled promise',async()=>{
 holder.auth.isAuthenticated.value=false
 holder.auth.loginWithRedirect.mockRejectedValue(new Error('offline'))
 mount();await flushPromises();expect(host.textContent).toContain('Sign in could not start')
 expect(instances).toHaveLength(0)
})
