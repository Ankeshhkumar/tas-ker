-include .env
export
BACKEND_PORT ?= 8080
FRONTEND_PORT ?= 3000

.DEFAULT_GOAL := help
.PHONY: help setup up down logs ps reset smoke test

help: ## Show this help
	grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-8s %s\n", $$1, $$2}'

setup: ## Create .env from .env.example if missing
	test -f .env || (cp .env.example .env && echo "Created .env")

up: setup ## Build and start everything
	docker compose up -d --build
	echo "Open http://localhost:$(FRONTEND_PORT)"

down: ## Stop everything (keeps data)
	docker compose down

logs: ## Follow logs
	docker compose logs -f

ps: ## Show running containers
	docker compose ps

reset: ## Stop and delete all data
	docker compose down -v

smoke: ## Check the API and the frontend proxy respond
	curl -fsS http://localhost:$(BACKEND_PORT)/health && echo
	curl -fsS http://localhost:$(BACKEND_PORT)/projects > /dev/null && echo "projects ok"
	curl -fsS http://localhost:$(FRONTEND_PORT)/api/projects > /dev/null && echo "frontend proxy ok"

test: ## Run backend and frontend tests
	cd backend && go test ./...
	cd frontend && npm ci && npm run lint && npm test
