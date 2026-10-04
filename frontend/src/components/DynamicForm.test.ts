// ABOUTME: Form save coordination, draft retention, and stale AI response tests.
// ABOUTME: Uses mounted Vue components with isolated Pinia state and deferred promises.
import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {createApp, defineComponent, h, nextTick, reactive, ref, type App} from 'vue'
import {createPinia} from 'pinia'
import DynamicForm from './DynamicForm.vue'
import {analyzeItem} from '../api/client'
import type {Item, ModuleSchema} from '../api/types'

vi.mock('../api/client', () => ({getAllTags: vi.fn(async () => []), getAIStatus: vi.fn(async () => ({enabled: true})), analyzeItem: vi.fn()}))
vi.mock('./ImageAttach.vue', () => ({default: defineComponent({emits: ['busy', 'update:images'], setup(_, {emit}) {return () => h('button', {type: 'button', 'data-upload': '', onClick: () => emit('busy', true)}, 'Upload')}})}))

let app: App | undefined
let host: HTMLDivElement
const flush = async () => {for (let n = 0; n < 10; n++) await nextTick()}
const schema: ModuleSchema = {id: 'books', displayName: 'Books', attributes: [{name: 'count', type: 'number'}]}
function item(id: string): Item {return {id, moduleId: 'books', title: id, images: ['cover.png'], tags: [], attributes: {count: 1, legacy: 'keep'}, purchasePrice: null, createdAt: '2020-01-01T00:00:00Z', updatedAt: '2025-01-01T00:00:00.123456Z'}}
function mount(saveItem: (item: Item) => Promise<boolean>) {
  const props = reactive({schema, item: item('first'), saveItem})
  const form = ref<InstanceType<typeof DynamicForm>>()
  app = createApp({render: () => h(DynamicForm, {...props, ref: form})})
  app.use(createPinia()).mount(host)
  return {props, form}
}
function submit() {host.querySelector('form')!.dispatchEvent(new Event('submit', {bubbles: true, cancelable: true}))}
beforeEach(() => {host = document.createElement('div'); document.body.appendChild(host); vi.mocked(analyzeItem).mockReset()})
afterEach(() => {app?.unmount(); host.remove()})

it('sends preserved attributes/version once and keeps the draft after failure', async () => {
  let finish!: (saved: boolean) => void
  const save = vi.fn((_item: Item) => new Promise<boolean>(resolve => {finish = resolve}))
  const {form} = mount(save)
  await flush()
  const title = host.querySelector<HTMLInputElement>('[placeholder="Item title"]')!
  title.value = 'Unsaved title'; title.dispatchEvent(new Event('input', {bubbles: true}))
  submit(); submit()
  await flush()
  expect(save).toHaveBeenCalledTimes(1)
  expect(save.mock.calls[0][0]).toMatchObject({title: 'Unsaved title', attributes: {legacy: 'keep'}, updatedAt: item('first').updatedAt})
  expect(host.querySelector('fieldset')?.disabled).toBe(true)
  finish(false); await flush()
  expect(form.value?.dirty).toBe(true)
  expect(title.value).toBe('Unsaved title')
  expect(form.value?.submitting).toBe(false)
})

it('locks repeat creation after a simulated committed save loses its acknowledgement', async () => {
  let committed = 0
  const save = vi.fn(async () => {committed++; throw new Error('lost acknowledgement')})
  const {props} = mount(save)
  props.item = {...item(''), updatedAt: '', createdAt: ''}
  await flush()
  const title = host.querySelector<HTMLInputElement>('[placeholder="Item title"]')!
  title.value = 'New draft'; title.dispatchEvent(new Event('input', {bubbles: true}))
  submit(); await flush(); submit(); await flush()
  expect(committed).toBe(1)
  expect(title.value).toBe('New draft')
  expect(host.textContent).toContain('may have completed')
  expect(host.querySelector<HTMLButtonElement>('button[type=submit]')?.disabled).toBe(true)
})
it('allows correction after an explicit validation rejection', async () => {
  const save = vi.fn(async () => {throw Object.assign(new Error('invalid field'), {status: 400})})
  mount(save); await flush(); submit(); await flush(); submit(); await flush()
  expect(save).toHaveBeenCalledTimes(2)
})
it('blocks saving while an upload is pending', async () => {
  const save = vi.fn(async () => true)
  const {form} = mount(save)
  host.querySelector<HTMLButtonElement>('[data-upload]')!.click()
  submit(); await flush()
  expect(save).not.toHaveBeenCalled()
  expect(form.value?.working).toBe(true)
})

it('aborts and ignores AI results for a replaced draft', async () => {
  let finish!: (result: any) => void
  vi.mocked(analyzeItem).mockImplementation(() => new Promise(resolve => {finish = resolve}))
  const {props} = mount(async () => true)
  await flush()
  host.querySelector<HTMLButtonElement>('.btn-smart-scan')!.click()
  const signal = vi.mocked(analyzeItem).mock.calls[0][2]
  props.item = item('second')
  await flush()
  expect(signal?.aborted).toBe(true)
  finish({title: 'Stale title', attributes: {count: 999}, warnings: []})
  await flush()
  expect(host.querySelector<HTMLInputElement>('[placeholder="Item title"]')!.value).toBe('second')
  expect(host.textContent).not.toContain('Stale title')
})
