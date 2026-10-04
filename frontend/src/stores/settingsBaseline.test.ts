// ABOUTME: Unknown or malformed saved settings must never be overwritten as empty views.
// ABOUTME: Successful loading establishes a copied baseline before writes can begin.
import {afterEach,beforeEach,expect,it,vi} from 'vitest'
import {createPinia,setActivePinia,disposePinia,type Pinia} from 'pinia'
import {useSmartFolderStore} from './smartFolderStore'
import * as api from '../api/client'
vi.mock('../api/client',()=>({put:vi.fn().mockResolvedValue(undefined)}))
let pinia:Pinia
beforeEach(()=>{pinia=createPinia();setActivePinia(pinia);vi.clearAllMocks()})
afterEach(()=>disposePinia(pinia))
it('blocks create, rename, delete and retry-save without a loaded baseline',async()=>{
 const store=useSmartFolderStore()
 expect(store.create('new','','',{},[])).toBeNull()
 expect(store.rename('old','new')).toBe(false)
 await expect(store.remove('old')).rejects.toThrow()
 await store.saveToSettings();expect(api.put).not.toHaveBeenCalled()
 expect(store.loadFromSettings({smartFolders:[{id:'old',name:'Saved'}]})).toBe(true)
 expect(store.create('new','','',{},[])).not.toBeNull()
 await store.saveToSettings()
 expect(vi.mocked(api.put).mock.calls.at(-1)?.[1]).toMatchObject({smartFolders:[{id:'old',name:'Saved'},{name:'new'}]})
})
it('rejects malformed baselines and does not silently drop unknown views',async()=>{
 const store=useSmartFolderStore()
 for(const settings of [null,[],{smartFolders:null},{smartFolders:'bad'},{smartFolders:[null]},{smartFolders:[{id:'a',name:'A',filters:{field:null}}]},{smartFolders:[{id:'a',name:'A'},{id:'a',name:'B'}]}]){
  expect(store.loadFromSettings(settings)).toBe(false)
  expect(store.baselineReady).toBe(false)
  await store.saveToSettings()
 }
 expect(api.put).not.toHaveBeenCalled()
})
it('copies loaded views and invalidates them before refresh',async()=>{
 const store=useSmartFolderStore(),settings={smartFolders:[{id:'a',name:'A'}]}
 store.loadFromSettings(settings);settings.smartFolders[0].name='changed outside'
 expect(store.folders[0].name).toBe('A')
 store.invalidateBaseline();expect(store.create('new','','',{},[])).toBeNull()
 expect(store.loadFromSettings({})).toBe(true);expect(store.folders).toEqual([])
})
