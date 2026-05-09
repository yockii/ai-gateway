import { http, HttpResponse } from 'msw'
import type { AdminUser, User, ExternalModel, Supplier, SystemMetrics, Alert } from '@/types/models'

// Mock admin login response
const mockAdminUser: AdminUser = {
  id: 'admin-1',
  email: 'admin@example.com',
  name: 'Test Admin',
  role: 'superadmin',
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

// Mock users list
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

// Mock models list
const mockModels: ExternalModel[] = [
  {
    id: 'model-1',
    name: 'gpt-4',
    display_name: 'GPT-4',
    model_type: 'chat',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'model-2',
    name: 'claude-3-opus',
    display_name: 'Claude 3 Opus',
    model_type: 'chat',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

// Mock suppliers list
const mockSuppliers: Supplier[] = [
  {
    id: 'supplier-1',
    name: 'openai',
    display_name: 'OpenAI',
    provider: 'openai',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'supplier-2',
    name: 'anthropic',
    display_name: 'Anthropic',
    provider: 'anthropic',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

// Mock system metrics
const mockMetrics: SystemMetrics = {
  qps: 1250,
  avg_latency: 45,
  error_rate: 0.02,
  active_users: 42,
  total_requests: 150000,
  cpu_usage: 65,
  memory_usage: 72,
}

// Mock alerts
const mockAlerts: Alert[] = [
  {
    id: 'alert-1',
    type: 'error',
    message: 'High error rate detected on OpenAI provider',
    created_at: '2024-01-01T12:00:00Z',
    resolved: false,
  },
  {
    id: 'alert-2',
    type: 'warning',
    message: 'API quota usage at 80%',
    created_at: '2024-01-01T10:00:00Z',
    resolved: false,
  },
]

export const handlers = [
  // Admin auth endpoints
  http.post('/v1/admin/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string }

    if (body.email === 'admin@example.com' && body.password === 'password123') {
      return HttpResponse.json({
        admin: mockAdminUser,
        token: 'mock-jwt-token-12345',
      })
    }

    return HttpResponse.json(
      { message: 'Invalid credentials' },
      { status: 401 }
    )
  }),

  http.post('/v1/admin/logout', () => {
    return HttpResponse.json({ message: 'Logged out successfully' })
  }),

  http.get('/v1/admin/profile', () => {
    return HttpResponse.json(mockAdminUser)
  }),

  // Users endpoints
  http.get('/v1/admin/users', ({ request }) => {
    const url = new URL(request.url)
    const search = url.searchParams.get('search')

    let filteredUsers = mockUsers
    if (search) {
      filteredUsers = mockUsers.filter(
        (user) =>
          user.email.includes(search) || user.name.includes(search)
      )
    }

    return HttpResponse.json({
      data: filteredUsers,
      total: filteredUsers.length,
    })
  }),

  http.get('/v1/admin/users/:id', ({ params }) => {
    const user = mockUsers.find((u) => u.id === params.id)
    if (!user) {
      return HttpResponse.json(
        { message: 'User not found' },
        { status: 404 }
      )
    }
    return HttpResponse.json(user)
  }),

  http.delete('/v1/admin/users/:id', ({ params }) => {
    const user = mockUsers.find((u) => u.id === params.id)
    if (!user) {
      return HttpResponse.json(
        { message: 'User not found' },
        { status: 404 }
      )
    }
    return HttpResponse.json({ message: 'User deleted successfully' })
  }),

  http.patch('/v1/admin/users/:id/status', ({ params }) => {
    const user = mockUsers.find((u) => u.id === params.id)
    if (!user) {
      return HttpResponse.json(
        { message: 'User not found' },
        { status: 404 }
      )
    }
    return HttpResponse.json({
      ...user,
      is_active: !user.is_active,
    })
  }),

  // Models endpoints
  http.get('/v1/admin/models', () => {
    return HttpResponse.json({
      data: mockModels,
      total: mockModels.length,
    })
  }),

  // Suppliers endpoints
  http.get('/v1/admin/suppliers', () => {
    return HttpResponse.json({
      data: mockSuppliers,
      total: mockSuppliers.length,
    })
  }),

  // Monitoring endpoints
  http.get('/v1/monitoring/metrics', () => {
    return HttpResponse.json(mockMetrics)
  }),

  http.get('/v1/monitoring/alerts', () => {
    return HttpResponse.json({
      data: mockAlerts,
      total: mockAlerts.length,
    })
  }),
]
