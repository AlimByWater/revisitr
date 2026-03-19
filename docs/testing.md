# Testing Strategy — Revisitr

Спецификация тестирования для SaaS-платформы Revisitr. Три уровня: unit, integration, E2E.

## Обзор архитектуры тестирования

```
┌─────────────────────────────────────────────────────────┐
│                    E2E Tests (Playwright)                │
│  Frontend + Backend + DB — реальные user flows           │
├─────────────────────────────────────────────────────────┤
│              Integration Tests (Go httptest)             │
│  API handlers + usecase + real DB (docker-compose)       │
├─────────────────┬───────────────────────────────────────┤
│  Backend Unit   │  Frontend Unit (Vitest)               │
│  Go test + mock │  Component + hook tests               │
├─────────────────┼───────────────────────────────────────┤
│  Bot Unit Tests │  Bot Integration (mock Telegram API)  │
│  telego mock    │  webhook handler + real DB             │
└─────────────────┴───────────────────────────────────────┘
```

---

## 1. Backend Unit Tests

**Что**: usecase-слой с замоканными репозиториями.
**Уже есть**: `campaigns`, `clients`, `dashboard`.
**Инструменты**: стандартный `testing`, ручные моки через struct с function fields.

### Паттерн (уже установлен)

```go
// backend/internal/usecase/<domain>/<domain>_test.go
type mockRepo struct {
    createFn func(ctx context.Context, ...) error
    // ...
}

func TestUsecase_Create(t *testing.T) {
    repo := &mockRepo{
        createFn: func(ctx context.Context, ...) error { return nil },
    }
    uc := New(repo)
    err := uc.Create(ctx, ...)
    if err != nil {
        t.Fatalf("expected nil, got %v", err)
    }
}
```

### Что покрыть дополнительно

| Domain   | File                                    | Priority |
|----------|-----------------------------------------|----------|
| auth     | `usecase/auth/auth_test.go`             | P0       |
| bots     | `usecase/bots/bots_test.go`             | P0       |
| loyalty  | `usecase/loyalty/loyalty_test.go`       | P1       |
| pos      | `usecase/pos/pos_test.go`               | P1       |

### Запуск

```bash
cd backend && go test -race ./internal/usecase/...
```

---

## 2. Backend Integration Tests (API + DB)

**Что**: HTTP handlers через `httptest.Server` с реальной PostgreSQL + Redis из docker-compose.
**Инструменты**: Go `testing`, `httptest`, `net/http`, docker-compose infra.

### Инфраструктура

Используем существующий `infra/docker-compose.yml` (postgres:5433, redis:6380).
Тестовая БД создаётся отдельно: `revisitr_test`.

**Файл**: `infra/postgres/init.sql` — добавить:
```sql
CREATE DATABASE revisitr_test;
GRANT ALL PRIVILEGES ON DATABASE revisitr_test TO revisitr;
```

### Структура

```
backend/
  tests/
    integration/
      testmain_test.go      — setup/teardown, DB подключение, миграции
      auth_test.go           — register → login → refresh → logout
      bots_test.go           — CRUD ботов
      loyalty_test.go        — программы лояльности
      campaigns_test.go      — кампании и сценарии
      clients_test.go        — клиентская база
      pos_test.go            — точки продаж
      dashboard_test.go      — виджеты и графики
      helpers_test.go        — утилиты: createTestUser, authHeader, etc.
```

### Паттерн TestMain

```go
// backend/tests/integration/testmain_test.go
package integration_test

import (
    "database/sql"
    "fmt"
    "log"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
)

var (
    testServer *httptest.Server
    testDB     *sql.DB
)

func TestMain(m *testing.M) {
    // 1. Connect to test DB (docker-compose postgres on :5433)
    dsn := fmt.Sprintf(
        "postgres://revisitr:devpassword@localhost:5433/revisitr_test?sslmode=disable",
    )
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("connect to test db: %v", err)
    }
    testDB = db

    // 2. Run migrations
    if err := goose.Up(db, "../../migrations"); err != nil {
        log.Fatalf("run migrations: %v", err)
    }

    // 3. Build full app with test config → httptest.Server
    // (собираем Gin engine с реальными usecase + repo, но тестовая БД)
    server := buildTestServer(db)
    testServer = httptest.NewServer(server)

    // 4. Run tests
    code := m.Run()

    // 5. Cleanup
    testServer.Close()
    goose.Down(db, "../../migrations")
    db.Close()
    os.Exit(code)
}
```

