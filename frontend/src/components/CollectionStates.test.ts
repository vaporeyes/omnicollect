// ABOUTME: Lists/grids distinguish unavailable results from genuinely empty or filtered results.
// ABOUTME: Search stays usable without showing stale rows during unavailable states.
import {afterEach,expect,it} from 'vitest'
import {mount,type VueWrapper} from '@vue/test-utils'
import {createPinia} from 'pinia'
import ItemList from './ItemList.vue'
import CollectionGrid from './CollectionGrid.vue'
let wrapper:VueWrapper|undefined
afterEach(()=>wrapper?.unmount())
for(const component of [ItemList,CollectionGrid]){
 it(`${component.__name} suppresses empty hints and stale rows while unavailable`,async()=>{
  wrapper=mount(component,{props:{items:[],modules:[],unavailable:true},global:{plugins:[createPinia()]}})
  expect(wrapper.find('.empty-state').exists()).toBe(false)
  await wrapper.setProps({items:[{id:'stale',title:'Private stale row'} as any]})
  expect(wrapper.text()).not.toContain('Private stale row')
  if(component===ItemList)expect(wrapper.find('input[aria-label="Search collection"]').exists()).toBe(true)
 })
 it(`${component.__name} distinguishes filters from an empty collection`,async()=>{
  wrapper=mount(component,{props:{items:[],modules:[{id:'a',displayName:'A',attributes:[]}],filtered:true},global:{plugins:[createPinia()]}})
  expect(wrapper.text()).toContain('No items match')
  expect(wrapper.find('.cta-btn').exists()).toBe(false)
  await wrapper.setProps({filtered:false})
  expect(wrapper.text()).toContain('No items in the loaded collection')
  expect(wrapper.find('.cta-btn').exists()).toBe(true)
 })
}
