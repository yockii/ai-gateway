import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/models'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const user = ref<User | null>(null)
  const loading = ref(false)

  // 从 localStorage 恢复用户信息
  try {
    const savedUser = localStorage.getItem('user_info')
    if (savedUser) {
      user.value = JSON.parse(savedUser)
    }
  } catch {
    localStorage.removeItem('user_info')
  }

  const isAuthenticated = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('access_token', newToken)
  }

  const setUser = (userData: User) => {
    user.value = userData
    localStorage.setItem('user_info', JSON.stringify(userData))
  }

  const login = async (email: string, password: string) => {
    loading.value = true
    try {
      const response = await authApi.login({ email, password }) as any
      setToken(response.token)
      setUser({
        id: response.user_id,
        email: response.email,
        name: response.name,
        user_group_id: '',
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      })
      return response
    } finally {
      loading.value = false
    }
  }

  const register = async (email: string, password: string, name: string) => {
    loading.value = true
    try {
      const response = await authApi.register({ email, password, name }) as any
      // 注册返回的数据不包含 token，需要手动构造
      const userObj = {
        id: response.id,
        email: response.email,
        name: response.name,
        user_group_id: '',
        is_active: response.is_active ?? true,
        created_at: response.created_at ?? new Date().toISOString(),
        updated_at: response.created_at ?? new Date().toISOString(),
      }
      setUser(userObj)
      return response
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    // JWT 是无状态的，只需清除本地 token
    // 不需要等待后端响应，避免因端点不存在而阻塞
    authApi.logout().catch(() => {})
    token.value = null
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('user_info')
  }

  const fetchProfile = async () => {
    const response = await authApi.profile() as any
    setUser({
      id: response.id,
      email: response.email,
      name: user.value?.name ?? 'User',
      user_group_id: '',
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    })
  }

  return {
    token,
    user,
    loading,
    isAuthenticated,
    setToken,
    setUser,
    login,
    register,
    logout,
    fetchProfile,
  }
})
