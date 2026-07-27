# ============================================================
# MAKE FILE FOR REST USER AGREGATOR
# ============================================================
# Commands:
#   make test-u        - Run unit tests (mocks, no DB) - fast
#   make test-int      - Run integration tests (with DB) - starts DB automatically
#   make test-all      - Run unit tests first, then integration tests
#   make run           - Run server locally
#   make build         - Build binary
#   make docker-up     - Start all containers
#   make docker-down   - Stop all containers
#   make docker-up-db  - Start only PostgreSQL
#   make help          - Show all available commands
# ============================================================

.PHONY: test-u test-int test-all run build docker-up docker-down docker-up-db help

# -------------------- TESTS --------------------

# Unit tests (mocks, no DB) - runs only tests with "_Mock" suffix
test-u:
	@echo "========================================="
	@echo "  RUNNING UNIT TESTS (mocks, no DB)"
	@echo "========================================="
	go test -run 'Test.*_Mock$$' ./... -v

# Integration tests (with DB) - runs tests with "Handler" suffix OR TestInitDB OR TestRun
test-int:
	@echo "========================================="
	@echo "  RUNNING INTEGRATION TESTS (with DB)"
	@echo "========================================="
	@echo "[RUN] Starting DB..."
	docker-compose up -d db
	@echo "[RUN] Waiting for DB to be ready (timeout: 30s)..."
	@timeout=30; \
	elapsed=0; \
	while [ $$elapsed -lt $$timeout ]; do \
		echo "SELECT 1" | docker exec -i subscription-db psql -U postgres -d subscriptions > /dev/null 2>&1; \
		if [ $$? -eq 0 ]; then \
			echo "[RUN] DB is ready."; \
			break; \
		fi; \
		sleep 1; \
		elapsed=$$((elapsed + 1)); \
		echo "[RUN] Waiting for DB... $$elapsed s"; \
	done; \
	if [ $$elapsed -ge $$timeout ]; then \
		echo "[ERR] Timeout: DB did not start within 30 seconds."; \
		exit 1; \
	fi
	go test -run 'Test.*Handler$$|TestInitDB|TestRun' ./... -v

# All tests: unit first, then integration
test-all:
	@echo "========================================="
	@echo "  RUNNING ALL TESTS"
	@echo "========================================="
	$(MAKE) test-u && $(MAKE) test-int

# -------------------- RUN --------------------

# Run the server locally
run:
	@echo "[RUN] Starting server..."
	go run cmd/api/main.go

# Build the binary
build:
	@echo "[RUN] Building binary..."
	go build -o bin/subscription_app cmd/api/main.go

# -------------------- DOCKER --------------------

# Start all containers
docker-up:
	@echo "[RUN] Starting containers..."
	docker-compose up -d

# Stop all containers
docker-down:
	@echo "[RUN] Stopping containers..."
	docker-compose down

# Start only the database
docker-up-db:
	@echo "[RUN] Starting DB..."
	docker-compose up -d db

# -------------------- HELP --------------------

# Show all available commands
help:
	@echo "========================================="
	@echo "  AVAILABLE COMMANDS"
	@echo "========================================="
	@echo "  make test-u        - Unit tests (mocks, no DB) - fast"
	@echo "  make test-int      - Integration tests (starts DB, waits 30s max)"
	@echo "  make test-all      - Unit tests first, then integration"
	@echo "  make run           - Run server locally"
	@echo "  make build         - Build binary"
	@echo "  make docker-up     - Start containers"
	@echo "  make docker-down   - Stop containers"
	@echo "  make docker-up-db  - Start only DB"
	@echo "  make help          - Show this help"
	@echo "========================================="