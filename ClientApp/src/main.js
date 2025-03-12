import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import App from './App.vue'

// Import routes
import ApplicationForm from './views/ApplicationForm.vue'
import AdminPanel from './views/AdminPanel.vue'
import ThankYou from './views/ThankYou.vue'

// Import CSS
import './assets/main.css'

// Create Pinia (state management)
const pinia = createPinia()

// Create router
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: ApplicationForm },
    { path: '/admin', component: AdminPanel },
    { path: '/thank-you', component: ThankYou },
  ],
})

// Create and mount app
const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app') 