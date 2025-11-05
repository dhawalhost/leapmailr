# Makefile for LeapMailR
# Simplifies common development tasks

.PHONY: help setup build run test clean docker-up docker-down deploy-staging deploy-prod

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := leapmailr
DOCKER_COMPOSE := docker-compose -f docker-compose/docker-compose.yml
GO_FILES := $(shell find . -name '*.go' -type f)
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Colors for help
CYAN := \033[0;36m
RESET := \033[0m

## help: Show this help message
help:
	@echo "$(CYAN)LeapMailR - Available Commands$(RESET)"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""

## setup: Run initial setup (install dependencies, create config)
setup:
	@echo "Setting up LeapMailR..."
	@chmod +x setup.sh
	@./setup.sh

## install: Install Go dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@go mod tidy

## build: Build the application
build: $(GO_FILES)
	@echo "Building $(APP_NAME)..."
	@go build $(LDFLAGS) -o $(APP_NAME) .
	@echo "Build complete: ./$(APP_NAME)"

## build-linux: Build for Linux (amd64)
build-linux:
	@echo "Building for Linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-linux-amd64 .

## build-windows: Build for Windows (amd64)
build-windows:
	@echo "Building for Windows..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-windows-amd64.exe .

## build-mac: Build for macOS (amd64 and arm64)
build-mac:
	@echo "Building for macOS (Intel)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-darwin-amd64 .
	@echo "Building for macOS (Apple Silicon)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(APP_NAME)-darwin-arm64 .

## build-all: Build for all platforms
build-all: build-linux build-windows build-mac
	@echo "All builds complete!"

## run: Run the application
run:
	@echo "Starting $(APP_NAME)..."
	@go run .

## dev: Run with hot reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out
	@echo ""
	@echo "Generate HTML report with: make coverage-html"

## coverage-html: Generate HTML coverage report
coverage-html: test-coverage
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(APP_NAME) $(APP_NAME)-*
	@rm -f coverage.out coverage.html
	@rm -rf logs/*.log
	@echo "Clean complete!"

## db-create: Create PostgreSQL database
db-create:
	@echo "Creating database..."
	@psql -U postgres -c "CREATE USER leapmailr WITH PASSWORD 'leapmailr_dev_123';" 2>/dev/null || true
	@psql -U postgres -c "CREATE DATABASE leapmailr OWNER leapmailr;" 2>/dev/null || true
	@psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE leapmailr TO leapmailr;"
	@echo "Database created!"

## db-drop: Drop PostgreSQL database
db-drop:
	@echo "Dropping database..."
	@psql -U postgres -c "DROP DATABASE IF EXISTS leapmailr;"
	@echo "Database dropped!"

## db-reset: Reset database (drop and create)
db-reset: db-drop db-create
	@echo "Database reset complete!"

## db-migrate: Run database migrations (if available)
db-migrate:
	@echo "Running migrations..."
	@# Add migration command here
	@echo "Migrations complete!"

## backup: Create database backup
backup:
	@echo "Creating backup..."
	@chmod +x scripts/backup.sh
	@./scripts/backup.sh

## restore: Restore database from backup
restore:
	@echo "Restoring from backup..."
	@chmod +x scripts/restore.sh
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Usage: make restore BACKUP_FILE=./backups/leapmailr_backup_YYYYMMDD_HHMMSS.sql.gz"; \
	else \
		./scripts/restore.sh $(BACKUP_FILE); \
	fi

## secrets-rotate: Rotate secrets
secrets-rotate:
	@echo "Rotating secrets..."
	@chmod +x scripts/rotate-secrets.sh
	@./scripts/rotate-secrets.sh $(TYPE)

## docker-build: Build Docker images
docker-build:
	@echo "Building Docker images..."
	@docker build -t $(APP_NAME):latest -t $(APP_NAME):$(VERSION) .

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting Docker services..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Services started!"
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"

## docker-down: Stop all Docker services
docker-down:
	@echo "Stopping Docker services..."
	@$(DOCKER_COMPOSE) down

## docker-logs: View Docker logs
docker-logs:
	@$(DOCKER_COMPOSE) logs -f

## docker-restart: Restart Docker services
docker-restart: docker-down docker-up

## docker-clean: Remove all Docker containers, images, and volumes
docker-clean:
	@echo "Cleaning Docker resources..."
	@$(DOCKER_COMPOSE) down -v --rmi all
	@echo "Docker cleanup complete!"

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@go version

## deps: Show dependency tree
deps:
	@go mod graph

## update-deps: Update all dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

## security-scan: Run security vulnerability scan
security-scan:
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

## profile-cpu: Run CPU profiling
profile-cpu:
	@echo "Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./...
	@echo "View profile with: go tool pprof cpu.prof"

## profile-mem: Run memory profiling
profile-mem:
	@echo "Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./...
	@echo "View profile with: go tool pprof mem.prof"

## generate: Run go generate
generate:
	@echo "Running go generate..."
	@go generate ./...

## mod-verify: Verify dependencies
mod-verify:
	@echo "Verifying dependencies..."
	@go mod verify

## todo: Show TODO comments in code
todo:
	@echo "TODO items:"
	@grep -rn "TODO" --include="*.go" . || echo "No TODOs found!"

## loc: Count lines of code
loc:
	@echo "Lines of code:"
	@find . -name '*.go' -type f | xargs wc -l | tail -1

## deploy-staging: Deploy to staging environment
deploy-staging:
	@echo "Deploying to staging..."
	@kubectl set image deployment/backend backend=your-registry/$(APP_NAME):$(VERSION) -n leapmailr-staging
	@kubectl rollout status deployment/backend -n leapmailr-staging

## deploy-prod: Deploy to production environment
deploy-prod:
	@echo "Deploying to production..."
	@echo "⚠️  This will deploy to PRODUCTION. Continue? [y/N]"
	@read -r response; \
	if [ "$$response" = "y" ]; then \
		kubectl set image deployment/backend backend=your-registry/$(APP_NAME):$(VERSION) -n leapmailr; \
		kubectl rollout status deployment/backend -n leapmailr; \
	else \
		echo "Deployment cancelled."; \
	fi

## health: Check application health
health:
	@echo "Checking application health..."
	@curl -f http://localhost:8080/health || echo "Application not running"

## metrics: View application metrics
metrics:
	@curl -s http://localhost:8080/metrics

## logs: Tail application logs
logs:
	@tail -f logs/application.log

## start-all: Start backend and frontend
start-all:
	@echo "Starting backend in background..."
	@./$(APP_NAME) &
	@echo "Backend started (PID: $$!)"
	@sleep 3
	@if [ -d "../leapmailr-ui" ]; then \
		echo "Starting frontend..."; \
		cd ../leapmailr-ui && npm run dev; \
	fi

## ci: Run CI checks locally
ci: fmt vet lint test-coverage
	@echo "✓ All CI checks passed!"

## pre-commit: Run pre-commit checks
pre-commit: fmt vet lint
	@echo "✓ Pre-commit checks passed!"
