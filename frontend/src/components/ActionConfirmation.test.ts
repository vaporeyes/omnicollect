// ABOUTME: Mutation confirmations freeze intent and guard pending or uncertain writes.
// ABOUTME: Native dialog APIs are stubbed; these are not browser focus tests.
import {afterEach, expect, it, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import ActionConfirmation from './ActionConfirmation.vue'
import TagManager from './TagManager.vue'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
let wrapper: VueWrapper | undefined, cleanup: (() => void) | undefined
afterEach(() => {wrapper?.unmount(); cleanup?.(); vi.restoreAllMocks()})
it('freezes the action and blocks duplicate writes, cancellation and unload', async () => {
 cleanup = stubNativeDialogs()
 let finish!: (value: string) => void
 const run = vi.fn(() => new Promise<string>(resolve => {finish = resolve}))
 wrapper = mount(ActionConfirmation, {props: {action: {title:'Move A',description:'Only A',label:'Move',run}}})
 const replacement = vi.fn()
 await wrapper.setProps({action:{title:'Move B',description:'Only B',label:'Move',run:replacement}})
 await wrapper.get('.confirm-action').trigger('click'); await wrapper.get('.confirm-action').trigger('click')
 expect(run).toHaveBeenCalledTimes(1); expect(replacement).not.toHaveBeenCalled()
 expect(wrapper.text()).toContain('Move A')
 wrapper.get('dialog').element.dispatchEvent(new Event('cancel',{cancelable:true}))
 expect(wrapper.emitted('close')).toBeUndefined()
 const unload=new Event('beforeunload',{cancelable:true});window.dispatchEvent(unload);expect(unload.defaultPrevented).toBe(true)
 finish('Saved');await flushPromises();expect(wrapper.emitted('confirmed')).toEqual([['Saved']])
})
it('retains uncertain failures and cannot repeat the mutation',async()=>{
 cleanup=stubNativeDialogs();const run=vi.fn().mockRejectedValue(new Error('offline'))
 wrapper=mount(ActionConfirmation,{props:{action:{title:'Delete tag',description:'Every collection',label:'Delete',run}}})
 await wrapper.get('.confirm-action').trigger('click');await flushPromises()
 expect(wrapper.get('[role=alert]').text()).toContain('may have completed')
 expect(wrapper.emitted('uncertain')).toHaveLength(1)
 await wrapper.get('.confirm-action').trigger('click');expect(run).toHaveBeenCalledTimes(1)
 expect(wrapper.emitted('confirmed')).toBeUndefined()
})
it('retains a rename draft until the parent confirms success and guards exit',async()=>{
 wrapper=mount(TagManager,{props:{tags:[{name:'old',count:2}]}})
 await wrapper.get('.rename-btn').trigger('click')
 await wrapper.get('input').setValue('new')
 await wrapper.get('.save-btn').trigger('click')
 expect(wrapper.emitted('rename')).toEqual([[{oldName:'old',newName:'new'}]])
 expect((wrapper.get('input').element as HTMLInputElement).value).toBe('new')
 const confirm=vi.spyOn(window,'confirm').mockReturnValue(false)
 expect((wrapper.vm as any).canLeave()).toBe(false);expect(confirm).toHaveBeenCalled()
 await wrapper.setProps({disabled:true});await wrapper.get('.save-btn').trigger('click')
 expect(wrapper.emitted('rename')).toHaveLength(1)
})
