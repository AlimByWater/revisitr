# Workflow: Phase 3 — Аналитика и интеграции

> Сгенерирован на основе `PHASE-3-analytics.md` и анализа существующих паттернов кодовой базы.
> Актуализирован: 2026-03-19 — исправления по ревью (MV, scheduler, тесты, FunnelChart).

## Анализ зависимостей и критический путь

```
3.1 Analytics ──────────────────────────────────► 3.1 Frontend
     │
     └──► 3.2 Segmentation (миграция segments)
               │
               └──► используется в 3.1 фильтрах
                    используется в 3.4 условиях акций

3.3 POS Integrations ──────────────────────────► независимо
     └──► external_orders питают 3.1 sales analytics

3.4 Promotions ─────────────────────────────────► зависит от 3.2 сегментов
```

**Рекомендуемый порядок**: `3.1 BE → 3.2 BE → 3.1 FE + 3.2 FE` параллельно → `3.4 BE → 3.4 FE` → `3.3 BE → 3.3 FE`

---

## Wave 1 — Backend Foundation (~6–8 дней)

### Step 1 — Миграции

Все миграции создать сразу, чтобы не блокировать параллельную работу.

**`migrations/00009_analytics_segments.sql`**
```sql
-- +goose Up
CREATE TABLE segments (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('rfm', 'custom')),
    filter      JSONB NOT NULL DEFAULT '{}',
    auto_assign BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_segments_org_id ON segments(org_id);

CREATE TABLE segment_clients (
    segment_id  INT NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    client_id   INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (segment_id, client_id)
);

ALTER TABLE bot_clients
    ADD COLUMN IF NOT EXISTS rfm_recency    INT,
    ADD COLUMN IF NOT EXISTS rfm_frequency  INT,
    ADD COLUMN IF NOT EXISTS rfm_monetary   NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS rfm_segment    VARCHAR(50),
    ADD COLUMN IF NOT EXISTS rfm_updated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS rfm_recency,
    DROP COLUMN IF EXISTS rfm_frequency,
    DROP COLUMN IF EXISTS rfm_monetary,
    DROP COLUMN IF EXISTS rfm_segment,
    DROP COLUMN IF EXISTS rfm_updated_at;
DROP TABLE IF EXISTS segment_clients;
DROP TABLE IF EXISTS segments;
```

**`migrations/00010_promotions.sql`**
```sql
-- +goose Up
CREATE TABLE promotions (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(30) NOT NULL CHECK (type IN ('discount','bonus','tag_update','campaign')),
    conditions  JSONB NOT NULL DEFAULT '{}',
    result      JSONB NOT NULL DEFAULT '{}',
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,
    usage_limit INT,
    combinable  BOOLEAN NOT NULL DEFAULT true,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_promotions_org_id ON promotions(org_id);

CREATE TABLE promo_codes (
    id               SERIAL PRIMARY KEY,
    org_id           INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    promotion_id     INT REFERENCES promotions(id) ON DELETE SET NULL,
    code             VARCHAR(50) NOT NULL,
    discount_percent NUMERIC(5,2),
    bonus_amount     INT,
    starts_at        TIMESTAMPTZ,
    ends_at          TIMESTAMPTZ,
    conditions       JSONB NOT NULL DEFAULT '{}',
    usage_count      INT NOT NULL DEFAULT 0,
    usage_limit      INT,
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, code)
);

CREATE TABLE promotion_usages (
    id            SERIAL PRIMARY KEY,
    promotion_id  INT NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    client_id     INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    promo_code_id INT REFERENCES promo_codes(id) ON DELETE SET NULL,
    used_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_promotion_usages_promotion_id ON promotion_usages(promotion_id);
CREATE INDEX idx_promotion_usages_client_id    ON promotion_usages(client_id);

-- +goose Down
DROP TABLE IF EXISTS promotion_usages;
DROP TABLE IF EXISTS promo_codes;
DROP TABLE IF EXISTS promotions;
```

