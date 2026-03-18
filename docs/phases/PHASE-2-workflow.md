# Фаза 2: Управление — Пошаговый Workflow

> Детальный порядок реализации с зависимостями, параллелизацией и критериями проверки.

---

## Обзор потоков

```
Step 1: Клиенты (Backend) ──────────────────────────────────────────────────────┐
        Клиенты (Frontend)                                                       │
                                                                                 │
Step 2: ┌──────────────────────────┐  ┌────────────────────────────────────────┐ │
        │ Дашборд                  │  │ Рассылки / Кампании                    │ │
        │ (Backend + Frontend)     │  │ (Backend + Frontend + Bot integration) │ │
        └────────────┬─────────────┘  └──────────────┬─────────────────────────┘ │
                     │                                │                          │
                     ▼                                ▼                          │
Step 3: ┌──────────────────────────────────────────────────────────────────────┐ │
        │  Интеграция: Dashboard ← Campaigns stats, Auto-scenarios via Bot   │ │
        └──────────────────────────┬───────────────────────────────────────────┘ │
                                   │                                             │
Step 4: ┌──────────────────────────┴───────────────────────────────────────────┐ │
        │  Polish + Testing + Deploy                                            │ │
        └───────────────────────────────────────────────────────────────────────┘ │
```

---

## Зависимости от Фазы 1

Фаза 2 строится на:
- **bot_clients** таблица (расширяем полями: пол, возраст, город, теги)
- **loyalty** модуль (баланс, уровни, транзакции для профиля клиента)
- **bots** модуль (отправка рассылок через Telegram)
- **pos** модуль (фильтр по заведению)
- **auth** (JWT middleware, org_id из контекста)

---

## Step 1: Клиенты

**Зависимости**: Фаза 1 завершена (bot_clients, loyalty, bots существуют)
**Это блокирующий шаг** — дашборд и рассылки оперируют клиентскими данными.

---

### Step 1A: Клиенты Backend

**Порядок**: последовательный (миграция → entity → repo → usecase → controller)

#### 1A.1 — Миграция

```
backend/migrations/YYYYMMDDHHMMSS_extend_bot_clients.sql
```

```sql
-- +goose Up
ALTER TABLE bot_clients
  ADD COLUMN gender VARCHAR(10),
  ADD COLUMN birth_date DATE,
  ADD COLUMN city VARCHAR(100),
  ADD COLUMN tags JSONB DEFAULT '[]'::jsonb,
  ADD COLUMN os VARCHAR(20);

CREATE INDEX idx_bot_clients_tags ON bot_clients USING GIN(tags);
CREATE INDEX idx_bot_clients_birth_date ON bot_clients(birth_date);

-- +goose Down
DROP INDEX IF EXISTS idx_bot_clients_birth_date;
DROP INDEX IF EXISTS idx_bot_clients_tags;
ALTER TABLE bot_clients
  DROP COLUMN IF EXISTS os,
  DROP COLUMN IF EXISTS tags,
  DROP COLUMN IF EXISTS city,
  DROP COLUMN IF EXISTS birth_date,
  DROP COLUMN IF EXISTS gender;
```

**Проверка**: `goose -dir migrations postgres "$DATABASE_URL" up` → успешно

#### 1A.2 — Entity: обновить BotClient

```
backend/internal/entity/bot_client.go — MODIFIED
backend/internal/entity/client.go     — NEW (запросы, фильтры)
```

Расширить `BotClient`:
- `Gender *string`, `BirthDate *time.Time`, `City *string`, `OS *string`
- `Tags Tags` (JSONB с Scan/Value, `[]string`)

Новые типы в `client.go`:
- `ClientFilter` — фильтры для списка:
  ```go
  type ClientFilter struct {
      BotID     *int    `form:"bot_id"`
      PosID     *int    `form:"pos_id"`
      Segment   *string `form:"segment"`
      Search    *string `form:"search"`     // имя или телефон
      SortBy    string  `form:"sort_by"`    // name, balance, purchases, registered_at
      SortOrder string  `form:"sort_order"` // asc, desc
      Limit     int     `form:"limit"`
      Offset    int     `form:"offset"`
  }
  ```
