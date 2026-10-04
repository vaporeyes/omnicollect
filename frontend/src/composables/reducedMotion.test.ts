// ABOUTME: Verifies live reduced-motion preferences and listener disposal.
// ABOUTME: Browser rendering itself is not simulated by this unit test.
import {expect,it,vi} from 'vitest'
import {createApp,h} from 'vue'
import {useReducedMotion} from './reducedMotion'
it('tracks preference changes and removes the listener',()=>{
 const add=vi.fn(),remove=vi.fn()
 vi.stubGlobal('matchMedia',vi.fn(()=>({matches:true,addEventListener:add,removeEventListener:remove})))
 const host=document.createElement('div')
 let reduced!:ReturnType<typeof useReducedMotion>
 const app=createApp({setup(){reduced=useReducedMotion();return ()=>h('div')}})
 try {
 app.mount(host);expect(reduced.value).toBe(true)
 const listener=add.mock.calls[0][1];listener({matches:false});expect(reduced.value).toBe(false)
 app.unmount();expect(remove).toHaveBeenCalledWith('change',listener)
 }finally{vi.unstubAllGlobals()}
})
