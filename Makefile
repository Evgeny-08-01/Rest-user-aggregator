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
# TESTING
# ============================================================

.PHONY: test-u
test-u: ## Run unit tests (with mocks, no database)
	@echo "========================================="
	@echo "  RUNNING UNIT TESTS (tag: unit)"
	@echo "========================================="
	go test ./internal/handlers -tags=unit -v

.PHONY: test-int
test-int: ## Run integration tests (with real DB via Docker)
	@echo "========================================="
	@echo "  RUNNING INTEGRATION TESTS (tag: integration)"
	@echo "========================================="
	go test -tags=integration -p 1 -count=1 ./...

.PHONY: test-all
test-all: test-u test-int ## Run unit tests first, then integration tests

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

.PHONY: run
run: ## Start the server locally (DB via Docker)
	@echo "[RUN] Checking Docker..."
	make docker-up-db
	@echo "[RUN] Starting server locally..."
	go run cmd/api/main.go

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
migrate-up: ## Apply all migrations (subscriptions + users)
	@echo "========================================="
	@echo "  APPLYING ALL MIGRATIONS"
	@echo "========================================="
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000001_create_subscriptions_table.up.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000002_create_users_table.up.sql
	@echo "✅ Migrations applied."

.PHONY: migrate-down
migrate-down: ## Rollback all migrations (users first, then subscriptions)
	@echo "========================================="
	@echo "  ROLLING BACK ALL MIGRATIONS"
	@echo "========================================="
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000002_create_users_table.down.sql
	docker exec -i subscription-db psql -U postgres -d subscriptions < migrations/000001_create_subscriptions_table.down.sql
	@echo "✅ All migrations rolled back."

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
# HELP
# ============================================================

.PHONY: help
help: ## Show all available commands
	@echo "========================================="
	@echo "  AVAILABLE COMMANDS"
	@echo "========================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'