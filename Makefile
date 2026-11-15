.PHONY: help setup-env build up down logs ps test-backend vet-backend lint-backend vulncheck-backend generate-mocks verify-mocks test-integration-gateway test-integration load-test test-frontend build-frontend dev-frontend test

SERVICES := api-gateway merchant-service notification-service product-service transaction-service user-service warehouse-service pkg/observability
MOCK_SERVICES := notification-service product-service transaction-service user-service warehouse-service

help:
	@echo "FMCG Multi-Warehouse & POS Platform - Available Commands:"
	@echo ""
	@echo "  make setup-env       Copy env.example -> .env across all backend services & frontend"
	@echo "  make build           Build all backend Docker images"
	@echo "  make up              Start full backend stack (PostgreSQL x6, Redis, RabbitMQ, 7 Go services)"
	@echo "  make down            Stop Docker Compose stack"
	@echo "  make ps              Show running Docker containers"
	@echo "  make logs            Tail Docker Compose logs"
	@echo "  make test-backend       Run go test -race across all 7 Go microservices"
	@echo "  make vet-backend        Run go vet across all 7 Go microservices"
	@echo "  make lint-backend       Run golangci-lint across all 7 Go microservices"
	@echo "  make vulncheck-backend  Run govulncheck vulnerability scan across all 7 Go microservices"
	@echo "  make generate-mocks     Regenerate GoMock files via go generate"
	@echo "  make verify-mocks       Ensure generated GoMock files match repository interfaces"
	@echo "  make test-integration-gateway  Run API Gateway HTTP integration tests with -race (no Docker needed)"
	@echo "  make test-integration          Run E2E integration tests against running Docker stack"
	@echo "  make load-test                 Run k6 load test against API Gateway"
	@echo "  make dev-frontend       Run Vue 3 Vite dev server"
	@echo "  make build-frontend     Build Vue 3 production bundle"
	@echo "  make test               Run backend tests (-race), vet, golangci-lint, frontend tests & build"

setup-env:
	@for svc in $(SERVICES); do \
		if [ -f backend/$$svc/env.example ] && [ ! -f backend/$$svc/.env ]; then \
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
		echo "=== Testing $$svc (-race) ==="; \
		(cd backend/$$svc && go test -race -v ./...) || exit 1; \
	done

vet-backend:
	@for svc in $(SERVICES); do \
		echo "=== Vetting $$svc ==="; \
		(cd backend/$$svc && go vet ./...) || exit 1; \
	done

lint-backend:
	@for svc in $(SERVICES); do \
		echo "=== Linting $$svc ==="; \
		(cd backend/$$svc && golangci-lint run --config $(CURDIR)/.golangci.yml ./...) || exit 1; \
	done

vulncheck-backend:
	@for svc in $(SERVICES); do \
		echo "=== Scanning vulnerabilities $$svc ==="; \
		(cd backend/$$svc && govulncheck ./...) || exit 1; \
	done

generate-mocks:
	@for svc in $(MOCK_SERVICES); do \
		echo "=== Generating mocks for $$svc ==="; \
		(cd backend/$$svc && go generate ./...) || exit 1; \
	done

verify-mocks: generate-mocks
	git diff --exit-code -- 'backend/*/mocks'
	@test -z "$$(git status --porcelain -- 'backend/*/mocks')" || (echo "Untracked mock files found"; git status --short -- 'backend/*/mocks'; exit 1)

dev-frontend:
	cd frontend && npm run dev

test-integration-gateway:
	cd backend/api-gateway && go test -tags=integration -race -v ./...

test-integration:
	cd backend/tests/integration && go test -tags=integration -v ./...

load-test:
	k6 run backend/tests/load/smoke_and_pos.js

test-frontend:
	cd frontend && npm test

build-frontend:
	cd frontend && npm run build

test: test-backend vet-backend lint-backend test-frontend build-frontend
