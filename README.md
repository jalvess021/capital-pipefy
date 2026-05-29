# Capital Pipefy — Teste Técnico Mundo Invest

API de gerenciamento de clientes com integração simulada ao Pipefy. Desenvolvida em Go com arquitetura limpa, resiliência de produção e observabilidade estruturada.

---

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- [k6](https://k6.io/docs/get-started/installation/) (para testes de carga)
- Go 1.25+ (apenas para desenvolvimento local sem Docker)

---

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.25 |
| HTTP | Gin |
| ORM | GORM |
| Banco | PostgreSQL 16 |
| Cache / Estado | Redis 7 |
| Proxy / Rate limit | NGINX 1.27 |
| Logs estruturados | Zap |
| Coleta de logs | Promtail 2.9 |
| Armazenamento de logs | Loki 2.9 |
| Dashboard | Grafana 10 |
| Migrations | golang-migrate |
| Testes de carga | k6 |

---

## Executando o projeto

```bash
# Copiar variáveis de ambiente e subir (1 instância, hot reload)
make dev

# Subir em produção com binário otimizado (1 instância)
make prod

# Subir em produção com 3 instâncias (simula escala horizontal)
make prod-scale

# Parar tudo
make down
```

A API estará disponível em `http://localhost:8000`.  
Swagger UI: `http://localhost:8000/swagger/index.html`

---

## Observabilidade

Todos os serviços sobem automaticamente com `make dev` ou `make prod`.

| Serviço | URL |
|---|---|
| **Dashboard de logs** | [`http://localhost:3000/d/capital-pipefy-logs`](http://localhost:3000/d/capital-pipefy-logs) |
| Grafana (home) | `http://localhost:3000` |

> Acesso anônimo habilitado por padrão — não é necessário login.

### Pipeline de logs

```
App (Zap — JSON) → Docker stdout → Promtail → Loki → Grafana
```

Cada log carrega os campos:

| Campo | Descrição |
|---|---|
| `level` | `INFO`, `WARN`, `ERROR` |
| `request_id` | UUID por requisição (`X-Request-ID`) |
| `method` / `path` | Rota HTTP |
| `status` | Código de resposta |
| `latency_ms` | Duração da requisição |
| `service` | Componente que gerou o log (`handler`, `pipefy`, etc.) |

O dashboard pré-provisionado exibe:
- Logs filtrados por nível, rota e `request_id`
- Painel separado para erros de infraestrutura (Pipefy, circuit breaker)
- Logs de requisição HTTP com latência

---

## Variáveis de ambiente

Copie `.env.example` para `.env` e ajuste conforme necessário.

| Variável | Padrão | Descrição |
|---|---|---|
| `PIPEFY_TOKEN` | `your_token_here` | Token real para integrar ao Pipefy |
| `PIPEFY_PIPE_ID` | `your_pipe_id_here` | ID do pipe no Pipefy |
| `RATE_LIMIT_ENABLED` | `true` | `false` para load tests locais (k6 usa 1 IP) |
| `NGINX_RATE_LIMIT_RPS` | `10` | Requisições/s por IP no NGINX |
| `DB_MAX_OPEN_CONNS` | `100` | Conexões simultâneas ao Postgres |
| `GRAFANA_PORT` | `3000` | Porta do Grafana |

---

## Rodando os testes

### Testes automatizados (unitários + integração)

Requer `make dev` ativo. O binário de produção não inclui o código fonte.

```bash
make dev   # obrigatório
make test
```

### Cobertura dos requisitos obrigatórios

| Requisito | Teste | Arquivo |
|---|---|---|
| 1 — Criação de cliente com payload válido e salvamento no banco | `TestClientHandler_Create_Success` | `internal/handler/client_handler_test.go` |
| 1 — Salvamento real no banco (integração) | `TestClientRepository_SaveAndFind` | `test/integration/` |
| 2 — Regra de prioridade: boundary exato (199.999 / 200.000) | `TestCalculatePriority` | `internal/service/client_service_test.go` |
| 2 — Regra de prioridade via fluxo HTTP completo | `TestWebhookHandler_Priority_Alta` / `_Normal` | `internal/handler/webhook_handler_test.go` |
| 3 — Bloqueio de `event_id` duplicado | `TestWebhookHandler_CardUpdated_DuplicateEvent_Returns200` | `internal/handler/webhook_handler_test.go` |

### Testes de carga (k6)

Simulam comportamento externo — usuários reais batendo na API. Funcionam contra qualquer ambiente.

```bash
# dev (1 instância)
make dev && make load-test-clients && make load-test-webhook

# prod com 3 instâncias (valida escala horizontal)
make prod-scale && make load-test-clients && make load-test-webhook
```

> Antes de rodar k6, ajuste o `.env` conforme o cenário desejado (ver comentários no arquivo).  
> `RATE_LIMIT_ENABLED=false` desabilita rate limit no app para medir capacidade real — rate limiter testado isoladamente via `make test`.

**Resultados esperados:** 0% erro, p99 < 500ms.

---

## Endpoints

### POST /clientes

Cria um cliente, salva no banco e sincroniza card no Pipefy (best-effort).

```bash
curl -X POST http://localhost:8000/clientes \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000
  }'
```

**Resposta 201:**
```json
{
  "id": "uuid",
  "nome": "João Silva",
  "email": "joao.silva@example.com",
  "tipo_solicitacao": "Atualização cadastral",
  "valor_patrimonio": 250000,
  "status": "Aguardando Análise",
  "prioridade": "prioridade_alta"
}
```

| `valor_patrimonio` | `prioridade` |
|---|---|
| ≥ 200.000 | `prioridade_alta` |
| < 200.000 | `prioridade_normal` |

---

### POST /webhooks/pipefy/card-updated

Simula recebimento de evento do Pipefy. Processa idempotentemente pelo `event_id`.

```bash
curl -X POST http://localhost:8000/webhooks/pipefy/card-updated \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-001",
    "card_id": "card-456",
    "cliente_email": "joao.silva@example.com",
    "timestamp": "2026-05-29T12:00:00Z"
  }'
```

**Respostas:**
- `200` — evento processado ou já processado (idempotente)
- `404` — cliente não encontrado
- `429` — rate limit excedido

---

## Arquitetura

```
cmd/api/
internal/
├── bootstrap/       # Wiring de dependências (DI manual)
├── config/          # Leitura de variáveis de ambiente
├── domain/          # Entidades de negócio (Client, ProcessedEvent)
├── dto/             # Request/response shapes
├── handler/         # Camada HTTP (Gin) — entrada e saída
├── service/         # Regras de negócio (calculatePriority, fluxos de cliente e webhook)
├── port/            # Interfaces de saída (Pipefy)
├── repository/      # Interfaces de persistência
│   └── postgres/    # Implementações GORM/Postgres
├── infrastructure/
│   └── pipefy/      # Client GraphQL com mutations reais, retry e circuit breaker Redis
│       └── card/    # createCard e updateCardField estruturados conforme doc Pipefy
├── apperrors/       # Erros de domínio (ErrNotFound, ErrConflict, ErrInternal)
├── middleware/       # Rate limiter (redis_rate) e request_id
├── logger/          # Log estruturado por camada (handler, infra, webhook, request)
└── route/           # Setup de rotas e middlewares
test/
├── integration/     # Testes com dependências reais (Redis, Postgres)
└── k6/              # Scripts de carga
infra/
├── nginx/           # Config NGINX (rate limit, proxy, envsubst)
├── loki/            # Config de armazenamento de logs
├── promtail/        # Config de coleta de logs (lê stdout Docker)
└── grafana/         # Dashboard e datasources pré-provisionados
migrations/          # SQL migrations (golang-migrate)
```

### Camadas e dependências

```
handler → service → repository (interface)
                 → pipefy port (interface)

repository/postgres  ← implementação concreta, só trocando providers.go
infrastructure/pipefy ← implementação concreta
```

Trocar Postgres por MongoDB ou GORM por raw SQL: apenas `internal/bootstrap/providers.go` muda.

---

## Resiliência

### Integração Pipefy (best-effort)

Falha do Pipefy não bloqueia resposta. Cliente é salvo no banco independente da integração. Mutations GraphQL estão estruturadas conforme documentação oficial:

- `createCard` — cria card com nome, email e patrimônio
- `updateCardField` — atualiza status para `"Processado"`

### Circuit Breaker (Redis — distribuído)

Estado compartilhado via Redis. Funciona corretamente em múltiplas instâncias.

```
CLOSED → N falhas consecutivas → OPEN (30s) → HALF-OPEN → testa 1 req
```

Configurável via `.env`: `PIPEFY_CB_THRESHOLD`, `PIPEFY_CB_OPEN_TIMEOUT`.

### Retry com backoff exponencial

Falhas transientes (502/503/504/timeout) repassam com delay crescente. Erros permanentes (401) abortam imediatamente.

### Rate Limiting (dois layers)

| Layer | Onde | Configuração |
|---|---|---|
| NGINX | Antes de chegar na app | `NGINX_RATE_LIMIT_RPS` / `NGINX_RATE_LIMIT_BURST` |
| App (redis_rate) | Middleware Go | `RATE_LIMIT_RPS` / `RATE_LIMIT_ENABLED` |

Ambos com estado Redis — funcionam em múltiplas instâncias.

### Idempotência atômica

`INSERT INTO processed_events ... ON CONFLICT (event_id) DO NOTHING`. Dois workers simultâneos com mesmo `event_id` — só um processa, sem race condition.

---

## Escalabilidade horizontal

```bash
make prod-scale  # build sem cache + sobe 3 instâncias
```

NGINX resolve `api` via DNS do Docker (round-robin automático). Redis e Postgres são compartilhados entre instâncias. Circuit breaker e rate limiter funcionam distribuídos via Redis.

**Validado com k6:** 50 VUs simultâneos, 0% erro, p99 < 500ms em ambiente com 3 instâncias.

> `make dev` usa Air (hot reload) — não suporta múltiplas instâncias no mesmo volume. Para escala horizontal use `make prod-scale`.

---

## Visão de Produção (AWS)

### Arquitetura sugerida

```
API Gateway → Lambda (Go) → RDS Postgres (Multi-AZ)
                          → ElastiCache Redis (circuit breaker + rate limit)
                          → SQS → Lambda worker → Pipefy GraphQL
```

### Componentes

**API Gateway + Lambda:** escala automaticamente, zero gerenciamento de servidores. Cold start mitigado com Provisioned Concurrency.

**RDS Postgres Multi-AZ:** réplica de leitura para consultas, failover automático. Migrations via CodePipeline na deploy.

**ElastiCache Redis:** circuit breaker e rate limiter distribuídos entre todas as instâncias Lambda. Sem estado local — safe para escala horizontal infinita.

**SQS para integração Pipefy:** substitui best-effort atual. Webhook recebe evento → publica na fila → worker Lambda consome → chama Pipefy com retry automático via SQS redrive policy. Nunca perde evento mesmo com Pipefy fora.

**DynamoDB (alternativa ao Postgres):** para `processed_events` (idempotência), DynamoDB com conditional write (`attribute_not_exists(event_id)`) é mais performático e serverless-native que Postgres para esse padrão.

### Escalabilidade

- Lambda escala de 0 a milhares de instâncias automaticamente
- SQS absorve picos sem perda de eventos
- Redis ElastiCache garante consistência de rate limit entre instâncias
- RDS com connection pooling via RDS Proxy (necessário com Lambda — muitas conexões)

---

## Limitações conhecidas

- **Sem fila para Pipefy:** falha do Pipefy = card não criado, sem retry posterior. Evolução: SQS/RabbitMQ.
- **Circuit breaker distribuído validado por k6:** integration tests usam instância única. Comportamento multi-instância validado via `make prod-scale + make load-test-*`.
- **Sem autenticação:** endpoints públicos. Em produção: JWT via API Gateway Authorizer.
