.PHONY: dev prod down test

dev:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile dev up -d

prod:
	@[ -f .env ] || cp .env.example .env
	docker compose --profile prod up -d

down:
	docker compose --profile dev --profile prod down --remove-orphans

test:
	docker exec capital-pipefy-api-dev-1 sh -c \
	  "go test ./internal/service/... ./internal/handler/... -v && \
	   go test -tags=integration ./test/integration/... -v"