**`migrations/00011_pos_integrations.sql`**
```sql
-- +goose Up
CREATE TABLE integrations (
    id           SERIAL PRIMARY KEY,
    org_id       INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type         VARCHAR(30) NOT NULL CHECK (type IN ('iiko','rkeeper','1c')),
    config       JSONB NOT NULL DEFAULT '{}',
    status       VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('active','inactive','error')),
    last_sync_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_integrations_org_id ON integrations(org_id);

CREATE TABLE external_orders (
    id             SERIAL PRIMARY KEY,
    integration_id INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_id    VARCHAR(255) NOT NULL,
    client_id      INT REFERENCES bot_clients(id) ON DELETE SET NULL,
    items          JSONB NOT NULL DEFAULT '[]',
    total          NUMERIC(12,2) NOT NULL DEFAULT 0,
    ordered_at     TIMESTAMPTZ,
    synced_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, external_id)
);
CREATE INDEX idx_external_orders_client_id      ON external_orders(client_id);
CREATE INDEX idx_external_orders_integration_id ON external_orders(integration_id);

-- +goose Down
DROP TABLE IF EXISTS external_orders;
DROP TABLE IF EXISTS integrations;
```

**`migrations/00012_analytics_views.sql`** — материализованные вью

> **Важно**: unique index обязателен для `REFRESH MATERIALIZED VIEW CONCURRENTLY`.
> `active_clients` убран из MV — `NOW()` в MV фиксирует время создания/рефреша,
> между рефрешами данные будут устаревшими. Считаем active_clients в runtime-запросе
> в `analytics.go` repo (WHERE last_activity > NOW() - INTERVAL '30 days').

```sql
-- +goose Up
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_daily_sales AS
SELECT
    DATE_TRUNC('day', lt.created_at) AS day,
    bc.bot_id,
    b.org_id,
    COUNT(DISTINCT lt.client_id)     AS unique_clients,
    COUNT(*)                         AS transaction_count,
    SUM(lt.amount)                   AS total_amount,
    AVG(lt.amount)                   AS avg_amount
FROM loyalty_transactions lt
JOIN bot_clients bc ON lt.client_id = bc.id
JOIN bots b ON bc.bot_id = b.id
WHERE lt.type = 'earn'
GROUP BY 1, 2, 3
WITH DATA;
CREATE UNIQUE INDEX ON mv_daily_sales(day, bot_id);

-- mv_loyalty_stats хранит только агрегаты регистраций по дням.
-- active_clients считается в runtime-запросе (analytics repo), а не в MV,
-- потому что NOW() в MV фиксируется на момент REFRESH и устаревает между рефрешами.
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_loyalty_stats AS
SELECT
    b.org_id,
    bc.bot_id,
    DATE_TRUNC('day', bc.registered_at) AS day,
    COUNT(*)                             AS new_clients
FROM bot_clients bc
JOIN bots b ON bc.bot_id = b.id
GROUP BY 1, 2, 3
WITH DATA;
CREATE UNIQUE INDEX ON mv_loyalty_stats(org_id, bot_id, day);

-- +goose Down
DROP MATERIALIZED VIEW IF EXISTS mv_loyalty_stats;
DROP MATERIALIZED VIEW IF EXISTS mv_daily_sales;
```

---

### Step 2 — Entity слой

