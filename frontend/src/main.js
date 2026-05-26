import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router.js'
import { clearLegacyAuthStorage } from './stores/auth.js'
import './fonts/roboto.css'
import './style.css'

clearLegacyAuthStorage()

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

app.use(router)
app.mount('#app')
