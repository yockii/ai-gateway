import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const qpsLatency = new Trend('qps_latency');

// Test configuration for maximum QPS
export const options = {
  scenarios: {
    constant_load: {
      executor: 'constant-arrival-rate',
      rate: 1000,            // Start at 1000 iterations/second
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      exec: 'loadTest',
    },
    ramping_load: {
      executor: 'ramping-arrival-rate',
      startRate: 1000,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 1000,
      stages: [
        { duration: '30s', target: 3000 },   // Ramp to 3000 RPS
        { duration: '30s', target: 5000 },   // Ramp to 5000 RPS
        { duration: '30s', target: 7000 },   // Ramp to 7000 RPS
        { duration: '30s', target: 10000 },  // Ramp to 10000 RPS
        { duration: '30s', target: 12000 },  // Ramp to 12000 RPS
        { duration: '30s', target: 15000 },  // Ramp to 15000 RPS (breaking point)
        { duration: '30s', target: 1000 },   // Ramp down
      ],
      exec: 'loadTest',
    },
  },
  thresholds: {
    'errors': ['rate<0.05'],           // Allow up to 5% errors at max load
    'http_req_duration': ['p(95)<100'], // P95 latency < 100ms
    'qps_latency': ['p(95)<50'],      // Target P95 < 50ms
  },
};

// Setup: Verify API is accessible
export function setup() {
  const apiURL = __ENV.API_URL || 'http://localhost:8080';
  const apiKey = __ENV.API_KEY || 'test-key';

  console.log(`Starting max QPS test against ${apiURL}`);
  console.log('WARNING: This test is designed to find breaking points');
  console.log('DO NOT run against production servers');

  // Health check
  const healthRes = http.get(`${apiURL}/health`);
  if (healthRes.status !== 200) {
    throw new Error(`API health check failed: ${healthRes.status}`);
  }

  return { apiURL, apiKey };
}

// Simple health check endpoint (fastest path)
export function healthCheckTest(data) {
  const url = `${data.apiURL}/health`;
  const res = http.get(url, { tags: { name: 'health' } });

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 20ms': (r) => r.timings.duration < 20,
  });

  errorRate.add(!success);
  qpsLatency.add(res.timings.duration);
}

// Models list endpoint
export function modelsTest(data) {
  const url = `${data.apiURL}/v1/models`;
  const res = http.get(url, {
    headers: { 'Authorization': `Bearer ${data.apiKey}` },
    tags: { name: 'models' },
  });

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 50ms': (r) => r.timings.duration < 50,
  });

  errorRate.add(!success);
  qpsLatency.add(res.timings.duration);
}

// Main load test - mixes endpoints for realistic load
export function loadTest(data) {
  // Use 80% health checks (fastest) and 20% models for QPS testing
  const useHealthCheck = Math.random() < 0.8;

  if (useHealthCheck) {
    healthCheckTest(data);
  } else {
    modelsTest(data);
  }
}

// Teardown: Print summary
export function teardown(data) {
  console.log('Max QPS test completed');
  console.log(`  API URL: ${data.apiURL}`);
  console.log('  Review metrics to find breaking point');
}
