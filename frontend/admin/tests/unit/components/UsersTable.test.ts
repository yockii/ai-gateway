import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Users from '@/views/Users.vue'
import * as usersApi from '@/api/users'
import type { User } from '@/types/models'

// Mock the users API
vi.mock('@/api/users', () => ({
  usersApi: {
    list: vi.fn(),
    delete: vi.fn(),
    toggleStatus: vi.fn(),
  },
}))

describe('Users Component', () => {
  let pinia: ReturnType<typeof createPinia>

  const mockUsers: User[] = [
    {
      id: 'user-1',
      email: 'user1@example.com',
      name: 'User One',
      user_group_id: 'group-1',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'user-2',
      email: 'user2@example.com',
      name: 'User Two',
      user_group_id: 'group-2',
      is_active: false,
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ]

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()

    // Mock successful API response
    vi.mocked(usersApi.usersApi.list).mockResolvedValue({
      data: mockUsers,
      total: 2,
    })
  })

  it('renders user list', async () => {
    const wrapper = mount(Users, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    // Wait for component to mount and fetch data
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(usersApi.usersApi.list).toHaveBeenCalled()
  })

  it('shows loading state', async () => {
    const wrapper = mount(Users, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    // Set loading state
    await wrapper.setData({ loading: true })
    await wrapper.vm.$nextTick()

    const loadingText = wrapper.find('.text-center')
    expect(loadingText.exists()).toBe(true)
  })

  it('handles empty state', async () => {
    vi.mocked(usersApi.usersApi.list).mockResolvedValue({
      data: [],
      total: 0,
    })

    const wrapper = mount(Users, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    // Should render table even if empty
    const table = wrapper.find('table')
    expect(table.exists()).toBe(true)
  })

  it('formats dates correctly', async () => {
    const wrapper = mount(Users, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    await wrapper.vm.$nextTick()

    // Check if formatDate function exists and works
    const testDate = '2024-01-01T00:00:00Z'
    const formatted = wrapper.vm.formatDate(testDate)

    expect(formatted).toContain('2024')
  })

  it('has search functionality', async () => {
    const wrapper = mount(Users, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    const searchInput = wrapper.find('input[type="text"]')
    expect(searchInput.exists()).toBe(true)

    await searchInput.setValue('test@example.com')
    expect(wrapper.vm.searchQuery).toBe('test@example.com')
  })
})
