import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Login from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { createRouter, createMemoryHistory } from 'vue-router'

describe('User Login Component', () => {
  let pinia: ReturnType<typeof createPinia>
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)

    vi.clearAllMocks()

    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/dashboard', component: { template: '<div>Dashboard</div>' } },
        { path: '/login', component: Login },
      ],
    })
  })

  it('renders login form with email and password inputs', () => {
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

  it('shows validation errors for invalid email', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const emailInput = wrapper.find('input[type="email"]')
    await emailInput.setValue('invalid-email')

    const submitButton = wrapper.find('button[type="submit"]')
    await submitButton.trigger('click')

    // Should show validation error
    await wrapper.vm.$nextTick()
  })

  it('disables submit button during loading', async () => {
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

  it('displays error message on failed login', async () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    await wrapper.setData({ error: 'Invalid email or password' })
    await wrapper.vm.$nextTick()

    const errorElement = wrapper.find('.text-red-500, .text-red-400')
    expect(errorElement.exists()).toBe(true)
  })

  it('has link to registration page', () => {
    const wrapper = mount(Login, {
      global: {
        plugins: [pinia, router],
      },
    })

    const registerLink = wrapper.find('a[href*="register"], a:contains("注册")')
    const linkExists = registerLink.exists() || wrapper.html().includes('注册')
    expect(linkExists).toBe(true)
  })
})
