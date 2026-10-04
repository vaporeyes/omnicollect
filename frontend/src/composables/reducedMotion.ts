// ABOUTME: Live operating-system motion preference for canvas animations.
// ABOUTME: CSS handles DOM motion; this listener is disposed with its component.
import {ref, onBeforeUnmount} from 'vue'
export function useReducedMotion() {
 const query = window.matchMedia('(prefers-reduced-motion: reduce)')
 const reduced = ref(query.matches)
 const update = (event: MediaQueryListEvent) => {reduced.value = event.matches}
 query.addEventListener('change', update)
 onBeforeUnmount(() => query.removeEventListener('change', update))
 return reduced
}
