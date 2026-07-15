.PHONY: dev dev-air build test lint openapi-check coverage-check migrate-up migrate-down migrate-status docker-up docker-down tidy backup restore observability-up miniapp-dev miniapp-build tunnel stack-up verify-webapp-auth package package-mac package-linux package-win

BINARY := lifeos
MAIN := ./cmd/lifeos
COMPOSE := docker compose -f deployments/docker-compose.yml
ifneq (,$(wildcard deployments/docker-compose.override.yml))
COMPOSE += -f deployments/docker-compose.override.yml
endif

VERSION ?= LifeOS_alpha_1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILT_AT ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/valentinezhov/lifeos/cmd/lifeos/cmd.Version=$(VERSION) -X github.com/valentinezhov/lifeos/cmd/lifeos/cmd.Commit=$(COMMIT) -X github.com/valentinezhov/lifeos/cmd/lifeos/cmd.BuiltAt=$(BUILT_AT)

dev:
	go run $(MAIN) serve

dev-air:
	air -c .air.toml

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(MAIN)

# Desktop/local click-to-run packages → dist/
package:
	./scripts/package.sh

package-mac:
	LIFEOS_VERSION=$(VERSION) ./scripts/package.sh darwin arm64

package-linux:
	LIFEOS_VERSION=$(VERSION) ./scripts/package.sh linux amd64

package-win:
	LIFEOS_VERSION=$(VERSION) ./scripts/package.sh windows amd64


test:
	GOTOOLCHAIN=local go test ./... -count=1 -coverprofile=coverage.out
	@GOTOOLCHAIN=local go tool cover -func=coverage.out | tail -1

coverage-check:
	./scripts/check-coverage.sh coverage.out

test-integration:
	GOTOOLCHAIN=local go test ./internal/transport/http/api/... ./e2e/... -count=1 -tags=integration

lint:
	golangci-lint run ./...

openapi-check:
	go test ./internal/transport/http/api -run TestOpenAPIParity -count=1

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
	$(COMPOSE) up -d --build

docker-down:
	$(COMPOSE) down

tidy:
	go mod tidy

ci: tidy lint openapi-check test coverage-check build

miniapp-dev:
	cd web/miniapp && npm run dev

miniapp-build:
	cd web/miniapp && npm ci && npm run build

tunnel:
	./scripts/https-tunnel.sh 8080

verify-webapp-auth:
	./scripts/verify-webapp-auth.sh

# Build Mini App, raise HTTPS tunnel, rebuild and restart full stack.
stack-up: miniapp-build
	$(COMPOSE) up -d --build
	./scripts/https-tunnel.sh 8080
	$(COMPOSE) up -d --build app
	@echo "Open Telegram → /start → reply keyboard «📱 Mini App» (or menu button)"
	@echo "Then: make verify-webapp-auth"