### Паттерн API-теста

```go
// backend/tests/integration/auth_test.go
func TestAuth_RegisterLoginRefresh(t *testing.T) {
    // Register
    body := `{"email":"test@example.com","password":"pass123","name":"Test"}`
    resp := doPost(t, "/api/v1/auth/register", body, "")
    assertStatus(t, resp, 201)

    var authResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&authResp)
    token := authResp["access_token"].(string)

    // Authenticated request
    resp = doGet(t, "/api/v1/bots", token)
    assertStatus(t, resp, 200)
}
```

### Запуск

```bash
# Требует: docker compose up -d (postgres + redis)
cd backend && go test -race -tags=integration ./tests/integration/...
```

Build tag `integration` для отделения от unit-тестов:
```go
//go:build integration
```

---

## 3. Frontend Unit Tests (Vitest)

**Что**: компоненты, хуки, stores, API-утилиты.
**Инструменты**: Vitest, @testing-library/react, MSW (Mock Service Worker).

### Зависимости

```bash
cd frontend && npm install -D vitest @testing-library/react @testing-library/jest-dom \
  @testing-library/user-event jsdom msw
```

### Конфигурация

**`frontend/vitest.config.ts`**:
```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
  },
})
```

**`frontend/src/test/setup.ts`**:
```typescript
import '@testing-library/jest-dom'
```

### MSW для API-моков

**`frontend/src/test/mocks/handlers.ts`**:
```typescript
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      access_token: 'test-jwt',
      refresh_token: 'test-refresh',
      user: { id: 1, email: 'test@example.com', name: 'Test' },
    })
  }),
  http.get('/api/v1/bots', () => {
    return HttpResponse.json({
      items: [{ id: 1, name: 'Test Bot', token: '123:ABC' }],
      total: 1,
    })
  }),
]
```

**`frontend/src/test/mocks/server.ts`**:
```typescript
import { setupServer } from 'msw/node'
import { handlers } from './handlers'

export const server = setupServer(...handlers)
```

### Структура тестов

```
frontend/src/
  test/
    setup.ts
    mocks/
      handlers.ts
      server.ts
  stores/
    auth.test.ts          — Zustand store тесты
  features/
    auth/
      api.test.ts         — login/register API hooks
    bots/
      api.test.ts         — bots CRUD hooks
  components/
    bots/
      BotForm.test.tsx    — компонент формы
    layout/
      Header.test.tsx     — layout компоненты
```

### Запуск

```bash
cd frontend && npx vitest              # watch mode
cd frontend && npx vitest run          # single run
cd frontend && npx vitest run --coverage  # с покрытием
```

---

## 4. E2E Tests (Playwright)

**Что**: полные user flows через браузер — frontend + backend + DB.
**Инструменты**: Playwright (через Playwright MCP или standalone).

### Предусловия

Для E2E тестов нужен полный стек:
1. PostgreSQL + Redis (docker-compose)
2. Backend API server (порт 8080)
3. Frontend dev server (порт 5173)

### Структура

```
e2e/
  playwright.config.ts
  tests/
    auth.spec.ts          — register, login, logout, token refresh
    bots.spec.ts          — создание, редактирование, удаление ботов
    loyalty.spec.ts       — программы лояльности CRUD
    pos.spec.ts           — точки продаж CRUD
    campaigns.spec.ts     — кампании и сценарии
    clients.spec.ts       — просмотр клиентов
    dashboard.spec.ts     — виджеты, графики
  fixtures/
    auth.ts               — login helper, test user factory
    db.ts                 — seed/cleanup helpers
```

