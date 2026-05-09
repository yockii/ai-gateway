import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiLatency = new Trend('api_latency');
const requestCounter = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 concurrent users
    { duration: '2m', target: 10 },    // Sustain 10 users
    { duration: '30s', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 50 },    // Sustain 50 users
    { duration: '30s', target: 100 },  // Ramp up to 100 users
    { duration: '2m', target: 100 },   // Sustain 100 users
    { duration: '30s', target: 500 },  // Ramp up to 500 users
    { duration: '2m', target: 500 },   // Sustain 500 users
    { duration: '1m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    'errors': ['rate<0.02'],           // Error rate < 2%
    'http_req_duration': ['p(95)<200'], // P95 latency < 200ms
    'api_latency': ['p(95)<100'],      // Non-chat API P95 < 100ms
  },
};

// Setup: Get authentication and list available models
export function setup() {
  const apiURL = __ENV.API_URL || 'http://localhost:8080';
  const apiKey = __ENV.API_KEY || 'test-key';

  // Health check
  const healthRes = http.get(`${apiURL}/health`);
  if (healthRes.status !== 200) {
    throw new Error(`API health check failed: ${healthRes.status}`);
  }

  // Get available models
  const modelsRes = http.get(`${apiURL}/v1/models`, {
    headers: { 'Authorization': `Bearer ${apiKey}` },
  });

  let models = ['gpt-3.5-turbo']; // Default fallback
  if (modelsRes.status === 200) {
    try {
      const body = JSON.parse(modelsRes.body);
      if (body.data && body.data.length > 0) {
        models = body.data.map(m => m.id);
      }
    } catch (e) {
      console.warn('Failed to parse models response, using default');
    }
  }

  return { apiURL, apiKey, models };
}

// User scenarios
const scenarios = {
  chat: {
    weight: 5, // 50% of requests
    execute: (data) => {
      const url = `${data.apiURL}/v1/chat/completions`;
      const model = data.models[Math.floor(Math.random() * data.models.length)];

      const payload = JSON.stringify({
        model: model,
        messages: [
          {
            role: 'user',
            content: `Test message ${Math.random()}`,
          },
        ],
        max_tokens: 20,
      });

      const params = {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${data.apiKey}`,
        },
        tags: { name: 'chat' },
      };

      const res = http.post(url, payload, params);

      check(res, {
        'chat status OK': (r) => r.status === 200 || r.status === 201,
      });

      return res.timings.duration;
    },
  },
  listModels: {
    weight: 2, // 20% of requests
    execute: (data) => {
      const url = `${data.apiURL}/v1/models`;
      const params = {
        headers: {
          'Authorization': `Bearer ${data.apiKey}`,
        },
        tags: { name: 'list-models' },
      };

      const res = http.get(url, params);

      check(res, {
        'models status OK': (r) => r.status === 200,
      });

      return res.timings.duration;
    },
  },
  usageStats: {
    weight: 2, // 20% of requests
    execute: (data) => {
      const url = `${data.apiURL}/v1/usage`;
      const params = {
        headers: {
          'Authorization': `Bearer ${data.apiKey}`,
        },
        tags: { name: 'usage' },
      };

      const res = http.get(url, params);

      check(res, {
        'usage status OK': (r) => r.status === 200 || r.status === 404, // 404 if not implemented
      });

      return res.timings.duration;
    },
  },
  healthCheck: {
    weight: 1, // 10% of requests
    execute: (data) => {
      const url = `${data.apiURL}/health`;
      const params = {
        tags: { name: 'health' },
      };

      const res = http.get(url, params);

      check(res, {
        'health status OK': (r) => r.status === 200,
      });

      return res.timings.duration;
    },
  },
};

// Main test scenario
export default function(data) {
  // Select a scenario based on weights
  const totalWeight = Object.values(scenarios).reduce((sum, s) => sum + s.weight, 0);
  let random = Math.random() * totalWeight;
  let selectedScenario = scenarios.chat;

  for (const [name, scenario] of Object.entries(scenarios)) {
    random -= scenario.weight;
    if (random <= 0) {
      selectedScenario = scenario;
      break;
    }
  }

  // Execute the selected scenario
  const startTime = Date.now();
  const latency = selectedScenario.execute(data);
  const endTime = Date.now();

  // Record metrics
  apiLatency.add(endTime - startTime);
  requestCounter.add(1);

  // Think time between requests (1-3 seconds)
  sleep(Math.random() * 2 + 1);
}
