# Phase 2: Core Features — Имплементационный документ

## 1. Обзор и цели

Phase 2 превращает Revisitr из инструмента управления лояльностью в полноценную маркетинговую платформу для HoReCa.
Четыре подфазы поставляют ключевую ценность для владельцев заведений:

| Подфаза | Цель | Бизнес-ценность |
|---------|------|-----------------|
| **2A** Integrations v1 | Импорт агрегатных данных из POS-систем (iiko) | Реальные данные о продажах, сравнение участников лояльности с остальными |
| **2B** Campaigns / Mailings | Целевые рассылки клиентам через Telegram | Прямой маркетинговый канал, отслеживание эффективности |
| **2C** Promotions & Promo Codes | Промо-коды как UTM-аналоги для каналов | Измеримый маркетинг, атрибуция каналов привлечения |
| **2D** Auto-Actions Engine | Автоматические действия по событиям/датам | Удержание клиентов без ручного труда |

---

## 2. Зависимости от Phase 1

Phase 2 полностью опирается на фундамент Phase 1. Перед началом работы над любой подфазой необходимо убедиться, что соответствующие компоненты Phase 1 завершены и протестированы.

### Требуемые компоненты Phase 1

| Компонент Phase 1 | Используется в | Статус (entity/migration) |
|-------------------|----------------|--------------------------|
| `organizations`, `users`, auth | Все подфазы | `00001_init.sql` — готово |
| `bots`, `bot_clients` | 2B, 2C, 2D | `00003_bots.sql`, `00006_extend_bot_clients.sql` — готово |
| `loyalty_programs`, `loyalty_levels`, `client_loyalty`, `loyalty_transactions` | 2A, 2D | `00004_loyalty.sql` — готово |
| `pos_terminals` | 2A | `00005_pos.sql` — готово |
| `integrations`, `external_orders` | 2A | `00011_pos_integrations.sql` — готово |
| `campaigns`, `campaign_messages`, `auto_scenarios` | 2B, 2D | `00007_campaigns.sql` — готово |
| `segments` | 2B, 2C | `00009_analytics_segments.sql` — готово |
| `promotions`, `promo_codes`, `promotion_usages` | 2C | `00010_promotions.sql` — готово |
| `phone_normalized` (bot_clients) | 2A (client matching) | `00006_extend_bot_clients.sql` — готово |
| MinIO / file storage | 2B (media upload) | Инфраструктура Phase 1 |
| Service `pos/` (POSProvider interface) | 2A | `service/pos/provider.go` — готово |
| Service `campaign/sender.go` | 2B | `service/campaign/sender.go` — готово |
| Service `campaign/scheduler.go` | 2D | `service/campaign/scheduler.go` — готово |
| `controller/scheduler/` | 2A, 2D | `scheduler.go` — готово |

### Что нужно доработать из Phase 1 перед Phase 2

1. **Scheduler — регистрация задач синхронизации** (2A): scheduler существует, но задача периодической синхронизации интеграций не зарегистрирована.
2. **Campaign Sender — поддержка медиа** (2B): текущий `sender.go` отправляет только текст, нужна поддержка фото/видео/GIF.
3. **Campaign Sender — кнопки и UTM-ссылки** (2B): текущий отправщик не поддерживает inline-кнопки для трекинга.
4. **Auto-scenario scheduler — расширение до auto-actions** (2D): текущий `campaign/scheduler.go` поддерживает только `birthday` и `inactive_days`, нужно расширить до полноценного Action Engine.

---

## 3. Подфаза 2A: Integrations v1 (iiko aggregate data)

### 3.1 Цель

Импортировать агрегатные бизнес-данные из POS-систем (начиная с iiko Cloud API): выручку, средний чек, количество транзакций. Сопоставлять клиентов iiko с клиентами Revisitr по номеру телефона. Показывать реальные данные на дашборде рядом с данными лояльности.

### 3.2 Задачи

#### 3.2.1 Расширение таблицы external_orders для агрегатных данных

Текущая таблица `external_orders` хранит отдельные заказы. Для агрегатов нужна таблица `integration_aggregates`.

**Критерии приёмки:**
- Таблица `integration_aggregates` создана и содержит суммарные данные по дням
- Индексы обеспечивают быструю фильтрацию по `integration_id` + `date`
- Уникальный ключ предотвращает дублирование при повторной синхронизации

#### 3.2.2 Расширение POS-коннектора (iiko)

Текущий `service/pos/iiko.go` реализует `POSProvider`. Нужно расширить для получения агрегатных данных.

**Критерии приёмки:**
- `iiko.go` получает дневные агрегаты (revenue, avg_check, tx_count) через iiko Cloud API
- Данные корректно маппятся в `IntegrationAggregate`
- Обработка ошибок API (rate limit, timeout, auth expiry)
- Логирование через `slog`

#### 3.2.3 Планировщик синхронизации

Регистрация задачи периодической синхронизации в существующем `controller/scheduler/`.

**Критерии приёмки:**
- Задача `sync_integrations` зарегистрирована в scheduler с настраиваемым интервалом (по умолчанию 60 мин)
- Синхронизация только для интеграций со статусом `active`
- При ошибке — статус интеграции обновляется на `error`, следующая попытка по расписанию
- `last_sync_at` обновляется после каждой успешной синхронизации

#### 3.2.4 Client matching по phone_normalized

Сопоставление клиентов iiko с `bot_clients` по нормализованному номеру телефона.

**Критерии приёмки:**
- Поле `client_id` в `external_orders` заполняется при совпадении `phone_normalized`
- Формат нормализации: `+7XXXXXXXXXX` (единый для iiko и bot_clients)
- Неcопоставленные заказы остаются с `client_id = NULL`
- Повторная синхронизация обновляет matching для новых клиентов

#### 3.2.5 Dashboard — реальные данные

Расширение дашборда для отображения данных из интеграций.

**Критерии приёмки:**
- Виджет «Продажи за период» с данными из `integration_aggregates`
- Виджет «Средний чек» с динамикой
- Сравнение: «Участники лояльности vs остальные» (avg_check, frequency)
- Фильтрация по интеграции и периоду

### 3.3 Схема БД

```sql
-- Миграция: 20260401120000_integration_aggregates.sql

-- +goose Up
CREATE TABLE integration_aggregates (
    id             SERIAL PRIMARY KEY,
    integration_id INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    date           DATE NOT NULL,
    revenue        NUMERIC(14,2) NOT NULL DEFAULT 0,
    avg_check      NUMERIC(10,2) NOT NULL DEFAULT 0,
    tx_count       INT NOT NULL DEFAULT 0,
    guest_count    INT NOT NULL DEFAULT 0,
    matched_count  INT NOT NULL DEFAULT 0,  -- сколько из tx_count сопоставлено с нашими клиентами
    synced_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, date)
);
CREATE INDEX idx_integration_aggregates_date ON integration_aggregates(integration_id, date);

-- Таблица для хранения маппинга внешних клиентов на наших
CREATE TABLE integration_client_map (
    id              SERIAL PRIMARY KEY,
    integration_id  INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_phone  VARCHAR(20) NOT NULL,
    client_id       INT REFERENCES bot_clients(id) ON DELETE SET NULL,
    matched_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, external_phone)
);
CREATE INDEX idx_integration_client_map_client ON integration_client_map(client_id);

-- +goose Down
DROP TABLE IF EXISTS integration_client_map;
DROP TABLE IF EXISTS integration_aggregates;
```