- `ClientProfile` — расширенный профиль с loyalty данными:
  ```go
  type ClientProfile struct {
      BotClient
      LoyaltyBalance  float64          `json:"loyalty_balance"`
      LoyaltyLevel    *string          `json:"loyalty_level"`
      TotalPurchases  float64          `json:"total_purchases"`
      PurchaseCount   int              `json:"purchase_count"`
      Transactions    []LoyaltyTransaction `json:"transactions,omitempty"`
  }
  ```
- `ClientStats` — виджеты:
  ```go
  type ClientStats struct {
      TotalClients   int     `json:"total_clients"`
      TotalBalance   float64 `json:"total_balance"`
      NewThisMonth   int     `json:"new_this_month"`
      ActiveThisWeek int     `json:"active_this_week"`
  }
  ```
- `UpdateClientRequest` — редактирование тегов:
  ```go
  type UpdateClientRequest struct {
      Tags *Tags `json:"tags,omitempty"`
  }
  ```

#### 1A.3 — Repository: Clients

```
backend/internal/repository/postgres/clients.go — NEW
```

Использует существующий `*Module`, работает с `bot_clients` + JOIN на `client_loyalty`, `loyalty_levels`.

Методы:
- `GetByOrgID(ctx, orgID, filter ClientFilter) ([]ClientProfile, int, error)`
  - JOIN: `bot_clients → bots (org_id check) → client_loyalty → loyalty_levels`
  - Динамическая фильтрация (bot_id, search ILIKE, tags @> operator)
  - Серверная пагинация + total count
  - Серверная сортировка
- `GetByID(ctx, orgID, clientID) (*ClientProfile, error)`
  - Полный профиль с JOIN на loyalty
- `Update(ctx, orgID, clientID, req UpdateClientRequest) error`
  - Проверка принадлежности через JOIN на bots.org_id
- `GetStats(ctx, orgID) (*ClientStats, error)`
  - Агрегация: COUNT, SUM(balance), COUNT WHERE registered_at > start_of_month, etc.
- `GetByFilter(ctx, orgID, filter ClientFilter) ([]int, error)`
  - Возвращает IDs клиентов по фильтру (для рассылок — подсчёт охвата)

#### 1A.4 — Usecase: Clients

```
backend/internal/usecase/clients/clients.go — NEW
backend/internal/usecase/clients/errors.go  — NEW
```

Зависимости: clients repo (interface)

Методы:
- `List(ctx, orgID, filter) ([]ClientProfile, int, error)`
- `GetProfile(ctx, orgID, clientID) (*ClientProfile, error)`
- `UpdateTags(ctx, orgID, clientID, req) error`
- `GetStats(ctx, orgID) (*ClientStats, error)`
- `CountByFilter(ctx, orgID, filter) (int, error)` — для превью рассылок

Sentinel errors: `ErrClientNotFound`

#### 1A.5 — Controller: Clients Group

```
backend/internal/controller/http/group/clients/clients.go — NEW
```

- `Path()` → `"/clients"`
- `Auth()` → JWT middleware
- Handlers:
  - `GET /` → list clients (query params: bot_id, search, sort_by, sort_order, limit, offset)
  - `GET /stats` → client stats widgets
  - `GET /:id` → client profile with transactions
  - `PATCH /:id` → update tags/segments
  - `GET /count` → count by filter (для preview рассылок)

#### 1A.6 — Wiring в main.go

```
backend/cmd/server/main.go — MODIFIED
```

- `clientsRepo := pgRepo.NewClients(pg)`
- `clientsUsecase := clientsUC.New(clientsRepo)`
- `clientsGrp := clientsGroup.New(clientsUsecase, jwtSecret)`
- Добавить в HTTP module и Application

**Проверка 1A**:
```bash
go build ./cmd/server

curl "localhost:8080/api/v1/clients?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
# → 200 {items: [...], total: N}

curl "localhost:8080/api/v1/clients/stats" \
  -H "Authorization: Bearer $TOKEN"
# → 200 {total_clients: N, total_balance: N, ...}
```

---

### Step 1B: Клиенты Frontend

**Зависимости**: Step 1A (backend endpoints работают)

#### 1B.1 — Feature: Clients

