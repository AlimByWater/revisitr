# Revisitr

SaaS-платформа лояльности для HoReCa (бары, кафе, рестораны) на базе Telegram.

## Stack

- **Backend**: Go 1.23+, Gin, sqlx/PostgreSQL, go-redis, telego (Telegram bot)
- **Frontend**: React, TanStack (Router + Query + Table + Form), shadcn/ui, Tailwind, Vite
- **Infrastructure**: Docker Compose, PostgreSQL 16, Redis 7, GitHub Actions
- **Testing**: Playwright (E2E), Vitest (frontend), Go standard testing (backend + bot), Telethon (bot E2E)

## Services

| Service | Описание | Команда |
|---------|----------|---------|
| **server** | REST API для админ-панели (`/api/v1/`) | `make backend-run` |
| **bot** | Telegram-бот для конечных пользователей | `make bot-run` |
| **frontend** | React SPA (админ-панель) | `make frontend-dev` |

## Quick Start

```bash
# 1. Запустить инфраструктуру (PostgreSQL + Redis)
make dev-up

# 2. Накатить миграции
DATABASE_URL="postgres://revisitr:devpassword@localhost:6281/revisitr?sslmode=disable" make migrate-up

# 3. Запустить backend
make backend-run

# 4. (в другом терминале) Запустить frontend
make frontend-dev
```

Порты: backend `:9721`, frontend `:5921`, PostgreSQL `:6281`, Redis `:7392`.

## Architecture

```
backend/
  cmd/server/      — API-сервер
  cmd/bot/         — Telegram-бот
  internal/
    application/   — bootstrap, config
    controller/    — HTTP handlers (Gin)
    usecase/       — бизнес-логика
    repository/    — PostgreSQL, Redis
    entity/        — доменные модели
  migrations/      — goose SQL-миграции
frontend/
  src/             — React-приложение (file-based routing)
telegram/          — Telethon-инфраструктура для E2E тестов бота
infra/             — Docker Compose, nginx, скрипты деплоя
```

## Документация

- `docs/testing.md` — полная спецификация тестирования
- `docs/userdocs/` — дизайн-документ, Figma-анализ
- `docs/e2e/` — план E2E тестов (Playwright)
- `.ai-memory/` — shared knowledge для AI агентов

## Команды

| Команда | Описание |
|---------|----------|
| `make dev-up` | Запустить PostgreSQL + Redis |
| `make dev-down` | Остановить инфру |
| `make backend-run` | Запустить API-сервер |
| `make bot-run` | Запустить Telegram-бот |
| `make frontend-dev` | Запустить Vite dev server |
| `make migrate-up` | Накатить миграции |
| `make test` | Все unit-тесты |
| `make test-all` | Unit + integration |
| `make lint` | Линтинг (go vet + eslint) |
| `make build-backend` | Сборка Go бинарников |
| `make docker-prod` | Запуск production сборки |