**`internal/entity/analytics.go`**
```go
package entity

import "time"

// Единые фильтры для всех аналитических endpoints
type AnalyticsFilter struct {
    OrgID     int
    BotID     *int
    POSID     *int
    SegmentID *int
    From      time.Time
    To        time.Time
}

// --- Sales ---
type SalesMetrics struct {
    TransactionCount int64   `json:"transaction_count" db:"transaction_count"`
    UniqueClients    int64   `json:"unique_clients"    db:"unique_clients"`
    TotalAmount      float64 `json:"total_amount"      db:"total_amount"`
    AvgAmount        float64 `json:"avg_amount"        db:"avg_amount"`
    BuyFrequency     float64 `json:"buy_frequency"     db:"buy_frequency"`
}
type SalesChartPoint struct {
    Day   time.Time `json:"day"   db:"day"`
    Value float64   `json:"value" db:"value"`
}
type SalesAnalytics struct {
    Metrics    SalesMetrics                 `json:"metrics"`
    Charts     map[string][]SalesChartPoint `json:"charts"` // "transactions","revenue","avg_amount"
    Comparison *LoyaltyComparison           `json:"comparison,omitempty"`
}
type LoyaltyComparison struct {
    ParticipantsAvgAmount    float64 `json:"participants_avg_amount"`
    NonParticipantsAvgAmount float64 `json:"non_participants_avg_amount"`
}

// --- Loyalty ---
type LoyaltyAnalytics struct {
    NewClients    int64              `json:"new_clients"`
    ActiveClients int64              `json:"active_clients"`
    BonusEarned   float64            `json:"bonus_earned"`
    BonusSpent    float64            `json:"bonus_spent"`
    Demographics  ClientDemographics `json:"demographics"`
    BotFunnel     []FunnelStep       `json:"bot_funnel"`
}
type ClientDemographics struct {
    ByGender       []PieSlice `json:"by_gender"`
    ByAgeGroup     []PieSlice `json:"by_age_group"`
    ByOS           []PieSlice `json:"by_os"`
    LoyaltyPercent float64    `json:"loyalty_percent"`
}
type PieSlice struct {
    Label   string  `json:"label"`
    Value   int64   `json:"value"`
    Percent float64 `json:"percent"`
}
type FunnelStep struct {
    Step    string  `json:"step"`
    Count   int64   `json:"count"`
    Percent float64 `json:"percent"`
}

// --- Campaigns ---
type CampaignAnalytics struct {
    TotalSent   int64          `json:"total_sent"`
    TotalOpened int64          `json:"total_opened"`
    OpenRate    float64        `json:"open_rate"`
    Conversions int64          `json:"conversions"`
    ConvRate    float64        `json:"conv_rate"`
    ByCampaign  []CampaignStat `json:"by_campaign"`
}
type CampaignStat struct {
    CampaignID   int     `json:"campaign_id"`
    CampaignName string  `json:"campaign_name"`
    Sent         int64   `json:"sent"`
    OpenRate     float64 `json:"open_rate"`
    Conversions  int64   `json:"conversions"`
}
```

**`internal/entity/segment.go`**
```go
package entity

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// RFM категории
const (
    RFMChampions   = "champions"
    RFMLoyal       = "loyal"
    RFMAtRisk      = "at_risk"
    RFMCantLose    = "cant_lose"
    RFMHibernating = "hibernating"
    RFMLost        = "lost"
)

type Segment struct {
    ID          int           `json:"id"           db:"id"`
    OrgID       int           `json:"org_id"       db:"org_id"`
    Name        string        `json:"name"         db:"name"`
    Type        string        `json:"type"         db:"type"` // "rfm"|"custom"
    Filter      SegmentFilter `json:"filter"       db:"filter"`
    AutoAssign  bool          `json:"auto_assign"  db:"auto_assign"`
    ClientCount *int          `json:"client_count,omitempty" db:"-"`
    CreatedAt   time.Time     `json:"created_at"   db:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at"   db:"updated_at"`
}

