import { http, HttpResponse } from 'msw'
import type { User, UserAPIKey, UsageRecord, Bill, MembershipPlan, DashboardStats } from '@/types/models'

// Mock user
const mockUser: User = {
  id: 'user-1',
  email: 'user@example.com',
  name: 'Test User',
  user_group_id: 'group-1',
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

// Mock API keys
const mockApiKeys: UserAPIKey[] = [
  {
    id: 'key-1',
    user_id: 'user-1',
    name: 'Development Key',
    quota_daily: 1000,
    quota_monthly: 30000,
    concurrency_limit: 5,
    model_concurrency: {},
    expires_at: '2025-12-31T23:59:59Z',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'key-2',
    user_id: 'user-1',
    name: 'Production Key',
    quota_daily: 10000,
    quota_monthly: 300000,
    concurrency_limit: 10,
    model_concurrency: {},
    expires_at: '2025-12-31T23:59:59Z',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

// Mock usage records
const mockUsageRecords: UsageRecord[] = [
  {
    id: 'usage-1',
    request_id: 'req-1',
    user_id: 'user-1',
    model_id: 'gpt-4',
    supplier_id: 'openai',
    input_tokens: 100,
    output_tokens: 50,
    total_tokens: 150,
    cost_price: 0.002,
    selling_price: 0.003,
    profit: 0.001,
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 'usage-2',
    request_id: 'req-2',
    user_id: 'user-1',
    model_id: 'claude-3-opus',
    supplier_id: 'anthropic',
    input_tokens: 200,
    output_tokens: 100,
    total_tokens: 300,
    cost_price: 0.005,
    selling_price: 0.008,
    profit: 0.003,
    created_at: '2024-01-01T01:00:00Z',
  },
]

// Mock bills
const mockBills: Bill[] = [
  {
    id: 'bill-1',
    user_id: 'user-1',
    period: '2024-01',
    start_date: '2024-01-01T00:00:00Z',
    end_date: '2024-01-31T23:59:59Z',
    items: [],
    total_requests: 1500,
    total_cost: 5.50,
    total_revenue: 8.25,
    total_profit: 2.75,
    status: 'paid',
    created_at: '2024-02-01T00:00:00Z',
    updated_at: '2024-02-01T00:00:00Z',
  },
]

// Mock membership plans
const mockPlans: MembershipPlan[] = [
  {
    id: 'plan-free',
    name: 'Free',
    description: 'Basic access with limited requests',
    price_monthly: 0,
    price_yearly: 0,
    features: ['100 requests/day', 'Standard models', 'Community support'],
    is_active: true,
  },
  {
    id: 'plan-pro',
    name: 'Pro',
    description: 'Enhanced access for professionals',
    price_monthly: 29,
    price_yearly: 290,
    features: ['10000 requests/day', 'All models', 'Priority support', 'API access'],
    is_active: true,
  },
]

// Mock dashboard stats
const mockDashboardStats: DashboardStats = {
  total_keys: 2,
  monthly_requests: 1500,
  monthly_cost: 5.50,
  monthly_tokens: 45000,
}

export const handlers = [
  // User auth endpoints
  http.post('/v1/auth/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string }

    if (body.email === 'user@example.com' && body.password === 'password123') {
      return HttpResponse.json({
        user: mockUser,
        token: 'mock-user-jwt-token-12345',
      })
    }

    return HttpResponse.json(
      { message: 'Invalid credentials' },
      { status: 401 }
    )
  }),

  http.post('/v1/auth/register', async ({ request }) => {
    const body = await request.json() as { email: string; password: string; name: string }

    if (body.email && body.password && body.name) {
      return HttpResponse.json({
        user: {
          ...mockUser,
          email: body.email,
          name: body.name,
        },
        token: 'mock-user-jwt-token-67890',
      })
    }

    return HttpResponse.json(
      { message: 'Registration failed' },
      { status: 400 }
    )
  }),

  http.post('/v1/auth/logout', () => {
    return HttpResponse.json({ message: 'Logged out successfully' })
  }),

  http.get('/v1/auth/profile', () => {
    return HttpResponse.json(mockUser)
  }),

  // API Keys endpoints
  http.get('/v1/keys', () => {
    return HttpResponse.json({
      data: mockApiKeys,
      total: mockApiKeys.length,
    })
  }),

  http.post('/v1/keys', async ({ request }) => {
    const body = await request.json() as { name: string; quota_daily?: number }

    const newKey: UserAPIKey = {
      id: `key-${Date.now()}`,
      user_id: 'user-1',
      name: body.name,
      quota_daily: body.quota_daily || 1000,
      quota_monthly: (body.quota_daily || 1000) * 30,
      concurrency_limit: 5,
      model_concurrency: {},
      expires_at: '2025-12-31T23:59:59Z',
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }

    return HttpResponse.json({
      key: `sk-${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`,
      data: newKey,
    })
  }),

  http.delete('/v1/keys/:id', () => {
    return HttpResponse.json({ message: 'API key deleted successfully' })
  }),

  // Usage endpoints
  http.get('/v1/usage', () => {
    return HttpResponse.json({
      data: mockUsageRecords,
      total: mockUsageRecords.length,
    })
  }),

  // Bills endpoints
  http.get('/v1/bills', () => {
    return HttpResponse.json({
      data: mockBills,
      total: mockBills.length,
    })
  }),

  http.get('/v1/bills/:id/export', () => {
    return HttpResponse.json({
      download_url: '/downloads/bill-2024-01.pdf',
    })
  }),

  // Plans endpoints
  http.get('/v1/plans', () => {
    return HttpResponse.json({
      data: mockPlans,
      total: mockPlans.length,
    })
  }),

  // Dashboard endpoints
  http.get('/v1/dashboard/stats', () => {
    return HttpResponse.json(mockDashboardStats)
  }),
]