### 3.4 Backend: файлы и интерфейсы

#### Entity

```
backend/internal/entity/integration.go  (дополнить)
```

Добавить структуры:

```go
type IntegrationAggregate struct {
    ID            int       `json:"id"             db:"id"`
    IntegrationID int       `json:"integration_id" db:"integration_id"`
    Date          time.Time `json:"date"           db:"date"`
    Revenue       float64   `json:"revenue"        db:"revenue"`
    AvgCheck      float64   `json:"avg_check"      db:"avg_check"`
    TxCount       int       `json:"tx_count"       db:"tx_count"`
    GuestCount    int       `json:"guest_count"    db:"guest_count"`
    MatchedCount  int       `json:"matched_count"  db:"matched_count"`
    SyncedAt      time.Time `json:"synced_at"      db:"synced_at"`
}

type IntegrationClientMap struct {
    ID            int        `json:"id"              db:"id"`
    IntegrationID int        `json:"integration_id"  db:"integration_id"`
    ExternalPhone string     `json:"external_phone"  db:"external_phone"`
    ClientID      *int       `json:"client_id"       db:"client_id"`
    MatchedAt     *time.Time `json:"matched_at"      db:"matched_at"`
    CreatedAt     time.Time  `json:"created_at"      db:"created_at"`
}

type DashboardAggregates struct {
    Revenue       float64 `json:"revenue"        db:"revenue"`
    AvgCheck      float64 `json:"avg_check"      db:"avg_check"`
    TxCount       int     `json:"tx_count"       db:"tx_count"`
    LoyaltyAvg    float64 `json:"loyalty_avg"    db:"loyalty_avg"`
    NonLoyaltyAvg float64 `json:"non_loyalty_avg" db:"non_loyalty_avg"`
}
```

#### Repository

```
backend/internal/repository/postgres/integrations.go  (дополнить)
```

Добавить методы в интерфейс `integrationsRepo`:

```go
UpsertAggregate(ctx context.Context, agg *entity.IntegrationAggregate) error
GetAggregates(ctx context.Context, integrationID int, from, to time.Time) ([]entity.IntegrationAggregate, error)
GetDashboardAggregates(ctx context.Context, orgID int, from, to time.Time) (*entity.DashboardAggregates, error)
UpsertClientMap(ctx context.Context, mapping *entity.IntegrationClientMap) error
MatchClients(ctx context.Context, integrationID int) (int, error)  // returns matched count
```

#### Service

```
backend/internal/service/pos/iiko.go  (дополнить)
```

Расширить `IikoProvider`:

```go
GetDailyAggregates(ctx context.Context, from, to time.Time) ([]IntegrationAggregate, error)
```

```
backend/internal/service/pos/sync.go  (дополнить)
```

Расширить `SyncService.Sync()`:
1. Запросить агрегаты за период с последней синхронизации
2. Сопоставить клиентов через `integration_client_map`
3. Сохранить агрегаты через `UpsertAggregate`
4. Обновить `last_sync_at`

#### Usecase

```
backend/internal/usecase/integrations/integrations.go  (дополнить)
```

Добавить методы:

```go
func (uc *Usecase) GetAggregates(ctx context.Context, id, orgID int, from, to time.Time) ([]entity.IntegrationAggregate, error)
func (uc *Usecase) GetDashboardData(ctx context.Context, orgID int, from, to time.Time) (*entity.DashboardAggregates, error)
```

#### Controller

```
backend/internal/controller/http/group/integrations/integrations.go  (дополнить)
```

Новые эндпоинты:

```
GET /api/v1/integrations/:id/aggregates?from=&to=
GET /api/v1/dashboard/sales?from=&to=
```

#### Scheduler

```
backend/cmd/server/main.go  (дополнить регистрацию задачи)
```

```go
sched.Register(scheduler.Task{
    Name:     "sync_integrations",
    Interval: 60 * time.Minute,
    Fn:       syncService.SyncAll,
})
```

### 3.5 Frontend

```
frontend/src/features/integrations/types.ts     (дополнить IntegrationAggregate)
frontend/src/features/integrations/api.ts       (дополнить getAggregates)
frontend/src/features/integrations/queries.ts   (дополнить useAggregates)
frontend/src/features/dashboard/types.ts        (дополнить DashboardAggregates)
frontend/src/features/dashboard/api.ts          (дополнить getSalesData)
frontend/src/routes/dashboard/index.tsx          (дополнить виджетами продаж)
frontend/src/routes/dashboard/integrations/$integrationId.tsx  (добавить вкладку Данные)
```

Компоненты:
- `SalesWidget` — карточка с revenue, avg_check, tx_count
- `LoyaltyComparisonChart` — bar chart «участники vs остальные»
- `AggregatesTable` — таблица дневных агрегатов на странице интеграции

### 3.6 Тестирование

**Unit-тесты:**
- `backend/internal/usecase/integrations/integrations_test.go` — дополнить тестами для `GetAggregates`, `GetDashboardData`
- `backend/internal/service/pos/iiko_test.go` — тест маппинга ответа iiko API в `IntegrationAggregate`
- `backend/internal/service/pos/sync_test.go` — тест client matching логики

**Integration-тесты:**
- `backend/tests/integration/integrations_test.go` — CRUD агрегатов, client matching через БД

---

## 4. Подфаза 2B: Campaigns / Mailings System

### 4.1 Цель

Превратить существующий скелет рассылок (MVP из Phase 1) в полноценную систему таргетированных рассылок через Telegram с поддержкой медиа, расписания, отслеживания кликов и аналитики.

### 4.2 Задачи

#### 4.2.1 Redis-очередь сообщений с интерфейсом

Текущая реализация отправляет сообщения синхронно в `sender.go`. Нужна очередь через Redis с интерфейсом для будущей миграции на Kafka.

**Критерии приёмки:**
- Интерфейс `MessageQueue` определён и реализован через Redis (RPUSH/BLPOP)
- Worker-пул забирает сообщения из очереди и отправляет через telego
- Rate limiting: не более 30 msg/sec на один бот-токен
- При падении worker — сообщения остаются в очереди (at-least-once delivery)
- Интерфейс позволяет подключить Kafka без изменений в usecase

#### 4.2.2 Расширение Campaign Sender — медиа и кнопки

**Критерии приёмки:**
- Поддержка типов медиа: фото (JPEG/PNG), видео (MP4), GIF, документ
- `media_url` указывает на файл в MinIO, sender скачивает и отправляет через telego
- Inline-кнопки: каждому сообщению можно добавить до 3 кнопок с URL
- UTM-параметры автоматически добавляются к URL в кнопках

#### 4.2.3 Расширение audience selection

**Критерии приёмки:**
- Выбор аудитории: по сегменту (Segment), по уровню лояльности, по боту, ручной список client_ids
- Предпросмотр аудитории: показать количество получателей до отправки
- `AudienceFilter` расширен полями `segment_id`, `level_id`, `client_ids`

#### 4.2.4 Scheduled sending

