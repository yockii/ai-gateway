import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import * as authApi from '@/api/auth'

// Mock the auth API
vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    logout: vi.fn(),
    profile: vi.fn(),
  },
}))

describe('Auth Store', () => {
  let authStore: ReturnType<typeof useAuthStore>

  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
    authStore = useAuthStore()

    // Clear localStorage
    vi.clearAllMocks()
    localStorage.clear()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('has correct initial state', () => {
    expect(authStore.token).toBeNull()
    expect(authStore.admin).toBeNull()
    expect(authStore.loading).toBe(false)
    expect(authStore.isAuthenticated).toBe(false)
  })

  it('login action sets token and admin', async () => {
    const mockResponse = {
      admin: {
        id: 'admin-1',
        email: 'admin@example.com',
        name: 'Test Admin',
        role: 'superadmin' as const,
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      token: 'test-token-123',
    }

    vi.mocked(authApi.authApi.login).mockResolvedValue(mockResponse)

    await authStore.login('admin@example.com', 'password')

    expect(authStore.token).toBe('test-token-123')
    expect(authStore.admin).toEqual(mockResponse.admin)
    expect(authStore.isAuthenticated).toBe(true)
    expect(authApi.authApi.login).toHaveBeenCalledWith({
      email: 'admin@example.com',
      password: 'password',
    })
  })

  it('logout action clears state', async () => {
    // Set initial state
    authStore.setToken('test-token')
    authStore.setAdmin({
      id: 'admin-1',
      email: 'admin@example.com',
      name: 'Test Admin',
      role: 'superadmin',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })

    vi.mocked(authApi.authApi.logout).mockResolvedValue(undefined)

    await authStore.logout()

    expect(authStore.token).toBeNull()
    expect(authStore.admin).toBeNull()
    expect(authStore.isAuthenticated).toBe(false)
    expect(authApi.authApi.logout).toHaveBeenCalled()
  })

  it('setToken updates localStorage', () => {
    const setItemSpy = vi.spyOn(Storage.prototype, 'setItem')

    authStore.setToken('new-token')

    expect(authStore.token).toBe('new-token')
    expect(setItemSpy).toHaveBeenCalledWith('admin_token', 'new-token')
  })

  it('setAdmin updates localStorage', () => {
    const mockAdmin = {
      id: 'admin-1',
      email: 'admin@example.com',
      name: 'Test Admin',
      role: 'superadmin' as const,
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    const setItemSpy = vi.spyOn(Storage.prototype, 'setItem')

    authStore.setAdmin(mockAdmin)

    expect(authStore.admin).toEqual(mockAdmin)
    expect(setItemSpy).toHaveBeenCalledWith('admin_info', JSON.stringify(mockAdmin))
  })

  it('loading state updates during login', async () => {
    vi.mocked(authApi.authApi.login).mockImplementation(
      () => new Promise((resolve) => {
        setTimeout(() => {
          resolve({
            admin: {
              id: 'admin-1',
              email: 'admin@example.com',
              name: 'Test Admin',
              role: 'superadmin' as const,
              is_active: true,
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          })
        }, 100)
      })
    )

    const loginPromise = authStore.login('admin@example.com', 'password')

    expect(authStore.loading).toBe(true)

    await loginPromise

    expect(authStore.loading).toBe(false)
  })
})
