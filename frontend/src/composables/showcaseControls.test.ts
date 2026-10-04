// ABOUTME: Sharing consent, duplicate-request and clipboard failure regressions.
// ABOUTME: Tests do not publish real collections or access a real clipboard.
import {afterEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, type App} from 'vue'
import {useShowcaseControls} from './showcaseControls'
import * as api from '../api/client'
vi.mock('../api/client', () => ({getAIStatus: vi.fn(), listShowcases: vi.fn(), toggleShowcase: vi.fn()}))
let app: App | undefined, host: HTMLElement, state: ReturnType<typeof useShowcaseControls>
const mod = {id: 'books', displayName: 'Books', attributes: []}
const flush = async () => {await Promise.resolve(); await Promise.resolve(); await nextTick()}
afterEach(() => {app?.unmount(); host?.remove(); vi.restoreAllMocks(); vi.unstubAllGlobals(); vi.clearAllMocks()})
async function mount() {
 vi.mocked(api.getAIStatus).mockResolvedValue({cloudMode: true} as any)
 vi.mocked(api.listShowcases).mockResolvedValue([])
 host = document.createElement('div');document.body.append(host)
 app = createApp({setup(){state = useShowcaseControls();return () => h('div')}});app.mount(host);await flush()
}
it('requires informed consent, prevents duplicate toggles, and fails closed after ambiguous errors', async () => {
 await mount()
 const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
 await state.onToggleShowcase(mod)
 expect(api.toggleShowcase).not.toHaveBeenCalled()
 expect(confirm.mock.calls[0][0]).toContain('GPS')
 expect(confirm.mock.calls[0][0]).toContain('Future items')
 confirm.mockReturnValue(true)
 let reject!: (error: Error) => void
 vi.mocked(api.toggleShowcase).mockReturnValue(new Promise((_resolve, fail) => {reject = fail}))
 const pending = state.onToggleShowcase(mod)
 await state.onToggleShowcase(mod)
 expect(api.toggleShowcase).toHaveBeenCalledTimes(1)
 reject(new Error('offline'));await pending
 expect(state.sharingReady.value).toBe(false)
 expect(state.sharingError.value).toContain('not confirmed')
})
it('never shows copied success when clipboard rejects', async () => {
 await mount()
 state.showcaseMap.value.books = {slug: 'books', url: '/showcase/books'} as any
 const writeText = vi.fn().mockRejectedValue(new Error('denied'))
 vi.stubGlobal('navigator', {clipboard: {writeText}})
 await state.copyShowcaseUrl(mod)
 expect(state.copiedSlug.value).toBe(null)
 expect(state.sharingError.value).toContain('Clipboard')
})