**Критерии приёмки:**
- Рассылка со статусом `scheduled` + заполненным `scheduled_at` запускается автоматически
- Scheduler проверяет scheduled-рассылки каждую минуту
- При наступлении времени — статус `scheduled` → `sending`, запуск отправки
- Отмена запланированной рассылки: возврат в `draft`

#### 4.2.5 Tracking — inline buttons + UTM links

**Критерии приёмки:**
- Два режима трекинга (переключаемые и комбинируемые):
  - **Inline buttons**: кнопка с callback_data содержит `campaign_id:client_id`, при нажатии — запись клика
  - **UTM links**: URL с параметрами `utm_source=revisitr&utm_medium=telegram&utm_campaign={campaign_id}`
- Трекинг кликов: таблица `campaign_clicks`
- Подсчёт: sent / delivered / clicked

#### 4.2.6 Campaign analytics

**Критерии приёмки:**
- Статистика по рассылке: total, sent, failed, clicked, click_rate
- Агрегированная статистика по всем рассылкам организации за период
- API-эндпоинт `GET /api/v1/campaigns/:id/analytics`
- UI: страница рассылки с графиками доставки и кликов

#### 4.2.7 Campaign statuses — полный цикл

Текущие статусы: `draft`, `scheduled`, `sending`, `sent`, `failed`.
Добавить: `completed` (sent + analytics собрана).

**Критерии приёмки:**
- Переходы статусов: `draft` → `scheduled` → `sending` → `sent` → `completed`
- `completed` ставится автоматически через 24 часа после `sent` (финализация аналитики)
- `failed` — если 100% сообщений не доставлены

### 4.3 Схема БД

```sql
-- Миграция: 20260410120000_campaigns_v2.sql

-- +goose Up

-- Расширить audience_filter для новых типов выбора аудитории
-- (audience_filter уже JSONB, расширение на уровне entity)

-- Добавить поддержку inline-кнопок
ALTER TABLE campaigns
    ADD COLUMN buttons JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN tracking_mode VARCHAR(20) DEFAULT 'utm' CHECK (tracking_mode IN ('utm', 'buttons', 'both', 'none'));

-- Добавить completed к допустимым статусам
ALTER TABLE campaigns
    DROP CONSTRAINT IF EXISTS campaigns_status_check;
ALTER TABLE campaigns
    ADD CONSTRAINT campaigns_status_check
    CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'completed', 'failed'));

-- Трекинг кликов
CREATE TABLE campaign_clicks (
    id          SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    client_id   INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    button_idx  INT,  -- индекс нажатой кнопки (NULL для UTM)
    url         TEXT,
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_campaign_clicks_campaign ON campaign_clicks(campaign_id);
CREATE INDEX idx_campaign_clicks_client   ON campaign_clicks(client_id);

-- Очередь сообщений в Redis не требует SQL, но добавим таблицу статуса очереди для мониторинга
CREATE TABLE campaign_queue_status (
    id          SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    queued_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    worker_id   VARCHAR(50),
    status      VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'done', 'failed'))
);

-- +goose Down
DROP TABLE IF EXISTS campaign_queue_status;
DROP TABLE IF EXISTS campaign_clicks;
ALTER TABLE campaigns DROP COLUMN IF EXISTS buttons;
ALTER TABLE campaigns DROP COLUMN IF EXISTS tracking_mode;
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_status_check;
ALTER TABLE campaigns ADD CONSTRAINT campaigns_status_check
    CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed'));
```

### 4.4 Backend: файлы и интерфейсы

#### Entity

```
backend/internal/entity/campaign.go  (дополнить)
```

```go
type CampaignButton struct {
    Text string `json:"text"`
    URL  string `json:"url"`
}

type CampaignButtons []CampaignButton

// Scan/Value для JSONB — по установленному паттерну проекта

// Расширить AudienceFilter:
type AudienceFilter struct {
    BotID     *int     `json:"bot_id,omitempty"`
    Tags      []string `json:"tags,omitempty"`
    SegmentID *int     `json:"segment_id,omitempty"`     // NEW
    LevelID   *int     `json:"level_id,omitempty"`       // NEW
    ClientIDs []int    `json:"client_ids,omitempty"`     // NEW (ручной список)
}

// Расширить Campaign:
// + Buttons       CampaignButtons `db:"buttons" json:"buttons"`
// + TrackingMode  string          `db:"tracking_mode" json:"tracking_mode"`

type CampaignClick struct {
    ID         int       `db:"id" json:"id"`
    CampaignID int       `db:"campaign_id" json:"campaign_id"`
    ClientID   int       `db:"client_id" json:"client_id"`
    ButtonIdx  *int      `db:"button_idx" json:"button_idx,omitempty"`
    URL        *string   `db:"url" json:"url,omitempty"`
    ClickedAt  time.Time `db:"clicked_at" json:"clicked_at"`
}

type CampaignAnalyticsDetail struct {
    Total     int     `json:"total"`
    Sent      int     `json:"sent"`
    Failed    int     `json:"failed"`
    Clicked   int     `json:"clicked"`
    ClickRate float64 `json:"click_rate"`
}
```

#### Repository — Redis queue

```
backend/internal/repository/redis/campaign_queue.go  (новый файл)
```

```go
// MessageQueue — абстракция очереди сообщений.
// Реализация через Redis, интерфейс позволяет подключить Kafka.
type MessageQueue interface {
    Enqueue(ctx context.Context, msg *QueueMessage) error
    Dequeue(ctx context.Context, timeout time.Duration) (*QueueMessage, error)
    Ack(ctx context.Context, msgID string) error
    Len(ctx context.Context) (int64, error)
}

type QueueMessage struct {
    ID         string `json:"id"`
    CampaignID int    `json:"campaign_id"`
    ClientID   int    `json:"client_id"`
    TelegramID int64  `json:"telegram_id"`
    Text       string `json:"text"`
    MediaURL   string `json:"media_url,omitempty"`
    MediaType  string `json:"media_type,omitempty"`  // "photo","video","gif","document"
    Buttons    []CampaignButton `json:"buttons,omitempty"`
}
```

Redis-реализация использует:
- `RPUSH revisitr:campaign:{campaign_id}:queue` для enqueue
- `BLPOP` для dequeue с таймаутом
- `DEL revisitr:campaign:{campaign_id}:processing:{msg_id}` для ack

#### Repository — Postgres

```
backend/internal/repository/postgres/campaigns.go  (дополнить)
```

Новые методы:

```go
CreateClick(ctx context.Context, click *entity.CampaignClick) error
GetClicksByCampaign(ctx context.Context, campaignID int) ([]entity.CampaignClick, error)
GetCampaignAnalytics(ctx context.Context, campaignID int) (*entity.CampaignAnalyticsDetail, error)
GetScheduledCampaigns(ctx context.Context, before time.Time) ([]entity.Campaign, error)
```

#### Service

```
backend/internal/service/campaign/sender.go  (переработать)
```

Изменения:
- Вместо прямой отправки — постановка в Redis-очередь
- Worker-пул (configurable, default 3 workers)
- Поддержка `telego.SendPhoto`, `telego.SendVideo`, `telego.SendAnimation`, `telego.SendDocument`
- Добавление inline keyboard с tracking callback_data
- UTM-параметры в URL кнопок