```
frontend/src/features/clients/types.ts   — NEW
frontend/src/features/clients/api.ts     — NEW
frontend/src/features/clients/queries.ts — NEW
```

Types:
- `ClientProfile` — matching backend response
- `ClientFilter` — query params interface
- `ClientStats` — stats widget data

API:
- `clientsApi.list(filter)` → GET /clients
- `clientsApi.getById(id)` → GET /clients/:id
- `clientsApi.getStats()` → GET /clients/stats
- `clientsApi.updateTags(id, tags)` → PATCH /clients/:id
- `clientsApi.countByFilter(filter)` → GET /clients/count

Queries:
- `useClientsQuery(filter)` — with keepPreviousData for pagination
- `useClientProfileQuery(id)`
- `useClientStatsQuery()`
- `useUpdateTagsMutation()`

#### 1B.2 — Таблица клиентов

```
frontend/src/routes/dashboard/clients/index.tsx — NEW
```

TanStack Table:
- Колонки: имя (FirstName + LastName), бот, уровень, сегмент, баланс, покупок, зарегистрирован
- Серверная пагинация (limit/offset из query params)
- Фильтры: dropdown бота, поиск по имени/телефону
- Сортировка по клику на заголовок колонки
- Виджеты статистики сверху (total clients, total balance, new this month)
- Клик на строку → открыть профиль

Зависимости npm:
- `@tanstack/react-table` (уже в проекте)

#### 1B.3 — Профиль клиента

```
frontend/src/routes/dashboard/clients/$clientId.tsx — NEW
```

Или модальное окно (решить по UX):
- Основная информация (имя, телефон, город, ОС)
- Текущий уровень и баланс (визуальная карточка)
- История транзакций (таблица: дата, тип, сумма, описание)
- Теги/сегменты (badges с возможностью добавления/удаления)

#### 1B.4 — Sidebar обновление

```
frontend/src/components/layout/Sidebar.tsx — MODIFIED
```

- Добавить пункт "Клиенты" в навигацию
- Badge с общим кол-вом клиентов (из stats)

**Проверка 1B**:
```bash
npm run dev
# → /dashboard/clients → таблица с пагинацией
# Поиск → фильтрация
# Клик на строку → профиль
# Виджеты статистики отображаются
```

---

## Step 2: Два параллельных модуля

> После завершения Step 1 (Clients) — Дашборд и Рассылки **независимы** и могут разрабатываться параллельно.

---

### Step 2A: Дашборд

**Зависимости**: Step 1 (clients для агрегации)
**Сложность**: средняя

#### 2A.1 — Entity: Dashboard

```
backend/internal/entity/dashboard.go — NEW
```

Типы:
- `DashboardWidgets` — агрегированные данные:
  ```go
  type DashboardWidgets struct {
      Revenue        DashboardMetric `json:"revenue"`
      AvgCheck       DashboardMetric `json:"avg_check"`
      NewClients     DashboardMetric `json:"new_clients"`
      ActiveClients  DashboardMetric `json:"active_clients"`
      CampaignsSent  DashboardMetric `json:"campaigns_sent"`
  }
  type DashboardMetric struct {
      Value    float64 `json:"value"`
      Previous float64 `json:"previous"`   // за предыдущий период
      Trend    float64 `json:"trend"`      // % изменения
  }
  ```
- `DashboardFilter`:
  ```go
  type DashboardFilter struct {
      Period  string `form:"period"`   // 7d, 30d, 90d, custom
      From    *time.Time `form:"from"`
      To      *time.Time `form:"to"`
      BotID   *int   `form:"bot_id"`
      PosID   *int   `form:"pos_id"`
      Segment *string `form:"segment"`
  }
  ```
- `DashboardChart` — данные для графиков:
  ```go
  type ChartPoint struct {
      Date  string  `json:"date"`
      Value float64 `json:"value"`
  }
  type DashboardCharts struct {
      Revenue       []ChartPoint `json:"revenue"`
      NewClients    []ChartPoint `json:"new_clients"`
      AvgCheck      []ChartPoint `json:"avg_check"`
  }
  ```

#### 2A.2 — Repository: Dashboard

```
backend/internal/repository/postgres/dashboard.go — NEW
```

