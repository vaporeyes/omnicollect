import {createApp, defineComponent, h} from 'vue'
import {AuthGuard} from './auth/guard'
import {createPinia} from 'pinia'
import App from './App.vue'
import {auth0Plugin} from './auth/plugin'
import './style.css';

const Root = defineComponent({render: () => auth0Plugin ? h(AuthGuard, null, {default: () => h(App)}) : h(App)})
const app = createApp(Root)
app.use(createPinia())
if (auth0Plugin) {
  app.use(auth0Plugin)
}
app.mount('#app')
