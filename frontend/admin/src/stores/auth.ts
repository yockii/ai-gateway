import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminUser } from '@/types/models'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const admin = ref<AdminUser | null>(null)
  const loading = ref(false)

  try {
    const saved = localStorage.getItem('admin_info')
    if (saved) admin.value = JSON.parse(saved)
  } catch {
    localStorage.removeItem('admin_info')
  }

  const isAuthenticated = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('admin_token', newToken)
  }

  const setAdmin = (data: AdminUser) => {
    admin.value = data
    localStorage.setItem('admin_info', JSON.stringify(data))
  }

  const login = async (email: string, password: string) => {
    loading.value = true
    try {
      const response = await authApi.login({ email, password }) as any
      setToken(response.token)
      // 后端直接返回 admin_id, email, name 字段
      setAdmin({
        id: response.admin_id,
        email: response.email,
        name: response.name,
      })
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } finally {
      token.value = null
      admin.value = null
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
    }
  }

  return { token, admin, loading, isAuthenticated, setToken, setAdmin, login, logout }
})
