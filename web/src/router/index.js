import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', redirect: '/dashboard' },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/links',
    name: 'LinksManage',
    component: () => import('../views/LinksManage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/links/create',
    name: 'LinkCreate',
    component: () => import('../views/LinkCreate.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/links/:code',
    name: 'LinkDetail',
    component: () => import('../views/LinkDetail.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.guest && auth.isLoggedIn) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