Методы:
- `GetWidgets(ctx, orgID, filter DashboardFilter) (*DashboardWidgets, error)`
  - SQL агрегация: JOIN loyalty_transactions, bot_clients, campaigns
  - Вычисление trend: текущий период vs предыдущий (такой же длины)
- `GetCharts(ctx, orgID, filter DashboardFilter) (*DashboardCharts, error)`
  - GROUP BY date для каждого графика
  - `generate_series` для заполнения пустых дней нулями

#### 2A.3 — Usecase: Dashboard

```
backend/internal/usecase/dashboard/dashboard.go — NEW
```

Зависимости: dashboard repo

Методы:
- `GetWidgets(ctx, orgID, filter) (*DashboardWidgets, error)`
- `GetCharts(ctx, orgID, filter) (*DashboardCharts, error)`

#### 2A.4 — Controller: Dashboard Group

```
backend/internal/controller/http/group/dashboard/dashboard.go — NEW
```

- `Path()` → `"/dashboard"`
- `Auth()` → JWT middleware
- Handlers:
  - `GET /widgets` → виджеты с трендами
  - `GET /charts` → данные для графиков

#### 2A.5 — Wiring

```
backend/cmd/server/main.go — MODIFIED
```

#### 2A.6 — Frontend: Дашборд

```
frontend/src/features/dashboard/types.ts   — NEW
frontend/src/features/dashboard/api.ts     — NEW
frontend/src/features/dashboard/queries.ts — NEW
frontend/src/routes/dashboard/index.tsx    — MODIFIED (replace placeholder)
```

Компоненты:
```
frontend/src/components/dashboard/WidgetCard.tsx  — NEW (число + тренд ↑↓)
frontend/src/components/dashboard/ChartCard.tsx   — NEW (обёртка для графика)
frontend/src/components/dashboard/FilterBar.tsx   — NEW (период, бот, заведение)
```

Зависимости npm:
- `recharts` — графики (line chart, bar chart)

Страница:
- Фильтр-бар сверху: период (7д / 30д / 90д / свой), бот, заведение
- Виджеты-карточки: выручка, средний чек, новые клиенты, активные, рассылки
- Графики: выручка по дням (area chart), новые клиенты (bar chart), средний чек (line chart)

**Проверка 2A**:
```bash
curl "localhost:8080/api/v1/dashboard/widgets?period=30d" \
  -H "Authorization: Bearer $TOKEN"
# → 200 {revenue: {value, previous, trend}, ...}

npm run dev
# → /dashboard → виджеты + графики + фильтры
```

---

### Step 2B: Рассылки (Campaigns)

**Зависимости**: Step 1 (clients для аудитории), Фаза 1 Bot Service (для отправки)
**Сложность**: высокая (очередь отправки, авто-сценарии, статистика)

#### 2B.1 — Миграции

```
backend/migrations/YYYYMMDDHHMMSS_campaigns.sql
```

```sql
-- +goose Up
CREATE TABLE campaigns (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id),
    bot_id INT NOT NULL REFERENCES bots(id),
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('manual', 'auto')),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed')),
    audience_filter JSONB DEFAULT '{}'::jsonb,
    message TEXT NOT NULL,
    media_url VARCHAR(500),
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    stats JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE campaign_messages (
    id SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    client_id INT NOT NULL REFERENCES bot_clients(id),
    telegram_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auto_scenarios (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id),
    bot_id INT NOT NULL REFERENCES bots(id),
    name VARCHAR(200) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('inactive_days', 'visit_count', 'bonus_threshold', 'level_up', 'birthday')),
    trigger_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    message TEXT NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_org_id ON campaigns(org_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaign_messages_campaign_id ON campaign_messages(campaign_id);
CREATE INDEX idx_campaign_messages_status ON campaign_messages(status);
CREATE INDEX idx_auto_scenarios_org_id ON auto_scenarios(org_id);
CREATE INDEX idx_auto_scenarios_trigger_type ON auto_scenarios(trigger_type);

-- +goose Down
DROP TABLE IF EXISTS auto_scenarios;
DROP TABLE IF EXISTS campaign_messages;
DROP TABLE IF EXISTS campaigns;
```

#### 2B.2 — Entity

