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

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

export default function () {
  // Test Catalog DCAT dataset endpoint
  const catalogRes = http.get(`${BASE_URL}/api/dsp/2025-1/catalog/datasets`);
  check(catalogRes, {
    'catalog status is 200': (r) => r.status === 200,
  });

  // Test Negotiate contract endpoint (GET/POST validation endpoint ping)
  const negotiateRes = http.get(`${BASE_URL}/api/dsp/2025-1/contracts/negotiate`);
  check(negotiateRes, {
    'negotiate endpoint responded': (r) => r.status === 200 || r.status === 405 || r.status === 404,
  });

  sleep(1);
}
