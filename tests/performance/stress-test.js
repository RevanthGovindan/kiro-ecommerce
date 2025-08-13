import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Stress test configuration - gradually increase load to find breaking point
export const options = {
  stages: [
    { duration: '1m', target: 10 },   // Warm up
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 300 },  // Push to 300 users
    { duration: '5m', target: 300 },  // Stay at 300 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1s
    http_req_failed: ['rate<0.2'],     // Error rate must be below 20%
    errors: ['rate<0.2'],
  },
};

const BASE_URL = 'http://localhost:8080';

export default function() {
  // High-load scenarios focusing on most critical endpoints
  const scenarios = [
    { endpoint: '/api/products', weight: 50 },
    { endpoint: '/api/products/search?q=laptop', weight: 20 },
    { endpoint: '/api/categories', weight: 15 },
    { endpoint: '/api/auth/login', weight: 10, method: 'POST' },
    { endpoint: '/api/cart', weight: 5 }
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  
  let response;
  if (scenario.method === 'POST') {
    response = http.post(`${BASE_URL}${scenario.endpoint}`, JSON.stringify({
      email: 'test@example.com',
      password: 'password123'
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  } else {
    response = http.get(`${BASE_URL}${scenario.endpoint}`);
  }

  check(response, {
    'status is not 5xx': (r) => r.status < 500,
    'response time < 2s': (r) => r.timings.duration < 2000,
  }) || errorRate.add(1);

  sleep(0.1); // Minimal sleep for stress testing
}