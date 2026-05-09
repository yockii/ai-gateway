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
  (response: AxiosResponse) => response.data,
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

export default client
