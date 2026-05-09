import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const chatLatency = new Trend('chat_latency');

// Test configuration
export const options = {
  stages: [
    { duration: '1m', target: 100 },   // Ramp up to 100 users
    { duration: '3m', target: 100 },   // Sustain 100 users
    { duration: '1m', target: 500 },   // Ramp up to 500 users
    { duration: '3m', target: 500 },   // Sustain 500 users
    { duration: '1m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    'errors': ['rate<0.01'],           // Error rate < 1%
    'http_req_duration': ['p(95)<500'], // P95 latency < 500ms
    'chat_latency': ['p(95)<500'],     // Chat P95 latency < 500ms
  },
};

// Setup: Get authentication token
export function setup() {
  const apiURL = __ENV.API_URL || 'http://localhost:8080';
  const apiKey = __ENV.API_KEY || 'test-key';

  // Verify API is accessible
  const healthRes = http.get(`${apiURL}/health`, {
    headers: { 'Accept': 'application/json' },
  });

  if (healthRes.status !== 200) {
    throw new Error(`API health check failed: ${healthRes.status}`);
  }

  return { apiURL, apiKey };
}

// Teardown: Clean up resources
export function teardown(data) {
  console.log('Test completed. Final stats:');
  console.log(`  API URL: ${data.apiURL}`);
}

// Main test scenario
export default function(data) {
  const url = `${data.apiURL}/v1/chat/completions`;
  const payload = JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [
      {
        role: 'user',
        content: 'Hello, this is a load test message.',
      },
    ],
    temperature: 0.7,
    max_tokens: 50,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.apiKey}`,
      'Accept': 'application/json',
    },
    tags: { name: 'chat-completions' },
  };

  const startTime = Date.now();
  const res = http.post(url, payload, params);
  const endTime = Date.now();

  // Record custom metrics
  chatLatency.add(endTime - startTime);

  // Validate response
  const success = check(res, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'has valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.object === 'chat.completion' || body.object === 'chat.completion.chunk';
      } catch {
        // Allow non-JSON responses (upstream errors)
        return r.status < 500;
      }
    },
    'response time < 1s': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);

  // Small think time between requests
  sleep(Math.random() * 2 + 1); // 1-3 seconds
}