```
backend/internal/entity/campaign.go       — NEW
backend/internal/entity/auto_scenario.go  — NEW
```

Campaign:
- `Campaign` struct с `AudienceFilter AudienceFilter` (JSONB Scan/Value)
- `CampaignMessage` struct
- `CampaignStats` (JSONB):
  ```go
  type CampaignStats struct {
      Total    int `json:"total"`
      Sent     int `json:"sent"`
      Failed   int `json:"failed"`
  }
  ```
- `AudienceFilter`:
  ```go
  type AudienceFilter struct {
      BotID    *int      `json:"bot_id,omitempty"`
      PosID    *int      `json:"pos_id,omitempty"`
      Tags     []string  `json:"tags,omitempty"`
      Segments []string  `json:"segments,omitempty"`
  }
  ```
- Request types: `CreateCampaignRequest`, `UpdateCampaignRequest`

AutoScenario:
- `AutoScenario` struct
- `TriggerConfig` (JSONB Scan/Value):
  ```go
  type TriggerConfig struct {
      Days      *int `json:"days,omitempty"`       // inactive_days
      Count     *int `json:"count,omitempty"`       // visit_count
      Threshold *int `json:"threshold,omitempty"`   // bonus_threshold
  }
  ```
- Request types: `CreateScenarioRequest`, `UpdateScenarioRequest`

#### 2B.3 — Repository: Campaigns

```
backend/internal/repository/postgres/campaigns.go       — NEW
backend/internal/repository/postgres/auto_scenarios.go   — NEW
```

Campaigns:
- `Create(ctx, campaign) (*Campaign, error)`
- `GetByID(ctx, id) (*Campaign, error)`
- `GetByOrgID(ctx, orgID, limit, offset) ([]Campaign, int, error)`
- `Update(ctx, campaign) error`
- `UpdateStatus(ctx, id, status) error`
- `UpdateStats(ctx, id, stats) error`
- `Delete(ctx, id) error`

CampaignMessages:
- `CreateBatch(ctx, messages []CampaignMessage) error` — bulk insert
- `UpdateStatus(ctx, id, status, errorMsg) error`
- `GetByCampaignID(ctx, campaignID, limit, offset) ([]CampaignMessage, int, error)`
- `GetStatsByCampaignID(ctx, campaignID) (*CampaignStats, error)`

AutoScenarios:
- `Create(ctx, scenario) (*AutoScenario, error)`
- `GetByOrgID(ctx, orgID) ([]AutoScenario, error)`
- `GetByID(ctx, id) (*AutoScenario, error)`
- `Update(ctx, scenario) error`
- `Delete(ctx, id) error`
- `GetActiveByTriggerType(ctx, triggerType) ([]AutoScenario, error)`

#### 2B.4 — Usecase: Campaigns

```
backend/internal/usecase/campaigns/campaigns.go — NEW
backend/internal/usecase/campaigns/errors.go    — NEW
backend/internal/usecase/campaigns/sender.go    — NEW
```

Зависимости: campaigns repo, clients repo (для подсчёта аудитории), Redis (очередь)

Методы:
- `Create(ctx, orgID, req) (*Campaign, error)`
- `List(ctx, orgID, limit, offset) ([]Campaign, int, error)`
- `GetByID(ctx, orgID, id) (*Campaign, error)`
- `Update(ctx, orgID, id, req) error`
- `Delete(ctx, orgID, id) error`
- `Send(ctx, orgID, id) error`
  1. Получить кампанию → проверить статус (draft/scheduled)
  2. Получить клиентов по audience_filter
  3. Создать campaign_messages (bulk insert)
  4. Поставить задачу в Redis queue
  5. Обновить статус → sending
- `PreviewAudience(ctx, orgID, filter) (int, error)` — кол-во клиентов

AutoScenarios:
- `CreateScenario(ctx, orgID, req) (*AutoScenario, error)`
- `ListScenarios(ctx, orgID) ([]AutoScenario, error)`
- `UpdateScenario(ctx, orgID, id, req) error`
- `DeleteScenario(ctx, orgID, id) error`
- `ToggleScenario(ctx, orgID, id, active) error`

