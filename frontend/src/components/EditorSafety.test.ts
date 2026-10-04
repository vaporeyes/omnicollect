// ABOUTME: Mounted regressions for schema and settings draft safety.
// ABOUTME: Exercises immediate saves, invalid JSON, duplicate writes, exit guards and theme rollback.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, type App} from 'vue'
import SchemaBuilder from './SchemaBuilder.vue'
import SettingsPage from './SettingsPage.vue'
import * as api from '../api/client'
import {applyTheme} from '../theme'
vi.mock('../api/client', () => ({post: vi.fn(), put: vi.fn()}))
vi.mock('../theme', () => ({applyTheme: vi.fn(), DEFAULT_CONFIG: {mode: 'light'}}))
vi.mock('./SchemaCodeEditor.vue', () => ({default: {props: ['modelValue', 'disabled'], emits: ['update:modelValue'], render(this: any) {return h('textarea', {value: this.modelValue, disabled: this.disabled, onInput: (e: Event) => this.$emit('update:modelValue', (e.target as HTMLTextAreaElement).value)})}}}))
vi.mock('./SchemaVisualEditor.vue', () => ({default: {render: () => h('div')}}))
vi.mock('./SchemaFormPreview.vue', () => ({default: {render: () => h('div')}}))
let app: App | undefined, host: HTMLElement, exposed: any
beforeEach(() => {vi.clearAllMocks(); host = document.createElement('div'); document.body.append(host); vi.spyOn(window, 'confirm').mockReturnValue(false)})
afterEach(() => {app?.unmount(); app = undefined; host.remove(); vi.restoreAllMocks()})
const flush = async () => {await Promise.resolve(); await Promise.resolve(); await nextTick()}
function mount(component: any, props: any) {app = createApp({render: () => h(component, {...props, ref: (value: any) => {exposed = value}})}); app.mount(host)}
function button(text: string) {return [...host.querySelectorAll('button')].find(b => b.textContent?.trim() === text)!}
function code(text: string) {const el = host.querySelector('textarea')!; el.value = text; el.dispatchEvent(new Event('input', {bubbles: true}))}
const schema = {id: 'books', displayName: 'Books', attributes: []}
it('saves latest code before debounce, blocks duplicate writes and exits, retains failure', async () => {
  let reject!: (e: Error) => void
  vi.mocked(api.post).mockReturnValue(new Promise((_resolve, fail) => {reject = fail}))
  const saved = vi.fn()
  mount(SchemaBuilder, {initialJSON: JSON.stringify(schema), onSaved: saved})
  code(JSON.stringify({...schema, displayName: 'Latest'}))
  button('Save').click(); button('Save').click()
  expect(api.post).toHaveBeenCalledTimes(1)
  expect(vi.mocked(api.post).mock.calls[0][1]).toMatchObject({displayName: 'Latest'})
  expect(exposed.canLeave()).toBe(false)
  const unload = new Event('beforeunload', {cancelable: true}); window.dispatchEvent(unload); expect(unload.defaultPrevented).toBe(true)
  reject(new Error('Schema would invalidate an item')); await flush()
  expect(host.textContent).toContain('Schema would invalidate an item')
  expect(host.querySelector('textarea')!.value).toContain('Latest')
  expect(exposed.canLeave()).toBe(false)
  vi.mocked(window.confirm).mockReturnValue(true)
  expect(exposed.canLeave()).toBe(true)
  expect(saved).not.toHaveBeenCalled()
})
it('rejects invalid JSON shape without sending a stale visual draft', async () => {
  mount(SchemaBuilder, {initialJSON: JSON.stringify(schema)})
  for (const text of ['null', '[]', '{', '{"id":1}', '{"id":"a","displayName":"A","attributes":[null]}']) {
    code(text); button('Save').click(); await flush()
    expect(api.post).not.toHaveBeenCalled()
    expect(host.querySelector('.save-error')).not.toBeNull()
  }
})
it('blocks settings duplicates and closes, retains failed preview and restores committed theme', async () => {
  let reject!: (e: Error) => void
  vi.mocked(api.put).mockReturnValue(new Promise((_resolve, fail) => {reject = fail}))
  const close = vi.fn()
  mount(SettingsPage, {initialConfig: {mode: 'light'}, systemDark: false, onClose: close})
  button('Dark').click(); await flush()
  expect(applyTheme).toHaveBeenLastCalledWith(true)
  expect(exposed.canLeave()).toBe(false)
  button('Save').click(); button('Save').click(); await flush()
  expect(api.put).toHaveBeenCalledTimes(1)
  expect(exposed.canLeave()).toBe(false)
  reject(new Error('offline')); await flush()
  expect(host.textContent).toContain('Error: offline')
  expect(host.querySelector('.mode-btn.active')?.textContent?.trim()).toBe('Dark')
  expect(close).not.toHaveBeenCalled()
  app?.unmount(); app = undefined
  expect(applyTheme).toHaveBeenLastCalledWith(false)
})
it('successful settings saves establish a clean baseline and preserve the committed theme', async () => {
  vi.mocked(api.put).mockResolvedValue(undefined)
  mount(SettingsPage, {initialConfig: {mode: 'light'}, systemDark: false})
  button('Dark').click(); await flush(); button('Save').click(); await flush()
  expect(exposed.canLeave()).toBe(true)
  expect(window.confirm).not.toHaveBeenCalled()
  app?.unmount(); app = undefined
  expect(applyTheme).toHaveBeenLastCalledWith(true)
})
