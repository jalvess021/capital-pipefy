.PHONY: dev prod prod-scale down test load-test-clients load-test-webhook

dev:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile dev up -d

prod:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile prod up -d

prod-scale:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile prod build --no-cache
	docker compose --profile prod up -d --scale api=3

down:
	docker compose --profile dev --profile prod down --remove-orphans

load-test-clients:
	k6 run test/k6/create_client.js

load-test-webhook:
	k6 run test/k6/webhook.js

test:
	docker compose --profile dev exec api-dev sh -c \
	  "go test ./internal/handler/... ./internal/service/... ./internal/infrastructure/... -v && \
	   go test -tags=integration ./test/integration/... -v"
