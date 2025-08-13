import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up to 10 users
    { duration: '5m', target: 10 }, // Stay at 10 users
    { duration: '2m', target: 20 }, // Ramp up to 20 users
    { duration: '5m', target: 20 }, // Stay at 20 users
    { duration: '2m', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data
const testUser = {
  email: 'loadtest@example.com',
  password: 'loadtest123',
  firstName: 'Load',
  lastName: 'Test'
};

export function setup() {
  // Register test user
  const registerResponse = http.post(`${BASE_URL}/api/auth/register`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (registerResponse.status === 201 || registerResponse.status === 409) {
    // Login to get token
    const loginResponse = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
      email: testUser.email,
      password: testUser.password
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (loginResponse.status === 200) {
      const loginData = JSON.parse(loginResponse.body);
      return { token: loginData.data.token };
    }
  }
  
  return { token: null };
}

export default function(data) {
  const headers = data.token ? {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json'
  } : {
    'Content-Type': 'application/json'
  };

  // Test scenarios with different weights
  const scenarios = [
    { name: 'browse_products', weight: 40 },
    { name: 'search_products', weight: 20 },
    { name: 'view_product_details', weight: 15 },
    { name: 'cart_operations', weight: 15 },
    { name: 'user_operations', weight: 10 }
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  
  switch(scenario.name) {
    case 'browse_products':
      browseProducts();
      break;
    case 'search_products':
      searchProducts();
      break;   
 case 'view_product_details':
      viewProductDetails();
      break;
    case 'cart_operations':
      cartOperations(headers);
      break;
    case 'user_operations':
      userOperations(headers);
      break;
  }

  sleep(1);
}

function browseProducts() {
  // Get products list
  const response = http.get(`${BASE_URL}/api/products?page=1&page_size=20`);
  
  check(response, {
    'products list status is 200': (r) => r.status === 200,
    'products list response time < 200ms': (r) => r.timings.duration < 200,
    'products list has data': (r) => {
      const body = JSON.parse(r.body);
      return body.success && body.data.products.length > 0;
    }
  }) || errorRate.add(1);

  // Get categories
  const categoriesResponse = http.get(`${BASE_URL}/api/categories`);
  
  check(categoriesResponse, {
    'categories status is 200': (r) => r.status === 200,
    'categories response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);
}

function searchProducts() {
  const searchTerms = ['laptop', 'phone', 'shirt', 'book', 'watch'];
  const term = searchTerms[Math.floor(Math.random() * searchTerms.length)];
  
  const response = http.get(`${BASE_URL}/api/products/search?q=${term}&page=1&page_size=10`);
  
  check(response, {
    'search status is 200': (r) => r.status === 200,
    'search response time < 300ms': (r) => r.timings.duration < 300,
    'search has results': (r) => {
      const body = JSON.parse(r.body);
      return body.success;
    }
  }) || errorRate.add(1);
}

function viewProductDetails() {
  // First get products list to get a product ID
  const productsResponse = http.get(`${BASE_URL}/api/products?page=1&page_size=5`);
  
  if (productsResponse.status === 200) {
    const productsData = JSON.parse(productsResponse.body);
    if (productsData.success && productsData.data.products.length > 0) {
      const productId = productsData.data.products[0].id;
      
      const response = http.get(`${BASE_URL}/api/products/${productId}`);
      
      check(response, {
        'product details status is 200': (r) => r.status === 200,
        'product details response time < 150ms': (r) => r.timings.duration < 150,
        'product details has data': (r) => {
          const body = JSON.parse(r.body);
          return body.success && body.data.product;
        }
      }) || errorRate.add(1);
    }
  }
}

function cartOperations(headers) {
  if (!headers.Authorization) return;

  // Get cart
  let response = http.get(`${BASE_URL}/api/cart`, { headers });
  
  check(response, {
    'get cart status is 200': (r) => r.status === 200,
    'get cart response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Add item to cart (get a product first)
  const productsResponse = http.get(`${BASE_URL}/api/products?page=1&page_size=1`);
  if (productsResponse.status === 200) {
    const productsData = JSON.parse(productsResponse.body);
    if (productsData.success && productsData.data.products.length > 0) {
      const productId = productsData.data.products[0].id;
      
      response = http.post(`${BASE_URL}/api/cart/add`, JSON.stringify({
        productId: productId,
        quantity: 1
      }), { headers });
      
      check(response, {
        'add to cart status is 200': (r) => r.status === 200,
        'add to cart response time < 200ms': (r) => r.timings.duration < 200,
      }) || errorRate.add(1);
    }
  }
}

function userOperations(headers) {
  if (!headers.Authorization) return;

  // Get user profile
  const response = http.get(`${BASE_URL}/api/users/profile`, { headers });
  
  check(response, {
    'get profile status is 200': (r) => r.status === 200,
    'get profile response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Get user orders
  const ordersResponse = http.get(`${BASE_URL}/api/users/orders`, { headers });
  
  check(ordersResponse, {
    'get orders status is 200': (r) => r.status === 200,
    'get orders response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);
}

export function teardown(data) {
  // Cleanup if needed
  console.log('Load test completed');
}