import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import App from './App.vue'

// Import routes
import ApplicationForm from './views/ApplicationForm.vue'
import AdminPanel from './views/AdminPanel.vue'
import MobileAdminPanel from './views/MobileAdminPanel.vue'
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
    { path: '/admin/mobile', component: MobileAdminPanel },
    { path: '/thank-you', component: ThankYou },
  ],
})

// Navigation guard for mobile redirection
router.beforeEach((to, from, next) => {
  if (to.path === '/admin') {
    // Check if the device is mobile-like (width <= 768px)
    const isMobile = window.innerWidth <= 768;
    if (isMobile) {
      next('/admin/mobile');
    } else {
      next();
    }
  } else {
    next();
  }
});

// Create and mount app
const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app') 