```
backend/internal/service/campaign/worker.go  (новый файл)
```

```go
type Worker struct {
    id       string
    queue    MessageQueue
    bot      *telego.Bot
    campaigns campaignsRepo
    logger   *slog.Logger
}

func (w *Worker) Run(ctx context.Context) {
    for {
        msg, err := w.queue.Dequeue(ctx, 5*time.Second)
        // ... send via telego, ack, update status
        time.Sleep(35 * time.Millisecond) // rate limit
    }
}
```

```
backend/internal/service/campaign/scheduler.go  (дополнить)
```

Добавить в `evaluate()`:
```go
func (s *Scheduler) checkScheduledCampaigns(ctx context.Context) {
    campaigns, _ := s.campaigns.GetScheduledCampaigns(ctx, time.Now())
    for _, c := range campaigns {
        s.sender.SendCampaign(ctx, c.ID)
    }
}
```

#### Usecase

```
backend/internal/usecase/campaigns/campaigns.go  (дополнить)
```

Новые/изменённые методы:

```go
func (uc *Usecase) Send(ctx context.Context, orgID, id int) error         // переработать: через очередь
func (uc *Usecase) Schedule(ctx context.Context, orgID, id int, at time.Time) error
func (uc *Usecase) CancelScheduled(ctx context.Context, orgID, id int) error
func (uc *Usecase) RecordClick(ctx context.Context, campaignID, clientID int, buttonIdx *int, url *string) error
func (uc *Usecase) GetAnalytics(ctx context.Context, orgID, id int) (*entity.CampaignAnalyticsDetail, error)
```

#### Controller

```
backend/internal/controller/http/group/campaigns/campaigns.go  (дополнить)
```

Новые эндпоинты:

```
POST   /api/v1/campaigns/:id/schedule     { "scheduled_at": "..." }
DELETE /api/v1/campaigns/:id/schedule      (отмена)
GET    /api/v1/campaigns/:id/analytics
POST   /api/v1/campaigns/:id/click         (callback от бота)
```

### 4.5 Frontend

```
frontend/src/features/campaigns/types.ts      (дополнить CampaignButton, CampaignClick, analytics)
frontend/src/features/campaigns/api.ts        (дополнить schedule, cancel, analytics, click)
frontend/src/features/campaigns/queries.ts    (дополнить hooks)
frontend/src/routes/dashboard/campaigns/create.tsx  (переработать: добавить media upload, buttons, audience)
frontend/src/routes/dashboard/campaigns/$campaignId.tsx  (дополнить аналитикой)
```

Компоненты:
- `MediaUploader` — drag-and-drop загрузка фото/видео/GIF в MinIO
- `ButtonEditor` — добавление до 3 inline-кнопок с URL
- `AudienceSelector` — выбор аудитории: по сегменту, уровню, боту, ручной список
- `AudiencePreview` — показ количества получателей
- `CampaignScheduler` — выбор даты/времени отправки
- `CampaignAnalyticsPanel` — sent/delivered/clicked с графиком

### 4.6 Тестирование

**Unit-тесты:**
- `backend/internal/usecase/campaigns/campaigns_test.go` — дополнить тестами Schedule, CancelScheduled, RecordClick, GetAnalytics
- `backend/internal/service/campaign/sender_test.go` — тест постановки в очередь, media type detection
- `backend/internal/service/campaign/worker_test.go` — тест dequeue + send + ack
- `backend/internal/repository/redis/campaign_queue_test.go` — тест Enqueue/Dequeue/Ack

**Integration-тесты:**
- `backend/tests/integration/campaigns_test.go` — полный цикл: create → schedule → send → click → analytics

---

## 5. Подфаза 2C: Promotions & Promo Codes

### 5.1 Цель

Расширить существующий скелет промо-акций (Phase 1) до полноценной системы промо-кодов с атрибуцией маркетинговых каналов. Промо-коды выступают UTM-аналогами для офлайн и онлайн каналов.

### 5.2 Задачи

#### 5.2.1 Расширение PromoCode — channel tracking

**Критерии приёмки:**
- Поле `channel` (varchar) для указания источника: `smm`, `targeting`, `yandex_maps`, `flyer`, `partner`, `custom`
- Генерация кодов: автоматическая (6 символов, alphanum uppercase) и пользовательская (например, `BIRTHDAY2024`)
- Уникальность кода в рамках организации (уже есть UNIQUE constraint)
- Валидация пользовательских кодов: 3-20 символов, только A-Z0-9

#### 5.2.2 Per-user limit

**Критерии приёмки:**
- Поле `per_user_limit` — сколько раз один клиент может использовать промо-код
- Проверка при активации: `COUNT(*) FROM promotion_usages WHERE promo_code_id = ? AND client_id = ?`
- По умолчанию — 1 (одноразовый для каждого клиента)

#### 5.2.3 Minimum order amount

**Критерии приёмки:**
- Поле `min_amount` в `PromoCodeConditions` (уже существует в entity)
- Проверка при активации промо-кода: сумма заказа >= min_amount
- Ошибка `ErrMinAmountNotMet` если условие не выполнено

#### 5.2.4 Analytics — conversion tracking

**Критерии приёмки:**
- Подсчёт активаций по каждому промо-коду: `usage_count` (уже есть)
- Аналитика по каналам: group by `channel`, count, total_discount
- Конверсия: сколько клиентов пришли через промо-код и сделали повторные покупки
- API-эндпоинт для аналитики промо-кодов

#### 5.2.5 Promotion entity — расширение

**Критерии приёмки:**
- Тип `recurring` (ежедневная/еженедельная акция) в дополнение к `one-time`
- Привязка промо-кодов к акции (уже есть `promotion_id` в promo_codes)
- Список промо-кодов акции на странице акции

### 5.3 Схема БД

```sql
-- Миграция: 20260415120000_promo_codes_v2.sql

-- +goose Up

-- Канал распространения промо-кода
ALTER TABLE promo_codes
    ADD COLUMN channel VARCHAR(50),
    ADD COLUMN per_user_limit INT DEFAULT 1,
    ADD COLUMN description TEXT;

-- Тип акции: one-time или recurring
ALTER TABLE promotions
    ADD COLUMN recurrence VARCHAR(20) DEFAULT 'one_time' CHECK (recurrence IN ('one_time', 'daily', 'weekly', 'monthly'));

-- Аналитика промо-кодов по каналам
CREATE VIEW promo_channel_analytics AS
SELECT
    pc.org_id,
    pc.channel,
    COUNT(DISTINCT pc.id) AS code_count,
    COALESCE(SUM(pc.usage_count), 0) AS total_usages,
    COUNT(DISTINCT pu.client_id) AS unique_clients
FROM promo_codes pc
LEFT JOIN promotion_usages pu ON pu.promo_code_id = pc.id
GROUP BY pc.org_id, pc.channel;

-- +goose Down
DROP VIEW IF EXISTS promo_channel_analytics;
ALTER TABLE promotions DROP COLUMN IF EXISTS recurrence;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS description;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS per_user_limit;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS channel;
```

### 5.4 Backend: файлы и интерфейсы

#### Entity

```
backend/internal/entity/promotion.go  (дополнить)
```

