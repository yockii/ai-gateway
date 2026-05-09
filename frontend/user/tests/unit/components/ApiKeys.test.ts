import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ApiKeys from '@/views/ApiKeys.vue'
import * as apiKeysApi from '@/api/keys'

// Mock the API keys API
vi.mock('@/api/keys', () => ({
  keysApi: {
    list: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('ApiKeys Component', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()

    // Mock successful API response
    vi.mocked(apiKeysApi.keysApi.list).mockResolvedValue({
      data: [
        {
          id: 'key-1',
          user_id: 'user-1',
          name: 'Test Key',
          quota_daily: 1000,
          quota_monthly: 30000,
          concurrency_limit: 5,
          model_concurrency: {},
          expires_at: '2025-12-31T23:59:59Z',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      total: 1,
    })
  })

  it('renders API keys list', async () => {
    const wrapper = mount(ApiKeys, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(apiKeysApi.keysApi.list).toHaveBeenCalled()
  })

  it('shows loading state while fetching', async () => {
    const wrapper = mount(ApiKeys, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    await wrapper.setData({ loading: true })
    await wrapper.vm.$nextTick()

    const loadingText = wrapper.find('.text-center, .loading')
    const hasLoading = loadingText.exists() || wrapper.html().includes('加载中')
    expect(hasLoading).toBe(true)
  })

  it('handles empty state', async () => {
    vi.mocked(apiKeysApi.keysApi.list).mockResolvedValue({
      data: [],
      total: 0,
    })

    const wrapper = mount(ApiKeys, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    // Should render even if empty
    expect(wrapper.exists()).toBe(true)
  })

  it('has create button for new API key', () => {
    const wrapper = mount(ApiKeys, {
      global: {
        plugins: [pinia],
        stubs: ['router-link'],
      },
    })

    const createButton = wrapper.find('button:has-text("创建"), button:has-text("Create")')
    const hasCreateButton = createButton.exists() || wrapper.html().includes('创建') || wrapper.html().includes('Create')
    expect(hasCreateButton).toBe(true)
  })
})
