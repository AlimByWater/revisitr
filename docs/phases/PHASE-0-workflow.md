# Фаза 0: Workflow — пошаговая реализация

> Детальный порядок выполнения задач Фазы 0 с зависимостями и параллелизацией.

---

## Обзор потоков работы

```
Step 1: Git Init ─────────────────────────────────────────────────────┐
                                                                      │
Step 2: ┌─────────────────┐  ┌──────────────────┐  ┌───────────────┐ │
        │ Backend Skeleton │  │ Frontend Skeleton │  │ Infra: Docker │ │
        │ (Go code)        │  │ (React/TanStack)  │  │ (compose,     │ │
        └────────┬────────┘  └────────┬─────────┘  │  Dockerfiles) │ │
                 │                     │            └───────┬───────┘ │
                 ▼                     ▼                    ▼         │
Step 3: ┌─────────────────────────────────────────────────────────┐   │
        │            Локальная проверка: всё стартует              │   │
        └────────────────────────┬────────────────────────────────┘   │
                                 │                                    │
Step 4: ┌────────────────────────┴────────────────────────────────┐   │
        │     CI/CD + nginx + Production Deploy                     │   │
        └────────────────────────┬────────────────────────────────┘   │
                                 │                                    │
Step 5: ┌────────────────────────┴────────────────────────────────┐   │
        │     Верификация: elysium.fm/revisitr работает             │   │
        └─────────────────────────────────────────────────────────┘   │
```

---

## Step 1: Git Init + Первый коммит

**Зависимости**: нет
**Параллелизация**: нет

### Задачи
1. `git init` в корне проекта
2. Проверить `.gitignore` (уже создан)
3. Первый коммит: структура + docs + конфиги
4. Создать GitHub репозиторий
5. `git remote add origin` + `git push`

### Команды
```bash
cd /Users/admin/go/src/revisitr
git init
git add .
git commit -m "feat: initial project structure, docs, and configs"
# Создать repo на GitHub, затем:
git remote add origin git@github.com:<org>/revisitr.git
git push -u origin main
```

### Критерий: repo на GitHub с начальной структурой

---

## Step 2: Три параллельных потока

> Эти три потока **независимы** и могут выполняться параллельно.

---

### Step 2A: Backend Skeleton

**Зависимости**: Step 1
**Оценка**: ~20 файлов

#### 2A.1 — Go Module + Dependencies
```
backend/go.mod
backend/go.sum
```
Зависимости: gin, sqlx, go-redis, telego, jwt, godotenv, goose, slog

#### 2A.2 — Entity Layer
```
backend/internal/entity/user.go          # User struct (id, email, phone, name, password_hash, role, org_id)
backend/internal/entity/organization.go  # Organization struct (id, name, owner_id)
```

#### 2A.3 — Application Bootstrap
```
backend/internal/application/env/env.go          # .env loading (godotenv)
backend/internal/application/config/config.go     # Config structs (Http, Postgres, Redis, Auth)
backend/internal/application/app.go               # Application lifecycle
```

Паттерн из driptech:
- `env.Module` загружает `.env` файл
- `config.Module` хранит typed config structs
- `app.Application` оркестрирует Init → Run → Shutdown

#### 2A.4 — Repository Layer
```
backend/internal/repository/postgres/postgres.go  # sqlx connect, Init/Close
backend/internal/repository/redis/redis.go        # go-redis connect, Init/Close
```

#### 2A.5 — HTTP Controller
```
backend/internal/controller/http/http.go                  # Gin server, group registration
backend/internal/controller/http/middleware/cors.go        # CORS
backend/internal/controller/http/middleware/recovery.go    # Panic recovery
backend/internal/controller/http/middleware/logger.go      # Request logging (slog)
backend/internal/controller/http/middleware/auth.go        # JWT auth (заглушка — проверка только наличия токена)
backend/internal/controller/http/group/health/health.go   # GET /healthz
```

#### 2A.6 — Entrypoints
```
backend/cmd/server/main.go  # API server — wires everything, runs app
backend/cmd/bot/main.go     # Telegram bot — minimal telego setup, /start handler
```

#### 2A.7 — Migrations
```
backend/migrations/00001_init.sql  # users + organizations tables
```

#### Порядок внутри 2A
```
2A.1 (go.mod) → 2A.2 (entity) → 2A.3 (application) → 2A.4 (repository)
                                                            ↓
                                          2A.5 (http controller) → 2A.6 (main.go)
                                                                        ↓
                                                                  2A.7 (migrations)
```

#### Проверка 2A
```bash
cd backend
go build ./cmd/server    # компилируется
go build ./cmd/bot       # компилируется
go vet ./...             # нет ошибок
```

