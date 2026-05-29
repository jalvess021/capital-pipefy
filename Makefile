.PHONY: dev prod down test load-test-clients load-test-webhook

dev:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile dev up -d

prod:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile prod up -d

down:
	docker compose --profile dev --profile prod down --remove-orphans

load-test-clients:
	k6 run test/k6/create_client.js

load-test-webhook:
	k6 run test/k6/webhook.js

test:
	docker exec capital-pipefy-api-dev-1 sh -c \
	  "go test ./internal/handler/... ./internal/service/... ./internal/infrastructure/... -v && \
	   go test -tags=integration ./test/integration/... -v"
