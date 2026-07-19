# Revisitr

SaaS-платформа лояльности для HoReCa (бары, кафе, рестораны) на базе Telegram.

## Stack

- **Backend**: Go 1.26+, Gin, sqlx/PostgreSQL, go-redis, telego (Telegram bot)
- **Frontend**: React, TanStack (Router + Query + Table + Form), shadcn/ui, Tailwind, Vite
- **Infrastructure**: Docker Compose, PostgreSQL 16, Redis 7

## Быстрый старт

```bash
# 1. Инфраструктура
make dev-up

# 2. Окружение
cp .env.example .env

# 3. Миграции
source .env && cd backend && goose -dir migrations postgres "$DATABASE_URL" up

# 4. Бэкенд
make backend-run

# 5. Фронтенд
make frontend-dev
```

## Порты (локальная разработка)

| Сервис    | Порт  |
|-----------|-------|
| Frontend  | 5921  |
| Backend   | 9721  |
| PostgreSQL| 6281  |
| Redis     | 7392  |

## Структура

```
backend/
  cmd/server/   — API-сервер (админ-панель)
  cmd/bot/      — Telegram-бот
  internal/
    application/ — bootstrap, config
    controller/  — HTTP handlers
    usecase/     — бизнес-логика
    repository/  — PostgreSQL, Redis
    entity/      — доменные модели
frontend/       — React SPA
infra/          — Docker Compose
migrations/     — goose SQL-миграции
```
