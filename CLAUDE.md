## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

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
# Docker (production) — на сервере в /opt/revisitr/infra/
docker compose -f docker-compose.prod.yml up -d            # all services
infra/scripts/deploy.sh backend|bot|frontend|infra               # independent deploy
```

## Ports

| Service    | Local dev | Production (host → container) |
|------------|-----------|-------------------------------|
| Backend    | 9721      | 8090 → 8080                   |
| Frontend   | 5921      | 3340 → 80                     |
| PostgreSQL | 6281      | 5433 → 5432                   |
| Redis      | 7392      | 6380 → 6379                   |

Production nginx routes: `/revisitr/api/*` → `:8090`, `/revisitr/*` → `:3340`

## Remote Access (tuna tunnels)

Для удалённого доступа к локальной dev-среде используется [tuna](https://tuna.am/docs/tunnels/http/).

```bash
# Основной тунель — фронтенд + API (Vite проксирует /api → localhost:8080)
tuna http 5921 --basic-auth="login:password"
# Доступ: https://<subdomain>.tuna.am/revisitr/

# Бот (если нужно тестировать Telegram webhook-и)
tuna http 9722 --basic-auth="login:password"
```

**Как это работает**: один тунель на Vite dev server (5173) покрывает и фронтенд, и API — Vite proxy серверно перенаправляет `/api` на backend (localhost:8080). Отдельный тунель для backend не нужен.

**Требования**: локально должны быть запущены backend (`go run ./cmd/server`), frontend (`npm run dev`), и инфра (`docker compose up`).

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

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
