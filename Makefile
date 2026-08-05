# ============================================================
# MAKE FILE FOR REST USER AGREGATOR
# ============================================================
# Commands:
#   make test-u        - Run unit tests (tag: unit)
#   make test-int      - Run integration tests (tag: integration) - starts DB
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

# Unit tests (tag: unit) - runs only tests with build tag "unit"
test-u:
	@echo "========================================="
	@echo "  RUNNING UNIT TESTS (tag: unit)"
	@echo "========================================="
	go test ./internal/handlers -tags=unit -v
	go test ./pkg/logger -v 

# Integration tests (tag: integration) - starts DB and waits (max 30s)
test-int:
	@echo "========================================="
	@echo "  RUNNING INTEGRATION TESTS (tag: integration)"
	@echo "========================================="
	@echo "[RUN] Starting DB..."
	docker-compose up -d db
	@echo "[RUN] Waiting for DB to be ready (timeout: 30s)..."
	@for i in $$(seq 1 30); do \
		if docker exec subscription-db pg_isready -U postgres; then \
			echo "[RUN] DB is ready."; \
			break; \
		fi; \
		sleep 1; \
		if [ $$i -eq 30 ]; then \
			echo "[ERR] Timeout: DB did not start within 30 seconds."; \
			exit 1; \
		fi; \
	done
	go test ./internal/handlers ./cmd/api -tags=integration -v

# All tests: unit first, then integration
test-all:
	@echo "========================================="
	@echo "  RUNNING ALL TESTS"
	@echo "========================================="
	$(MAKE) test-u && $(MAKE) test-
	
# -------------------- RUN --------------------

# Run the server locally with database
run:
	@echo "[RUN] Checking Docker..."
	@if ! docker info > /dev/null 2>&1; then \
		echo "========================================="; \
		echo "  ❌ DOCKER IS NOT RUNNING"; \
		echo "========================================="; \
		echo "  Please start Docker Desktop manually:"; \
		echo "  1. Find Docker Desktop in the Start menu"; \
		echo "  2. Launch it"; \
		echo "  3. Wait for the green dot in the system tray"; \
		echo "  4. Run 'make run' again"; \
		echo "========================================="; \
		exit 1; \
	fi
	@$(MAKE) docker-up-db
	@echo "[RUN] Starting server locally..."
	go run cmd/api/main.go
# -------------------- BUILD --------------------

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

# View logs from all containers
docker-logs:
	@echo "[RUN] Showing logs (Ctrl+C to exit)..."
	docker-compose logs -f

# View logs only from server
docker-logs-server:
	@echo "[RUN] Showing server logs (Ctrl+C to exit)..."
	docker-compose logs -f server

# -------------------- CLEAN --------------------

# Clean all build artifacts (bin/, coverage/, cache)
clean:
	@echo "========================================="
	@echo "  CLEANING BUILD ARTIFACTS"
	@echo "========================================="
	@echo "[RUN] Removing bin/..."
	rm -rf bin/
	@echo "[RUN] Removing coverage/..."
	rm -rf coverage/
	@echo "[RUN] Removing coverage files..."
	rm -f coverage.out coverage.html
	@echo "[RUN] Cleaning Go cache..."
	go clean -cache
	go clean -testcache
	@echo "[RUN] Tidying go.mod..."
	go mod tidy
	@echo "========================================="
	@echo "  CLEAN COMPLETE! 
	@echo "========================================="

# -------------------- HELP --------------------

# Show all available commands
help:
	@echo "========================================="
	@echo "  AVAILABLE COMMANDS"
	@echo "========================================="
	@echo "  make test-u        - Unit tests (tag: unit) - fast, no DB"
	@echo "  make test-int      - Integration tests (tag: integration) - starts DB"
	@echo "  make test-all      - Unit tests first, then integration"
	@echo "  make run           - Run server locally"
	@echo "  make build         - Build binary"
	@echo "  make docker-up     - Start containers"
	@echo "  make docker-down   - Stop containers"
	@echo "  make docker-up-db  - Start only DB"
	@echo "  make help          - Show this help"
	@echo "========================================="