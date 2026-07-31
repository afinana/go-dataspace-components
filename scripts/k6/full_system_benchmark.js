import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 15 },
    { duration: '15s', target: 15 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
  },
};

const CONTROL_PLANE_URL = __ENV.CONTROL_PLANE_URL || 'http://localhost:8081';
const DATA_PLANE_URL = __ENV.DATA_PLANE_URL || 'http://localhost:8082';
const IDENTITY_HUB_URL = __ENV.IDENTITY_HUB_URL || 'http://localhost:8080';
const DASHBOARD_URL = __ENV.DASHBOARD_URL || 'http://localhost:8084';

export default function () {
  // 1. Check Identity Hub health / DID query
  const idHubRes = http.get(`${IDENTITY_HUB_URL}/.well-known/did.json`);
  check(idHubRes, {
    'identity hub status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });

  // 2. Query Catalog on Control Plane
  const catalogRes = http.get(`${CONTROL_PLANE_URL}/api/dsp/2025-1/catalog/datasets`);
  check(catalogRes, {
    'control plane catalog status is 200': (r) => r.status === 200,
  });

  // 3. Check Data Plane status
  const dataPlaneRes = http.get(`${DATA_PLANE_URL}/health`);
  check(dataPlaneRes, {
    'data plane responded': (r) => r.status === 200 || r.status === 404,
  });

  // 4. Query Dashboard GUI
  const dashboardRes = http.get(`${DASHBOARD_URL}/`);
  check(dashboardRes, {
    'dashboard responded': (r) => r.status === 200,
  });

  sleep(1);
}