```go
// Расширить PromoCode:
// + Channel      *string `json:"channel"       db:"channel"`
// + PerUserLimit *int    `json:"per_user_limit" db:"per_user_limit"`
// + Description  *string `json:"description"    db:"description"`

// Расширить Promotion:
// + Recurrence string `json:"recurrence" db:"recurrence"` // "one_time","daily","weekly","monthly"

// Расширить CreatePromoCodeRequest:
// + Channel      *string `json:"channel,omitempty"`
// + PerUserLimit *int    `json:"per_user_limit,omitempty"`
// + Description  *string `json:"description,omitempty"`

type PromoChannelAnalytics struct {
    Channel       string `json:"channel"        db:"channel"`
    CodeCount     int    `json:"code_count"     db:"code_count"`
    TotalUsages   int    `json:"total_usages"   db:"total_usages"`
    UniqueClients int    `json:"unique_clients" db:"unique_clients"`
}

type PromoCodeValidation struct {
    Valid   bool   `json:"valid"`
    Reason  string `json:"reason,omitempty"`
    Promo   *PromoResult `json:"promo,omitempty"`
}
```

#### Repository

```
backend/internal/repository/postgres/promotions.go  (дополнить)
```

Новые методы:

```go
GetChannelAnalytics(ctx context.Context, orgID int) ([]entity.PromoChannelAnalytics, error)
CountUsagesByClient(ctx context.Context, promoCodeID, clientID int) (int, error)
GenerateUniqueCode(ctx context.Context, orgID int) (string, error)
GetPromoCodesByPromotion(ctx context.Context, promotionID int) ([]entity.PromoCode, error)
```

#### Usecase

```
backend/internal/usecase/promotions/promotions.go  (дополнить)
```

Новые/изменённые методы:

```go
func (uc *Usecase) ValidatePromoCode(ctx context.Context, orgID int, code string, clientID int, orderAmount float64) (*entity.PromoCodeValidation, error)
func (uc *Usecase) ActivatePromoCode(ctx context.Context, orgID int, code string, clientID int) (*entity.PromoResult, error)
func (uc *Usecase) GetChannelAnalytics(ctx context.Context, orgID int) ([]entity.PromoChannelAnalytics, error)
func (uc *Usecase) GenerateCode(ctx context.Context, orgID int) (string, error)
```

Логика `ActivatePromoCode`:
1. Найти промо-код по `code` + `org_id`
2. Проверить `active`, `starts_at`, `ends_at`
3. Проверить `usage_limit` (глобальный)
4. Проверить `per_user_limit` (для данного клиента)
5. Проверить `conditions.min_amount` (если передана сумма)
6. Создать `promotion_usage`
7. Инкрементировать `usage_count`
8. Вернуть `PromoResult`

#### Controller

```
backend/internal/controller/http/group/promotions/promotions.go  (дополнить)
```

Новые эндпоинты:

```
POST /api/v1/promo-codes/validate    { "code": "...", "client_id": 1, "order_amount": 500 }
POST /api/v1/promo-codes/activate    { "code": "...", "client_id": 1 }
GET  /api/v1/promo-codes/generate    → { "code": "A1B2C3" }
GET  /api/v1/promo-codes/analytics   → channel analytics
GET  /api/v1/promotions/:id/codes    → list of promo codes for promotion
```

### 5.5 Frontend

```
frontend/src/features/promotions/types.ts    (новый файл)
frontend/src/features/promotions/api.ts      (новый файл)
frontend/src/features/promotions/queries.ts  (новый файл)
frontend/src/routes/dashboard/promotions/index.tsx   (переработать)
frontend/src/routes/dashboard/promotions/codes.tsx   (переработать)
```

Компоненты:
- `CreatePromoCodeModal` — форма создания с выбором канала, генерацией кода
- `PromoCodeTable` — таблица промо-кодов с фильтрацией по каналу и статусу
- `ChannelAnalyticsChart` — bar chart по каналам (usages, unique_clients)
- `PromotionCodesTab` — вкладка со списком промо-кодов акции

### 5.6 Тестирование

**Unit-тесты:**
- `backend/internal/usecase/promotions/promotions_test.go` — дополнить: ValidatePromoCode (все сценарии: expired, limit exceeded, min_amount, per_user_limit), ActivatePromoCode, GenerateCode
- Тест генерации уникальных кодов: формат, уникальность

**Integration-тесты:**
- `backend/tests/integration/promotions_test.go` — create promo code → activate → check usage_count → check per_user_limit

---

## 6. Подфаза 2D: Auto-Actions Engine

### 6.1 Цель

Заменить упрощённый `auto_scenarios` из Phase 1 полноценным движком автоматических действий. Авто-действие отличается от авто-рассылки тем, что рассылка — лишь одно из возможных действий. Авто-действие может: начислить бонусы, создать промо-код, отправить сообщение, изменить уровень лояльности.

### 6.2 Ключевое архитектурное решение

Текущая таблица `auto_scenarios` и entity `AutoScenario` остаются, но расширяются:
- `trigger_config` (уже JSONB) — расширяется новыми полями
- Добавляется `actions` (JSONB) — массив действий вместо одного `message`
- `message` остаётся как fallback / shorthand для простого сценария «отправить сообщение»

Гибкая JSONB-структура позволяет описывать сложные авто-действия:

```json
{
  "trigger": "birthday",
  "timing": { "days_before": 7, "days_after": 7 },
  "condition": { "type": "order", "min_amount": 0 },
  "actions": [
    { "type": "bonus", "amount": 500 },
    { "type": "campaign", "template_id": 1 },
    { "type": "promo_code", "discount": 15 }
  ]
}
```

### 6.3 Задачи

#### 6.3.1 Расширение auto_scenarios → auto_actions

**Критерии приёмки:**
- Новые trigger types: `holiday`, `registration`, `level_change` (в дополнение к существующим `birthday`, `inactive_days`, `visit_count`, `bonus_threshold`, `level_up`)
- Поле `actions` (JSONB) для массива действий
- Поле `timing` (JSONB) для настройки временных рамок триггера
- Поле `conditions` (JSONB) для дополнительных условий выполнения
- Обратная совместимость: если `actions` пустой, использовать `message` как action type `campaign`

#### 6.3.2 ActionExecutor service

**Критерии приёмки:**
- Сервис `ActionExecutor` интерпретирует `actions` JSONB и делегирует соответствующим usecase-ам
- Поддерживаемые типы действий:
  - `bonus` — начислить бонусы через `usecase/loyalty`
  - `campaign` — отправить шаблонное сообщение через `service/campaign/sender`
  - `promo_code` — создать персональный промо-код через `usecase/promotions`
  - `level_change` — изменить уровень лояльности через `usecase/loyalty`
- Каждое действие выполняется независимо (ошибка одного не блокирует другие)
- Логирование каждого выполненного действия

#### 6.3.3 Расширение scheduler для date-based triggers

**Критерии приёмки:**
- `birthday`: за N дней до / в день / через N дней после (timing JSONB)
- `holiday`: настраиваемые даты (8 Марта, Новый год, день рождения заведения)
- `inactivity`: N дней без визита (на основе `loyalty_transactions` или `external_orders`)
- Дедупликация: авто-действие не выполняется повторно для того же клиента в том же триггерном периоде

#### 6.3.4 Event-driven triggers

