import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration', true);

export const options = {
  stages: [
    { duration: '10s', target: 10 },  // ramp up para 10 usuarios
    { duration: '30s', target: 50 },  // sobe para 50 usuarios
    { duration: '10s', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

let counter = 0;

export default function () {
  counter++;
  const email = `load-test-${__VU}-${__ITER}-${Date.now()}@example.com`;

  const payload = JSON.stringify({
    cliente_nome: `Load Test User ${__VU}`,
    cliente_email: email,
    tipo_solicitacao: 'investimento',
    valor_patrimonio: Math.random() > 0.5 ? 250000 : 100000,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(`${BASE_URL}/clientes`, payload, params);

  const success = check(res, {
    'status 201': (r) => r.status === 201,
    'tem id na resposta': (r) => {
      try {
        return JSON.parse(r.body).id !== undefined;
      } catch {
        return false;
      }
    },
    'tem prioridade na resposta': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.prioridade === 'prioridade_alta' || body.prioridade === 'prioridade_normal';
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
  requestDuration.add(res.timings.duration);

  sleep(0.1);
}
