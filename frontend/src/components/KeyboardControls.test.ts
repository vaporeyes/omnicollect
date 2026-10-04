// ABOUTME: Keyboard/focus contracts for menus, image inspection and collection actions.
// ABOUTME: Native dialog is stubbed; real-browser trapping/layout is not asserted.
import {afterEach,expect,it,vi} from 'vitest'
import {mount,flushPromises,type VueWrapper} from '@vue/test-utils'
import {createPinia} from 'pinia'
import ContextMenu from './ContextMenu.vue'
import ImageLightbox from './ImageLightbox.vue'
import CollectionGrid from './CollectionGrid.vue'
import ItemList from './ItemList.vue'
import {useSelectionStore} from '../stores/selectionStore'
import {stubNativeDialogs} from '../testSupport/nativeDialog'
let wrapper:VueWrapper|undefined,opener:HTMLButtonElement|undefined,cleanup:(()=>void)|undefined
function makeOpener(){opener=document.createElement('button');document.body.append(opener);opener.focus()}
afterEach(()=>{wrapper?.unmount();opener?.remove();cleanup?.();vi.restoreAllMocks()})
it('focuses menu options, wraps arrows and returns focus on Escape without bubbling',async()=>{
 makeOpener()
 wrapper=mount(ContextMenu,{attachTo:document.body,props:{visible:true,x:-10,y:-20,options:[{label:'Edit',action:'edit'},{label:'Delete',action:'delete'}]}})
 await flushPromises()
 const menu=document.querySelector<HTMLElement>('[role=menu]')!,buttons=menu.querySelectorAll('button')
 expect(document.activeElement).toBe(buttons[0]);expect(menu.style.left).toBe('8px')
 buttons[0].dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowUp',bubbles:true,cancelable:true}))
 expect(document.activeElement).toBe(buttons[1])
 buttons[1].dispatchEvent(new KeyboardEvent('keydown',{key:'Home',bubbles:true,cancelable:true}))
 expect(document.activeElement).toBe(buttons[0])
 const globalKey=vi.fn();document.addEventListener('keydown',globalKey)
 buttons[0].dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,cancelable:true}))
 document.removeEventListener('keydown',globalKey)
 expect(globalKey).not.toHaveBeenCalled();expect(wrapper.emitted('close')).toHaveLength(1);expect(document.activeElement).toBe(opener)
})
it('selects exactly once and closes on Tab',async()=>{
 makeOpener();wrapper=mount(ContextMenu,{attachTo:document.body,props:{visible:true,x:0,y:0,options:[{label:'Edit',action:'edit'}]}})
 await flushPromises();const button=document.querySelector<HTMLButtonElement>('[role=menuitem]')!
 button.click();button.click();expect(wrapper.emitted('select')).toEqual([['edit']]);expect(wrapper.emitted('close')).toHaveLength(1)
 await wrapper.setProps({visible:false});await wrapper.setProps({visible:true});await flushPromises()
 document.querySelector('[role=menuitem]')!.dispatchEvent(new KeyboardEvent('keydown',{key:'Tab',bubbles:true}))
 expect(wrapper.emitted('close')).toHaveLength(2)
})
it('uses native modal hooks and supports keyboard image zoom, pan and reset',async()=>{
 cleanup=stubNativeDialogs()
 wrapper=mount(ImageLightbox,{props:{visible:true,filename:'a.png'},global:{stubs:{MediaImage:true}}})
 expect(wrapper.get('dialog').attributes('aria-label')).toBe('Image viewer')
 await wrapper.get('.loupe-container').trigger('keydown',{key:'Enter'})
 expect(wrapper.get('.loupe-container').attributes('aria-pressed')).toBe('true')
 await wrapper.get('dialog').trigger('keydown',{key:'+'})
 await wrapper.get('dialog').trigger('keydown',{key:'ArrowRight'})
 expect(wrapper.get('[role=status]').text()).toContain('2.8')
 expect(wrapper.get('media-image-stub').attributes('style')).toContain('60%')
 await wrapper.setProps({filename:'b.png'})
 expect(wrapper.get('.loupe-container').attributes('aria-pressed')).toBe('false')
 wrapper.get('dialog').element.dispatchEvent(new Event('cancel',{cancelable:true}))
 expect(wrapper.emitted('close')).toHaveLength(1)
})
it('exposes grid selection, item and context actions as named buttons',async()=>{
 const pinia=createPinia(),item={id:'a',title:'Example',moduleId:'m',images:[],updatedAt:''}
 wrapper=mount(CollectionGrid,{props:{items:[item as any],modules:[]},global:{plugins:[pinia]}})
 await wrapper.get('button[aria-label="Select Example"]').trigger('click')
 expect(useSelectionStore(pinia).isSelected('a')).toBe(true);expect(wrapper.emitted('select')).toBeUndefined()
 await wrapper.get('.card-open').trigger('click');expect(wrapper.emitted('select')).toHaveLength(1)
 await wrapper.get('button[aria-haspopup=menu]').trigger('click');expect(wrapper.emitted('itemContextMenu')).toHaveLength(1)
})
it('sorts using header buttons with aria-sort and named selection controls',async()=>{
 wrapper=mount(ItemList,{props:{items:[{id:'b',title:'B',moduleId:'m'},{id:'a',title:'A',moduleId:'m'}] as any,modules:[]},global:{plugins:[createPinia()]}})
 await wrapper.get('.sort-control').trigger('click')
 expect(wrapper.findAll('.item-open').map(button=>button.text())).toEqual(['A','B'])
 expect(wrapper.get('th[aria-sort=ascending]').text()).toContain('Title')
 expect(wrapper.find('input[aria-label="Select all displayed items"]').exists()).toBe(true)
 await wrapper.setProps({modules:[{id:'m',displayName:'Module',attributes:[]}],activeModuleId:'m'})
 expect(wrapper.get('select').attributes('aria-label')).toBe('Collection type')
 expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('m')
 await wrapper.setProps({activeModuleId:''})
 expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('')
})
