.PHONY: help setup build up down logs clean rebuild

# Load environment variables from .env
include .env
export

help: ## Show this help message
	@echo "🐳 Docker Commands for VEN API"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "📚 Documentation: QUICKSTART.md or DOCKER_SETUP.md"

setup: ## Initial setup - creates .env from template
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		chmod +x build.sh setup.sh; \
		echo "✅ Created .env file"; \
		echo "⚠️  Please edit .env and add your GITHUB_TOKEN"; \
		echo "   Get token: https://github.com/settings/tokens/new"; \
	else \
		echo "⚠️  .env already exists"; \
	fi

check-token: ## Check if GitHub token is configured
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ GITHUB_TOKEN not set in .env"; \
		echo "   Run: make setup"; \
		exit 1; \
	else \
		echo "✅ GitHub token configured (length: $$(echo $(GITHUB_TOKEN) | wc -c) chars)"; \
	fi

build: check-token ## Build Docker images
	@echo "🔨 Building Docker images..."
	docker compose build

up: check-token ## Start services (foreground)
	@echo "🚀 Starting services..."
	@echo "   API:       http://localhost:8000"
	@echo "   MailHog:   http://localhost:8025"
	@echo "   MinIO:     http://localhost:9001"
	docker compose up

up-d: check-token ## Start services (background/detached)
	@echo "🚀 Starting services in background..."
	docker compose up -d
	@echo "✅ Services started"
	@echo "   API:       http://localhost:8000"
	@echo "   MailHog:   http://localhost:8025"
	@echo "   MinIO:     http://localhost:9001"

down: ## Stop services
	@echo "🛑 Stopping services..."
	docker compose down

logs: ## View logs (follow mode)
	docker compose logs -f

logs-api: ## View API logs only
	docker compose logs -f ven-api

ps: ## Show running containers
	docker compose ps

clean: ## Stop services and remove volumes
	@echo "🧹 Cleaning up..."
	docker compose down -v
	@echo "✅ Services stopped and volumes removed"

rebuild: ## Rebuild from scratch (no cache)
	@echo "🔨 Rebuilding from scratch..."
	docker compose build --no-cache

restart: ## Restart all services
	@echo "🔄 Restarting services..."
	docker compose restart

restart-api: ## Restart API service only
	@echo "🔄 Restarting API..."
	docker compose restart ven-api

shell-api: ## Open shell in API container
	docker compose exec ven-api sh

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U ven_user -d ven_api

redis-cli: ## Open Redis CLI
	docker compose exec redis redis-cli -a ven-redis-pass

install: setup build up-d ## Complete installation (setup + build + run)
	@echo "✅ Installation complete!"
	@echo "   API running at: http://localhost:8000"
