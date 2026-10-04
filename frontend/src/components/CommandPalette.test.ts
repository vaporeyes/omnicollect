// ABOUTME: Palette search ordering, cancellation, failure states, and Escape regressions.
// ABOUTME: Uses deferred search promises; native modal mechanics require browser verification.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, reactive, type App} from 'vue'
import CommandPalette from './CommandPalette.vue'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
const {search} = vi.hoisted(() => ({search: vi.fn()}))
vi.mock('../stores/collectionStore', async importOriginal => ({...await importOriginal<any>(), useCollectionStore: () => ({searchAllItems: search})}))
vi.mock('../stores/moduleStore', () => ({useModuleStore: () => ({getModuleById: () => ({displayName: 'Books'})})}))
let app: App
let host: HTMLElement
let restore: () => void
const flush = async () => {for (let i = 0; i < 5; i++) await nextTick()}
function input(value: string) {
  const field = document.querySelector<HTMLInputElement>('.palette-input')!
  field.value = value; field.dispatchEvent(new Event('input', {bubbles: true}))
}
beforeEach(() => {vi.useFakeTimers(); search.mockReset(); restore = stubNativeDialogs(); host = document.createElement('div'); document.body.appendChild(host)})
afterEach(() => {app.unmount(); host.remove(); restore(); vi.useRealTimers(); vi.restoreAllMocks()})
function mount() {
  const state = reactive({visible: true})
  app = createApp({render: () => h(CommandPalette, {visible: state.visible, onClose: () => {state.visible = false}})})
  app.mount(host)
  return state
}
it('ignores late old responses and aborts immediately when input changes', async () => {
  const pending: Array<(value: any[]) => void> = []
  search.mockImplementation(() => new Promise(resolve => pending.push(resolve)))
  mount(); input('older'); await vi.advanceTimersByTimeAsync(200)
  const oldSignal = search.mock.calls[0][1]
  input('newer')
  expect(oldSignal.aborted).toBe(true)
  await vi.advanceTimersByTimeAsync(200)
  pending[1]([{id: 'new', title: 'Current result', moduleId: 'books', images: []}]); await flush()
  pending[0]([{id: 'old', title: 'Obsolete result', moduleId: 'books', images: []}]); await flush()
  expect(document.querySelector('.palette-dialog')?.textContent).toContain('Current result')
  expect(document.querySelector('.palette-dialog')?.textContent).toContain('up to 20 matching items')
  expect(document.querySelector('.palette-dialog')?.textContent).not.toContain('Obsolete result')
})
it('shows failures distinctly from an empty search and cancels on close', async () => {
  search.mockRejectedValue(new Error('Search unavailable'))
  const state = mount(); input('query'); await vi.advanceTimersByTimeAsync(200); await flush()
  expect(document.querySelector('[role="alert"]')?.textContent).toBe('Search unavailable')
  expect(document.querySelector('.palette-dialog')?.textContent).not.toContain('No results for')
  input('another'); state.visible = false; await flush(); await vi.advanceTimersByTimeAsync(300)
  expect(search).toHaveBeenCalledTimes(1)
})
it('does not bubble Escape into the underlying editor shortcut handler', async () => {
  mount(); await flush()
  const listener = vi.fn(); document.addEventListener('keydown', listener)
  document.querySelector('.palette-input')!.dispatchEvent(new KeyboardEvent('keydown', {key: 'Escape', bubbles: true}))
  expect(listener).not.toHaveBeenCalled()
  document.removeEventListener('keydown', listener)
})
