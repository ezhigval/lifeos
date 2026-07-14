.PHONY: dev dev-air build test lint migrate-up migrate-down migrate-status docker-up docker-down tidy backup restore observability-up miniapp-dev miniapp-build

BINARY := lifeos
MAIN := ./cmd/lifeos

dev:
	GOTOOLCHAIN=local go run $(MAIN) serve

dev-air:
	air -c .air.toml

build:
	GOTOOLCHAIN=local go build -o bin/$(BINARY) $(MAIN)

test:
	GOTOOLCHAIN=local go test ./... -count=1 -coverprofile=coverage.out
	@GOTOOLCHAIN=local go tool cover -func=coverage.out | tail -1

test-integration:
	GOTOOLCHAIN=local go test ./internal/transport/http/api/... ./e2e/... -count=1 -tags=integration

lint:
	golangci-lint run ./...

backup:
	./scripts/backup.sh

restore:
	@test -n "$(FILE)" || (echo "usage: make restore FILE=backups/lifeos_....dump" && exit 1)
	./scripts/restore.sh "$(FILE)"

observability-up:
	docker compose -f deployments/docker-compose.yml --profile observability up -d prometheus grafana jaeger

migrate-up:
	go run $(MAIN) migrate up

migrate-down:
	go run $(MAIN) migrate down

migrate-status:
	go run $(MAIN) migrate status

sqlc:
	sqlc generate

docker-up:
	docker compose -f deployments/docker-compose.yml up -d --build

docker-down:
	docker compose -f deployments/docker-compose.yml down

tidy:
	go mod tidy

ci: tidy lint test build

miniapp-dev:
	cd web/miniapp && npm run dev

miniapp-build:
	cd web/miniapp && npm ci && npm run build
