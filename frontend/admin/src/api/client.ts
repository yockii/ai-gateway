import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/v1/admin'

const rawClient = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

rawClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('admin_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

rawClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // 如果响应包含 data 字段，自动提取
    const data = response.data as any
    if (data && typeof data === 'object' && 'data' in data) {
      return data.data
    }
    return response.data
  },
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      window.location.href = '/login'
    }
    return Promise.reject(error.response?.data || { message: error.message })
  }
)

const client = {
  get: <T>(url: string, config?: any) => rawClient.get<any, T>(url, config),
  post: <T>(url: string, data?: any, config?: any) => rawClient.post<any, T>(url, data, config),
  put: <T>(url: string, data?: any, config?: any) => rawClient.put<any, T>(url, data, config),
  delete: <T>(url: string, config?: any) => rawClient.delete<any, T>(url, config),
  patch: <T>(url: string, data?: any, config?: any) => rawClient.patch<any, T>(url, data, config),
}

// 公开 API 客户端（用于登录等不需要认证的操作）
const publicClient = axios.create({
  baseURL: '/v1/public',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

publicClient.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  (error: AxiosError) => {
    return Promise.reject(error.response?.data || { message: error.message })
  }
)

export { publicClient }
export default client