### Конфигурация

**`e2e/playwright.config.ts`**:
```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,  // sequential — tests share DB state
  retries: 1,
  timeout: 30_000,
  use: {
    baseURL: 'http://localhost:5173/revisitr',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: [
    {
      command: 'cd ../backend && go run ./cmd/server',
      url: 'http://localhost:8080/api/v1/health',
      timeout: 30_000,
      reuseExistingServer: true,
      env: {
        POSTGRES_PORT: '5433',
        REDIS_PORT: '6380',
      },
    },
    {
      command: 'cd ../frontend && npm run dev',
      url: 'http://localhost:5173/revisitr/',
      timeout: 15_000,
      reuseExistingServer: true,
    },
  ],
})
```

### Паттерн E2E-теста

```typescript
// e2e/tests/auth.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('register → login → see dashboard', async ({ page }) => {
    // Register
    await page.goto('/auth/register')
    await page.fill('[name="email"]', 'e2e@test.com')
    await page.fill('[name="password"]', 'password123')
    await page.fill('[name="name"]', 'E2E User')
    await page.click('button[type="submit"]')

    // Should redirect to dashboard
    await expect(page).toHaveURL(/dashboard/)
    await expect(page.locator('h1')).toContainText('Dashboard')
  })

  test('login with wrong password → error', async ({ page }) => {
    await page.goto('/auth/login')
    await page.fill('[name="email"]', 'wrong@test.com')
    await page.fill('[name="password"]', 'wrong')
    await page.click('button[type="submit"]')

    await expect(page.locator('[role="alert"]')).toBeVisible()
  })
})
```

### Запуск

```bash
cd e2e && npx playwright test                    # все тесты
cd e2e && npx playwright test auth.spec.ts       # только auth
cd e2e && npx playwright test --ui               # UI mode для дебага
```

### Использование с Claude Code MCP

Claude Code может запускать E2E тесты двумя способами:

**1. Playwright MCP** (browser_navigate, browser_snapshot, browser_click):
```
# Для ручного exploratory testing и проверки UI
# Claude открывает localhost:5173/revisitr, делает snapshot, проверяет элементы
```

**2. Claude Preview MCP** (preview_start, preview_screenshot, preview_snapshot):
```
# Для быстрой визуальной проверки через launch.json dev server
# Требует .claude/launch.json с конфигурацией серверов
```

**3. Bash → Playwright CLI** (для автоматизированного прогона):
```bash
cd e2e && npx playwright test --reporter=list
```

---

## 5. Bot Tests

### 5.1 Unit Tests (Go)

Telegram бот тестируется через mock Telegram API.

```
backend/
  internal/service/
    bot_manager_test.go     — BotManager unit tests с mock telego
  cmd/bot/
    handlers_test.go        — webhook handler unit tests
```

**Паттерн**:
```go
// Мокаем telego.Bot через интерфейс
type mockTelegramBot interface {
    SendMessage(params *telego.SendMessageParams) (*telego.Message, error)
}

func TestBotHandler_Start(t *testing.T) {
    bot := &mockBot{
        sendMessageFn: func(params *telego.SendMessageParams) (*telego.Message, error) {
            if params.Text == "" {
                t.Fatal("expected non-empty message")
            }
            return &telego.Message{}, nil
        },
    }
    handler := NewHandler(bot, usecases...)
    handler.HandleStart(ctx, update)
}
```

### 5.2 Bot Integration Tests

Бот-интеграционные тесты проверяют обработку webhook → usecase → DB.

```go
//go:build integration

func TestBot_LoyaltyAccrual(t *testing.T) {
    // 1. Seed: создать бота, программу лояльности, клиента в тестовой БД
    // 2. Симулировать webhook update от Telegram
    // 3. Проверить что баллы начислены в БД
}
```

### Запуск

