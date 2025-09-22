# Voting App Makefile
# Comprehensive build, test, and deployment commands

.PHONY: help build test test-unit test-integration test-e2e test-all clean deps run dev setup-db setup-test-db docker-build docker-run lint format

# Default target
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  dev             - Run in development mode with hot reload"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-integration- Run integration tests"
	@echo "  test-e2e        - Run end-to-end tests"
	@echo "  test-all        - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  setup-db        - Setup main database"
	@echo "  setup-test-db   - Setup test database"
	@echo "  migrate         - Run database migrations"
	@echo "  seed            - Seed database with test data"
	@echo "  deps            - Install dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  lint            - Run linters"
	@echo "  format          - Format code"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run in Docker container"

# Variables
APP_NAME = voting-app
BUILD_DIR = ./build
GO_FILES = $(shell find . -name "*.go" -not -path "./vendor/*" -not -path "./tests/*")
TEST_FILES = $(shell find ./tests -name "*_test.go")

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@echo "Checking Go modules..."
	@go mod tidy || (echo "Failed to tidy modules" && exit 1)
	@go build -o $(BUILD_DIR)/$(APP_NAME) . || (echo "Build failed" && exit 1)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Run the application
run: build
	@echo "Starting $(APP_NAME)..."
	@./$(BUILD_DIR)/$(APP_NAME)

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

# Database setup
setup-db:
	@echo "Setting up main database..."
	@psql -h localhost -U postgres -c "CREATE DATABASE voting_app;" || echo "Database might already exist"
	@echo "Database setup complete"

setup-test-db:
	@echo "Setting up test database..."
	@psql -h localhost -U postgres -c "CREATE DATABASE voting_app_test;" || echo "Test database might already exist"
	@echo "Test database setup complete"

# Run enhanced migrations
migrate:
	@echo "Running database migrations..."
	@psql -h localhost -U postgres -d voting_app -f enhanced_schema.sql
	@echo "Migrations complete"

migrate-test:
	@echo "Running test database migrations..."
	@psql -h localhost -U postgres -d voting_app_test -f enhanced_schema.sql
	@echo "Test migrations complete"

# Seed database with sample data
seed:
	@echo "Seeding database with sample data..."
	@go run scripts/seed.go
	@echo "Database seeded"

# Test commands
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./app/...
	@echo "Unit tests complete"

test-integration: setup-test-db migrate-test
	@echo "Running integration tests..."
	@go test -v -run "TestSuite" ./tests/
	@echo "Integration tests complete"

test-e2e: setup-test-db migrate-test
	@echo "Running end-to-end tests..."
	@go test -v -timeout 30m ./tests/
	@echo "End-to-end tests complete"

test-all: test-unit test-integration test-e2e
	@echo "All tests complete"

test-coverage: setup-test-db migrate-test
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	@go test -v -coverprofile=coverage/coverage.out ./app/... ./tests/...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

# Continuous testing (watch mode)
test-watch:
	@echo "Starting test watch mode..."
	@if command -v gotestsum > /dev/null; then \
		gotestsum --watch --format testname; \
	else \
		echo "gotestsum not found. Install with: go install gotest.tools/gotestsum@latest"; \
		echo "Falling back to regular test mode..."; \
		make test-all; \
	fi

# Code quality
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"; \
		echo "Running basic vet and fmt checks..."; \
		go vet ./...; \
		go fmt ./...; \
	fi
	@echo "Linting complete"

format:
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w $(GO_FILES); \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@echo "Formatting complete"

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .
	@echo "Docker image built: $(APP_NAME):latest"

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file local.env $(APP_NAME):latest

docker-test:
	@echo "Running tests in Docker..."
	@docker build -f Dockerfile.test -t $(APP_NAME)-test:latest .
	@docker run --rm $(APP_NAME)-test:latest

# Database operations
db-reset: setup-db migrate seed
	@echo "Database reset complete"

db-backup:
	@echo "Creating database backup..."
	@mkdir -p backups
	@pg_dump -h localhost -U postgres voting_app > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/"

db-restore:
	@echo "Restoring database from backup..."
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make db-restore FILE=backups/backup_file.sql"; \
		exit 1; \
	fi
	@psql -h localhost -U postgres -d voting_app < $(FILE)
	@echo "Database restored"

# Performance testing
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./app/...
	@echo "Benchmarks complete"

load-test:
	@echo "Running load tests..."
	@if command -v hey > /dev/null; then \
		hey -n 1000 -c 10 http://localhost:8080/v1/venues/search; \
	else \
		echo "hey not found. Install from: https://github.com/rakyll/hey"; \
	fi

# API documentation
docs:
	@echo "Generating API documentation..."
	@if command -v swag > /dev/null; then \
		swag init; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi
	@echo "Documentation generated"

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		go generate ./...; \
	else \
		echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"; \
	fi
	@echo "Mocks generated"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf coverage/
	@rm -rf tmp/
	@go clean -testcache
	@echo "Clean complete"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install gotest.tools/gotestsum@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Development tools installed"

# CI/CD helpers
ci-test: deps lint test-all
	@echo "CI tests complete"

ci-build: deps build
	@echo "CI build complete"

# Quick development setup
quick-setup: deps install-tools setup-db migrate seed
	@echo "Quick setup complete! Run 'make dev' to start developing"

# Production deployment helpers
prod-build:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Production build complete"

# Health check
health:
	@echo "Checking application health..."
	@curl -f http://localhost:8080/v1/utils/health || echo "Application not responding"

# Show project statistics
stats:
	@echo "Project Statistics:"
	@echo "Go files: $(shell find . -name "*.go" -not -path "./vendor/*" | wc -l)"
	@echo "Test files: $(shell find . -name "*_test.go" | wc -l)"
	@echo "Lines of code: $(shell find . -name "*.go" -not -path "./vendor/*" -exec cat {} \; | wc -l)"
	@echo "Git commits: $(shell git rev-list --count HEAD 2>/dev/null || echo 'N/A')"

# Release preparation
release: clean lint test-all prod-build
	@echo "Release preparation complete"

# Default environment check
env-check:
	@echo "Checking environment..."
	@echo "Go version: $(shell go version)"
	@echo "PostgreSQL available: $(shell pg_config --version 2>/dev/null || echo 'Not found')"
	@echo "Docker available: $(shell docker --version 2>/dev/null || echo 'Not found')"
	@if [ -f "local.env" ]; then \
		echo "local.env: Found"; \
	else \
		echo "local.env: Not found (copy from local.env.example)"; \
	fi