**Критерии приёмки:**
- `registration` — при регистрации нового клиента в боте
- `level_change` — при изменении уровня лояльности
- `visit_count` — при достижении N визитов (event из loyalty transactions)
- Event-driven триггеры вызываются из соответствующих usecase-ов, не через scheduler

#### 6.3.5 Template auto-actions

**Критерии приёмки:**
- Предустановленные шаблоны: «День рождения», «8 Марта», «Новый год», «Неактивность 30 дней»
- Шаблоны создаются через seed migration
- Пользователь может клонировать шаблон и настроить под себя
- Полностью кастомное создание авто-действий

#### 6.3.6 Auto-action execution log

**Критерии приёмки:**
- Таблица `auto_action_log` для аудита выполненных действий
- Запись: scenario_id, client_id, action_type, result, executed_at
- UI: лог выполнения на странице авто-действия

### 6.4 Схема БД

```sql
-- Миграция: 20260420120000_auto_actions_v2.sql

-- +goose Up

-- Расширить auto_scenarios
ALTER TABLE auto_scenarios
    ADD COLUMN actions JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN timing JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN conditions JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN is_template BOOLEAN DEFAULT false,
    ADD COLUMN template_key VARCHAR(50);

-- Расширить допустимые trigger types
ALTER TABLE auto_scenarios
    DROP CONSTRAINT IF EXISTS auto_scenarios_trigger_type_check;
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_trigger_type_check
    CHECK (trigger_type IN (
        'inactive_days', 'visit_count', 'bonus_threshold', 'level_up',
        'birthday', 'holiday', 'registration', 'level_change'
    ));

-- Лог выполнения авто-действий
CREATE TABLE auto_action_log (
    id           SERIAL PRIMARY KEY,
    scenario_id  INT NOT NULL REFERENCES auto_scenarios(id) ON DELETE CASCADE,
    client_id    INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    action_type  VARCHAR(30) NOT NULL,
    action_data  JSONB DEFAULT '{}'::jsonb,
    result       VARCHAR(20) NOT NULL CHECK (result IN ('success', 'failed', 'skipped')),
    error_msg    TEXT,
    executed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_auto_action_log_scenario ON auto_action_log(scenario_id);
CREATE INDEX idx_auto_action_log_client   ON auto_action_log(client_id);
CREATE INDEX idx_auto_action_log_date     ON auto_action_log(executed_at);

-- Дедупликация: предотвращение повторного выполнения
CREATE TABLE auto_action_dedup (
    id           SERIAL PRIMARY KEY,
    scenario_id  INT NOT NULL REFERENCES auto_scenarios(id) ON DELETE CASCADE,
    client_id    INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    trigger_key  VARCHAR(100) NOT NULL,  -- например "birthday:2026" или "inactive:2026-03"
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scenario_id, client_id, trigger_key)
);

-- +goose Down
DROP TABLE IF EXISTS auto_action_dedup;
DROP TABLE IF EXISTS auto_action_log;
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_trigger_type_check;
ALTER TABLE auto_scenarios ADD CONSTRAINT auto_scenarios_trigger_type_check
    CHECK (trigger_type IN ('inactive_days', 'visit_count', 'bonus_threshold', 'level_up', 'birthday'));
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS template_key;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS is_template;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS conditions;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS timing;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS actions;
```

```sql
-- Миграция: 20260420120001_auto_actions_seed_templates.sql

-- +goose Up
INSERT INTO auto_scenarios (org_id, bot_id, name, trigger_type, trigger_config, message, is_template, template_key, actions, timing, is_active)
VALUES
-- Шаблоны (org_id=0, bot_id=0 — глобальные шаблоны)
(0, 0, 'День рождения клиента', 'birthday',
 '{}', '{name}, с днём рождения! 🎂',
 true, 'tpl_birthday',
 '[{"type":"bonus","amount":500},{"type":"campaign","template":"birthday"}]',
 '{"days_before":0,"days_after":0}',
 false),

(0, 0, '8 Марта', 'holiday',
 '{}', '{name}, поздравляем с 8 Марта! 💐',
 true, 'tpl_8march',
 '[{"type":"bonus","amount":300},{"type":"campaign","template":"holiday"}]',
 '{"month":3,"day":8}',
 false),

(0, 0, 'Новый год', 'holiday',
 '{}', '{name}, с Новым годом! 🎄',
 true, 'tpl_newyear',
 '[{"type":"promo_code","discount":10},{"type":"campaign","template":"newyear"}]',
 '{"month":12,"day":31}',
 false),

(0, 0, 'Неактивность 30 дней', 'inactive_days',
 '{"days":30}', '{name}, мы скучаем! Вот вам подарок 🎁',
 true, 'tpl_inactive30',
 '[{"type":"bonus","amount":200},{"type":"campaign","template":"comeback"}]',
 '{}',
 false);

-- +goose Down
DELETE FROM auto_scenarios WHERE is_template = true AND template_key IN ('tpl_birthday', 'tpl_8march', 'tpl_newyear', 'tpl_inactive30');
```

### 6.5 Backend: файлы и интерфейсы

#### Entity

```
backend/internal/entity/auto_scenario.go  (переработать)
```

```go
// Расширить TriggerConfig:
type TriggerConfig struct {
    Days      *int `json:"days,omitempty"`       // inactive_days
    Count     *int `json:"count,omitempty"`      // visit_count
    Threshold *int `json:"threshold,omitempty"`  // bonus_threshold
    Month     *int `json:"month,omitempty"`      // holiday
    Day       *int `json:"day,omitempty"`         // holiday
}

type ActionTiming struct {
    DaysBefore *int `json:"days_before,omitempty"`
    DaysAfter  *int `json:"days_after,omitempty"`
    Month      *int `json:"month,omitempty"`
    Day        *int `json:"day,omitempty"`
}
// Scan/Value по установленному паттерну

type ActionCondition struct {
    Type      string   `json:"type,omitempty"`       // "order", "level", "segment"
    MinAmount *float64 `json:"min_amount,omitempty"`
    LevelID   *int     `json:"level_id,omitempty"`
    SegmentID *int     `json:"segment_id,omitempty"`
}
// Scan/Value по установленному паттерну

type ActionDef struct {
    Type       string  `json:"type"`                   // "bonus","campaign","promo_code","level_change"
    Amount     *int    `json:"amount,omitempty"`        // for bonus
    TemplateID *int    `json:"template_id,omitempty"`   // for campaign
    Template   *string `json:"template,omitempty"`      // template key for campaign
    Discount   *int    `json:"discount,omitempty"`      // for promo_code
    LevelID    *int    `json:"level_id,omitempty"`      // for level_change
}

type ActionDefs []ActionDef
// Scan/Value по установленному паттерну

// Расширить AutoScenario:
// + Actions     ActionDefs     `db:"actions" json:"actions"`
// + Timing      ActionTiming   `db:"timing" json:"timing"`
// + Conditions  ActionCondition `db:"conditions" json:"conditions"`
// + IsTemplate  bool           `db:"is_template" json:"is_template"`
// + TemplateKey *string        `db:"template_key" json:"template_key,omitempty"`

type AutoActionLog struct {
    ID         int       `db:"id" json:"id"`
    ScenarioID int       `db:"scenario_id" json:"scenario_id"`
    ClientID   int       `db:"client_id" json:"client_id"`
    ActionType string    `db:"action_type" json:"action_type"`
    ActionData json.RawMessage `db:"action_data" json:"action_data"`
    Result     string    `db:"result" json:"result"`
    ErrorMsg   *string   `db:"error_msg" json:"error_msg,omitempty"`
    ExecutedAt time.Time `db:"executed_at" json:"executed_at"`
}
```

