APP_NAME  := pos-backend
CMD_PATH  := ./cmd/api
BIN_PATH  := ./bin/$(APP_NAME)
MIGRATION_DIR := ./migrations
DB_URL    ?= $(shell grep ^DB_URL .env | cut -d '=' -f2-)

.PHONY: all dev run build test lint swagger \
        migrate-up migrate-down migrate-create migrate-force \
        docker-up docker-down docker-logs clean help

# ── Development ────────────────────────────────────────────────────────────────

## dev: [LOCAL] generate swagger + apply migrations + start server (all-in-one)
## NOTE: Gunakan ini saat local dev. Jangan dipakai di production/CI — gunakan "make run".
dev: swagger migrate-up run

## run: start the application only (reads .env via godotenv)
run:
	go run $(CMD_PATH)/main.go

## build: compile binary to ./bin/
build:
	go build -o $(BIN_PATH) $(CMD_PATH)/main.go
	@echo "Binary built: $(BIN_PATH)"

## test: run all unit tests with race detector and coverage report
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: run golangci-lint
lint:
	golangci-lint run ./...

# ── Swagger ────────────────────────────────────────────────────────────────────

## swagger: regenerate Swagger documentation from code annotations
swagger:
	@which swag > /dev/null 2>&1 || \
		(echo "❌ swag CLI not found. Install: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g cmd/api/main.go -o ./docs --parseDependency --parseInternal
	@echo "✅ Swagger docs generated at ./docs"

# ── Migration ──────────────────────────────────────────────────────────────────

## migrate-up: apply all pending migrations
migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up

## migrate-down: rollback the last applied migration
migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1

## migrate-create name=<migration_name>: create a new migration file pair
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)
	@echo "Migration files created in $(MIGRATION_DIR)/"

## migrate-force version=<version>: force migration version (use when migration is dirty)
migrate-force:
	@if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=<version>"; exit 1; fi
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(version)

# ── Docker ─────────────────────────────────────────────────────────────────────

## docker-up: start all services (Postgres, Redis, Grafana stack) in background
docker-up:
	docker compose up -d
	@echo "Services started. Grafana: http://localhost:3000 (admin/admin)"

## docker-down: stop all services
docker-down:
	docker compose down

## docker-logs: follow logs of all services
docker-logs:
	docker compose logs -f

# ── Cleanup ────────────────────────────────────────────────────────────────────

## clean: remove build artifacts and coverage reports
clean:
	rm -rf $(BIN_PATH) coverage.out coverage.html

# ── Help ───────────────────────────────────────────────────────────────────────

## help: display this help message
help:
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""
