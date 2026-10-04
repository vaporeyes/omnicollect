// ABOUTME: Lazy view boundaries expose loading and load-failure feedback.
// ABOUTME: Verifies both successful deferred resolution and rejected chunks.
import {afterEach, expect, it, vi} from 'vitest'
import {createApp, defineComponent, h, type App} from 'vue'
import {flushPromises} from '@vue/test-utils'
import {lazyComponent} from './lazyComponent'
let app: App | undefined, host: HTMLElement
function mount(view: any) {host=document.createElement('div'); document.body.append(host); app=createApp({render:()=>h(view)});app.mount(host)}
afterEach(()=>{app?.unmount();host?.remove();vi.useRealTimers();vi.restoreAllMocks()})
it('shows loading then resolves the requested view',async()=>{
 vi.useFakeTimers()
 const resolved=defineComponent({render:()=>h('p','Loaded editor')})
 let finish!:(value:{default:typeof resolved})=>void
 const view=lazyComponent(()=>new Promise<{default:typeof resolved}>(resolve=>{finish=resolve}))
 mount(view);await vi.advanceTimersByTimeAsync(160)
 expect(host.textContent).toContain('Loading view')
 finish({default:resolved});await vi.advanceTimersByTimeAsync(0)
 expect(host.textContent).toContain('Loaded editor')
})
it('shows a recoverable error instead of a blank view',async()=>{
 vi.spyOn(console,'warn').mockImplementation(()=>{})
 vi.spyOn(console,'error').mockImplementation(()=>{})
 mount(lazyComponent(()=>Promise.reject(new Error('chunk unavailable'))))
 await flushPromises()
 expect(host.querySelector('[role=alert]')?.textContent).toContain('could not load')
 expect(host.querySelector('button')?.textContent).toBe('Reload application')
})
