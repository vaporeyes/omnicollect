// ABOUTME: Verifies modal dismissal policy and focus restoration around native dialog calls.
// ABOUTME: Browser-native focus trapping and inertness are intentionally not simulated.
import {afterEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, reactive, type App} from 'vue'
import ModalSurface from './ModalSurface.vue'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
let app: App | undefined
let cleanup: (() => void) | undefined
let host: HTMLElement | undefined
let opener: HTMLButtonElement | undefined
afterEach(() => {app?.unmount(); app = undefined; host?.remove(); opener?.remove(); cleanup?.()})
it('blocks native cancel while busy and restores the invoking control on removal', async () => {
  cleanup = stubNativeDialogs()
  host = document.createElement('div'); document.body.appendChild(host)
  opener = document.createElement('button'); document.body.appendChild(opener); opener.focus()
  const state = reactive({busy: true}), close = vi.fn()
  app = createApp({render: () => h(ModalSurface, {label: 'Test modal', busy: state.busy, onClose: close}, () => h('button', 'Inside'))})
  app.mount(host)
  const dialog = host.querySelector('dialog')!
  expect(dialog.open).toBe(true)
  const cancelled = new Event('cancel', {cancelable: true})
  dialog.dispatchEvent(cancelled)
  expect(cancelled.defaultPrevented).toBe(true)
  expect(close).not.toHaveBeenCalled()
  state.busy = false; await nextTick()
  dialog.dispatchEvent(new Event('cancel', {cancelable: true}))
  expect(close).toHaveBeenCalledTimes(1)
  dialog.querySelector('button')!.focus()
  app.unmount(); app = undefined; await nextTick()
  expect(document.activeElement).toBe(opener)
})
it('wraps explicit Tab boundaries in both directions', async () => {
  cleanup = stubNativeDialogs()
  host = document.createElement('div'); document.body.appendChild(host)
  app = createApp({render: () => h(ModalSurface, {label: 'Test modal'}, () => [h('button', 'First'), h('button', 'Last')])})
  app.mount(host); await nextTick()
  const buttons = host.querySelectorAll('button')
  // jsdom has no layout; these rectangles test only our boundary policy.
  buttons.forEach(button => {button.getClientRects = () => [{}] as any})
  buttons[1].focus()
  const forward = new KeyboardEvent('keydown', {key: 'Tab', bubbles: true, cancelable: true})
  buttons[1].dispatchEvent(forward)
  expect(forward.defaultPrevented).toBe(true); expect(document.activeElement).toBe(buttons[0])
  buttons[0].dispatchEvent(new KeyboardEvent('keydown', {key: 'Tab', shiftKey: true, bubbles: true, cancelable: true}))
  expect(document.activeElement).toBe(buttons[1])
})
it('repairs lost focus after slot mounting without stealing valid child focus', async () => {
  cleanup = stubNativeDialogs()
  host = document.createElement('div'); document.body.appendChild(host)
  opener = document.createElement('button'); document.body.appendChild(opener); opener.focus()
  const state = reactive({busy: false})
  app = createApp({render: () => h(ModalSurface, {label: 'Test modal', busy: state.busy}, () => h('button', 'Inside'))})
  app.mount(host)
  const dialog = host.querySelector('dialog')!
  const focus = vi.spyOn(dialog, 'focus')
  await nextTick()
  expect(focus).toHaveBeenCalledOnce()
  focus.mockClear()
  dialog.querySelector('button')!.focus()
  state.busy = true; await nextTick(); await nextTick()
  expect(focus).not.toHaveBeenCalled()
  expect(document.activeElement).toBe(dialog.querySelector('button'))
  dialog.querySelector('button')!.disabled = true
  state.busy = false; await nextTick(); await nextTick()
  expect(focus).toHaveBeenCalledOnce()
})