Sender (горутина для обработки очереди):
- `ProcessQueue(ctx) error` — читает из Redis, отправляет через Telegram API
- Rate limiting: 30 сообщений/сек (лимит Telegram)
- Обновление статусов по мере отправки
- Обновление campaign stats после завершения

Sentinel errors: `ErrCampaignNotFound`, `ErrNotCampaignOwner`, `ErrCampaignAlreadySent`

#### 2B.5 — Controller: Campaigns Group

```
backend/internal/controller/http/group/campaigns/campaigns.go — NEW
```

- `Path()` → `"/campaigns"`
- `Auth()` → JWT middleware
- Handlers:
  - `POST /` → create campaign
  - `GET /` → list campaigns (with pagination)
  - `GET /:id` → get campaign with message stats
  - `PATCH /:id` → update campaign
  - `DELETE /:id` → delete campaign (only draft)
  - `POST /:id/send` → trigger send
  - `POST /preview` → preview audience count
  - `GET /scenarios` → list auto scenarios
  - `POST /scenarios` → create scenario
  - `PATCH /scenarios/:id` → update scenario
  - `DELETE /scenarios/:id` → delete scenario
  - `PATCH /scenarios/:id/toggle` → activate/deactivate

#### 2B.6 — Sender Worker integration

Отправка рассылок происходит в горутине внутри server-процесса (для MVP).
Sender читает задачи из Redis list (`campaign_queue:{orgID}`) и отправляет сообщения.

В `cmd/server/main.go`:
- Создать sender goroutine
- Graceful shutdown через context

Паттерн отправки через бот:
- Campaigns usecase получает `botToken` из bots repo по `campaign.BotID`
- Использует `telego.Bot` для отправки `SendMessage`
- Rate limiting через `time.Ticker` (30 msg/sec)

#### 2B.7 — Wiring

```
backend/cmd/server/main.go — MODIFIED
```

#### 2B.8 — Frontend: Рассылки

```
frontend/src/features/campaigns/types.ts   — NEW
frontend/src/features/campaigns/api.ts     — NEW
frontend/src/features/campaigns/queries.ts — NEW
```

Страницы:
```
frontend/src/routes/dashboard/campaigns/index.tsx      — NEW (архив + список)
frontend/src/routes/dashboard/campaigns/create.tsx     — NEW (создание)
frontend/src/routes/dashboard/campaigns/$campaignId.tsx — NEW (детали)
frontend/src/routes/dashboard/campaigns/scenarios.tsx  — NEW (авто-сценарии)
```

Компоненты:
```
frontend/src/components/campaigns/AudienceFilter.tsx    — NEW (фильтр аудитории)
frontend/src/components/campaigns/MessagePreview.tsx    — NEW (предпросмотр)
frontend/src/components/campaigns/CampaignStats.tsx     — NEW (статистика)
frontend/src/components/campaigns/ScenarioCard.tsx      — NEW (карточка сценария)
```

**Создание рассылки** (create.tsx):
1. Шаг 1: Выбор аудитории (фильтры + показ количества подходящих клиентов)
2. Шаг 2: Текст сообщения + файлы (media_url)
3. Шаг 3: Время отправки (сейчас / запланировать)
4. Шаг 4: Предпросмотр + подтверждение

**Архив рассылок** (index.tsx):
- Таблица: название, дата, тип (manual/auto), охват, отправлено/ошибки, статус
- Фильтр по статусу и типу

**Авто-сценарии** (scenarios.tsx):
- Список карточек-шаблонов (день рождения, не был N дней, N-й визит, новый уровень, порог бонусов)
- Toggle вкл/выкл для каждого
- Клик → настройка условий + текст сообщения

**Проверка 2B**:
```bash
# Создание кампании
curl -X POST localhost:8080/api/v1/campaigns \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"bot_id":1,"name":"Тест","message":"Привет!","audience_filter":{}}'
# → 201

# Preview аудитории
curl -X POST localhost:8080/api/v1/campaigns/preview \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"bot_id":1}'
# → 200 {count: N}

# Отправка
curl -X POST localhost:8080/api/v1/campaigns/1/send \
  -H "Authorization: Bearer $TOKEN"
# → 200

npm run dev
# → /dashboard/campaigns → архив
# → /dashboard/campaigns/create → wizard
# → /dashboard/campaigns/scenarios → авто-сценарии
```

