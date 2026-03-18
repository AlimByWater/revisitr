# Фаза 0: Фундамент

> Цель: рабочий скелет проекта — код компилируется, деплоится, запускается пустой.

## Статус: 🔄 В процессе

---

## Задачи

### 0.1 Инициализация репозитория
- [x] Создать структуру директорий
- [x] CLAUDE.md
- [x] docs/architecture.md
- [x] .gitignore (Go + Node + Docker + IDE)
- [x] git init, первый коммит
- [ ] GitHub remote, push

### 0.2 Backend: скелет
- [x] `go.mod` (module revisitr)
- [x] `cmd/server/main.go` — минимальный Gin-сервер с health check
- [x] `cmd/bot/main.go` — минимальный telego-бот с /start
- [x] `internal/application/app.go` — lifecycle (Init/Run/Shutdown)
- [x] `internal/application/env/` — загрузка .env (без Vault для MVP)
- [x] `internal/application/config/` — config structs
- [x] `internal/controller/http/http.go` — Gin setup, group interface
- [x] `internal/controller/http/middleware/` — CORS, recovery, logging, JWT auth
- [x] `internal/repository/postgres/postgres.go` — sqlx connect, schema create
- [x] `internal/repository/redis/redis.go` — go-redis connect
- [x] `internal/entity/` — базовые entity (User, Organization)
- [x] Health check endpoint: `GET /healthz`
- [x] Structured logging (slog)

### 0.3 Frontend: скелет
- [x] `package.json` с зависимостями (React, TanStack Router/Query, shadcn/ui, Tailwind, Vite)
- [x] Vite config + TanStack Router plugin
- [x] `tsconfig.json`
- [x] Tailwind config + кастомные цвета (accent, sidebar, surface)
- [x] Root layout (`__root.tsx`) с QueryClient context
- [x] Auth layout (login page с формой)
- [x] Dashboard layout с sidebar + header + Outlet
- [x] API client (`lib/api.ts`) с auth interceptor
- [x] `.env.example`

### 0.4 Infrastructure: local dev
- [x] `infra/docker-compose.yml` — PostgreSQL 16 + Redis 7
- [x] `infra/postgres/init.sql` — расширения (uuid-ossp, pg_trgm)
- [x] `.env.example` для всего проекта
- [x] `Makefile` в корне (dev-up, dev-down, backend-run, bot-run, frontend-dev)

### 0.5 Migrations
- [x] goose как зависимость в go.mod
- [x] Первая миграция: `users` + `organizations` таблицы
- [ ] Интеграция goose в `cmd/server` как subcommand (или отдельно)

### 0.6 Docker: production builds
- [x] `backend/Dockerfile` — multi-stage (builder + alpine runtime)
- [x] `backend/Dockerfile.bot` — multi-stage для бота
- [x] `frontend/Dockerfile` — multi-stage (node builder + nginx)
- [x] `frontend/nginx.conf` — SPA routing, health check, gzip, cache
- [x] `infra/docker-compose.prod.yml` — Traefik + backend + bot + frontend + PG + Redis

### 0.7 CI/CD
- [x] `.github/workflows/backend.yml` — lint, test, build, deploy
- [x] `.github/workflows/bot.yml` — lint, test, build, deploy
- [x] `.github/workflows/frontend.yml` — lint, build, deploy
- [x] `.github/workflows/infrastructure.yml` — migrations
- [x] `scripts/deploy.sh` — deploy backend|bot|frontend|infra
- [ ] Self-hosted runner настроен и работает

### 0.8 Traefik
- [x] `infra/traefik/traefik.yml` — entrypoints, Let's Encrypt
- [x] Routing rules (path-prefix based, strip prefix)
- [x] Health check labels для всех сервисов

---

## Критерии завершения

- [ ] `docker compose up` локально поднимает PG + Redis
- [ ] Backend стартует, отвечает на `/healthz`
- [ ] Bot стартует, отвечает на `/start` в Telegram
- [ ] Frontend стартует, показывает layout с sidebar
- [ ] Push в main → GitHub Actions → деплой на сервер
- [ ] `elysium.fm/revisitr/` показывает фронтенд
- [ ] `elysium.fm/revisitr/api/healthz` возвращает 200

## Оставшиеся задачи
1. Создать GitHub repo + push
2. Настроить self-hosted runner на сервере
3. Первый деплой + верификация