#### Service — ActionExecutor

```
backend/internal/service/autoaction/executor.go  (новый файл)
```

```go
type ActionExecutor struct {
    loyalty    loyaltyUsecase
    campaigns  campaignSender
    promotions promotionsUsecase
    scenarios  scenariosRepo
    logger     *slog.Logger
}

// Execute выполняет массив действий для клиента.
func (e *ActionExecutor) Execute(ctx context.Context, scenario entity.AutoScenario, client entity.BotClient) error {
    for _, action := range scenario.Actions {
        result := e.executeAction(ctx, action, scenario, client)
        e.logAction(ctx, scenario.ID, client.ID, action, result)
    }
    return nil
}

func (e *ActionExecutor) executeAction(ctx context.Context, action entity.ActionDef, scenario entity.AutoScenario, client entity.BotClient) actionResult {
    switch action.Type {
    case "bonus":
        return e.executeBonus(ctx, action, client)
    case "campaign":
        return e.executeCampaign(ctx, action, scenario, client)
    case "promo_code":
        return e.executePromoCode(ctx, action, client)
    case "level_change":
        return e.executeLevelChange(ctx, action, client)
    default:
        return actionResult{Result: "skipped", Error: "unknown action type"}
    }
}
```

```
backend/internal/service/autoaction/scheduler.go  (новый файл)
```

Заменяет / расширяет текущий `campaign/scheduler.go`:

```go
type AutoActionScheduler struct {
    scenarios scenariosRepo
    executor  *ActionExecutor
    dedup     dedupRepo
    bots      botsRepo
    clients   clientsRepo
    logger    *slog.Logger
}

func (s *AutoActionScheduler) Evaluate(ctx context.Context) error {
    // 1. Получить все активные сценарии
    // 2. Для date-based triggers: проверить timing
    // 3. Для каждого клиента, подходящего под trigger:
    //    a. Проверить dedup (trigger_key уникален для периода)
    //    b. Проверить conditions
    //    c. Выполнить actions через ActionExecutor
    //    d. Записать dedup
}
```

#### Repository

```
backend/internal/repository/postgres/auto_scenarios.go  (дополнить)
```

Новые методы:

```go
GetTemplates(ctx context.Context) ([]entity.AutoScenario, error)
GetActiveDateBased(ctx context.Context) ([]entity.AutoScenario, error)
CreateActionLog(ctx context.Context, log *entity.AutoActionLog) error
GetActionLog(ctx context.Context, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error)
CheckDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) (bool, error)
CreateDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) error
```

#### Usecase

```
backend/internal/usecase/campaigns/campaigns.go  (дополнить)
```

Новые методы для авто-действий:

```go
func (uc *Usecase) CloneTemplate(ctx context.Context, orgID int, templateKey string, botID int) (*entity.AutoScenario, error)
func (uc *Usecase) GetTemplates(ctx context.Context) ([]entity.AutoScenario, error)
func (uc *Usecase) GetActionLog(ctx context.Context, orgID, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error)
```

#### Event-driven hooks

Для event-driven триггеров (`registration`, `level_change`, `visit_count`) нужно вызывать `ActionExecutor` из соответствующих usecase-ов:

```
backend/internal/usecase/loyalty/loyalty.go:
  - При начислении бонусов: проверить visit_count и bonus_threshold triggers
  - При смене уровня: вызвать level_change trigger

backend/internal/service/botmanager/handler.go:
  - При /start (регистрации): вызвать registration trigger
```

Паттерн вызова — через интерфейс `AutoActionHook`:

```go
type AutoActionHook interface {
    OnEvent(ctx context.Context, eventType string, clientID int, data map[string]interface{}) error
}
```

#### Controller

```
backend/internal/controller/http/group/campaigns/campaigns.go  (дополнить)
```

Новые эндпоинты:

```
GET  /api/v1/auto-actions/templates                      → список шаблонов
POST /api/v1/auto-actions/templates/:key/clone            → клонировать шаблон
GET  /api/v1/auto-actions/:id/log?limit=20&offset=0      → лог выполнения
```

### 6.6 Frontend

```
frontend/src/features/campaigns/types.ts      (дополнить ActionDef, AutoActionLog, ActionTiming)
frontend/src/features/campaigns/api.ts        (дополнить templates, cloneTemplate, getActionLog)
frontend/src/features/campaigns/queries.ts    (дополнить hooks)
frontend/src/routes/dashboard/campaigns/scenarios.tsx  (переработать → auto-actions с actions JSONB editor)
```

Компоненты:
- `ActionEditor` — визуальный редактор массива действий (drag-and-drop, add/remove)
- `TriggerSelector` — выбор типа триггера с настройкой timing
- `ConditionEditor` — настройка условий выполнения
- `TemplateGallery` — галерея шаблонов для клонирования
- `ActionLogTable` — таблица лога выполнения авто-действий

### 6.7 Тестирование

**Unit-тесты:**
- `backend/internal/service/autoaction/executor_test.go` — тест каждого типа действия: bonus, campaign, promo_code, level_change
- `backend/internal/service/autoaction/scheduler_test.go` — тест date-based evaluation, dedup, timing
- `backend/internal/usecase/campaigns/campaigns_test.go` — тест CloneTemplate

**Integration-тесты:**
- `backend/tests/integration/auto_actions_test.go` — полный цикл: create scenario → trigger event → check action log → check dedup

---

## 7. Стратегия тестирования (общая)

### 7.1 Unit-тесты (без инфраструктуры)

| Подфаза | Файл | Что тестируется |
|---------|------|-----------------|
| 2A | `usecase/integrations/integrations_test.go` | GetAggregates, GetDashboardData, client matching logic |
| 2A | `service/pos/iiko_test.go` | Маппинг ответа iiko API |
| 2B | `usecase/campaigns/campaigns_test.go` | Schedule, CancelScheduled, RecordClick, GetAnalytics |
| 2B | `service/campaign/sender_test.go` | Enqueue, media type detection |
| 2B | `service/campaign/worker_test.go` | Dequeue + send + ack |
| 2B | `repository/redis/campaign_queue_test.go` | Redis queue operations |
| 2C | `usecase/promotions/promotions_test.go` | ValidatePromoCode, ActivatePromoCode, GenerateCode |
| 2D | `service/autoaction/executor_test.go` | Каждый тип действия |
| 2D | `service/autoaction/scheduler_test.go` | Date-based evaluation, dedup |

Паттерн мок-объектов — ручные моки через struct с function fields (установленный паттерн проекта):

```go
type mockCampaignsRepo struct {
    CreateFn       func(ctx context.Context, campaign *entity.Campaign) error
    GetByIDFn      func(ctx context.Context, id int) (*entity.Campaign, error)
    // ...
}

func (m *mockCampaignsRepo) Create(ctx context.Context, campaign *entity.Campaign) error {
    return m.CreateFn(ctx, campaign)
}
```

