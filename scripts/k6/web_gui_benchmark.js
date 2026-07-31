import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 10 },
    { duration: '10s', target: 10 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE_URL = __ENV.DASHBOARD_URL || 'http://localhost:8084';

export default function () {
  // Test Dashboard root page / static assets load
  const indexRes = http.get(`${BASE_URL}/`);
  check(indexRes, {
    'dashboard UI status is 200': (r) => r.status === 200,
  });

  // Test Dashboard status API endpoint
  const statusRes = http.get(`${BASE_URL}/api/status`);
  check(statusRes, {
    'dashboard status API is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });

  sleep(1);
}
