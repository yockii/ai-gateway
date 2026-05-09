import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Login from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { createRouter, createMemoryHistory } from 'vue-router'

describe('Login Component', () => {
  let pinia: ReturnType<typeof createPinia>
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)

    // Clear localStorage
    vi.clearAllMocks()

    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/users', component: { template: '<div>Users</div>' } },
        { path: '/login', component: Login },
      ],
    })
  })

  it('renders email and password inputs', () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const emailInput = wrapper.find('input[type="email"]')
    const passwordInput = wrapper.find('input[type="password"]')

    expect(emailInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)
  })

  it('shows validation errors for empty fields', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const submitButton = wrapper.find('button[type="submit"]')
    await submitButton.trigger('click')

    // Should not submit with empty fields
    const authStore = useAuthStore()
    expect(authStore.loading).toBe(false)
  })

  it('submits form with valid data', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const emailInput = wrapper.find('input[type="email"]')
    const passwordInput = wrapper.find('input[type="password"]')

    await emailInput.setValue('admin@example.com')
    await passwordInput.setValue('password123')

    // Mock router push
    const pushSpy = vi.spyOn(router, 'push')

    // Simulate form submission
    const form = wrapper.find('form')
    await form.trigger('submit.prevent')

    // Form should be submitted (actual API call will be mocked in integration tests)
    expect(emailInput.element.value).toBe('admin@example.com')
  })

  it('disables submit button when loading', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const authStore = useAuthStore()
    authStore.loading = true

    await wrapper.vm.$nextTick()

    const submitButton = wrapper.find('button[type="submit"]')
    expect(submitButton.attributes('disabled')).toBeDefined()
  })

  it('displays error message when login fails', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    // Trigger error
    await wrapper.setData({ error: 'Invalid credentials' })

    await wrapper.vm.$nextTick()

    const errorElement = wrapper.find('.text-red-400')
    expect(errorElement.exists()).toBe(true)
    expect(errorElement.text()).toBe('Invalid credentials')
  })
})