type SegmentFilter struct {
    Gender      *string  `json:"gender,omitempty"`
    AgeFrom     *int     `json:"age_from,omitempty"`
    AgeTo       *int     `json:"age_to,omitempty"`
    MinVisits   *int     `json:"min_visits,omitempty"`
    MaxVisits   *int     `json:"max_visits,omitempty"`
    MinSpend    *float64 `json:"min_spend,omitempty"`
    MaxSpend    *float64 `json:"max_spend,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    RFMCategory *string  `json:"rfm_category,omitempty"`
}

func (f SegmentFilter) Value() (driver.Value, error) { return json.Marshal(f) }
func (f *SegmentFilter) Scan(src interface{}) error {
    b, ok := src.([]byte)
    if !ok {
        return fmt.Errorf("SegmentFilter.Scan: expected []byte")
    }
    return json.Unmarshal(b, f)
}

type CreateSegmentRequest struct {
    Name       string        `json:"name"        binding:"required"`
    Type       string        `json:"type"        binding:"required,oneof=rfm custom"`
    Filter     SegmentFilter `json:"filter"`
    AutoAssign bool          `json:"auto_assign"`
}
type UpdateSegmentRequest struct {
    Name       *string        `json:"name"`
    Filter     *SegmentFilter `json:"filter"`
    AutoAssign *bool          `json:"auto_assign"`
}
```

**`internal/entity/promotion.go`** — аналогично с JSONB Scan/Value для `PromotionConditions` и `PromotionResult`.

**`internal/entity/integration.go`** — аналогично с JSONB для `IntegrationConfig` и `ExternalOrder.Items`.

---

### Step 3 — Repository слой

| Файл | Ключевые методы |
|------|----------------|
| `postgres/analytics.go` | `GetSalesMetrics`, `GetSalesCharts`, `GetLoyaltyAnalytics` (включая runtime-подсчёт active_clients), `GetCampaignAnalytics`, `RefreshMaterializedViews` (использовать `CONCURRENTLY`) |
| `postgres/segments.go` | CRUD + `GetClients(segmentID, limit, offset)`, `SyncClients(segmentID, []clientIDs)` |
| `postgres/integrations.go` | CRUD + `UpdateLastSync`, `UpsertOrder` (ON CONFLICT DO UPDATE) |
| `postgres/promotions.go` | CRUD promotions + CRUD promo_codes + `GetApplicable(orgID, clientID)`, `RecordUsage` |

**Паттерны из существующего кода:**
- `clients.go` → динамический SQL с параметрами `$N` для аналитических фильтров
- `campaigns.go` → JSONB Scan/Value для conditions/result полей
- `clients.go` → `COUNT(*) OVER()` для пагинации без второго запроса

---

### Step 4 — Usecase слой

**`internal/usecase/analytics/analytics.go`**
```go
// Тонкий слой — делегирует в repo, добавляет org ownership + валидацию фильтров
func (uc *Usecase) GetSalesAnalytics(ctx, orgID int, filter entity.AnalyticsFilter) (*entity.SalesAnalytics, error)
func (uc *Usecase) GetLoyaltyAnalytics(ctx, orgID int, filter entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error)
func (uc *Usecase) GetCampaignAnalytics(ctx, orgID int, filter entity.AnalyticsFilter) (*entity.CampaignAnalytics, error)
```

**`internal/usecase/segments/segments.go`**
```go
var (
    ErrSegmentNotFound = fmt.Errorf("segment not found")
    ErrNotSegmentOwner = fmt.Errorf("not authorized")
)
func (uc *Usecase) Create(ctx, orgID int, req *entity.CreateSegmentRequest) (*entity.Segment, error)
func (uc *Usecase) GetByOrgID(ctx, orgID int) ([]entity.Segment, error)
func (uc *Usecase) GetByID(ctx, id, orgID int) (*entity.Segment, error)
func (uc *Usecase) Update(ctx, id, orgID int, req *entity.UpdateSegmentRequest) (*entity.Segment, error)
func (uc *Usecase) Delete(ctx, id, orgID int) error
func (uc *Usecase) GetClients(ctx, segmentID, orgID, limit, offset int) ([]entity.BotClient, int, error)
func (uc *Usecase) RecalculateCustom(ctx, segmentID, orgID int) error
```

**`internal/usecase/promotions/promotions.go`**
```go
func (uc *Usecase) Create/List/GetByID/Update/Delete  // promotions CRUD
func (uc *Usecase) CreatePromoCode/ListPromoCodes     // promo codes
func (uc *Usecase) ApplyPromoCode(ctx, orgID, clientID int, code string) (*entity.PromoResult, error)
```

**`internal/usecase/integrations/integrations.go`**
```go
func (uc *Usecase) Create/List/Update/Delete
func (uc *Usecase) SyncNow(ctx, id, orgID int) error  // вызывает POS-клиент
```

---

### Step 5 — Service: RFM

**`internal/service/rfm/rfm.go`** — новый сервис, вызывается шедулером каждые 24ч
```go
type Service struct {
    clientsRepo rfmClientsRepo  // GetAll + UpdateRFM
    txRepo      rfmTxRepo       // GetPerClient (из loyalty_transactions)
}

func (s *Service) RecalculateAll(ctx context.Context, orgID int) error
// Scoring: 1–5 по каждому измерению
// Recency:   5=<7d, 4=<30d, 3=<90d, 2=<180d, 1=else
// Frequency: 5=10+, 4=7+, 3=4+, 2=2+, 1=else
// Monetary:  5=top20%, 4=top40%, 3=top60%, 2=top80%, 1=else
// Segment:   champions(≥13), loyal(10–12), at_risk(7–9), cant_lose(5–6), hibernating(3–4), lost(<3)
```

---

### Step 6 — Controller слой

| Group | Path | Handlers |
|-------|------|----------|
| `analytics` | `/api/v1/analytics` | `GET /sales`, `GET /loyalty`, `GET /campaigns` |
| `segments` | `/api/v1/segments` | CRUD + `GET /:id/clients` |
| `promotions` | `/api/v1/promotions` | CRUD promotions + `POST /promo-codes`, `GET /promo-codes` |
| `integrations` | `/api/v1/integrations` | CRUD + `POST /:id/sync` |

**Query params для аналитики** (единые для всех 3 endpoints):
```
GET /api/v1/analytics/sales?from=2024-01-01&to=2024-03-31&bot_id=1&segment_id=2
```

---

### Step 7 — Scheduler: `internal/controller/scheduler/`

Создать единый scheduler в слое controller — это input adapter (как HTTP handler),
принимает тик таймера → вызывает usecase/service. Тонкий оркестратор, без бизнес-логики.

> **Существующий** `service/campaign/scheduler.go` запускается из `cmd/bot/main.go`.
> Его не трогаем — он относится к боту. Новый scheduler — для `cmd/server`.

**`internal/controller/scheduler/scheduler.go`**
```go
type Task struct {
    Name     string
    Interval time.Duration
    Fn       func(ctx context.Context) error
}

type Scheduler struct {
    tasks  []Task
    logger *slog.Logger
}

func New(logger *slog.Logger) *Scheduler {
    return &Scheduler{logger: logger}
}

func (s *Scheduler) Register(task Task) {
    s.tasks = append(s.tasks, task)
}

func (s *Scheduler) Run(ctx context.Context) {
    for _, task := range s.tasks {
        go s.runTask(ctx, task)
    }
    <-ctx.Done()
}

func (s *Scheduler) runTask(ctx context.Context, task Task) {
    s.logger.Info("scheduler: task started", "task", task.Name, "interval", task.Interval)
    // Run immediately once
    if err := task.Fn(ctx); err != nil {
        s.logger.Error("scheduler: task failed", "task", task.Name, "error", err)
    }
    ticker := time.NewTicker(task.Interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := task.Fn(ctx); err != nil {
                s.logger.Error("scheduler: task failed", "task", task.Name, "error", err)
            }
        }
    }
}
```

**Регистрация задач в `cmd/server/main.go`:**
```go
sched := scheduler.New(logger)
sched.Register(scheduler.Task{
    Name:     "refresh_materialized_views",
    Interval: 6 * time.Hour,
    Fn:       func(ctx context.Context) error { return analyticsRepo.RefreshMaterializedViews(ctx) },
})
sched.Register(scheduler.Task{
    Name:     "rfm_recalculate",
    Interval: 24 * time.Hour,
    Fn:       func(ctx context.Context) error { return rfmSvc.RecalculateAll(ctx) },
})
go sched.Run(appCtx)
```

### Step 8 — DI в cmd/server/main.go

Подключить все новые компоненты по существующему паттерну:
```go
analyticsRepo    := pgRepo.NewAnalytics(pg)
segmentsRepo     := pgRepo.NewSegments(pg)
promotionsRepo   := pgRepo.NewPromotions(pg)
integrationsRepo := pgRepo.NewIntegrations(pg)

analyticsUC    := analyticsUC.New(analyticsRepo)
segmentsUC     := segmentsUC.New(segmentsRepo, clientsRepo)
promotionsUC   := promotionsUC.New(promotionsRepo, segmentsRepo)
integrationsUC := integrationsUC.New(integrationsRepo, ordersRepo, clientsRepo)
rfmSvc         := rfm.New(clientsRepo, txRepo)

// Groups
analyticsGrp    := analyticsGroup.New(analyticsUC, jwtSecret)
segmentsGrp     := segmentsGroup.New(segmentsUC, jwtSecret)
promotionsGrp   := promotionsGroup.New(promotionsUC, jwtSecret)
integrationsGrp := integrationsGroup.New(integrationsUC, jwtSecret)

// Scheduler (controller layer)
sched := scheduler.New(logger)
sched.Register(scheduler.Task{
    Name: "refresh_materialized_views", Interval: 6 * time.Hour,
    Fn: func(ctx context.Context) error { return analyticsRepo.RefreshMaterializedViews(ctx) },
})
sched.Register(scheduler.Task{
    Name: "rfm_recalculate", Interval: 24 * time.Hour,
    Fn: func(ctx context.Context) error { return rfmSvc.RecalculateAll(ctx) },
})
go sched.Run(appCtx)
```

### Step 9 — Unit Tests

По установленному паттерну `usecase/<domain>/<domain>_test.go` с ручными моками через struct + function fields.

```
internal/usecase/analytics/analytics_test.go  — фильтрация, валидация периодов, org ownership
internal/usecase/segments/segments_test.go    — CRUD, ownership проверки, ErrSegmentNotFound
internal/usecase/promotions/promotions_test.go — CRUD, ApplyPromoCode (лимиты, период, условия)
internal/usecase/integrations/integrations_test.go — CRUD, ownership
internal/service/rfm/rfm_test.go              — scoring логика: границы R/F/M, segment assignment
```

RFM service — приоритет для unit-тестов: нетривиальная логика скоринга, чистые функции,
легко тестируется табличными тестами (table-driven tests).

**Запуск** (без инфраструктуры):
```bash
cd backend && go test -race ./internal/usecase/... ./internal/service/rfm/...
```

### Step 10 — Integration Tests

По паттерну `backend/tests/integration/`:
```
tests/integration/
    analytics_test.go    — GET /analytics/sales, /loyalty, /campaigns с реальными данными
    segments_test.go     — CRUD + GetClients + RFM score присваивается
    promotions_test.go   — CRUD + ApplyPromoCode + usage записывается
    integrations_test.go — CRUD + SyncNow (с mock iiko HTTP)
```

> **Важно**: в `setup_test.go` нужно добавить новые группы (analyticsGroup, segmentsGroup,
> promotionsGroup, integrationsGroup) в DI-секцию и в массив `groups`.
> Также рекомендуется добавить автоматический прогон goose-миграций в TestMain,
> чтобы Claude Code мог запускать integration tests автономно без ручной подготовки БД.

**Запуск** (требует `make dev-up`):
```bash
cd backend && go test -race -tags=integration -v ./tests/integration/...
```

---

## Wave 2 — POS Integrations (~3–4 дня)

### Step 11 — iiko API Client

**`internal/service/pos/iiko/client.go`**
```go
type Client struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}
func (c *Client) GetOrders(from, to time.Time) ([]IikoOrder, error)
func (c *Client) GetCustomerByPhone(phone string) (*IikoCustomer, error)
func (c *Client) MapToExternalOrder(o IikoOrder, clientID *int) entity.ExternalOrder
```

### Step 12 — SyncService

**`internal/service/pos/sync.go`**
```go
type SyncService struct {
    integrationsRepo integrationsRepo
    ordersRepo       ordersRepo
    clientsRepo      clientsRepo
    iiko             *iiko.Client
}
func (s *SyncService) Sync(ctx context.Context, integration *entity.Integration) error
// 1. Определить тип → выбрать клиент
// 2. Получить заказы начиная с last_sync_at
// 3. Привязать клиентов по телефону (GetByPhone в bot_clients)
// 4. UpsertOrder для каждого заказа
// 5. UpdateLastSync(id, "active")
```

### Step 13 — POS Sync в scheduler

Добавить задачу POS-синхронизации в `controller/scheduler`:
```go
sched.Register(scheduler.Task{
    Name:     "pos_sync",
    Interval: 4 * time.Hour,
    Fn:       func(ctx context.Context) error { return posSyncSvc.SyncAll(ctx) },
})
```

### Step 14 — POS Integration Tests (mock iiko HTTP)

```go
// tests/integration/integrations_test.go или отдельный pos_sync_test.go
// 1. Поднять httptest.NewServer с mock iiko API (возвращает фиксированные заказы)
// 2. Создать integration запись в БД с URL mock-сервера
// 3. Вызвать SyncNow → проверить что external_orders появились в БД
// 4. Повторный sync → проверить что дубликаты не создались (UNIQUE constraint)
```

---

## Wave 3 — Frontend (~5–6 дней)

### Step 15 — API Hooks

**`frontend/src/features/analytics/api.ts`**
```ts
export function useSalesAnalytics(filter: AnalyticsFilter) {
    return useQuery({
        queryKey: ['analytics', 'sales', filter],
        queryFn: () => api.get('/analytics/sales', { params: filter }),
    })
}
// аналогично: useLoyaltyAnalytics, useCampaignAnalytics
```

### Step 16 — Страницы аналитики

**`routes/dashboard/analytics/sales.tsx`**
```tsx
// DateRangeFilter — общий компонент для всех 3 страниц
// SalesWidgets    — 4 числовых виджета (transactions, clients, revenue, avg)
// SalesCharts     — LineChart для каждого показателя (recharts, уже в deps)
// LoyaltyComparison — BarChart: участники vs не-участники
```

**`routes/dashboard/analytics/loyalty.tsx`**
```tsx
// LoyaltyWidgets   — новые, активные, бонусы накоплено/списано
// DemographicsPie  — 4 PieChart (пол, возраст, ОС, % лояльности)
// BotFunnel        — горизонтальный BarChart (recharts не имеет нативного FunnelChart)
//                    каждый шаг — bar с шириной пропорциональной % от предыдущего шага
```

**`routes/dashboard/analytics/campaigns.tsx`**
```tsx
// CampaignWidgets  — отправлено, открыто, конверсии
// CampaignTable    — TanStack Table: name, sent, open_rate, conv_rate
```

### Step 17 — Сегменты

**`routes/dashboard/segments/index.tsx`**
```tsx
// RFMPieChart          — распределение по 6 категориям
// RFMTable             — сегмент + кол-во + % + ср. чек + покупки
// SegmentList          — кастомные сегменты
// SegmentFilterBuilder — конструктор условий:
//   [поле ▼] [оператор ▼] [значение] [+]
//   Поля: gender, age, visits, spend, tags, rfm_category
//   Preview: "Попадает 142 клиента"
```

### Step 18 — Акции и промокоды

**`routes/dashboard/promotions/create.tsx`** — 3-шаговый wizard
```tsx
// Step 1: PromotionBasicInfo  — name, period, usage_limit, combinable
// Step 2: PromotionConditions — segment selector + purchase conditions
// Step 3: PromotionResult     — type: discount|bonus|tag_update|campaign
//                               + параметры в зависимости от типа
```

**`routes/dashboard/promotions/index.tsx`** — список активных + архив

**`routes/dashboard/promotions/promo-codes.tsx`** — создание, список, QR-коды

### Step 19 — Интеграции

**`routes/dashboard/integrations/index.tsx`**
```tsx
// IntegrationCard — карточка для iiko/rkeeper/1c
//   Connect/Disconnect кнопка
//   Форма настройки (API key, URL — зависит от типа)
//   SyncStatusBadge — last_sync_at + статус
//   SyncLogsTable   — последние N синхронизаций
```

---

## Wave 4 — Bot Integration (~2 дня)

### Step 20 — Применение акций в боте

```go
// cmd/bot — расширить обработчик транзакции earn:
// 1. promotionsUC.GetApplicable(ctx, orgID, clientID)
// 2. Для каждой акции: проверить conditions (мин. сумма, сегмент, период)
// 3. Применить result:
//    - bonus → начислить доп. баллы
//    - tag_update → обновить теги клиента
//    - campaign → поставить в очередь рассылку

// Команда /promo <code>:
// promotionsUC.ApplyPromoCode(ctx, orgID, clientID, code)
// → сообщение с результатом: "Скидка 10% активирована"
```

---

## Итоговый порядок задач

```
┌─ Wave 1: Backend Foundation ──────────────────────────── ~7–9 дней ─┐
│  [ ] 1.  Миграции 00009–00012 (все сразу)                            │
│  [ ] 2.  Entity: analytics.go, segment.go, promotion.go, integration │
│  [ ] 3.  Repos: analytics, segments, promotions, integrations        │
│  [ ] 4.  Usecase: analytics, segments, promotions, integrations      │
│  [ ] 5.  Service: internal/service/rfm/rfm.go                        │
│  [ ] 6.  Controllers: analytics, segments, promotions, integrations  │
│  [ ] 7.  Scheduler: internal/controller/scheduler/ (тонкий оркестр.) │
│  [ ] 8.  DI в cmd/server/main.go (включая scheduler)                │
│  [ ] 9.  Unit tests: usecase/* + service/rfm (без инфры)             │
│  [ ] 10. Integration tests: обновить setup_test.go + новые тесты     │
└──────────────────────────────────────────────────────────────────────┘
┌─ Wave 2: POS Integration ──────────────────────────────── ~3–4 дня ─┐
│  [ ] 11. iiko API client (service/pos/iiko/)                         │
│  [ ] 12. SyncService оркестратор (service/pos/sync.go)               │
│  [ ] 13. POS sync задача в controller/scheduler                      │
│  [ ] 14. Integration tests (mock iiko HTTP server)                   │
└──────────────────────────────────────────────────────────────────────┘
┌─ Wave 3: Frontend ─────────────────────────────────────── ~5–6 дней ─┐
│  [ ] 15. features/analytics/ — API hooks + types                     │
│  [ ] 16. Страница «Продажи» (widgets + line charts)                  │
│  [ ] 17. Страница «Лояльность» (widgets + pie + horiz. BarChart)     │
│  [ ] 18. Страница «Рассылки» (widgets + campaign table)              │
│  [ ] 19. features/segments/ + RFM + SegmentFilterBuilder             │
│  [ ] 20. Конструктор акций (3-step wizard) + список + архив          │
│  [ ] 21. Промокоды — создание, список, деактивация                   │
│  [ ] 22. Страница «Интеграции» — карточки + настройка + логи         │
└──────────────────────────────────────────────────────────────────────┘
┌─ Wave 4: Bot ──────────────────────────────────────────── ~2 дня ───┐
│  [ ] 23. Применение акций при earn-транзакциях                       │
│  [ ] 24. Команда /promo <code> для промокодов                        │
└──────────────────────────────────────────────────────────────────────┘
```

## Локальное тестирование (Claude Code self-testing)

### Что можно тестировать без инфраструктуры

```bash
# Unit tests — работают сразу, без Docker
cd backend && go test -race ./internal/usecase/...
cd backend && go test -race ./internal/service/rfm/...
cd backend && go test -race ./internal/controller/scheduler/...
cd frontend && npx vitest run
```

### Что требует Docker (PG + Redis)

```bash
# 1. Поднять инфру
make dev-up

# 2. Integration tests
make test-integration
```

### Что требует full stack (для E2E)

```bash
# 1. Docker
make dev-up

# 2. Backend + Frontend
cd backend && go run ./cmd/server &
cd frontend && npm run dev &

# 3. E2E
cd e2e && npx playwright test
```

### Что нельзя протестировать локально

| Компонент | Почему | Workaround |
|-----------|--------|------------|
| iiko API (реальный) | Нужен доступ к iiko cloud | Mock HTTP server в integration tests — полностью покрывает логику sync |
| Telegram bot (реальный) | Нужен webhook от Telegram | Unit tests с mock telego + integration tests с httptest |
| Materialized view refresh (production timing) | 6h interval нецелесообразно ждать | Вызвать `RefreshMaterializedViews()` напрямую в тесте |
| RFM scheduler (24h interval) | Аналогично | Вызвать `rfmSvc.RecalculateAll()` напрямую в тесте |

> Все компоненты тестируемы локально. External APIs (iiko, Telegram) покрываются mock'ами.
> Scheduler-задачи тестируются через прямой вызов функций, без ожидания таймеров.

### Подготовка setup_test.go для Phase 3

В `backend/tests/integration/setup_test.go` нужно добавить:
1. Импорты новых групп (`analyticsGroup`, `segmentsGroup`, `promotionsGroup`, `integrationsGroup`)
2. Создание новых repo + usecase + groups в TestMain
3. (Рекомендация) Автоматический прогон goose-миграций в TestMain для автономности

---

## Критерии завершения фазы

- [ ] Unit tests: все новые usecase + RFM service покрыты (запуск без Docker)
- [ ] Integration tests: все новые endpoints проходят с реальной БД
- [ ] Аналитика: все виджеты и графики из дизайн-документа работают с реальными данными
- [ ] RFM: автоматический пересчёт по шедулеру + визуализация в UI
- [ ] Кастомные сегменты: конструктор фильтров + предпросмотр кол-ва клиентов
- [ ] iiko интеграция: полный цикл синхронизации заказов → клиенты в базе (mock в тестах)
- [ ] Конструктор акций: полный 3-шаговый процесс сохраняется и применяется
- [ ] Промокоды: создание в UI + применение через бота командой /promo
- [ ] Scheduler: `controller/scheduler/` запускается из cmd/server, задачи регистрируются
