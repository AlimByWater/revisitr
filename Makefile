.PHONY: dev-up dev-down backend-run bot-run frontend-dev migrate-up migrate-status lint test

# ─── Local Development ───────────────────────────────────────

dev-up:
	cd infra && docker compose up -d

dev-down:
	cd infra && docker compose down

backend-run:
	cd backend && go run ./cmd/server

bot-run:
	cd backend && go run ./cmd/bot

frontend-dev:
	cd frontend && npm run dev

# ─── Database ────────────────────────────────────────────────

migrate-up:
	cd backend && goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	cd backend && goose -dir migrations postgres "$$DATABASE_URL" down

migrate-status:
	cd backend && goose -dir migrations postgres "$$DATABASE_URL" status

migrate-create:
	cd backend && goose -dir migrations create $(name) sql

# ─── Quality ─────────────────────────────────────────────────

lint:
	cd backend && go vet ./... && staticcheck ./...
	cd frontend && npm run lint

test:
	cd backend && go test -race ./...
	cd frontend && npm run test -- --run

test-backend:
	cd backend && go test -race ./...

test-frontend:
	cd frontend && npm run test -- --run

# Integration tests — requires: cd infra && docker compose up -d
test-integration:
	cd backend && go test -race -tags=integration -v ./tests/integration/...

# All tests (unit + integration). Integration requires docker compose.
test-all: test-backend test-frontend test-integration

# Quick check — unit only, no infra needed
test-quick: test-backend test-frontend

# ─── Build ───────────────────────────────────────────────────

build-backend:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/bot ./cmd/bot

build-frontend:
	cd frontend && npm run build

# ─── Docker ──────────────────────────────────────────────────

docker-build:
	docker build -t revisitr-backend -f backend/Dockerfile backend/
	docker build -t revisitr-bot -f backend/Dockerfile.bot backend/
	docker build -t revisitr-frontend -f frontend/Dockerfile frontend/

docker-prod:
	cd infra && docker compose -f docker-compose.prod.yml up -d

docker-prod-down:
	cd infra && docker compose -f docker-compose.prod.yml down