```bash
cd backend && go test -race ./internal/service/...       # unit
cd backend && go test -race -tags=integration ./cmd/bot/... # integration
```

---

## 6. Docker-compose для тестов

**`infra/docker-compose.test.yml`** (переопределение для CI):

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: revisitr_test
      POSTGRES_USER: revisitr
      POSTGRES_PASSWORD: devpassword
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data  # RAM для скорости
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U revisitr"]
      interval: 2s
      timeout: 2s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    tmpfs:
      - /data
```

---

## 7. Makefile targets

```makefile
# Unit tests
test-unit-backend:
	cd backend && go test -race ./internal/usecase/...

test-unit-frontend:
	cd frontend && npx vitest run

# Integration tests (требует docker compose)
test-integration:
	cd backend && go test -race -tags=integration ./tests/integration/...

# E2E tests (требует full stack)
test-e2e:
	cd e2e && npx playwright test

# All tests
test-all: test-unit-backend test-unit-frontend test-integration test-e2e

# Quick check (unit only, без инфры)
test-quick: test-unit-backend test-unit-frontend
```

---

## 8. CI Pipeline

### GitHub Actions: test job (расширение существующего)

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  unit-backend:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - run: cd backend && go test -race ./internal/usecase/...

  unit-frontend:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: cd frontend && npm ci && npx vitest run

  integration:
    runs-on: self-hosted
    needs: [unit-backend]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: revisitr_test
          POSTGRES_USER: revisitr
          POSTGRES_PASSWORD: devpassword
        ports:
          - 5433:5432
        options: >-
          --health-cmd "pg_isready -U revisitr"
          --health-interval 5s
          --health-timeout 3s
          --health-retries 10
      redis:
        image: redis:7-alpine
        ports:
          - 6380:6379
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - run: cd backend && go test -race -tags=integration ./tests/integration/...
        env:
          POSTGRES_PORT: "5433"
          REDIS_PORT: "6380"
```

---

## 9. Claude Code Testing Workflow

### .claude/launch.json (для Preview MCP)

```json
{
  "version": "0.0.1",
  "configurations": [
    {
      "name": "frontend",
      "runtimeExecutable": "npm",
      "runtimeArgs": ["run", "dev"],
      "port": 5173
    },
    {
      "name": "backend",
      "runtimeExecutable": "go",
      "runtimeArgs": ["run", "./cmd/server"],
      "port": 8080
    }
  ]
}
```

### Доступные MCP для тестирования

| MCP Server      | Использование                                       |
|-----------------|-----------------------------------------------------|
| Playwright MCP  | Exploratory E2E testing, UI snapshot, accessibility  |
| Claude Preview  | Dev server management + visual verification          |
| Bash            | `go test`, `npx vitest`, `npx playwright test`      |
| Postgres MCP    | Прямые SQL-запросы для проверки данных после тестов  |

### Типичный workflow Claude Code

```
1. Поднять инфру:         cd infra && docker compose up -d
2. Unit тесты бэкенда:    cd backend && go test -race ./internal/usecase/...
3. Unit тесты фронтенда:  cd frontend && npx vitest run
4. Integration тесты:     cd backend && go test -race -tags=integration ./tests/integration/...
5. E2E (если нужно):      cd e2e && npx playwright test
6. Visual check (MCP):    preview_start → preview_screenshot → preview_snapshot
```

---

## 10. Приоритеты реализации

| Фаза | Что                                  | Зависимости        |
|------|--------------------------------------|---------------------|
| P0   | Backend integration tests (auth+CRUD)| docker-compose, goose |
| P0   | Frontend Vitest setup + store tests  | vitest, msw         |
| P1   | Backend unit tests (auth, bots)      | —                   |
| P1   | E2E auth flow (Playwright)           | full stack          |
| P2   | E2E CRUD flows                       | full stack          |
| P2   | Bot unit + integration tests         | telego mocks        |
| P3   | CI pipeline (GitHub Actions)         | all above           |
| P3   | Coverage reporting                   | —                   |
