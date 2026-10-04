// ABOUTME: Detail deletion uses the shared parent confirmation; schema reorders remain detached drafts.
// ABOUTME: Verifies control semantics without treating stubbed images or drag events as browser evidence.
import {expect,it} from 'vitest'
import {mount} from '@vue/test-utils'
import draggable from 'vuedraggable'
import ItemDetail from './ItemDetail.vue'
import SchemaVisualEditor from './SchemaVisualEditor.vue'
import ComparisonView from './ComparisonView.vue'

const item:any={id:'one',title:'Fixture',moduleId:'m',attributes:{},images:['one.png','two.png'],tags:[],createdAt:'2026-01-01',updatedAt:'2026-01-01'}
function schema(){return {id:'m',displayName:'Fixture',description:'',attributes:['a','b'].map(name=>({name,type:'string',required:false,options:[],display:{label:'',placeholder:'',widget:''}}))}}
it('routes detail deletion directly to the parent recovery-aware confirmation',async()=>{
 const view=mount(ItemDetail,{props:{item,schema:null},global:{stubs:{MediaImage:true}}})
 await view.get('.action-delete').trigger('click')
 expect(view.emitted('delete')).toHaveLength(1)
 expect(document.body.textContent).not.toContain('This action cannot be undone')
 await view.get('button[aria-label="Show image 2"]').trigger('click')
 await view.get('.gallery-open').trigger('click')
 expect(view.emitted('viewImage')).toEqual([['two.png']])
 expect(view.get('button[aria-label="Next image"]').attributes('disabled')).toBeDefined()
 view.unmount()
})
it('applies one immutable drag reorder through modelValue',async()=>{
 const original=schema(),view=mount(SchemaVisualEditor,{props:{schema:original}})
 view.getComponent(draggable).vm.$emit('update:modelValue',[original.attributes[1],original.attributes[0]])
 expect((view.emitted('update:schema')?.[0]?.[0] as any)?.attributes.map((a:any)=>a.name)).toEqual(['b','a'])
 expect(original.attributes.map(a=>a.name)).toEqual(['a','b'])
 view.unmount()
})
it('keeps editing input identity and provides immutable keyboard reorder controls',async()=>{
 const original=schema(),view=mount(SchemaVisualEditor,{props:{schema:original}})
 const input=view.get('input[aria-label="Field 1 name"]'),element=input.element
 await input.setValue('renamed')
 await view.setProps({schema:view.emitted('update:schema')![0][0] as any})
 expect(view.get('input[aria-label="Field 1 name"]').element).toBe(element)
 await view.get('button[aria-label="Move field 1 down"]').trigger('click')
 const reordered=view.emitted('update:schema')!.at(-1)![0] as any
 expect(reordered.attributes.map((a:any)=>a.name)).toEqual(['b','renamed'])
 expect(original.attributes.map(a=>a.name)).toEqual(['a','b'])
 view.unmount()
})
it('provides native named comparison image controls',async()=>{
 const view=mount(ComparisonView,{props:{itemA:item,itemB:{...item,title:'Other'},schemaA:null,schemaB:null},global:{stubs:{MediaImage:true}}})
 await view.get('button[aria-label="Open full-size image for Other"]').trigger('click')
 expect(view.emitted('viewImage')).toEqual([['one.png']])
 view.unmount()
})