---

## Step 3: Интеграция

**Зависимости**: Steps 2A + 2B завершены

### 3.1 — Dashboard ← Campaign Stats

Обновить `dashboard.go` repo:
- Виджет "Рассылки": общее кол-во отправленных за период
- Тренд по сравнению с предыдущим периодом

### 3.2 — Auto-scenarios ↔ Bot Service

Интеграция авто-сценариев с bot service:
- **Birthday**: scheduler (cron/ticker) проверяет `birth_date = TODAY` → отправляет
- **Inactive days**: scheduler проверяет `last_activity < NOW() - N days` → отправляет
- **Level up**: hook в loyalty usecase → при смене уровня → отправляет
- **Visit count**: hook в loyalty usecase → при N-й транзакции → отправляет
- **Bonus threshold**: hook в loyalty usecase → при достижении порога → отправляет

Реализация:
```
backend/internal/service/scheduler/scheduler.go — NEW
```

Scheduler запускается в server main.go, проверяет time-based сценарии с интервалом (1 час).
Event-based сценарии (level_up, visit_count, bonus_threshold) вызываются из loyalty usecase через callback interface.

### 3.3 — Sidebar обновление

```
frontend/src/components/layout/Sidebar.tsx — MODIFIED
```

- Добавить пункты: "Дашборд" (home icon), "Рассылки"
- Дашборд как главная страница после логина

---

## Step 4: Polish + Testing + Deploy

**Зависимости**: Steps 1-3

### 4.1 — Frontend polish

- Skeleton loading для таблицы клиентов и графиков дашборда
- Error states для API ошибок
- Responsive: таблицы с горизонтальным скроллом на мобильных
- Виджеты дашборда: 2 колонки на мобильных, 4-5 на десктопе

### 4.2 — Backend tests

```
backend/internal/usecase/clients/clients_test.go    — NEW
backend/internal/usecase/dashboard/dashboard_test.go — NEW
backend/internal/usecase/campaigns/campaigns_test.go — NEW
```

Unit-тесты для usecases с mock repositories.

### 4.3 — Frontend build check

```bash
cd frontend && npm run lint && npm run build
```

### 4.4 — Deploy

```bash
git push origin main
# GitHub Actions → build → deploy
```

### 4.5 — Production Verification

- [ ] `/dashboard` → виджеты с реальными данными + графики
- [ ] `/dashboard/clients` → таблица с пагинацией, поиск работает
- [ ] Клик на клиента → профиль с историей транзакций
- [ ] Создание ручной рассылки → отправка → получение в Telegram
- [ ] Авто-сценарий "День рождения" активен → срабатывает
- [ ] Фильтры дашборда работают (период, бот, заведение)

---

## Граф зависимостей

```
Step 1A (Clients Backend) ───────────────────────┐
Step 1B (Clients Frontend) ←── depends on 1A     │
                                                   │
Step 2A (Dashboard)    ←── depends on Step 1 ─────┤
Step 2B (Campaigns)    ←── depends on Step 1 ─────┼──▶ Step 3 (Integration) ──▶ Step 4 (Polish + Deploy)
                                                   │
                                                   │

Параллелизация:
- 2A + 2B — полная параллелизация (Dashboard + Campaigns)
- 2A.6 + 2B.8 (frontend) — полная параллелизация
```

---

## Оптимальная стратегия с Claude Code

### Порядок реализации

1. **Step 1A** → `backend-architect` agent → Clients backend (миграция → entity → repo → usecase → controller → wiring)
2. **Step 1B** → `frontend-architect` agent → Clients frontend (feature → таблица → профиль → sidebar)
3. **Step 2A + 2B** → два параллельных агента:
   - `backend-architect` → Dashboard backend + frontend
   - `backend-architect` → Campaigns backend + frontend (более сложный)
4. **Step 3** → `backend-architect` → Integration (dashboard stats, auto-scenarios scheduler)
5. **Step 4** → `quality-engineer` → Tests + Polish + Deploy

### Файлы, которые будут созданы/изменены

