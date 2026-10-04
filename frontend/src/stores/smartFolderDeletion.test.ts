// ABOUTME: Saved-view deletion waits for persistence and preserves local data on failure.
// ABOUTME: Blocks racing settings edits and stale account completions.
import {afterEach,beforeEach,expect,it,vi} from 'vitest'
import {createPinia,setActivePinia,disposePinia,type Pinia} from 'pinia'
import {flushPromises} from '@vue/test-utils'
import {useSmartFolderStore} from './smartFolderStore'
import * as api from '../api/client'
vi.mock('../api/client',()=>({put:vi.fn()}))
let pinia:Pinia
beforeEach(()=>{pinia=createPinia();setActivePinia(pinia);vi.clearAllMocks()})
afterEach(()=>disposePinia(pinia))
function fixture(){const store=useSmartFolderStore();store.loadFromSettings({smartFolders:[{id:'a',name:'A'},{id:'b',name:'B'}]});store.setActive('a');return store}
it('keeps the target until acknowledgement and blocks concurrent edits',async()=>{
 let finish!:()=>void
 vi.mocked(api.put).mockReturnValue(new Promise<void>(resolve=>{finish=resolve}))
 const store=fixture();const request=store.removePersisted('a');await flushPromises()
 expect(store.folders.map(f=>f.id)).toEqual(['a','b']);expect(store.activeSmartFolderId).toBe('a')
 expect(store.create('racing','','',{},[])).toBeNull();expect(store.rename('b','racing')).toBe(false)
 await expect(store.removePersisted('b')).rejects.toThrow('Wait')
 finish();await request
 expect(api.put).toHaveBeenCalledWith('/api/v1/settings',{smartFolders:[{id:'b',name:'B'}]})
 expect(store.folders.map(f=>f.id)).toEqual(['b']);expect(store.activeSmartFolderId).toBeNull()
})
it('preserves the original and locks further writes after an uncertain deletion',async()=>{
 vi.mocked(api.put).mockRejectedValue(new Error('lost acknowledgement'))
 const store=fixture();await expect(store.removePersisted('a')).rejects.toThrow('lost acknowledgement')
 expect(store.folders).toHaveLength(2);expect(store.deletionUncertain).toBe(true)
 await store.saveToSettings();await expect(store.remove('a')).rejects.toThrow('reload');expect(store.rename('b','changed')).toBe(false)
 expect(api.put).toHaveBeenCalledTimes(1);expect(store.folders).toHaveLength(2)
})
it('cannot repopulate a disposed account on a late acknowledgement',async()=>{
 let finish!:()=>void
 vi.mocked(api.put).mockReturnValue(new Promise<void>(resolve=>{finish=resolve}))
 const store=fixture();const request=store.removePersisted('a');const rejected=expect(request).rejects.toMatchObject({name:'AbortError'})
 await flushPromises();store.$dispose();finish();await rejected
 expect(store.folders).toEqual([])
})
