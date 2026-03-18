# Revisitr

SaaS-платформа лояльности для HoReCa (бары, кафе, рестораны) на базе Telegram.

## Stack

- **Backend**: Go 1.23+, Gin, sqlx/PostgreSQL, go-redis, telego (Telegram bot)
- **Frontend**: React, TanStack (Router + Query + Table + Form), shadcn/ui, Tailwind, Vite
- **Infrastructure**: Docker Compose, Traefik, PostgreSQL 16, Redis 7, GitHub Actions (self-hosted runner)
- **Hosting**: Dedicated server, domain elysium.fm/revisitr

## Architecture

Clean Architecture (адаптировано из driptech):

```
backend/
  cmd/server/   — API-сервер (админ-панель)
  cmd/bot/      — Telegram-бот (отдельный сервис)
  internal/
    application/ — bootstrap, config, env
    controller/  — HTTP handlers (Gin), scheduler
    usecase/     — бизнес-логика
    repository/  — PostgreSQL, Redis
    entity/      — доменные модели
  migrations/    — goose SQL-миграции
```

Два сервиса в одном репозитории:
1. **server** — REST API для фронтенда (админка)
2. **bot** — Telegram-бот для конечных пользователей (telego)

## Commands

```bash
# Local development
cd infra && docker compose up -d                    # PostgreSQL + Redis
cd backend && go run ./cmd/server                   # API server
cd backend && go run ./cmd/bot                      # Telegram bot
cd frontend && npm run dev                          # Frontend dev server

# Build
cd backend && go build -o bin/server ./cmd/server
cd backend && go build -o bin/bot ./cmd/bot
cd frontend && npm run build

# Test
cd backend && go test ./...
cd frontend && npm run test

# Migrations
cd backend && goose -dir migrations postgres "$DATABASE_URL" up
cd backend && goose -dir migrations postgres "$DATABASE_URL" status

# Lint
cd backend && go vet ./... && staticcheck ./...
cd frontend && npm run lint

# Docker
docker compose -f infra/docker-compose.yml up -d          # dev
docker compose -f infra/docker-compose.prod.yml up -d      # prod
```

## Conventions

- **Go**: стандартный Go-стиль, slog для логирования, явный DI (без фреймворков)
- **Frontend**: file-based routing (TanStack Router), feature-based организация
- **API**: RESTful, JSON, префикс `/api/v1/`
- **Миграции**: goose, формат `YYYYMMDDHHMMSS_description.sql`
- **Коммиты**: conventional commits (feat/fix/refactor/docs/chore)
- **Ветки**: feature branches → PR → main

## Project Docs

- `userdocs/` — дизайн-документ, презентация, Figma-анализ
- `docs/phases/` — план разработки по фазам
- `docs/architecture.md` — архитектурные решения

## Figma

- File: `FaPibi9jR7Lqns9FGLcVIc` (REVISITR-DOC)
- Спроектировано: auth flow, layout (header+sidebar), боты (CRUD)
- Стиль: минимализм, ч/б + noise grain, акцент красный/оранжевый
