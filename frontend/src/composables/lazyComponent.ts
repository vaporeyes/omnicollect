// ABOUTME: Lazy view boundaries with visible loading and recoverable load-error states.
// ABOUTME: Keeps charts/editors out of the initial application entry module.
import {defineAsyncComponent, defineComponent, h, type Component} from 'vue'
const loading = defineComponent({render: () => h('p', {role: 'status'}, 'Loading view…')})
const failed = defineComponent({render: () => h('div', {role: 'alert'}, [
 h('p', 'This view could not load. Check your connection and reload to try again.'),
 h('button', {onClick: () => window.location.reload()}, 'Reload application'),
])})
export function lazyComponent<T extends Component>(loader: () => Promise<{default: T}>) {
 return defineAsyncComponent<T>({loader: () => loader().then(module => module.default), delay: 150, timeout: 20000, loadingComponent: loading, errorComponent: failed})
}
