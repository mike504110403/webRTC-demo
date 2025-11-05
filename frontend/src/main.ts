import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import Broadcaster from './views/Broadcaster.vue'
import Viewer from './views/Viewer.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/broadcaster' },
    { path: '/broadcaster', component: Broadcaster },
    { path: '/viewer', component: Viewer }
  ]
})

const app = createApp(App)
app.use(router)
app.mount('#app')

