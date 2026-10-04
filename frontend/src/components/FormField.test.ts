// ABOUTME: Dynamic field accessibility and explicit disabled-state contracts.
// ABOUTME: Supplements real Chrome inspection without claiming screen-reader verification.
import {expect, it, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import FormField from './FormField.vue'
import MarkdownEditor from './MarkdownEditor.vue'
import SchemaFormPreview from './SchemaFormPreview.vue'
vi.mock('vue-codemirror',()=>({Codemirror:{props:['disabled'],template:'<div data-code :data-disabled="disabled" />'}}))

it.each(['string', 'number', 'date', 'enum'])('associates %s labels and errors with unique controls', type => {
  const props = {attribute: {name:'value',type,required:true,options:['one'],display:{label:'Readable label'}},modelValue:null,errorMessage:'Invalid value'}
  const a=mount(FormField,{props}), b=mount(FormField,{props})
  const input=a.get('input,select')
  expect(a.get('label').attributes('for')).toBe(input.attributes('id'))
  expect(input.attributes('id')).toBeTruthy()
  expect(input.attributes('id')).not.toBe(b.get('input,select').attributes('id'))
  expect(input.attributes('aria-invalid')).toBe('true')
  expect(input.attributes('aria-describedby')).toBe(a.get('[role=alert]').attributes('id'))
  a.unmount();b.unmount()
})
it('forwards labels, errors and disabled state to Markdown content',()=>{
  const view=mount(FormField,{props:{attribute:{name:'notes',type:'string',display:{widget:'textarea'}},modelValue:'Draft',disabled:true,errorMessage:'Invalid notes'},global:{stubs:{MarkdownEditor:true}}})
  const editor=view.getComponent(MarkdownEditor)
  expect(editor.props('disabled')).toBe(true)
  expect(editor.props('labelledBy')).toBe(view.get('label').attributes('id'))
  expect(editor.props('describedBy')).toBe(view.get('[role=alert]').attributes('id'))
  expect(editor.props('invalid')).toBe(true)
  view.unmount()
})
it('disables Markdown toolbar and editor while writes are pending',()=>{
  const view=mount(MarkdownEditor,{props:{modelValue:'Draft',disabled:true},global:{stubs:{Codemirror:{props:['disabled'],template:'<div data-code :data-disabled="disabled" />'}}}})
  expect(view.findAll('button').every(b=>b.attributes('disabled')!==undefined)).toBe(true)
  expect(view.get('[data-code]').attributes('data-disabled')).toBe('true')
  view.unmount()
})
it('makes schema previews non-editable to keyboard users',()=>{
  const view=mount(SchemaFormPreview,{props:{schema:{displayName:'Example',attributes:[{name:'field',type:'string'}]}}})
  expect(view.getComponent(FormField).props('disabled')).toBe(true)
  expect(view.findAll('input').every(input=>input.attributes('disabled')!==undefined)).toBe(true)
  view.unmount()
})
