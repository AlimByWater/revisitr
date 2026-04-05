# Revisitr

SaaS-платформа лояльности для HoReCa (бары, кафе, рестораны) на базе Telegram.

## Stack

- **Backend**: Go 1.23+, Gin, sqlx/PostgreSQL, go-redis, telego (Telegram bot)
- **Frontend**: React, TanStack (Router + Query + Table + Form), shadcn/ui, Tailwind, Vite
- **Infrastructure**: Docker Compose, host nginx, PostgreSQL 16, Redis 7, GitHub Actions (self-hosted runner)
- **Hosting**: Dedicated server, domain elysium.fm/revisitr (nginx routing, no Traefik)

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

# Test — unit (быстро, без инфры)
cd backend && go test -race ./internal/usecase/...     # backend unit
cd frontend && npx vitest run                          # frontend unit

# Test — integration (требует docker compose up)
cd backend && go test -race -tags=integration ./tests/integration/...

# Test — E2E (требует full stack: docker + backend + frontend)
cd e2e && npx playwright test

# Test — all
make test-all                                          # unit + integration + e2e
make test-quick                                        # unit only

# Migrations
cd backend && goose -dir migrations postgres "$DATABASE_URL" up
cd backend && goose -dir migrations postgres "$DATABASE_URL" status

# Lint
cd backend && go vet ./... && staticcheck ./...
cd frontend && npm run lint

# Docker (local dev)
make dev-up                                                # PG + Redis
make dev-down                                              # stop

# Docker (production) — на сервере в /opt/revisitr/infra/
docker compose -f docker-compose.prod.yml up -d            # all services
infra/scripts/deploy.sh backend|bot|frontend|infra               # independent deploy
```

## Ports

| Service    | Local dev | Production (host → container) |
|------------|-----------|-------------------------------|
| Backend    | 8080      | 8090 → 8080                   |
| Frontend   | 5173      | 3340 → 80                     |
| PostgreSQL | 5433      | 5433 → 5432                   |
| Redis      | 6380      | 6380 → 6379                   |

Production nginx routes: `/revisitr/api/*` → `:8090`, `/revisitr/*` → `:3340`

## Conventions

- **Go**: стандартный Go-стиль, slog для логирования, явный DI (без фреймворков)
- **Frontend**: file-based routing (TanStack Router), feature-based организация
- **API**: RESTful, JSON, префикс `/api/v1/`
- **Миграции**: goose, формат `YYYYMMDDHHMMSS_description.sql`
- **Коммиты**: conventional commits (feat/fix/refactor/docs/chore)
- **Ветки**: feature branches → PR → main

## Testing

Полная спецификация: `docs/testing.md`

### Уровни тестирования

| Уровень       | Инструмент          | Scope                              | Требует инфру |
|---------------|---------------------|------------------------------------|---------------|
| Backend unit  | Go `testing`        | usecase-слой, ручные моки          | Нет           |
| Frontend unit | Vitest + MSW        | компоненты, stores, API hooks      | Нет           |
| Integration   | Go `httptest` + DB  | API handlers + real DB             | Docker        |
| E2E           | Playwright          | полные user flows через браузер    | Full stack    |
| Bot unit      | Go `testing`        | bot handlers, mock telego          | Нет           |
| Bot integ.    | Go `httptest` + DB  | webhook → usecase → DB             | Docker        |

### Паттерны

- **Backend unit**: ручные моки через struct с function fields (установленный паттерн)
- **Integration**: build tag `//go:build integration`, тестовая БД `revisitr_test`
- **Frontend**: Vitest + @testing-library/react + MSW для API-моков
- **E2E**: Playwright, `e2e/` директория в корне проекта
- **Тесты бота**: отдельно от server, mock telego.Bot через интерфейс

### Claude Code MCP для тестирования

- **Playwright MCP**: exploratory E2E, UI snapshot, accessibility проверки
- **Claude Preview**: `preview_start` → `preview_screenshot` → visual verification
- **Postgres MCP**: прямые SQL-запросы для проверки данных после тестов
- **Bash**: `go test`, `npx vitest`, `npx playwright test`

## Project Docs

- `docs/userdocs/` — дизайн-документ, презентация, Figma-анализ
- `docs/e2e/` — план E2E тестов (Playwright)

## AI Agent Memory

Shared project knowledge for AI agents is in `.ai-memory/`:

- `patterns.md` — established code patterns (backend + frontend)
- `gotchas.md` — known pitfalls, common bugs, deployment traps
- `decisions.md` — key architecture and business logic decisions
- `status.md` — current dev status, phases, pending work
- `testing.md` — test coverage map, testing patterns

Read `.ai-memory/` files when working on non-trivial tasks to avoid known pitfalls.

## Figma

- File: `FaPibi9jR7Lqns9FGLcVIc` (REVISITR-DOC)
- Спроектировано: auth flow, layout (header+sidebar), боты (CRUD)
- Стиль: минимализм, ч/б + noise grain, акцент красный/оранжевый
