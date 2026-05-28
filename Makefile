.PHONY: dev prod down test test-integration

dev:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile dev up -d

prod:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile prod up -d

down:
	docker compose --profile dev --profile prod down --remove-orphans

test:
	@echo "→ unit + feature"
	go test ./internal/service/... ./internal/handler/... -v
	@echo "→ integration (requer DATABASE_URL)"
	go test -tags=integration ./test/integration/... -v
