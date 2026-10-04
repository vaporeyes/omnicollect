// ABOUTME: Upload lifecycle regressions for duplicate selection and cancelled editors.
// ABOUTME: Ensures obsolete uploads cannot attach images to replacement drafts.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, nextTick, type App} from 'vue'
import ImageAttach from './ImageAttach.vue'
import {postFile} from '../api/client'

vi.mock('../api/client', () => ({postFile: vi.fn(), getMedia: vi.fn()}))
let app: App | undefined
let host: HTMLDivElement
const flush = async () => {for (let i = 0; i < 6; i++) await nextTick()}
beforeEach(() => {host = document.createElement('div'); document.body.appendChild(host); vi.mocked(postFile).mockReset()})
afterEach(() => {app?.unmount(); app = undefined; host.remove()})
function select() {
  const input = host.querySelector('input')!
  Object.defineProperty(input, 'files', {configurable: true, value: [new File(['image'], 'photo.png', {type: 'image/png'})]})
  input.dispatchEvent(new Event('change', {bubbles: true}))
}
it('rejects a second selection until the first upload completes', async () => {
  let finish!: (result: any) => void
  vi.mocked(postFile).mockImplementation(() => new Promise(resolve => {finish = resolve}))
  const update = vi.fn(), busy = vi.fn()
  app = createApp(ImageAttach, {images: [], 'onUpdate:images': update, onBusy: busy})
  app.mount(host)
  select(); select()
  expect(postFile).toHaveBeenCalledTimes(1)
  expect(busy).toHaveBeenCalledWith(true)
  finish({filename: 'new.png'}); await flush()
  expect(update).toHaveBeenCalledWith(['new.png'])
  expect(busy).toHaveBeenLastCalledWith(false)
})
it('aborts on unmount and ignores late successful responses', async () => {
  let finish!: (result: any) => void
  vi.mocked(postFile).mockImplementation(() => new Promise(resolve => {finish = resolve}))
  const update = vi.fn()
  app = createApp(ImageAttach, {images: [], 'onUpdate:images': update})
  app.mount(host); select()
  const signal = vi.mocked(postFile).mock.calls[0][3]
  app.unmount(); app = undefined
  expect(signal?.aborted).toBe(true)
  finish({filename: 'stale.png'}); await flush()
  expect(update).not.toHaveBeenCalled()
})
