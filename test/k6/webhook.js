import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// 404 é esperado — cliente nao existe em load test sem setup prévio
http.setResponseCallback(http.expectedStatuses(200, 404));

const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration', true);

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '30s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],
    http_req_failed: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

export default function () {
  const payload = JSON.stringify({
    event_id: `evt-${__VU}-${__ITER}-${Date.now()}`,
    card_id: `card-${__VU}`,
    cliente_email: `load-test-${__VU}@example.com`,
    timestamp: new Date().toISOString(),
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(`${BASE_URL}/webhooks/pipefy/card-updated`, payload, params);

  // 200 = processado ou duplicado (idempotente)
  // 404 = cliente nao existe — esperado em load test sem setup
  const success = check(res, {
    'status 200 ou 404': (r) => r.status === 200 || r.status === 404,
    'nao retorna 500': (r) => r.status !== 500,
  });

  errorRate.add(!success);
  requestDuration.add(res.timings.duration);

  sleep(0.1);
}
