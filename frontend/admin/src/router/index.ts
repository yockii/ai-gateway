import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  Users,
  Cpu,
  Truck,
  CreditCard,
  Activity,
  BarChart3,
  FileText,
  DollarSign,
} from 'lucide-vue-next'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false, title: 'Login' },
    },
    {
      path: '/',
      component: () => import('@/components/layout/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/users' },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/views/Users.vue'),
          meta: { requiresAuth: true, title: 'User Management', icon: Users },
        },
        {
          path: 'models',
          name: 'Models',
          component: () => import('@/views/Models.vue'),
          meta: { requiresAuth: true, title: 'Model Management', icon: Cpu },
        },
        {
          path: 'suppliers',
          name: 'Suppliers',
          component: () => import('@/views/SupplierManagement.vue'),
          meta: { requiresAuth: true, title: 'Supplier Management', icon: Truck },
        },
        {
          path: 'suppliers/:id',
          name: 'SupplierDetail',
          component: () => import('@/views/SupplierManagement.vue'),
          meta: { requiresAuth: true, title: 'Supplier Details' },
        },
        {
          path: 'plans',
          name: 'Plans',
          component: () => import('@/views/Plans.vue'),
          meta: { requiresAuth: true, title: 'Plan Management', icon: CreditCard },
        },
        {
          path: 'pricing',
          name: 'Pricing',
          component: () => import('@/views/Pricing.vue'),
          meta: { requiresAuth: true, title: 'Pricing Management', icon: DollarSign },
        },
        {
          path: 'audit',
          name: 'AuditLogs',
          component: () => import('@/views/AuditLogs.vue'),
          meta: { requiresAuth: true, title: 'Audit Logs', icon: FileText },
        },
        {
          path: 'monitoring',
          name: 'Monitoring',
          component: () => import('@/views/Monitoring.vue'),
          meta: { requiresAuth: true, title: 'System Monitoring', icon: Activity },
        },
        {
          path: 'operations',
          name: 'Operations',
          component: () => import('@/views/Operations.vue'),
          meta: { requiresAuth: true, title: 'Operations Dashboard', icon: BarChart3 },
        },
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
    // Update document title
    if (to.meta.title) {
      document.title = `${to.meta.title} - Admin Portal`
    }
    next()
  }
})

export default router
