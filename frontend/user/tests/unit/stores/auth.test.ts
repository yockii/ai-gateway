import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import * as authApi from '@/api/auth'

// Mock the auth API
vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    profile: vi.fn(),
  },
}))

describe('User Auth Store', () => {
  let authStore: ReturnType<typeof useAuthStore>

  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
    authStore = useAuthStore()

    vi.clearAllMocks()
    localStorage.clear()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('has correct initial state', () => {
    expect(authStore.token).toBeNull()
    expect(authStore.user).toBeNull()
    expect(authStore.loading).toBe(false)
    expect(authStore.isAuthenticated).toBe(false)
  })

  it('login action sets token and user', async () => {
    const mockResponse = {
      user: {
        id: 'user-1',
        email: 'user@example.com',
        name: 'Test User',
        user_group_id: 'group-1',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      token: 'test-user-token-123',
    }

    vi.mocked(authApi.authApi.login).mockResolvedValue(mockResponse)

    await authStore.login('user@example.com', 'password')

    expect(authStore.token).toBe('test-user-token-123')
    expect(authStore.user).toEqual(mockResponse.user)
    expect(authStore.isAuthenticated).toBe(true)
    expect(authApi.authApi.login).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'password',
    })
  })

  it('register action creates new user', async () => {
    const mockResponse = {
      user: {
        id: 'user-1',
        email: 'newuser@example.com',
        name: 'New User',
        user_group_id: 'group-1',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      token: 'test-user-token-456',
    }

    vi.mocked(authApi.authApi.register).mockResolvedValue(mockResponse)

    await authStore.register('newuser@example.com', 'password', 'New User')

    expect(authStore.token).toBe('test-user-token-456')
    expect(authStore.user).toEqual(mockResponse.user)
    expect(authStore.isAuthenticated).toBe(true)
    expect(authApi.authApi.register).toHaveBeenCalledWith({
      email: 'newuser@example.com',
      password: 'password',
      name: 'New User',
    })
  })

  it('logout action clears state', async () => {
    // Set initial state
    authStore.setToken('test-token')
    authStore.setUser({
      id: 'user-1',
      email: 'user@example.com',
      name: 'Test User',
      user_group_id: 'group-1',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })

    vi.mocked(authApi.authApi.logout).mockResolvedValue(undefined)

    await authStore.logout()

    expect(authStore.token).toBeNull()
    expect(authStore.user).toBeNull()
    expect(authStore.isAuthenticated).toBe(false)
    expect(authApi.authApi.logout).toHaveBeenCalled()
  })

  it('fetchProfile updates user data', async () => {
    const mockUser = {
      id: 'user-1',
      email: 'user@example.com',
      name: 'Updated Name',
      user_group_id: 'group-1',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(authApi.authApi.profile).mockResolvedValue(mockUser)

    await authStore.fetchProfile()

    expect(authStore.user).toEqual(mockUser)
    expect(authApi.authApi.profile).toHaveBeenCalled()
  })

  it('loading state updates during async operations', async () => {
    vi.mocked(authApi.authApi.login).mockImplementation(
      () => new Promise((resolve) => {
        setTimeout(() => {
          resolve({
            user: {
              id: 'user-1',
              email: 'user@example.com',
              name: 'Test User',
              user_group_id: 'group-1',
              is_active: true,
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          })
        }, 100)
      })
    )

    const loginPromise = authStore.login('user@example.com', 'password')

    expect(authStore.loading).toBe(true)

    await loginPromise

    expect(authStore.loading).toBe(false)
  })
})
