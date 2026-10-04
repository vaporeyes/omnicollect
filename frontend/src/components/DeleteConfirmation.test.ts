// ABOUTME: Tests fixed deletion targets, duplicate protection and recoverable errors.
// ABOUTME: Native modal methods are stubbed; browser focus trapping is not simulated.
import {afterEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, reactive, type App} from 'vue'
import DeleteConfirmation from './DeleteConfirmation.vue'
import {APIError} from '../api/client'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
let app: App | undefined, host: HTMLElement, cleanup: (() => void) | undefined
const flush = async () => {await Promise.resolve(); await Promise.resolve(); await nextTick()}
afterEach(() => {app?.unmount(); app = undefined; host?.remove(); cleanup?.()})
it('uses the confirmed target snapshot and blocks duplicates, Escape and unload during writes', async () => {
  cleanup = stubNativeDialogs(); host = document.createElement('div'); document.body.append(host)
  let resolve!: (count: number) => void
  const remove = vi.fn(() => new Promise<number>(done => {resolve = done})), close = vi.fn(), deleted = vi.fn()
  const state = reactive({items: [{id: 'one', title: 'First'}]})
  app = createApp({render: () => h(DeleteConfirmation, {items: state.items, remove, onClose: close, onDeleted: deleted})}); app.mount(host)
  state.items = [{id: 'two', title: 'Other'}]; await nextTick()
  const submit = host.querySelector('button.danger') as HTMLButtonElement
  submit.click(); submit.click(); await nextTick()
  expect(remove).toHaveBeenCalledTimes(1); expect(remove).toHaveBeenCalledWith(['one'])
  host.querySelector('dialog')!.dispatchEvent(new Event('cancel', {cancelable: true}))
  expect(close).not.toHaveBeenCalled()
  const unload = new Event('beforeunload', {cancelable: true}); window.dispatchEvent(unload); expect(unload.defaultPrevented).toBe(true)
  resolve(1); await flush()
  expect(deleted).toHaveBeenCalledWith(1, ['one'])
  submit.click(); await flush(); expect(remove).toHaveBeenCalledTimes(1)
})
it('retains failures and blocks uncertain deletion repeats until inspection', async () => {
  cleanup = stubNativeDialogs(); host = document.createElement('div'); document.body.append(host)
  const remove = vi.fn().mockRejectedValueOnce(new APIError('Invalid selection', 400)).mockRejectedValueOnce(new Error('offline'))
  const deleted = vi.fn()
  app = createApp({render: () => h(DeleteConfirmation, {items: [{id: 'one', title: 'First'}], remove, onDeleted: deleted})}); app.mount(host)
  const submit = host.querySelector('button.danger') as HTMLButtonElement
  submit.click(); await flush(); expect(host.textContent).toContain('Invalid selection')
  submit.click(); await flush(); expect(remove).toHaveBeenCalledTimes(1)
  expect(submit.disabled).toBe(true)
  expect(host.textContent).toContain('Inspect Deletion recovery')
  expect(deleted).not.toHaveBeenCalled()
})
