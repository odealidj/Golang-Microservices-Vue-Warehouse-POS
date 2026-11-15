.PHONY: help setup-env build up down logs ps test-backend vet-backend test-integration-gateway test-integration test-frontend build-frontend dev-frontend test

SERVICES := api-gateway merchant-service notification-service product-service transaction-service user-service warehouse-service pkg/observability

help:
	@echo "FMCG Multi-Warehouse & POS Platform - Available Commands:"
	@echo ""
	@echo "  make setup-env       Copy env.example -> .env across all backend services & frontend"
	@echo "  make build           Build all backend Docker images"
	@echo "  make up              Start full backend stack (PostgreSQL x6, Redis, RabbitMQ, 7 Go services)"
	@echo "  make down            Stop Docker Compose stack"
	@echo "  make ps              Show running Docker containers"
	@echo "  make logs            Tail Docker Compose logs"
	@echo "  make test-backend    Run go test across all 7 Go microservices"
	@echo "  make vet-backend     Run go vet across all 7 Go microservices"
	@echo "  make test-integration-gateway  Run API Gateway HTTP integration tests (no Docker needed)"
	@echo "  make test-integration          Run E2E integration tests against running Docker stack"
	@echo "  make dev-frontend    Run Vue 3 Vite dev server"
	@echo "  make build-frontend  Build Vue 3 production bundle"
	@echo "  make test            Run all backend tests/vet and frontend production build"

setup-env:
	@for svc in $(SERVICES); do \
		if [ ! -f backend/$$svc/.env ]; then \
			cp backend/$$svc/env.example backend/$$svc/.env && echo "Created backend/$$svc/.env"; \
		fi; \
	done
	@if [ ! -f frontend/.env ]; then \
		cp frontend/.env.example frontend/.env && echo "Created frontend/.env"; \
	fi

build:
	cd backend && docker compose build

up:
	cd backend && docker compose up -d

down:
	cd backend && docker compose down

ps:
	cd backend && docker compose ps

logs:
	cd backend && docker compose logs -f

test-backend:
	@for svc in $(SERVICES); do \
		echo "=== Testing $$svc ==="; \
		(cd backend/$$svc && go test -v ./...) || exit 1; \
	done

vet-backend:
	@for svc in $(SERVICES); do \
		echo "=== Vetting $$svc ==="; \
		(cd backend/$$svc && go vet ./...) || exit 1; \
	done

dev-frontend:
	cd frontend && npm run dev

test-integration-gateway:
	cd backend/api-gateway && go test -tags=integration -v ./...

test-integration:
	cd backend/tests/integration && go test -tags=integration -v ./...

test-frontend:
	cd frontend && npm test

build-frontend:
	cd frontend && npm run build

test: test-backend vet-backend test-frontend build-frontend
