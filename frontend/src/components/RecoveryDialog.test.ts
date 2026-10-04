// ABOUTME: Mounted recovery/discard consent, duplicate guards, uncertain outcomes and refresh tests.
// ABOUTME: Native dialog methods are stubbed; these are not browser focus or inertness checks.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, type App} from 'vue'
import RecoveryDialog from './RecoveryDialog.vue'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
const api = vi.hoisted(() => ({get: vi.fn(), post: vi.fn(), del: vi.fn()}))
vi.mock('../api/client', () => api)
const batch = {recoveryId: 'one', deleted: 2, createdAt: '2026-01-01T00:00:00Z', titles: ['First', 'Second']}
let app: App | undefined, host: HTMLElement, cleanup: () => void
const flush = async () => {for (let i=0;i<8;i++) await Promise.resolve(); await nextTick()}
function button(text: string) {return Array.from(host.querySelectorAll('button')).find(b => b.textContent?.trim() === text)!}
async function mount(refresh = vi.fn().mockResolvedValue(undefined), close = vi.fn()) {
  app = createApp({render: () => h(RecoveryDialog, {refresh, onClose: close})}); app.mount(host); await flush()
}
async function consent() {const input = host.querySelector('input')!; input.checked = true; input.dispatchEvent(new Event('change', {bubbles: true})); await nextTick()}
beforeEach(() => {vi.resetAllMocks(); api.get.mockResolvedValue([batch]); cleanup = stubNativeDialogs(); host = document.createElement('div'); document.body.append(host)})
afterEach(() => {app?.unmount(); app = undefined; host.remove(); cleanup()})
it('requires disclosure consent, blocks duplicate recovery/cancel/unload, and keeps commit success distinct from refresh failure', async () => {
  let resolve!: () => void
  api.post.mockImplementation(() => new Promise<void>(done => {resolve = done}))
  const refresh = vi.fn().mockRejectedValueOnce(new Error('items unavailable')).mockResolvedValue(undefined), close = vi.fn()
  await mount(refresh, close)
  button('Recover').click(); await nextTick()
  expect(host.textContent).toContain('enabled public showcases')
  expect(button('Recover batch').disabled).toBe(true)
  await consent(); button('Recover batch').click(); button('Recover batch').click(); await nextTick()
  expect(api.post).toHaveBeenCalledTimes(1)
  expect(api.post).toHaveBeenCalledWith('/api/v1/recovery/one', {})
  const cancel = new Event('cancel', {cancelable: true}); host.querySelector('dialog')!.dispatchEvent(cancel)
  expect(close).not.toHaveBeenCalled()
  const unload = new Event('beforeunload', {cancelable: true}); window.dispatchEvent(unload); expect(unload.defaultPrevented).toBe(true)
  api.get.mockResolvedValue([]); resolve(); await flush()
  expect(host.textContent).toContain('Recovered 2 item(s)')
  expect(host.textContent).toContain('Refresh failed')
  button('Refresh collection and recovery list').click(); await flush()
  expect(api.post).toHaveBeenCalledTimes(1)
  expect(host.textContent).toContain('No deleted batches available')
})
it('requires separate permanent discard consent and freezes uncertain repeats until refresh', async () => {
  api.del.mockRejectedValue(new Error('lost acknowledgement'))
  await mount()
  button('Discard permanently…').click(); await nextTick()
  expect(host.textContent).toContain('does not erase media files')
  expect(button('Permanently discard batch').disabled).toBe(true)
  await consent(); button('Permanently discard batch').click(); await flush()
  expect(host.textContent).toContain('It may have completed')
  expect(button('Permanently discard batch').disabled).toBe(true)
  button('Permanently discard batch').click(); expect(api.del).toHaveBeenCalledTimes(1)
  api.get.mockResolvedValue([]); button('Refresh collection and recovery list').click(); await flush()
  expect(host.textContent).toContain('No deleted batches available')
  expect(api.post).not.toHaveBeenCalled()
})
it('does not present failed lists as empty and discards late work after unmount', async () => {
  api.get.mockRejectedValueOnce(new Error('offline'))
  const refresh = vi.fn()
  await mount(refresh)
  expect(host.textContent).toContain('Refresh failed')
  expect(host.textContent).not.toContain('No deleted batches')
  let resolve!: (value: typeof batch[]) => void
  api.get.mockImplementation(() => new Promise(done => {resolve = done}))
  button('Refresh collection and recovery list').click(); await nextTick()
  const signal = api.get.mock.calls.at(-1)![1] as AbortSignal
  app!.unmount(); app = undefined; expect(signal.aborted).toBe(true)
  resolve([batch]); await flush(); expect(refresh).not.toHaveBeenCalled()
})