---

### Step 2B: Frontend Skeleton

**Зависимости**: Step 1
**Оценка**: ~15 файлов

#### 2B.1 — Project Init
```
frontend/package.json
frontend/tsconfig.json
frontend/tsconfig.app.json
frontend/vite.config.ts
frontend/tailwind.config.ts
frontend/postcss.config.js
frontend/index.html
frontend/src/main.tsx
frontend/src/index.css          # Tailwind directives + noise grain background
```

#### 2B.2 — shadcn/ui Init
```bash
cd frontend && npx shadcn@latest init
```
Это создаёт `components.json` и базовые утилиты.

#### 2B.3 — App Setup
```
frontend/src/App.tsx                           # RouterProvider + QueryClientProvider
frontend/src/lib/api.ts                        # Axios client + interceptors
frontend/src/lib/utils.ts                      # cn() helper (shadcn)
```

#### 2B.4 — Routes (File-Based)
```
frontend/src/routes/__root.tsx                 # <html> wrapper, QueryClient
frontend/src/routes/index.tsx                  # Redirect → /dashboard
frontend/src/routes/auth/login.tsx             # Login page (заглушка с формой)
frontend/src/routes/dashboard/route.tsx        # Dashboard layout: sidebar + header + <Outlet />
frontend/src/routes/dashboard/index.tsx        # Dashboard home: "Welcome" placeholder
```

#### 2B.5 — Layout Components
```
frontend/src/components/layout/Sidebar.tsx     # Sidebar navigation (из Figma)
frontend/src/components/layout/Header.tsx      # Top header (logo + nav + profile)
frontend/src/components/layout/RootLayout.tsx  # Wrapper
```

#### Порядок внутри 2B
```
2B.1 (init) → 2B.2 (shadcn) → 2B.3 (app setup) → 2B.4 (routes) → 2B.5 (layout)
```

#### Проверка 2B
```bash
cd frontend
npm run dev          # стартует на localhost:5173
npm run build        # сборка без ошибок
npm run lint         # нет ошибок
```

---

### Step 2C: Infrastructure

**Зависимости**: Step 1
**Оценка**: ~10 файлов

#### 2C.1 — Local Dev (уже частично создано)
```
infra/docker-compose.yml          # ✅ Уже создан (PG + Redis)
infra/postgres/init.sql           # ✅ Уже создан
```

#### 2C.2 — Dockerfiles
```
backend/Dockerfile                # Multi-stage: golang:1.23-alpine → alpine
backend/Dockerfile.bot            # Аналогичный, но собирает cmd/bot
frontend/Dockerfile               # Multi-stage: node:22-alpine → nginx:1.27-alpine
frontend/nginx.conf               # SPA routing, /health endpoint, cache static
```

#### 2C.3 — Production Compose
```
infra/docker-compose.prod.yml     # Traefik + backend + bot + frontend + PG + Redis
infra/traefik/traefik.yml         # Entrypoints, Let's Encrypt, Docker provider
```

#### 2C.4 — Deploy Scripts
```
scripts/deploy.sh                 # deploy backend|bot|frontend|infra
scripts/migrate.sh                # goose up|down|status
```

#### Порядок внутри 2C
```
2C.1 (local) → 2C.2 (Dockerfiles) → 2C.3 (prod compose) → 2C.4 (scripts)
```

#### Проверка 2C
```bash
cd infra && docker compose up -d        # PG + Redis стартуют
docker build -f backend/Dockerfile backend/    # образ собирается
docker build -f frontend/Dockerfile frontend/  # образ собирается
```

---

## Step 3: Локальная интеграция

**Зависимости**: Steps 2A + 2B + 2C завершены
**Параллелизация**: нет (последовательная проверка)

### Задачи
1. Поднять infra: `make dev-up`
2. Запустить миграции: `make migrate-up`
3. Запустить backend: `make backend-run`
4. Проверить: `curl localhost:8080/healthz` → 200
5. Запустить bot: `make bot-run`
6. Проверить: отправить /start боту в Telegram
7. Запустить frontend: `make frontend-dev`
8. Проверить: открыть localhost:5173 → sidebar + header

### Критерий: все 3 компонента работают локально вместе

---

## Step 4: CI/CD + Production Deploy

**Зависимости**: Step 3
**Параллелизация**: workflows можно создавать параллельно

### 4.1 — GitHub Actions Workflows
```
.github/workflows/backend.yml        # Path: backend/** → lint, test, build, push to GHCR, deploy
.github/workflows/bot.yml            # Path: backend/** → аналогично для cmd/bot
.github/workflows/frontend.yml       # Path: frontend/** → lint, build, push to GHCR, deploy
.github/workflows/infrastructure.yml # Path: backend/migrations/** → run migrations
```