### 7.2 Integration-тесты (требуют Docker)

| Подфаза | Файл | Что тестируется |
|---------|------|-----------------|
| 2A | `tests/integration/integrations_test.go` | CRUD aggregates, client matching через БД |
| 2B | `tests/integration/campaigns_test.go` | Полный цикл рассылки, scheduled campaigns |
| 2C | `tests/integration/promotions_test.go` | Promo code activation, per-user limits |
| 2D | `tests/integration/auto_actions_test.go` | Trigger → action → log → dedup |

Build tag: `//go:build integration`

### 7.3 Frontend unit-тесты

| Подфаза | Файл | Что тестируется |
|---------|------|-----------------|
| 2A | `src/features/integrations/*.test.ts` | API hooks, data transformation |
| 2B | `src/features/campaigns/*.test.ts` | Campaign form validation, audience preview |
| 2C | `src/features/promotions/*.test.ts` | Promo code validation, channel analytics display |
| 2D | `src/routes/dashboard/campaigns/scenarios.test.tsx` | Action editor, template gallery |

### 7.4 E2E-тесты (Playwright)

```
e2e/tests/campaigns.spec.ts     — создание рассылки, отправка, просмотр аналитики
e2e/tests/promo-codes.spec.ts   — создание промо-кода, проверка в UI
e2e/tests/auto-actions.spec.ts  — клонирование шаблона, настройка, просмотр лога
e2e/tests/integrations.spec.ts  — подключение интеграции, просмотр данных
```

---

## 8. Definition of Done

### 8.1 Подфаза 2A: Integrations v1

- [ ] Миграция `integration_aggregates` и `integration_client_map` применена
- [ ] iiko коннектор получает дневные агрегаты
- [ ] Scheduler `sync_integrations` зарегистрирован и работает
- [ ] Client matching по `phone_normalized` заполняет `client_id`
- [ ] Dashboard показывает revenue, avg_check, tx_count из интеграций
- [ ] Сравнение «участники лояльности vs остальные» отображается
- [ ] Unit-тесты usecase/integrations покрывают новые методы
- [ ] Integration-тест: sync → aggregates → matching
- [ ] Коммит: `feat(integrations): add aggregate data sync from iiko`

### 8.2 Подфаза 2B: Campaigns / Mailings

- [ ] Redis-очередь сообщений реализована с интерфейсом `MessageQueue`
- [ ] Worker-пул отправляет сообщения через telego с rate limiting
- [ ] Поддержка медиа: фото, видео, GIF, документ
- [ ] Inline-кнопки и UTM-ссылки для трекинга
- [ ] Audience selection: по сегменту, уровню, боту, ручной список
- [ ] Scheduled sending работает через scheduler
- [ ] Campaign analytics: sent / delivered / clicked
- [ ] Статус `completed` выставляется через 24 часа после `sent`
- [ ] Frontend: создание рассылки с медиа, кнопками, аудиторией, расписанием
- [ ] Frontend: страница аналитики рассылки
- [ ] Unit-тесты: sender, worker, queue, usecase
- [ ] Integration-тест: полный цикл рассылки
- [ ] Коммит: `feat(campaigns): add media support, redis queue, tracking and analytics`

### 8.3 Подфаза 2C: Promotions & Promo Codes

- [ ] Поле `channel` добавлено к промо-кодам
- [ ] Per-user limit работает корректно
- [ ] Генерация уникальных кодов (auto и user-defined)
- [ ] Validate/Activate API работают с проверкой всех условий
- [ ] Channel analytics: group by channel, usages, unique_clients
- [ ] Promotions: recurrence type (one_time, daily, weekly, monthly)
- [ ] Frontend: создание промо-кода с выбором канала, аналитика каналов
- [ ] Unit-тесты: все сценарии validation/activation
- [ ] Integration-тест: create → activate → check limits
- [ ] Коммит: `feat(promotions): add channel tracking, per-user limits and analytics`

### 8.4 Подфаза 2D: Auto-Actions Engine

- [ ] Миграция `auto_action_log` и `auto_action_dedup` применена
- [ ] ActionExecutor выполняет все типы действий: bonus, campaign, promo_code, level_change
- [ ] Date-based triggers: birthday (timing), holiday (configurable), inactivity
- [ ] Event-driven triggers: registration, level_change, visit_count
- [ ] Дедупликация предотвращает повторное выполнение
- [ ] Template auto-actions: 4 предустановленных шаблона
- [ ] Клонирование шаблонов работает
- [ ] Лог выполнения записывается и отображается в UI
- [ ] Frontend: action editor, trigger selector, template gallery, action log
- [ ] Unit-тесты: executor (все типы), scheduler (date-based, dedup)
- [ ] Integration-тест: trigger → action → log → dedup
- [ ] Коммит: `feat(auto-actions): add action engine with triggers, templates and execution log`

---

## 9. Граф зависимостей между подфазами

```
Phase 1 (завершена)
    │
    ├──► 2A: Integrations v1
    │        Зависимости: integrations, external_orders, pos/provider, scheduler
    │        Блокирует: 2D (данные для inactivity trigger)
    │
    ├──► 2B: Campaigns / Mailings
    │        Зависимости: campaigns, campaign_messages, bot_clients, MinIO
    │        Блокирует: 2D (campaign как тип action)
    │
    ├──► 2C: Promotions & Promo Codes
    │        Зависимости: promotions, promo_codes, promotion_usages
    │        Блокирует: 2D (promo_code как тип action)
    │
    └──► 2D: Auto-Actions Engine
             Зависимости: 2A (inactivity data), 2B (campaign sender), 2C (promo code creation)
             Зависимости Phase 1: auto_scenarios, loyalty, bot_clients
```

### Рекомендованный порядок реализации

```
  2A ──────┐
  2B ──────┼──► 2D
  2C ──────┘
```

Подфазы **2A**, **2B**, **2C** могут разрабатываться параллельно (независимы друг от друга).
Подфаза **2D** начинается после завершения 2B и 2C (использует campaign sender и promo code creation как типы действий). Подфаза 2A желательна для 2D (данные inactivity из внешних заказов), но не строго обязательна (можно определять inactivity по `loyalty_transactions`).

### Предлагаемые коммиты по этапам

```
feat(integrations): add integration_aggregates and client_map migrations
feat(integrations): implement iiko daily aggregates sync
feat(integrations): add scheduled sync task and client matching
feat(integrations): add dashboard sales widgets with loyalty comparison

feat(campaigns): add redis message queue with MessageQueue interface
feat(campaigns): implement worker pool with media support
feat(campaigns): add inline buttons and UTM tracking
feat(campaigns): implement scheduled sending and campaign completion
feat(campaigns): add campaign click tracking and analytics

feat(promotions): add channel tracking and per-user limits to promo codes
feat(promotions): implement validate and activate with all conditions
feat(promotions): add channel analytics and code generation

feat(auto-actions): add actions JSONB, timing, conditions to auto_scenarios
feat(auto-actions): implement ActionExecutor service
feat(auto-actions): add date-based and event-driven triggers with dedup
feat(auto-actions): add template auto-actions and clone functionality
feat(auto-actions): add execution log and UI
```
