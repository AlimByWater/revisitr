# Фаза 0: Фундамент

> Цель: рабочий скелет проекта — код компилируется, деплоится, запускается пустой.

## Статус: ⏳ Не начата

---

## Задачи

### 0.1 Инициализация репозитория
- [x] Создать структуру директорий
- [x] CLAUDE.md
- [x] docs/architecture.md
- [x] .gitignore (Go + Node + Docker + IDE)
- [ ] git init, первый коммит
- [ ] GitHub remote, push

### 0.2 Backend: скелет
- [ ] `go.mod` (module revisitr)
- [ ] `cmd/server/main.go` — минимальный Gin-сервер с health check
- [ ] `cmd/bot/main.go` — минимальный telego-бот с /start
- [ ] `internal/application/app.go` — lifecycle (Init/Run/Shutdown)
- [ ] `internal/application/env/` — загрузка .env (без Vault для MVP)
- [ ] `internal/application/config/` — config structs
- [ ] `internal/controller/http/http.go` — Gin setup, group interface
- [ ] `internal/controller/http/middleware/` — CORS, recovery, logging, JWT auth
- [ ] `internal/repository/postgres/postgres.go` — sqlx connect, schema create
- [ ] `internal/repository/redis/redis.go` — go-redis connect
- [ ] `internal/entity/` — базовые entity (User, Organization)
- [ ] Health check endpoint: `GET /healthz`
- [ ] Structured logging (slog)

### 0.3 Frontend: скелет
- [ ] `package.json` с зависимостями (React, TanStack Router/Query, shadcn/ui, Tailwind, Vite)
- [ ] Vite config + TanStack Router plugin
- [ ] `tsconfig.json`
- [ ] Tailwind config + shadcn/ui init
- [ ] Root layout (`__root.tsx`) с базовым sidebar + header
- [ ] Auth layout (login page — заглушка)
- [ ] Dashboard layout с пустым контентом
- [ ] API client (`lib/api.ts`) с proxy на backend
- [ ] `.env.example`

### 0.4 Infrastructure: local dev
- [x] `infra/docker-compose.yml` — PostgreSQL 16 + Redis 7
- [x] `infra/postgres/init.sql` — CREATE DATABASE revisitr
- [x] `.env.example` для всего проекта
- [x] `Makefile` в корне (dev-up, dev-down, backend-run, bot-run, frontend-dev)

### 0.5 Migrations
- [ ] Установка goose
- [ ] Первая миграция: `users` + `organizations` таблицы
- [ ] Интеграция goose в `cmd/server` как subcommand (или отдельно)

### 0.6 Docker: production builds
- [ ] `backend/Dockerfile` — multi-stage (builder + alpine runtime)
- [ ] `backend/Dockerfile.bot` — multi-stage для бота
- [ ] `frontend/Dockerfile` — multi-stage (node builder + nginx)
- [ ] `frontend/nginx.conf` — SPA routing, health check
- [ ] `infra/docker-compose.prod.yml` — Traefik + backend + bot + frontend + PG + Redis

### 0.7 CI/CD
- [ ] `.github/workflows/backend.yml` — lint, test, build, deploy (path filter: backend/**)
- [ ] `.github/workflows/bot.yml` — lint, test, build, deploy (path filter: backend/** с bot-specific)
- [ ] `.github/workflows/frontend.yml` — lint, test, build, deploy (path filter: frontend/**)
- [ ] `.github/workflows/infrastructure.yml` — migrations (path filter: backend/migrations/**)
- [ ] `scripts/deploy.sh` — deploy backend|bot|frontend|infra
- [ ] Self-hosted runner настроен и работает

### 0.8 Traefik
- [ ] `infra/traefik/traefik.yml` — entrypoints, Let's Encrypt (если нужен SSL)
- [ ] Routing rules:
  - `elysium.fm/revisitr/api/*` → backend:8080
  - `elysium.fm/revisitr/*` → frontend:80
- [ ] Health check labels для всех сервисов

---

## Критерии завершения

- [ ] `docker compose up` локально поднимает PG + Redis
- [ ] Backend стартует, отвечает на `/healthz`
- [ ] Bot стартует, отвечает на `/start` в Telegram
- [ ] Frontend стартует, показывает layout с sidebar
- [ ] Push в main → GitHub Actions → деплой на сервер
- [ ] `elysium.fm/revisitr/` показывает фронтенд
- [ ] `elysium.fm/revisitr/api/healthz` возвращает 200
