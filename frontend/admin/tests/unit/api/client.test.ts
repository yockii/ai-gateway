import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import axios from 'axios'
import client from '@/api/client'

// Mock axios
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      patch: vi.fn(),
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })),
  },
}))

describe('API Client', () => {
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Mock axios.create to return our instance
    vi.mocked(axios.create).mockReturnValue(mockAxiosInstance as any)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('creates axios instance with correct baseURL', () => {
    expect(axios.create).toHaveBeenCalledWith(
      expect.objectContaining({
        baseURL: '/v1/admin',
        timeout: 30000,
      })
    )
  })

  it('sets up request interceptor for auth token', () => {
    const requestUseSpy = mockAxiosInstance.interceptors.request.use

    expect(requestUseSpy).toHaveBeenCalled()
    expect(requestUseSpy).toHaveBeenCalledWith(
      expect.any(Function),
      expect.any(Function)
    )
  })

  it('sets up response interceptor for error handling', () => {
    const responseUseSpy = mockAxiosInstance.interceptors.response.use

    expect(responseUseSpy).toHaveBeenCalled()
    expect(responseUseSpy).toHaveBeenCalledWith(
      expect.any(Function),
      expect.any(Function)
    )
  })

  it('client.get method exists and calls axios get', async () => {
    mockAxiosInstance.get.mockResolvedValue({ data: { result: 'success' } })

    const result = await client.get('/test')

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/test', undefined)
    expect(result).toEqual({ result: 'success' })
  })

  it('client.post method exists and calls axios post', async () => {
    const postData = { email: 'test@example.com', password: 'password' }
    mockAxiosInstance.post.mockResolvedValue({ data: { token: 'abc123' } })

    const result = await client.post('/login', postData)

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/login', postData, undefined)
    expect(result).toEqual({ token: 'abc123' })
  })

  it('client.put method exists and calls axios put', async () => {
    const putData = { name: 'Updated Name' }
    mockAxiosInstance.put.mockResolvedValue({ data: { success: true } })

    const result = await client.put('/users/1', putData)

    expect(mockAxiosInstance.put).toHaveBeenCalledWith('/users/1', putData, undefined)
    expect(result).toEqual({ success: true })
  })

  it('client.delete method exists and calls axios delete', async () => {
    mockAxiosInstance.delete.mockResolvedValue({ data: { success: true } })

    const result = await client.delete('/users/1')

    expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/users/1', undefined)
    expect(result).toEqual({ success: true })
  })

  it('client.patch method exists and calls axios patch', async () => {
    const patchData = { is_active: false }
    mockAxiosInstance.patch.mockResolvedValue({ data: { success: true } })

    const result = await client.patch('/users/1/status', patchData)

    expect(mockAxiosInstance.patch).toHaveBeenCalledWith('/users/1/status', patchData, undefined)
    expect(result).toEqual({ success: true })
  })
})
