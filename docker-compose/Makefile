# LeapMailr - Makefile
# Convenient shortcuts for Docker operations

.PHONY: help
help: ## Show this help message
	@echo "LeapMailr - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: verify
verify: ## Verify Docker setup
	@echo "Verifying setup..."
	@chmod +x verify-docker.sh
	@./verify-docker.sh

.PHONY: setup
setup: verify ## Initial setup (creates .env from example)
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✓ Created .env file"; \
		echo "⚠ Please edit .env with your configuration"; \
	fi

.PHONY: up
up: ## Start all services
	docker-compose up -d
	@echo ""
	@echo "✓ Services started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:8080"

.PHONY: down
down: ## Stop all services
	docker-compose down

.PHONY: restart
restart: ## Restart all services
	docker-compose restart

.PHONY: logs
logs: ## View logs (all services)
	docker-compose logs -f

.PHONY: logs-backend
logs-backend: ## View backend logs
	docker-compose logs -f backend

.PHONY: logs-frontend
logs-frontend: ## View frontend logs
	docker-compose logs -f frontend

.PHONY: logs-db
logs-db: ## View database logs
	docker-compose logs -f postgres

.PHONY: ps
ps: ## Show service status
	docker-compose ps

.PHONY: build
build: ## Rebuild all services
	docker-compose build

.PHONY: rebuild
rebuild: ## Rebuild and restart all services
	docker-compose up -d --build

.PHONY: rebuild-backend
rebuild-backend: ## Rebuild and restart backend only
	docker-compose up -d --build backend

.PHONY: rebuild-frontend
rebuild-frontend: ## Rebuild and restart frontend only
	docker-compose up -d --build frontend

.PHONY: clean
clean: ## Stop and remove containers
	docker-compose down

.PHONY: clean-all
clean-all: ## Stop and remove containers, volumes, and images
	@echo "⚠ WARNING: This will delete all data!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose down -v; \
		docker system prune -f; \
		echo "✓ Cleaned everything"; \
	else \
		echo "Cancelled"; \
	fi

.PHONY: db-shell
db-shell: ## Access database shell
	docker-compose exec postgres psql -U leapmailr -d leapmailr

.PHONY: db-backup
db-backup: ## Backup database
	@mkdir -p backups
	docker-compose exec postgres pg_dump -U leapmailr leapmailr > backups/backup-$$(date +%Y%m%d-%H%M%S).sql
	@echo "✓ Database backed up to backups/"

.PHONY: db-restore
db-restore: ## Restore database from latest backup
	@LATEST=$$(ls -t backups/*.sql | head -1); \
	if [ -z "$$LATEST" ]; then \
		echo "No backup files found"; \
		exit 1; \
	fi; \
	echo "Restoring from $$LATEST"; \
	docker-compose exec -T postgres psql -U leapmailr leapmailr < $$LATEST; \
	echo "✓ Database restored"

.PHONY: backend-shell
backend-shell: ## Access backend container shell
	docker-compose exec backend sh

.PHONY: frontend-shell
frontend-shell: ## Access frontend container shell
	docker-compose exec frontend sh

.PHONY: test-backend
test-backend: ## Run backend tests
	cd .. && go test ./...

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	cd ../../leapmailr-ui && npm test

.PHONY: dev-backend
dev-backend: ## Run backend locally (requires DB running)
	@echo "Starting backend in development mode..."
	@echo "Make sure database is running: make up-db"
	cd .. && go run main.go

.PHONY: dev-frontend
dev-frontend: ## Run frontend locally (requires backend running)
	@echo "Starting frontend in development mode..."
	@echo "Make sure backend is running"
	cd ../../leapmailr-ui && npm run dev

.PHONY: up-db
up-db: ## Start only database
	docker-compose up -d postgres

.PHONY: health
health: ## Check service health
	@echo "Checking backend health..."
	@curl -s http://localhost:8080/api/v1/health | jq . || echo "Backend not responding"
	@echo ""
	@echo "Checking frontend..."
	@curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 || echo "Frontend not responding"
	@echo ""
	@echo "Checking database..."
	@docker-compose exec postgres pg_isready -U leapmailr

.PHONY: stats
stats: ## Show resource usage
	docker stats --no-stream

.PHONY: install-backend
install-backend: ## Install backend dependencies
	cd .. && go mod download

.PHONY: install-frontend
install-frontend: ## Install frontend dependencies
	cd ../../leapmailr-ui && npm install

.PHONY: format-backend
format-backend: ## Format backend code
	cd .. && go fmt ./...

.PHONY: format-frontend
format-frontend: ## Format frontend code
	cd ../../leapmailr-ui && npm run lint --fix

.PHONY: prod-build
prod-build: ## Build production images
	docker build -t leapmailr-backend:latest ..
	docker build -t leapmailr-frontend:latest ../../leapmailr-ui
	@echo "✓ Production images built"

.PHONY: prod-up
prod-up: ## Start production environment
	docker-compose -f docker-compose.yml up -d
	@echo "✓ Production environment started"

# Default target
.DEFAULT_GOAL := help
