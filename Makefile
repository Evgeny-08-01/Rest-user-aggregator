# ============================================================
# MAKE FILE FOR REST USER AGREGATOR
# ============================================================
#
# Commands:
#   make test-u              - Run unit tests (with mocks, no DB)
#   make test-int            - Run integration tests (with DB, auto-start Docker)
#   make test-all            - Run unit tests first, then integration tests
#   make build               - Build the Go binary
#   make run                 - Start the server locally (with DB via Docker)
#   make docker-up           - Start all containers
#   make docker-down         - Stop all containers
#   make docker-up-db        - Start only PostgreSQL
#   make docker-logs         - View logs from all containers
#   make docker-logs-server  - View logs from server container only
#   make clean               - Clean build artifacts (bin/, coverage/, cache)
#   make help                - Show all available commands
#   mart stop                - stop local server
#
# Migration commands:
#   make migrate-up          - Apply all migrations
#   make migrate-down        - Rollback all migrations
#   make migrate-down-users  - Rollback only users table
#   make migrate-down-subs   - Rollback only subscriptions table
#   
# ============================================================

# ============================================================
# VARIABLES
# ============================================================
BINARY_NAME=subscription_app
COVERAGE_FILE=coverage.out
# ============================================================
# LINT
# ============================================================

.PHONY: lint
lint: ## Run golangci-lint
	@echo "========================================="
	@echo "  RUNNING LINTER"
	@echo "========================================="
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "========================================="
	@echo "  RUNNING LINTER WITH AUTO-FIX"
	@echo "========================================="
	golangci-lint run --fix ./...

# ============================================================
# TESTING
# ============================================================

.PHONY: test-u
test-u: ## Run all unit tests (with mocks, no database)
	@echo "========================================="
	@echo "  RUNNING ALL UNIT TESTS (tag: unit)"
	@echo "========================================="
	go test -tags=unit -v ./...

.PHONY: test-int
test-int: ## Run all integration tests (with real DB via Docker)
	@echo "========================================="
	@echo "  RUNNING ALL INTEGRATION TESTS (tag: integration)"
	@echo "========================================="
	go test -tags=integration -p 1 -count=1 ./...

.PHONY: test-all
test-all: ## Run all tests (unit + integration + no tags)
	@echo "========================================="
	@echo "  RUNNING ALL TESTS (unit + integration)"
	@echo "========================================="
	go test -tags=unit -v ./... && go test -tags=integration -p 1 -count=1 ./... && go test -p 1 -count=1 ./...
	
# ============================================================
# BUILD
# ============================================================

.PHONY: build
build: ## Build the Go binary
	@echo "========================================="
	@echo "  BUILDING BINARY"
	@echo "========================================="
	go build -o $(BINARY_NAME) ./cmd/api/main.go

# ============================================================
# RUN (LOCAL)
# ============================================================

.PHONY: docker-up-redis
docker-up-redis: ## Start Redis only
	docker-compose up -d redis

.PHONY: run
run: docker-up-db docker-up-redis ## Start the server locally (DB + Redis via Docker)
	@echo "[RUN] Starting server locally..."
	go run cmd/api/main.go cmd/api/init.go cmd/api/servers.go cmd/api/helpers.go
.PHONY: stop
stop: ## Остановить локальный сервер
	@echo "Stopping server..."
	@taskkill /F /IM go.exe 2>nul || true
	@echo "Server stopped"

# ============================================================
# DOCKER
# ============================================================

.PHONY: docker-up
docker-up: ## Start all containers
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all containers
	docker-compose down

.PHONY: docker-up-db
docker-up-db: ## Start only PostgreSQL
	docker-compose up -d db

.PHONY: docker-logs
docker-logs: ## View logs from all containers
	docker-compose logs -f

.PHONY: docker-logs-server
docker-logs-server: ## View logs from server container only
	docker logs subscription-api -f

# ============================================================
# MIGRATIONS
# ============================================================

.PHONY: migrate-up
migrate-up:
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000001_create_users_table.up.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000004_create_subscription_templates_table.up.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000002_create_subscriptions_table.up.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000003_create_cache_control_user_table.up.sql

.PHONY: migrate-down
migrate-down:
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000004_create_subscription_templates_table.down.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000003_create_cache_control_user_table.down.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000002_create_subscriptions_table.down.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000001_create_users_table.down.sql

.PHONY: migrate-down-users
migrate-down-users: ## Rollback only users table
	@echo "========================================="
	@echo "  ROLLING BACK USERS TABLE"
	@echo "========================================="
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000002_create_users_table.down.sql
	@echo "✅ Users table dropped."

.PHONY: migrate-down-subs
migrate-down-subs: ## Rollback only subscriptions table
	@echo "========================================="
	@echo "  ROLLING BACK SUBSCRIPTIONS TABLE"
	@echo "========================================="
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000001_create_subscriptions_table.down.sql
	@echo "✅ Subscriptions table dropped."

# ============================================================
# CLEANUP
# ============================================================

.PHONY: clean
clean: ## Clean build artifacts (bin/, coverage/, cache)
	@echo "========================================="
	@echo "  CLEANING UP"
	@echo "========================================="
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	go clean -cache
	@echo "✅ Cleanup done."

# ============================================================
# MOBILE
# ============================================================

ADB = /c/AppData/Local/Android/Sdk/platform-tools/adb.exe

.PHONY: adb-reverse
adb-reverse: ## Пробросить порт для мобильного приложения (USB)
	@echo "========================================="
	@echo "  PROXY PORT FOR MOBILE (adb reverse)"
	@echo "========================================="
	"C:/AppData/Local/Android/Sdk/platform-tools/adb.exe" reverse tcp:8087 tcp:8087
	"C:/AppData/Local/Android/Sdk/platform-tools/adb.exe" reverse tcp:50051 tcp:50051
	@echo "✅ Ports 8087 and 50051 forwarded to mobile device"

# ============================================================
# HELP
# ============================================================
.PHONY: help
help: ## Show all available commands
	@echo "========================================="
	@echo "  AVAILABLE COMMANDS"
	@echo "========================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

