import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../components/LoginView.vue'
import RegisterView from '../components/RegisterView.vue'
import MapView from '../components/MapView.vue'

const routes = [
  { path: '/login', component: LoginView, meta: { title: 'Авторизация' } },
  { path: '/register', component: RegisterView, meta: { title: 'Регистрация' } },
  { path: '/', component: MapView, meta: { requiresAuth: true, title: 'Карта' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else {
    next()
  }
  if (to.meta.title) {
    document.title = to.meta.title
  }
})


export default router
