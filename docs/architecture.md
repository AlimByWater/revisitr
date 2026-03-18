# Архитектура Revisitr

## Обзор системы

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  API Server  │────▶│  PostgreSQL   │
│  (React SPA) │     │   (Go/Gin)   │────▶│    Redis      │
└─────────────┘     └──────────────┘     └──────────────┘
                           │
                           │ shared DB
                           │
                    ┌──────────────┐     ┌──────────────┐
                    │ Telegram Bot │────▶│ Telegram API  │
                    │  (Go/telego) │     └──────────────┘
                    └──────────────┘
```

## Два сервиса

### 1. API Server (`cmd/server`)
- REST API для админ-панели
- Аутентификация (JWT)
- CRUD для всех сущностей
- Аналитика и отчёты

### 2. Telegram Bot (`cmd/bot`)
- Обработка команд пользователей
- Программа лояльности (начисление/списание баллов)
- Рассылки
- Бронирование, меню, обратная связь

Сервисы **разделяют базу данных** (PostgreSQL + Redis), но запускаются как **отдельные процессы**.

## Backend: Clean Architecture

```
internal/
├── application/          # Bootstrap
│   ├── app.go            # Lifecycle: Init → Run → Shutdown
│   ├── config/           # Config structs (JSON unmarshal)
│   └── env/              # Environment loading (.env / vault)
│
├── controller/           # Input adapters
│   ├── http/             # Gin HTTP server
│   │   ├── http.go       # Server setup, TLS
│   │   ├── middleware/    # CORS, auth, logging, recovery
│   │   └── group/        # Route groups (auth, clients, loyalty, ...)
│   └── scheduler/        # Cron jobs (gocron)
│
├── usecase/              # Business logic
│   ├── auth/             # JWT, sessions
│   ├── loyalty/          # Программы лояльности, баллы, уровни
│   ├── clients/          # Клиенты, профили, сегментация
│   ├── bots/             # Управление ботами
│   ├── campaigns/        # Рассылки, авто-сценарии
│   └── pos/              # Точки продаж
│
├── repository/           # Data access
│   ├── postgres/         # sqlx, SQL queries
│   └── redis/            # Кэш, сессии
│
└── entity/               # Domain models
    ├── user.go
    ├── client.go
    ├── loyalty.go
    ├── bot.go
    └── ...
```

## Паттерны из driptech

### Dependency Injection
Явный, без фреймворков. Всё собирается в `main.go`:

```go
// Repositories
postgresRepo := postgres.New(pgCfg, clientsTable, botsTable, ...)
redisRepo := redis.New(redisCfg)

// Usecases
authUC := auth.New(authCfg, redisRepo)
loyaltyUC := loyalty.New(postgresRepo, redisRepo)

// Controllers
httpCtrl := http.New(httpCfg, authGroup, clientsGroup, ...)

// Application
app := application.New(postgresRepo, redisRepo, authUC, loyaltyUC, httpCtrl)
app.Run(ctx)
```

### Module Interface
Каждый модуль реализует:

```go
type repository interface {
    Init(ctx context.Context, logger *slog.Logger) error
    Close() error
}

type usecase interface {
    Init(ctx context.Context, logger *slog.Logger) error
}

type controller interface {
    Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) error
    Run()
    Shutdown() error
}
```

### HTTP Group Interface

```go
type group interface {
    Path() string
    Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc)
    Auth() gin.HandlerFunc
}
```

## Frontend: TanStack SPA

```
frontend/src/
├── routes/                # File-based routing
│   ├── __root.tsx         # Root layout
│   ├── index.tsx          # Redirect to dashboard
│   ├── auth/
│   │   └── login.tsx
│   └── dashboard/
│       ├── route.tsx      # Dashboard layout (sidebar + header)
│       ├── index.tsx      # Dashboard home
│       ├── clients/
│       ├── loyalty/
│       ├── campaigns/
│       ├── bots/
│       ├── analytics/
│       ├── pos/
│       └── integrations/
│
├── features/              # Feature-specific API + types
│   ├── clients/
│   │   ├── api.ts
│   │   ├── queries.ts     # TanStack Query hooks
│   │   ├── mutations.ts
│   │   └── types.ts
│   └── ...
│
├── components/
│   ├── ui/                # shadcn/ui
│   └── common/            # DataTable, PageHeader, etc.
│
├── lib/
│   ├── api.ts             # Axios/fetch client
│   └── utils.ts
│
└── stores/                # Zustand (UI state only)
```

## API Design

Все эндпоинты: `https://elysium.fm/revisitr/api/v1/`

```
POST   /auth/login
POST   /auth/register
POST   /auth/verify

GET    /clients
GET    /clients/:id
POST   /clients
PATCH  /clients/:id

GET    /loyalty/programs
POST   /loyalty/programs
PATCH  /loyalty/programs/:id

GET    /bots
POST   /bots
PATCH  /bots/:id
DELETE /bots/:id

GET    /campaigns
POST   /campaigns
PATCH  /campaigns/:id

GET    /pos
POST   /pos
PATCH  /pos/:id
DELETE /pos/:id

GET    /analytics/sales
GET    /analytics/loyalty
GET    /analytics/campaigns

GET    /dashboard/widgets
```

## Infrastructure

```
                    elysium.fm
                        │
                    ┌───────┐
                    │Traefik│ (SSL, routing)
                    └───┬───┘
                        │
           ┌────────────┼────────────┐
           │            │            │
    /revisitr/api   /revisitr    (internal)
           │            │            │
      ┌────────┐  ┌──────────┐  ┌────────┐
      │Backend │  │ Frontend │  │  Bot   │
      │ :8080  │  │  nginx   │  │ :8081  │
      └────┬───┘  └──────────┘  └────┬───┘
           │                         │
      ┌────┴─────────────────────────┴───┐
      │         PostgreSQL :5432          │
      │           Redis :6379             │
      └──────────────────────────────────┘
```

## Deployment

- GitHub Actions (self-hosted runner на том же сервере)
- GHCR для Docker-образов
- Независимые деплои: backend, bot, frontend, infra
- Traefik → zero-downtime через health checks
