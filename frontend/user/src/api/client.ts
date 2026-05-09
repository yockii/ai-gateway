import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import type { ApiError } from '@/types/api'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/v1'

const rawClient = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

rawClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

rawClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data
  },
  (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
      window.location.href = '/login'
    }
    return Promise.reject(error.response?.data || { message: error.message })
  }
)

// Export a typed client that returns response.data directly
const client = {
  get: <T>(url: string, config?: any) => rawClient.get<any, T>(url, config),
  post: <T>(url: string, data?: any, config?: any) => rawClient.post<any, T>(url, data, config),
  put: <T>(url: string, data?: any, config?: any) => rawClient.put<any, T>(url, data, config),
  delete: <T>(url: string, config?: any) => rawClient.delete<any, T>(url, config),
  patch: <T>(url: string, data?: any, config?: any) => rawClient.patch<any, T>(url, data, config),
}

export default client
