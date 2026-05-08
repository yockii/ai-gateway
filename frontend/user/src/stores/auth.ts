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
      setUser(response.user)
      return response
    } finally {
      loading.value = false
    }
  }

  const register = async (email: string, password: string, name: string) => {
    loading.value = true
    try {
      const response = await authApi.register({ email, password, name }) as any
      setToken(response.token)
      setUser(response.user)
      return response
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } finally {
      token.value = null
      user.value = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
    }
  }

  const fetchProfile = async () => {
    const userData = await authApi.profile()
    setUser(userData)
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
