import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = __ENV.API_GATEWAY_URL || 'http://localhost:8080';
const EMAIL = __ENV.K6_EMAIL || 'manager@mail.com';
const PASSWORD = __ENV.K6_PASSWORD || 'manager123';

export const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '10s', target: 5 },
    { duration: '20s', target: 15 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500'],
    errors: ['rate<0.05'],
  },
};

export function setup() {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health status is 200': (r) => r.status === 200,
  });

  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: EMAIL, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const ok = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login returns token': (r) => Boolean(r.json('data.token')),
  });

  if (!ok) {
    throw new Error(`Setup login failed (${loginRes.status}): ${loginRes.body}`);
  }

  return { token: loginRes.json('data.token') };
}

export default function (data) {
  const params = {
    headers: {
      Authorization: `Bearer ${data.token}`,
      'Content-Type': 'application/json',
    },
  };

  const endpoints = [
    '/health',
    '/api/v1/products?page=1&limit=10',
    '/api/v1/categories?page=1&limit=10',
    '/api/v1/warehouses?page=1&limit=10',
    '/api/v1/merchants?page=1&limit=10',
  ];

  for (const path of endpoints) {
    const res = path === '/health'
      ? http.get(`${BASE_URL}${path}`)
      : http.get(`${BASE_URL}${path}`, params);

    const passed = check(res, {
      [`${path} status < 400`]: (r) => r.status >= 200 && r.status < 400,
    });
    errorRate.add(!passed);
  }

  sleep(1);
}
