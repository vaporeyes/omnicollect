import {createApp} from 'vue'
import {createPinia} from 'pinia'
import App from './App.vue'
import {auth0Plugin} from './auth/plugin'
import './style.css';

const app = createApp(App)
app.use(createPinia())
if (auth0Plugin) {
  app.use(auth0Plugin)
}
app.mount('#app')
