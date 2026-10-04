// ABOUTME: Requests must never migrate from an old account into the next account.
// ABOUTME: Covers token waits, non-cooperative fetch/body promises and caller cancellation.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {get, post, getMedia, downloadFile, setTokenGetter} from '../api/client'
import {setSessionIdentity} from './session'
import {flushPromises} from '@vue/test-utils'
beforeEach(() => {setSessionIdentity('A'); setTokenGetter(null)})
afterEach(() => {setSessionIdentity('local'); setTokenGetter(null); vi.restoreAllMocks(); vi.unstubAllGlobals()})
it('rejects a token wait immediately on account change and never dispatches it', async () => {
 let token!: (value: string) => void
 setTokenGetter(() => new Promise(resolve => {token=resolve}))
 vi.stubGlobal('fetch', vi.fn())
 const request=post('/api/v1/settings',{private:'A'})
 const rejected=expect(request).rejects.toMatchObject({name:'AbortError'})
 setSessionIdentity('B');await rejected
 token('B-token');await flushPromises()
 expect(fetch).not.toHaveBeenCalled()
})
it('aborts active fetches and ignores a late response even if fetch ignores cancellation',async()=>{
 let finish!:(value:Response)=>void
 const fetcher=vi.fn(()=>new Promise<Response>(resolve=>{finish=resolve}))
 vi.stubGlobal('fetch',fetcher)
 const request=get('/api/v1/items');const rejected=expect(request).rejects.toMatchObject({name:'AbortError'})
 await flushPromises()
 const signal=fetcher.mock.calls[0][1].signal as AbortSignal
 setSessionIdentity('B');await rejected;expect(signal.aborted).toBe(true)
 finish(new Response('[{"title":"private A"}]'));await flushPromises()
})
it('does not create a download URL for an old account body',async()=>{
 let body!:(value:Blob)=>void
 vi.stubGlobal('fetch',vi.fn().mockResolvedValue({ok:true,headers:new Headers(),blob:()=>new Promise<Blob>(resolve=>{body=resolve})}))
 const create=vi.spyOn(URL,'createObjectURL')
 const request=downloadFile('/api/v1/export/backup');const rejected=expect(request).rejects.toMatchObject({name:'AbortError'})
 await flushPromises();setSessionIdentity('B');await rejected
 body(new Blob(['private']));await flushPromises();expect(create).not.toHaveBeenCalled()
})
it('combines caller cancellation with the session signal',async()=>{
 vi.stubGlobal('fetch',vi.fn(()=>new Promise(()=>{})))
 const controller=new AbortController()
 const request=getMedia('/originals/test.png',controller.signal)
 const rejected=expect(request).rejects.toMatchObject({name:'AbortError'})
 await flushPromises();controller.abort();await rejected
 expect(vi.mocked(fetch).mock.calls[0][1]?.signal?.aborted).toBe(true)
})
it('blocks all network work while the session is suspended',async()=>{
 setSessionIdentity(null);vi.stubGlobal('fetch',vi.fn())
 await expect(post('/api/v1/items',{})).rejects.toMatchObject({name:'AbortError'})
 expect(fetch).not.toHaveBeenCalled()
})