### 4.2 — Server Setup
1. Убедиться что self-hosted runner работает
2. Настроить Docker + GHCR login на сервере
3. Создать `/opt/revisitr/` на сервере
4. Скопировать `infra/docker-compose.prod.yml` и `infra/traefik/`
5. Создать `.env.prod` на сервере (не в repo!)
6. Настроить GitHub Secrets: `DEPLOY_SSH_KEY`, `VPS_HOST` (или прямой runner)

### 4.3 — Traefik + Routing
1. Traefik контейнер с Let's Encrypt (если нужен SSL для elysium.fm)
2. Path-based routing:
   - `elysium.fm/revisitr/api/` → strip prefix → backend:8080
   - `elysium.fm/revisitr/` → strip prefix → frontend:80
3. Health check labels

### 4.4 — Первый деплой
```bash
git add . && git commit -m "feat: phase 0 complete — full skeleton"
git push origin main
# GitHub Actions → build → deploy
```

### Критерий: push в main автоматически деплоит на сервер

---

## Step 5: Production Verification

**Зависимости**: Step 4

### Чеклист
- [ ] `https://elysium.fm/revisitr/` → фронтенд с sidebar
- [ ] `https://elysium.fm/revisitr/api/healthz` → `{"status":"ok"}`
- [ ] Telegram-бот отвечает на /start
- [ ] GitHub Actions: все workflows зелёные
- [ ] Docker images в GHCR

### Коммит
```bash
# Обновить PHASE-0-foundation.md — все чекбоксы ✅
git commit -m "docs: mark phase 0 as complete"
```

---

## Зависимости между шагами (граф)

```
Step 1 (git init)
  ├──▶ Step 2A (backend)  ──┐
  ├──▶ Step 2B (frontend) ──┼──▶ Step 3 (local test) ──▶ Step 4 (CI/CD) ──▶ Step 5 (verify)
  └──▶ Step 2C (infra)    ──┘
```

## Рекомендуемый порядок с Claude Code

Оптимальная стратегия — использовать параллельных агентов:

1. **Step 1** → ручной git init + GitHub repo
2. **Step 2A + 2B + 2C** → три параллельных агента:
   - `backend-architect` → весь Go-код
   - `frontend-architect` → весь React/TanStack код
   - `devops-architect` → Dockerfiles, compose, Traefik, scripts
3. **Step 3** → ручная проверка (запуск, curl, браузер)
4. **Step 4** → `devops-architect` → CI/CD workflows + server setup
5. **Step 5** → ручная верификация

## Файлы, которые будут созданы (полный список)

```
# Backend (~20 файлов)
backend/go.mod
backend/go.sum
backend/cmd/server/main.go
backend/cmd/bot/main.go
backend/internal/application/app.go
backend/internal/application/env/env.go
backend/internal/application/config/config.go
backend/internal/controller/http/http.go
backend/internal/controller/http/middleware/cors.go
backend/internal/controller/http/middleware/recovery.go
backend/internal/controller/http/middleware/logger.go
backend/internal/controller/http/middleware/auth.go
backend/internal/controller/http/group/health/health.go
backend/internal/repository/postgres/postgres.go
backend/internal/repository/redis/redis.go
backend/internal/entity/user.go
backend/internal/entity/organization.go
backend/migrations/00001_init.sql

# Frontend (~15 файлов)
frontend/package.json
frontend/tsconfig.json
frontend/tsconfig.app.json
frontend/vite.config.ts
frontend/tailwind.config.ts
frontend/postcss.config.js
frontend/index.html
frontend/src/main.tsx
frontend/src/App.tsx
frontend/src/index.css
frontend/src/lib/api.ts
frontend/src/lib/utils.ts
frontend/src/routes/__root.tsx
frontend/src/routes/index.tsx
frontend/src/routes/auth/login.tsx
frontend/src/routes/dashboard/route.tsx
frontend/src/routes/dashboard/index.tsx
frontend/src/components/layout/Sidebar.tsx
frontend/src/components/layout/Header.tsx

# Infrastructure (~10 файлов)
backend/Dockerfile
backend/Dockerfile.bot
frontend/Dockerfile
frontend/nginx.conf
infra/docker-compose.prod.yml
infra/traefik/traefik.yml
scripts/deploy.sh
scripts/migrate.sh

# CI/CD (~4 файла)
.github/workflows/backend.yml
.github/workflows/bot.yml
.github/workflows/frontend.yml
.github/workflows/infrastructure.yml
```

**Итого: ~49 файлов для Фазы 0**
