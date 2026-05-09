import { vi, beforeEach, afterEach } from 'vitest'
import { setupServer } from 'msw/node'
import { HttpResponse } from 'msw'
import { handlers } from './mocks/handlers'

// Setup MSW server
export const server = setupServer(...handlers)

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}

Object.defineProperty(global, 'localStorage', {
  value: localStorageMock,
})

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    href: '',
    assign: vi.fn(),
    reload: vi.fn(),
    replace: vi.fn(),
  },
  writable: true,
})

// Setup MSW server before all tests
beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' })
})

// Reset handlers after each test
afterEach(() => {
  server.resetHandlers()
  // Clear localStorage mocks
  vi.clearAllMocks()
})

// Close server after all tests
afterAll(() => {
  server.close()
})
