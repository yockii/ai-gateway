import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/components/layout/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/users' },
        { path: 'users', name: 'Users', component: () => import('@/views/Users.vue') },
        { path: 'models', name: 'Models', component: () => import('@/views/Models.vue') },
        { path: 'suppliers', name: 'Suppliers', component: () => import('@/views/Suppliers.vue') },
        { path: 'plans', name: 'Plans', component: () => import('@/views/Plans.vue') },
        { path: 'monitoring', name: 'Monitoring', component: () => import('@/views/Monitoring.vue') },
        { path: 'operations', name: 'Operations', component: () => import('@/views/Operations.vue') },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ name: 'Users' })
  } else {
    next()
  }
})

export default router