```
# Новые миграции (2 файла)
backend/migrations/XXXXXX_extend_bot_clients.sql
backend/migrations/XXXXXX_campaigns.sql

# Entity (4 файла: 1 изменён, 3 новых)
backend/internal/entity/bot_client.go     MODIFIED
backend/internal/entity/client.go         NEW
backend/internal/entity/dashboard.go      NEW
backend/internal/entity/campaign.go       NEW
backend/internal/entity/auto_scenario.go  NEW

# Repository (4 файла)
backend/internal/repository/postgres/clients.go        NEW
backend/internal/repository/postgres/dashboard.go      NEW
backend/internal/repository/postgres/campaigns.go      NEW
backend/internal/repository/postgres/auto_scenarios.go NEW

# Usecase (6 файлов)
backend/internal/usecase/clients/clients.go       NEW
backend/internal/usecase/clients/errors.go        NEW
backend/internal/usecase/dashboard/dashboard.go   NEW
backend/internal/usecase/campaigns/campaigns.go   NEW
backend/internal/usecase/campaigns/errors.go      NEW
backend/internal/usecase/campaigns/sender.go      NEW

# Controller (3 файла)
backend/internal/controller/http/group/clients/clients.go       NEW
backend/internal/controller/http/group/dashboard/dashboard.go   NEW
backend/internal/controller/http/group/campaigns/campaigns.go   NEW

# Service (1 файл)
backend/internal/service/scheduler/scheduler.go   NEW

# Entrypoint (1 файл изменён)
backend/cmd/server/main.go   MODIFIED

# Frontend Clients (5 файлов)
frontend/src/features/clients/types.ts              NEW
frontend/src/features/clients/api.ts                NEW
frontend/src/features/clients/queries.ts            NEW
frontend/src/routes/dashboard/clients/index.tsx     NEW
frontend/src/routes/dashboard/clients/$clientId.tsx NEW

# Frontend Dashboard (6 файлов)
frontend/src/features/dashboard/types.ts            NEW
frontend/src/features/dashboard/api.ts              NEW
frontend/src/features/dashboard/queries.ts          NEW
frontend/src/routes/dashboard/index.tsx             MODIFIED
frontend/src/components/dashboard/WidgetCard.tsx    NEW
frontend/src/components/dashboard/ChartCard.tsx     NEW
frontend/src/components/dashboard/FilterBar.tsx     NEW

# Frontend Campaigns (10 файлов)
frontend/src/features/campaigns/types.ts                    NEW
frontend/src/features/campaigns/api.ts                      NEW
frontend/src/features/campaigns/queries.ts                  NEW
frontend/src/routes/dashboard/campaigns/index.tsx           NEW
frontend/src/routes/dashboard/campaigns/create.tsx          NEW
frontend/src/routes/dashboard/campaigns/$campaignId.tsx     NEW
frontend/src/routes/dashboard/campaigns/scenarios.tsx       NEW
frontend/src/components/campaigns/AudienceFilter.tsx        NEW
frontend/src/components/campaigns/MessagePreview.tsx        NEW
frontend/src/components/campaigns/CampaignStats.tsx         NEW
frontend/src/components/campaigns/ScenarioCard.tsx          NEW

# Sidebar (1 файл)
frontend/src/components/layout/Sidebar.tsx   MODIFIED

# Tests (3 файла)
backend/internal/usecase/clients/clients_test.go      NEW
backend/internal/usecase/dashboard/dashboard_test.go   NEW
backend/internal/usecase/campaigns/campaigns_test.go   NEW

# npm dependency
recharts (charts library)

# Итого: ~40 новых файлов, ~4 изменённых
```

---

## Критерии завершения Фазы 2

- [ ] Таблица клиентов с серверной пагинацией, фильтрами, поиском
- [ ] Профиль клиента с историей транзакций и тегами
- [ ] Дашборд с виджетами (выручка, ср. чек, новые, активные) + тренды
- [ ] Графики: выручка по дням, регистрации по дням, средний чек
- [ ] Фильтры дашборда: период, бот, заведение
- [ ] Ручная рассылка: создать → выбрать аудиторию → отправить → увидеть статистику
- [ ] Хотя бы 3 авто-сценария работают (день рождения, не был N дней, новый уровень)
- [ ] Все данные фильтруются по org_id текущего пользователя
