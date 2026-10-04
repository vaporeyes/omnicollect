// ABOUTME: Import duplicate/dismissal guards, replace consent, retry and result visibility tests.
// ABOUTME: Native dialog APIs are stubbed; no archive uploads or real storage are touched.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, nextTick, type App} from 'vue'
import ImportDialog from './ImportDialog.vue'
import {analyzeBackup, executeImport} from '../api/client'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
vi.mock('../api/client', async original => ({...await original<typeof import('../api/client')>(), analyzeBackup: vi.fn(), executeImport: vi.fn()}))
let app: App
let host: HTMLElement
let restore: () => void
const flush = async () => {for (let i = 0; i < 8; i++) await nextTick()}
beforeEach(() => {
  restore = stubNativeDialogs(); host = document.createElement('div'); document.body.appendChild(host)
  vi.mocked(analyzeBackup).mockReset(); vi.mocked(executeImport).mockReset()
  vi.mocked(analyzeBackup).mockResolvedValue({format: 'cloud', itemCount: 1, imageCount: 0, moduleCount: 1, warnings: [], tempId: 'owned-handle'})
})
afterEach(() => {app.unmount(); host.remove(); restore()})
async function mount(extra: {refreshError?: string} = {}) {
  const close = vi.fn(), imported = vi.fn(), refresh = vi.fn()
  app = createApp(ImportDialog, {...extra, onClose: close, onImported: imported, onRefresh: refresh})
  const instance = app.mount(host) as unknown as {requestClose: () => void}
  const input = document.querySelector<HTMLInputElement>('input[type="file"]')!
  Object.defineProperty(input, 'files', {value: [new File(['zip'], 'backup.zip')]})
  input.dispatchEvent(new Event('change')); await flush()
  return {instance, close, imported, refresh}
}
function confirm() {document.querySelector<HTMLButtonElement>('.import-confirm-btn')!.click()}
it('cannot duplicate or dismiss a pending import and keeps committed results visible', async () => {
  let finish!: (result: any) => void
  vi.mocked(executeImport).mockImplementation(() => new Promise(resolve => {finish = resolve}))
  const {instance, close, imported} = await mount()
  confirm(); confirm(); await flush()
  instance.requestClose()
  document.querySelector('dialog')!.dispatchEvent(new Event('cancel', {cancelable: true}))
  expect(close).not.toHaveBeenCalled()
  expect(executeImport).toHaveBeenCalledTimes(1)
  finish({itemsImported: 1, imagesRestored: 0, modulesImported: 1, warnings: ['Review restored settings']}); await flush()
  expect(imported).toHaveBeenCalledTimes(1)
  expect(document.querySelector('.import-dialog')?.textContent).toContain('Items imported')
  expect(document.querySelector('.import-dialog')?.textContent).toContain('Review restored settings')
  expect(close).not.toHaveBeenCalled()
  confirm(); expect(close).toHaveBeenCalledTimes(1)
})
it('retries only the view refresh after a committed import', async () => {
  vi.mocked(executeImport).mockResolvedValue({itemsImported: 1, imagesRestored: 0, modulesImported: 1, warnings: []})
  const {refresh} = await mount({refreshError: 'The collection view could not refresh.'})
  confirm(); await flush()
  const retry = [...document.querySelectorAll<HTMLButtonElement>('button')].find(button => button.textContent === 'Retry view refresh')!
  retry.click()
  expect(refresh).toHaveBeenCalledTimes(1)
  expect(executeImport).toHaveBeenCalledTimes(1)
})

it('requires explicit consent before replacing metadata', async () => {
  await mount()
  document.querySelector<HTMLInputElement>('input[value="replace"]')!.click(); await flush()
  expect(document.querySelector<HTMLButtonElement>('.import-confirm-btn')!.disabled).toBe(true)
  confirm(); expect(executeImport).not.toHaveBeenCalled()
})
it('warns about unknown commit outcome on connection loss and retains the handle', async () => {
  vi.mocked(executeImport).mockRejectedValue(new TypeError('Network failure'))
  await mount(); confirm(); await flush()
  expect(document.querySelector('.import-dialog')?.textContent).toContain('may have completed')
  confirm(); await flush()
  expect(executeImport).toHaveBeenNthCalledWith(2, 'owned-handle', 'merge')
})
