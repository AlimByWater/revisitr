# Phase 3: Advanced Features

Углубление существующих функций и guided UX для Revisitr.

## 1. Обзор и цели

Phase 3 расширяет базовый функционал Phase 1-2, добавляя:

- **3A. Integrations v2** -- привязка POS-транзакций к профилям клиентов, импорт меню, поэлементная аналитика
- **3B. RFM-сегментация** -- автоматическое распределение клиентов по RFM-сегментам с cron-пересчётом
- **3C. Onboarding Wizard** -- пошаговый мастер настройки для новых пользователей (6 шагов)
- **3D. Bot Constructor UI** -- полноценный интерфейс конфигурации бота в админ-панели

Все подфазы независимы друг от друга и могут выполняться параллельно.

---

## 2. Зависимости

### От Phase 1 (Foundation)

| Компонент Phase 1 | Зависят |
|---|---|
| Иерархия сущностей (org -> program -> bot) | 3C, 3D |
| Таблица `organizations` | 3C (onboarding_completed) |
| Таблица `bots`, `BotSettings` JSONB | 3D |
| Таблица `bot_clients` | 3A, 3B |
| Таблица `pos_locations` | 3A, 3C |
| Auth flow (JWT, middleware) | 3C (redirect) |

### От Phase 2 (Core Features)

| Компонент Phase 2 | Зависят |
|---|---|
| Integrations v1 (таблица `integrations`, `external_orders`, sync-сервис) | 3A |
| Данные транзакций в `external_orders` | 3B (источник данных для RFM) |
| Campaigns, Promotions, Analytics | 3C (шаг "Следующие шаги") |
| Segments (таблица `segments`, `segment_clients`) | 3B (расширение) |

### Текущее состояние кода

Уже реализовано и доступно для Phase 3:

- `entity.Integration`, `IntegrationConfig`, `ExternalOrder`, `OrderItems` -- полная модель интеграций
- `entity.Segment`, `SegmentFilter` с поддержкой `rfm_category`
- `entity.BotSettings` с `Modules`, `Buttons`, `RegistrationForm`, `WelcomeMessage`
- `bot_clients` -- RFM-поля: `rfm_recency`, `rfm_frequency`, `rfm_monetary`, `rfm_segment`
- `controller/scheduler` -- фреймворк для cron-задач
- `service/pos` -- интерфейс POS (customers, menu, sync)
- Все CRUD-контроллеры и usecase для integrations, segments, bots

---

## 3. Подфаза 3A: Integrations v2 (Deep Integration)

### 3A.1. Цель

Связать отдельные POS-транзакции с конкретными профилями клиентов (по совпадению телефона), импортировать меню из POS-систем, обеспечить поэлементную аналитику заказов.

### 3A.2. Задачи

#### 3A.2.1. Привязка заказов к клиентам по телефону

**Описание**: при синхронизации заказов из POS автоматически искать `bot_clients` по телефону клиента из POS и проставлять `client_id` в `external_orders`.

**Реализация**:
- В `syncService.Sync()` после получения заказов из POS -- для каждого заказа, содержащего телефон клиента, искать `bot_clients.phone`
- Нормализация телефонов: удаление `+`, пробелов, скобок; приведение к формату `7XXXXXXXXXX`
- Если найден -- проставить `external_orders.client_id`
- Если не найден -- оставить `client_id = NULL`, инкрементировать счётчик `unmatched_orders`
- Периодический пересчёт: при добавлении нового клиента с телефоном -- попытка привязать ранее непривязанные заказы

**Acceptance criteria**:
- При синхронизации заказы автоматически привязываются к клиентам по совпадению телефона
- Непривязанные заказы отображаются отдельно со счётчиком
- При регистрации нового клиента с телефоном ранее непривязанные заказы привязываются автоматически

#### 3A.2.2. r-keeper коннектор

**Описание**: реализовать адаптер для r-keeper, реализующий тот же интерфейс `posService`, что и iiko-коннектор.

**Реализация**:
- Новый пакет `internal/service/pos/rkeeper/`
- Реализация интерфейсов: `Sync`, `TestConnection`, `GetCustomers`, `GetMenu`
- Аутентификация: username/password (уже в `IntegrationConfig`)
- Маппинг r-keeper API на внутренние типы `ExternalOrder`, `POSCustomer`, `POSMenu`

**Acceptance criteria**:
- r-keeper коннектор проходит `TestConnection` с валидными credentials
- Синхронизация заказов работает аналогично iiko
- Меню и клиенты импортируются корректно

#### 3A.2.3. 1C коннектор

**Описание**: аналогичный адаптер для 1С (REST API или OData).

**Реализация**:
- Новый пакет `internal/service/pos/onec/`
- Реализация того же интерфейса `posService`
- Конфигурация: `api_url`, `api_key` из `IntegrationConfig`

**Acceptance criteria**:
- 1C коннектор реализует полный интерфейс `posService`
- Тест подключения, синхронизация заказов, импорт меню работают

#### 3A.2.4. Импорт и авто-обновление меню

**Описание**: сохранять меню из POS локально, обновлять при синхронизации.

**Реализация**:
- Новая таблица `menus` и `menu_items` (см. раздел 4)
- При каждой синхронизации -- сравнение с текущим меню, обновление изменений
- API для получения меню организации: `GET /api/v1/menus`
- Возможность ручного редактирования (переименование позиций, добавление описаний)

**Acceptance criteria**:
- Меню импортируется из POS при синхронизации
- Изменения в POS-меню отражаются после следующей синхронизации
- Ручные правки не перезатираются при авто-обновлении

#### 3A.2.5. Поэлементная аналитика в профиле клиента

**Описание**: показывать, какие конкретно позиции заказывает клиент.

**Реализация**:
- Агрегация `external_orders.items` по `client_id` -- топ позиций, частота, средний чек по позиции
- Новый endpoint: `GET /api/v1/clients/:id/order-stats`
- Вкладка "Заказы" на странице клиента в админке

**Acceptance criteria**:
- На странице клиента видна история заказов из POS
- Отображаются топ-позиции клиента, количество заказов, общая сумма
- Данные обновляются после каждой синхронизации

#### 3A.2.6. Аналитика эффективности эксклюзивных позиций

**Описание**: оценка эффективности позиций, доступных только через бот (коктейль за регистрацию, спецменю, бизнес-ланчи).

**Реализация**:
- Тегирование позиций меню: `bot_exclusive`, `promo`, `lunch`
- Сравнение продаж тегированных позиций среди клиентов бота vs. общий поток
- Дашборд: конверсия эксклюзивных позиций, выручка, ROI акций

**Acceptance criteria**:
- Администратор может пометить позиции меню тегами
- Аналитика показывает продажи тегированных позиций
- Видно сравнение: клиенты бота vs. все клиенты

#### 3A.2.7. История транзакций на странице клиента

**Описание**: расширить существующую страницу `clients/$clientId` вкладкой с POS-транзакциями.

**Реализация**:
- Объединить `loyalty_transactions` и `external_orders` в единую timeline
- Фильтры: по типу (лояльность / POS), по дате, по сумме
- Пагинация

**Acceptance criteria**:
- На странице клиента видны все транзакции: и лояльности, и POS
- Работают фильтры и пагинация
- Каждый POS-заказ раскрывается до позиций

---

## 4. Подфаза 3B: RFM-сегментация

### 3B.1. Цель

Автоматическое распределение клиентов по RFM-сегментам на основе транзакционных данных. 70-80% пользователей используют готовые шаблоны -- UI проектируется под этот сценарий.

### 3B.2. Задачи

#### 3B.2.1. RFM Calculation Engine

**Описание**: движок расчёта RFM-показателей для каждого клиента.

**Алгоритм**:
1. Собрать данные из `external_orders` и `loyalty_transactions` за период (по умолчанию 365 дней)
2. Для каждого клиента вычислить:
   - **Recency (R)**: дней с последней транзакции
   - **Frequency (F)**: количество транзакций за период
   - **Monetary (M)**: общая сумма за период
3. Разбить на квинтили (1-5) по каждой метрике
4. Присвоить сегмент по комбинации R-F-M scores

**Маппинг сегментов**:

| Сегмент | R score | F score | M score | Описание |
|---|---|---|---|---|
| `champions` | 4-5 | 4-5 | 4-5 | Лучшие клиенты |
| `loyal` | 3-5 | 3-5 | 3-5 | Лояльные постоянные |
| `potential_loyalist` | 3-5 | 1-3 | 1-3 | Недавние, но мало покупок |
| `new_customers` | 4-5 | 1 | 1 | Новые клиенты |
| `at_risk` | 1-2 | 3-5 | 3-5 | Были активны, давно не приходили |
| `cant_lose` | 1-2 | 4-5 | 4-5 | Важные клиенты, которых теряем |
| `lost` | 1 | 1-2 | 1-2 | Потерянные |

**Реализация**:
- Новый пакет `internal/usecase/rfm/`
- Метод `Calculate(ctx, orgID, period)` -- полный пересчёт
- Запись результатов в `bot_clients.rfm_*` поля (уже есть в схеме)
- Обновление `segment_clients` для RFM-сегментов

**Acceptance criteria**:
- RFM-показатели корректно вычисляются для каждого клиента
- Сегменты присваиваются по таблице маппинга
- Результаты записываются в `bot_clients` и `segment_clients`

#### 3B.2.2. Cron-задача пересчёта

**Описание**: автоматический пересчёт RFM через `controller/scheduler`.

**Реализация**:
- Регистрация задачи `rfm_recalculation` в scheduler
- Интервал: ежедневно (настраиваемо через env `RFM_RECALC_INTERVAL`, по умолчанию `24h`)
- Логирование: количество обработанных клиентов, изменения сегментов
- Атомарность: пересчёт в транзакции, откат при ошибке

**Acceptance criteria**:
- RFM пересчитывается автоматически по расписанию
- В логах видно количество обработанных клиентов и миграции между сегментами
- При ошибке пересчёта старые данные сохраняются

#### 3B.2.3. Стандартные RFM-сегменты

**Описание**: предустановленные сегменты, создаваемые автоматически при первом запуске RFM.

**Реализация**:
- При первом расчёте RFM -- автоматическое создание записей в `segments` с `type='rfm'`
- Имена сегментов на русском: Чемпионы, Лояльные, Потенциально лояльные, Новые, В зоне риска, Нельзя потерять, Потерянные
- `auto_assign = true` для всех RFM-сегментов
- Связь через `segment_clients`

**Acceptance criteria**:
- 7 стандартных RFM-сегментов создаются автоматически
- Клиенты автоматически попадают в соответствующие сегменты
- Сегменты обновляются при каждом пересчёте

#### 3B.2.4. Применение сегментов в модулях

**Описание**: использование RFM-сегментов для таргетинга в кампаниях, акциях, фильтрах аналитики.

**Реализация**:
- Расширить `CampaignFilter` и `PromotionFilter` поддержкой `rfm_segment`
- В аналитике -- фильтр по RFM-сегменту
- В списке клиентов -- фильтр и иконка сегмента

**Acceptance criteria**:
- При создании кампании можно выбрать целевой RFM-сегмент
- Аналитика фильтруется по RFM-сегментам
- В списке клиентов виден RFM-сегмент каждого клиента

#### 3B.2.5. UI: обзорный дашборд сегментов

**Описание**: страница с визуализацией RFM-распределения.

**Реализация**:
- Маршрут: `/dashboard/clients/segments` (уже существует, расширить)
- Карточки сегментов с количеством клиентов и долей
- Цветовая кодировка (зелёный -- champions, красный -- lost)
- Переход из карточки в отфильтрованный список клиентов
- Трендовый график: динамика сегментов за последние 30/90 дней

**Acceptance criteria**:
- Дашборд отображает все RFM-сегменты с количеством клиентов
- Клик по сегменту открывает отфильтрованный список
- Видна динамика изменения сегментов

---

## 5. Подфаза 3C: Onboarding Wizard

### 3C.1. Цель

Провести нового пользователя через настройку системы после регистрации. 6 шагов, каждый можно пропустить. Повторный доступ через настройки.

### 3C.2. Шаги мастера

| Шаг | Название | Обязательный | Описание |
|---|---|---|---|
| 1 | Информация | Нет | О системе, FAQ, демо-видео/GIF |
| 2 | Программа лояльности | Нет | Создание и настройка программы лояльности |
| 3 | Создание бота | Нет | Создание бота в Telegram, ввод токена, настройка |
| 4 | Точки продаж | Нет | Добавление POS-локаций (skip для одной точки) |
| 5 | Интеграции | Нет | Подключение POS-систем (iiko, r-keeper, 1C) |
| 6 | Следующие шаги | Нет | Гайд по Analytics, Campaigns, Promotions; ссылка на тарифы |

**Порядок**: Лояльность перед Ботом -- программа лояльности проще, бот без программы не имеет смысла.

### 3C.3. Задачи

#### 3C.3.1. Backend: onboarding state

**Описание**: хранение состояния онбординга в БД.

**Реализация**:
- Добавить `onboarding_completed BOOLEAN DEFAULT false` в `organizations`
- Добавить `onboarding_state JSONB DEFAULT '{}'` в `organizations` для хранения прогресса по шагам
- API endpoints:
  - `GET /api/v1/onboarding` -- текущее состояние
  - `PATCH /api/v1/onboarding` -- обновление состояния (завершение шага)
  - `POST /api/v1/onboarding/complete` -- пометить онбординг завершённым
  - `POST /api/v1/onboarding/reset` -- сброс для повторного прохождения

**Acceptance criteria**:
- Состояние онбординга хранится в БД
- API корректно возвращает и обновляет состояние
- `onboarding_completed` блокирует redirect после завершения

#### 3C.3.2. Backend: onboarding controller и usecase

**Реализация**:
- `internal/controller/http/group/onboarding/onboarding.go`
- `internal/usecase/onboarding/onboarding.go`
- Usecase переиспользует существующие usecases: `loyalty.Create`, `bots.Create`, `pos.Create`, `integrations.Create`

**Acceptance criteria**:
- Controller/usecase следуют установленным паттернам проекта
- Каждый шаг вызывает соответствующий существующий usecase

#### 3C.3.3. Frontend: redirect-логика

**Описание**: перенаправлять новых пользователей на онбординг вместо дашборда.

**Реализация**:
- В auth flow (после login/register) -- проверка `onboarding_completed`
- Если `false` -- redirect на `/dashboard/onboarding`
- Если `true` -- стандартный redirect на `/dashboard`
- Кнопка "Пройти настройку снова" в settings

**Acceptance criteria**:
- Новый пользователь после регистрации попадает на онбординг
- После завершения онбординга -- на дашборд
- Из настроек можно вернуться к онбордингу

#### 3C.3.4. Frontend: Wizard-компонент

**Описание**: компонент мастера с 6 шагами, прогресс-баром, кнопками навигации.

**Реализация**:
- Новый маршрут: `frontend/src/routes/dashboard/onboarding/index.tsx`
- Компонент `OnboardingWizard` с state machine для шагов
- Прогресс-индикатор (stepper) вверху страницы
- Каждый шаг переиспользует существующие формы создания:
  - Шаг 2: форма из `loyalty/index.tsx` (создание программы)
  - Шаг 3: форма из `bots/index.tsx` (создание бота) + настройки из `bots/$botId.tsx`
  - Шаг 4: форма из `pos/index.tsx` (добавление точки)
  - Шаг 5: форма из `integrations/index.tsx` (подключение POS)
- Кнопка "Создать позже" на каждом шаге
- Кнопки "Назад" / "Далее"
- Шаг 1: статический контент (описание, FAQ, демо)
- Шаг 6: карточки с ссылками на Analytics, Campaigns, Promotions

**Acceptance criteria**:
- Все 6 шагов отображаются корректно
- Прогресс-бар показывает текущий шаг
- Каждый шаг можно пропустить кнопкой "Создать позже"
- Формы создания работают в контексте wizard
- После последнего шага -- redirect на дашборд

---

## 6. Подфаза 3D: Bot Constructor UI

### 3D.1. Цель

Полноценный UI для настройки бота в админ-панели. Сущность `BotSettings` существует в backend, но фронтенд для неё отсутствует.

### 3D.2. Текущее состояние

Backend уже поддерживает:
```go
type BotSettings struct {
    Modules          []string    `json:"modules"`
    Buttons          []BotButton `json:"buttons"`
    RegistrationForm []FormField `json:"registration_form"`
    WelcomeMessage   string      `json:"welcome_message"`
}

type UpdateBotSettingsRequest struct {
    Modules          *[]string    `json:"modules,omitempty"`
    Buttons          *[]BotButton `json:"buttons,omitempty"`
    RegistrationForm *[]FormField `json:"registration_form,omitempty"`
    WelcomeMessage   *string      `json:"welcome_message,omitempty"`
}
```

API endpoint для обновления настроек уже должен существовать (через `bots` controller).

### 3D.3. Задачи

#### 3D.3.1. Страница бота с вкладками

**Описание**: расширить `bots/$botId.tsx` вкладками конфигурации.

**Реализация**:
- Вкладки: Общее | Сообщения | Кнопки | Форма регистрации | Модули | Меню | Привью
- Общее: имя бота, статус, привязанная программа лояльности, привязанные POS
- Каждая вкладка -- отдельный компонент, lazy-loaded

**Acceptance criteria**:
- Страница бота содержит вкладки конфигурации
- Переключение между вкладками без перезагрузки страницы

#### 3D.3.2. Редактор приветственного сообщения

**Описание**: текстовое поле для welcome message с live-превью.

**Реализация**:
- Textarea с поддержкой Telegram Markdown (bold, italic, links)
- Переменные-шаблоны: `{first_name}`, `{bonus_balance}`, `{loyalty_level}`
- Превью справа: отображение как в Telegram (bubble-стиль)
- Валидация длины (до 4096 символов -- лимит Telegram)

**Acceptance criteria**:
- Сообщение сохраняется через `UpdateBotSettingsRequest`
- Live-превью отображает форматирование
- Переменные-шаблоны подставляются в превью с примерными значениями

#### 3D.3.3. Конструктор кнопок

**Описание**: drag-and-drop интерфейс для управления кнопками бота.

**Реализация**:
- Список кнопок с drag-and-drop сортировкой (dnd-kit или встроенный)
- Для каждой кнопки: label, type (`url` | `callback` | `webapp`), value
- Кнопки "Добавить" / "Удалить"
- Ограничение: до 10 кнопок (лимит Telegram inline keyboard)
- Превью: отображение кнопок в стиле Telegram

**Acceptance criteria**:
- Кнопки добавляются, удаляются, переупорядочиваются
- Типы кнопок: url, callback, webapp
- Превью отображает кнопки как в Telegram
- Данные сохраняются в `BotSettings.Buttons`

#### 3D.3.4. Конструктор формы регистрации

**Описание**: настройка полей формы регистрации, которую заполняет пользователь при первом входе в бот.

**Реализация**:
- Список полей с drag-and-drop сортировкой
- Для каждого поля: `name` (системное), `label` (отображение), `type` (`text` | `phone` | `email` | `date` | `select`), `required` (boolean)
- Предустановленные поля: Имя, Телефон, Дата рождения, Город
- Пользователь может добавлять свои поля

**Acceptance criteria**:
- Поля формы настраиваются через UI
- Порядок полей меняется drag-and-drop
- Данные сохраняются в `BotSettings.RegistrationForm`

#### 3D.3.5. Управление модулями

**Описание**: переключатели для включения/выключения модулей бота.

**Реализация**:
- Список доступных модулей: `loyalty`, `menu`, `promotions`, `feedback`, `booking`
- Toggle-переключатели для каждого модуля
- Описание модуля при hover/click
- Сохранение в `BotSettings.Modules`

**Acceptance criteria**:
- Модули включаются/выключаются через toggle
- Список активных модулей сохраняется в настройках бота

#### 3D.3.6. Загрузка меню

**Описание**: загрузка меню для бота -- файлом или ручной ввод. Если подключена POS-интеграция, меню подтягивается автоматически (см. 3A.2.4).

**Реализация**:
- Ручной ввод: категории + позиции (название, цена, описание, фото URL)
- Загрузка файла: CSV/Excel с парсингом
- Отображение из POS-интеграции (read-only, если есть автоимпорт)
- Превью: карточки позиций, как в WebApp бота

**Acceptance criteria**:
- Меню вводится вручную или загружается файлом
- Если есть POS-интеграция с авто-импортом, меню отображается read-only
- Превью показывает меню в стиле бота

#### 3D.3.7. Привязка к POS-локациям

**Описание**: выбор, к каким точкам продаж привязан бот.

**Реализация**:
- Multi-select из существующих `pos_locations` организации
- Отображение выбранных точек с адресами
- Требует расширения entity `Bot` -- добавить `pos_ids` или связующую таблицу

**Acceptance criteria**:
- Бот привязывается к одной или нескольким POS-локациям
- Изменения сохраняются и отображаются

#### 3D.3.8. Превью бота

**Описание**: визуализация бота так, как его видит конечный пользователь в Telegram.

**Реализация**:
- Компонент `BotPreview` -- имитация Telegram-чата
- Отображает: welcome message, кнопки, форму регистрации, меню
- Обновляется в реальном времени при редактировании настроек
- Стилизация под Telegram (шрифты, bubble-стиль, цвета)

**Acceptance criteria**:
- Превью визуально соответствует Telegram-интерфейсу
- Все настройки отражаются в превью в реальном времени

---

## 7. Схемы базы данных (SQL-миграции)

### Миграция 3A: меню и привязка ботов к POS

```sql
-- Файл: backend/migrations/20260401120000_menus.sql

-- +goose Up

-- Меню, импортированное из POS или созданное вручную
CREATE TABLE menus (
    id             SERIAL PRIMARY KEY,
    org_id         INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    integration_id INT REFERENCES integrations(id) ON DELETE SET NULL,
    name           VARCHAR(255) NOT NULL,
    source         VARCHAR(20) NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'pos_import')),
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menus_org_id ON menus(org_id);

-- Категории меню
CREATE TABLE menu_categories (
    id        SERIAL PRIMARY KEY,
    menu_id   INT NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    name      VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menu_categories_menu_id ON menu_categories(menu_id);

-- Позиции меню
CREATE TABLE menu_items (
    id            SERIAL PRIMARY KEY,
    category_id   INT NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    price         NUMERIC(12,2) NOT NULL DEFAULT 0,
    image_url     VARCHAR(500),
    tags          JSONB NOT NULL DEFAULT '[]',
    external_id   VARCHAR(255),
    is_available  BOOLEAN NOT NULL DEFAULT true,
    sort_order    INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menu_items_category_id ON menu_items(category_id);
CREATE INDEX idx_menu_items_external_id ON menu_items(external_id);

-- Привязка бота к POS-локациям (many-to-many)
CREATE TABLE bot_pos_locations (
    bot_id INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    pos_id INT NOT NULL REFERENCES pos_locations(id) ON DELETE CASCADE,
    PRIMARY KEY (bot_id, pos_id)
);

-- Добавить телефон в external_orders для матчинга
ALTER TABLE external_orders
    ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS customer_name  VARCHAR(255);

CREATE INDEX idx_external_orders_customer_phone ON external_orders(customer_phone);

-- +goose Down
DROP INDEX IF EXISTS idx_external_orders_customer_phone;
ALTER TABLE external_orders
    DROP COLUMN IF EXISTS customer_phone,
    DROP COLUMN IF EXISTS customer_name;
DROP TABLE IF EXISTS bot_pos_locations;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
DROP TABLE IF EXISTS menus;
```

### Миграция 3B: RFM-расширение

```sql
-- Файл: backend/migrations/20260401120001_rfm_config.sql

-- +goose Up

-- Конфигурация RFM-расчёта для организации
CREATE TABLE rfm_configs (
    id              SERIAL PRIMARY KEY,
    org_id          INT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    period_days     INT NOT NULL DEFAULT 365,
    recalc_interval VARCHAR(20) NOT NULL DEFAULT '24h',
    last_calc_at    TIMESTAMPTZ,
    clients_processed INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- История изменений RFM-сегментов (для трендового графика)
CREATE TABLE rfm_history (
    id           SERIAL PRIMARY KEY,
    org_id       INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    segment      VARCHAR(50) NOT NULL,
    client_count INT NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_rfm_history_org_date ON rfm_history(org_id, calculated_at);

-- +goose Down
DROP TABLE IF EXISTS rfm_history;
DROP TABLE IF EXISTS rfm_configs;
```

### Миграция 3C: onboarding

```sql
-- Файл: backend/migrations/20260401120002_onboarding.sql

-- +goose Up

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS onboarding_state JSONB NOT NULL DEFAULT '{}';

-- onboarding_state структура:
-- {
--   "current_step": 1,
--   "steps": {
--     "1": {"completed": true, "skipped": false},
--     "2": {"completed": true, "skipped": false, "program_id": 5},
--     "3": {"completed": false, "skipped": true},
--     "4": {"completed": false, "skipped": true},
--     "5": {"completed": false, "skipped": true},
--     "6": {"completed": false, "skipped": false}
--   }
-- }

-- +goose Down

ALTER TABLE organizations
    DROP COLUMN IF EXISTS onboarding_completed,
    DROP COLUMN IF EXISTS onboarding_state;
```

### Миграция 3D: расширение bot_settings (не требуется)

Таблица `bots` уже содержит `settings JSONB` с полной структурой `BotSettings`. Связь бот-POS добавлена в миграции 3A (`bot_pos_locations`). Дополнительные миграции для 3D не нужны.

---

## 8. Backend: детали реализации

### 8.1. Структура новых пакетов

```
backend/internal/
  entity/
    menu.go                      -- Menu, MenuCategory, MenuItem, CRUD requests
    onboarding.go                -- OnboardingState, OnboardingStepStatus
  usecase/
    rfm/
      rfm.go                    -- RFM calculation engine
      rfm_test.go               -- Unit-тесты RFM-расчёта
    onboarding/
      onboarding.go             -- Orchestration поверх существующих usecases
  controller/
    http/group/
      onboarding/
        onboarding.go           -- HTTP handlers для онбординга
      menus/
        menus.go                -- HTTP handlers для меню
  repository/
    postgres/
      menus.go                  -- Репозиторий меню
      rfm.go                    -- Репозиторий RFM-конфигов и истории
  service/
    pos/
      rkeeper/
        rkeeper.go              -- r-keeper адаптер
      onec/
        onec.go                 -- 1C адаптер
```

### 8.2. Entity: menu.go

```go
package entity

import "time"

type Menu struct {
    ID            int        `db:"id"             json:"id"`
    OrgID         int        `db:"org_id"         json:"org_id"`
    IntegrationID *int       `db:"integration_id" json:"integration_id,omitempty"`
    Name          string     `db:"name"           json:"name"`
    Source        string     `db:"source"         json:"source"` // "manual"|"pos_import"
    LastSyncedAt  *time.Time `db:"last_synced_at" json:"last_synced_at,omitempty"`
    CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
    UpdatedAt     time.Time  `db:"updated_at"     json:"updated_at"`
    Categories    []MenuCategory `db:"-" json:"categories,omitempty"`
}

type MenuCategory struct {
    ID        int        `db:"id"         json:"id"`
    MenuID    int        `db:"menu_id"    json:"menu_id"`
    Name      string     `db:"name"       json:"name"`
    SortOrder int        `db:"sort_order" json:"sort_order"`
    CreatedAt time.Time  `db:"created_at" json:"created_at"`
    Items     []MenuItem `db:"-"          json:"items,omitempty"`
}

type MenuItem struct {
    ID          int     `db:"id"           json:"id"`
    CategoryID  int     `db:"category_id"  json:"category_id"`
    Name        string  `db:"name"         json:"name"`
    Description string  `db:"description"  json:"description,omitempty"`
    Price       float64 `db:"price"        json:"price"`
    ImageURL    string  `db:"image_url"    json:"image_url,omitempty"`
    Tags        Tags    `db:"tags"         json:"tags"`
    ExternalID  string  `db:"external_id"  json:"external_id,omitempty"`
    IsAvailable bool    `db:"is_available" json:"is_available"`
    SortOrder   int     `db:"sort_order"   json:"sort_order"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CreateMenuRequest struct {
    Name string `json:"name" binding:"required"`
}

type CreateMenuItemRequest struct {
    Name        string  `json:"name" binding:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"min=0"`
    ImageURL    string  `json:"image_url"`
    Tags        Tags    `json:"tags"`
}
```

### 8.3. Entity: onboarding.go

```go
package entity

type OnboardingState struct {
    CurrentStep int                          `json:"current_step"`
    Steps       map[string]OnboardingStep    `json:"steps"`
}

type OnboardingStep struct {
    Completed bool  `json:"completed"`
    Skipped   bool  `json:"skipped"`
    EntityID  *int  `json:"entity_id,omitempty"` // ID созданной сущности
}

type UpdateOnboardingRequest struct {
    Step      int  `json:"step" binding:"required,min=1,max=6"`
    Completed bool `json:"completed"`
    Skipped   bool `json:"skipped"`
    EntityID  *int `json:"entity_id,omitempty"`
}
```

### 8.4. Usecase: rfm.go (ключевая логика)

```go
package rfm

// Calculate выполняет полный RFM-пересчёт для организации.
//
// Алгоритм:
// 1. Получить всех клиентов org через bot_clients JOIN bots
// 2. Для каждого клиента агрегировать транзакции за period:
//    - external_orders (POS-данные)
//    - loyalty_transactions (внутренние операции)
// 3. Вычислить R, F, M raw-значения
// 4. Разбить на квинтили (percentile-based, не фиксированные пороги)
// 5. Присвоить сегмент по R-F-M score комбинации
// 6. Обновить bot_clients.rfm_* поля
// 7. Обновить segment_clients для RFM-сегментов
// 8. Записать rfm_history для трендов
//
// Весь расчёт выполняется в одной SQL-транзакции.

func (uc *Usecase) Calculate(ctx context.Context, orgID int, periodDays int) error
```

Основной SQL-запрос для агрегации (выполняется на стороне БД для производительности):

```sql
WITH client_stats AS (
    SELECT
        bc.id AS client_id,
        -- Recency: дней с последней транзакции
        COALESCE(
            EXTRACT(DAY FROM NOW() - GREATEST(
                MAX(eo.ordered_at),
                MAX(lt.created_at)
            )),
            999
        )::INT AS recency,
        -- Frequency: количество транзакций
        (COUNT(DISTINCT eo.id) + COUNT(DISTINCT lt.id))::INT AS frequency,
        -- Monetary: сумма
        COALESCE(SUM(eo.total), 0) + COALESCE(SUM(
            CASE WHEN lt.type = 'earn' THEN lt.amount ELSE 0 END
        ), 0) AS monetary
    FROM bot_clients bc
    JOIN bots b ON b.id = bc.bot_id AND b.org_id = $1
    LEFT JOIN external_orders eo ON eo.client_id = bc.id
        AND eo.ordered_at >= NOW() - INTERVAL '1 day' * $2
    LEFT JOIN loyalty_transactions lt ON lt.client_id = bc.id
        AND lt.created_at >= NOW() - INTERVAL '1 day' * $2
    GROUP BY bc.id
)
SELECT
    client_id,
    recency,
    frequency,
    monetary,
    NTILE(5) OVER (ORDER BY recency ASC)  AS r_score,
    NTILE(5) OVER (ORDER BY frequency DESC) AS f_score,
    NTILE(5) OVER (ORDER BY monetary DESC)  AS m_score
FROM client_stats
WHERE frequency > 0 OR monetary > 0;
```

### 8.5. Нормализация телефонов

```go
package phone

import "strings"

// Normalize приводит телефон к формату 7XXXXXXXXXX.
func Normalize(phone string) string {
    digits := strings.Map(func(r rune) rune {
        if r >= '0' && r <= '9' {
            return r
        }
        return -1
    }, phone)

    switch {
    case len(digits) == 11 && digits[0] == '8':
        return "7" + digits[1:]
    case len(digits) == 11 && digits[0] == '7':
        return digits
    case len(digits) == 10:
        return "7" + digits
    default:
        return digits
    }
}
```

### 8.6. API endpoints (новые)

| Метод | Путь | Описание | Подфаза |
|---|---|---|---|
| `GET` | `/api/v1/clients/:id/order-stats` | Поэлементная статистика заказов клиента | 3A |
| `GET` | `/api/v1/menus` | Список меню организации | 3A |
| `POST` | `/api/v1/menus` | Создать меню | 3A |
| `GET` | `/api/v1/menus/:id` | Получить меню с категориями и позициями | 3A |
| `PATCH` | `/api/v1/menus/:id` | Обновить меню | 3A |
| `DELETE` | `/api/v1/menus/:id` | Удалить меню | 3A |
| `POST` | `/api/v1/menus/:id/categories` | Добавить категорию | 3A |
| `POST` | `/api/v1/menus/:menuId/categories/:catId/items` | Добавить позицию | 3A |
| `PATCH` | `/api/v1/menus/items/:id` | Обновить позицию (теги, описание) | 3A |
| `GET` | `/api/v1/rfm/dashboard` | RFM-дашборд: сегменты + клиенты + тренды | 3B |
| `POST` | `/api/v1/rfm/recalculate` | Ручной запуск пересчёта RFM | 3B |
| `GET` | `/api/v1/rfm/config` | Текущие настройки RFM | 3B |
| `PATCH` | `/api/v1/rfm/config` | Обновить настройки RFM (период, интервал) | 3B |
| `GET` | `/api/v1/onboarding` | Состояние онбординга | 3C |
| `PATCH` | `/api/v1/onboarding` | Обновить шаг онбординга | 3C |
| `POST` | `/api/v1/onboarding/complete` | Завершить онбординг | 3C |
| `POST` | `/api/v1/onboarding/reset` | Сброс онбординга | 3C |
| `PATCH` | `/api/v1/bots/:id/settings` | Обновить настройки бота (уже есть?) | 3D |
| `GET` | `/api/v1/bots/:id/pos-locations` | POS-локации бота | 3D |
| `PUT` | `/api/v1/bots/:id/pos-locations` | Установить POS-локации бота | 3D |

---

## 9. Frontend: детали реализации

### 9.1. Новые маршруты

```
frontend/src/routes/dashboard/
  onboarding/
    index.tsx                    -- OnboardingWizard (3C)
  clients/
    segments.tsx                 -- Расширить RFM-дашбордом (3B)
    $clientId.tsx                -- Расширить вкладками заказов (3A)
  bots/
    $botId.tsx                   -- Расширить вкладками конструктора (3D)
  menus/
    index.tsx                    -- Список меню (3A)
    $menuId.tsx                  -- Редактор меню (3A)
```

### 9.2. Новые компоненты

```
frontend/src/components/
  onboarding/
    OnboardingWizard.tsx         -- Stepper + step content
    OnboardingStep.tsx           -- Обёртка одного шага
    StepInfo.tsx                 -- Шаг 1: информация
    StepLoyalty.tsx              -- Шаг 2: создание программы
    StepBot.tsx                  -- Шаг 3: создание + настройка бота
    StepPOS.tsx                  -- Шаг 4: добавление точек
    StepIntegrations.tsx         -- Шаг 5: подключение POS
    StepNextSteps.tsx            -- Шаг 6: дальнейшие действия
  bot-constructor/
    WelcomeMessageEditor.tsx     -- Редактор welcome message
    ButtonConstructor.tsx        -- Drag-and-drop конструктор кнопок
    RegistrationFormBuilder.tsx  -- Конструктор формы регистрации
    ModuleToggles.tsx            -- Управление модулями
    MenuUploader.tsx             -- Загрузка/ввод меню
    POSLocationSelector.tsx      -- Привязка к POS
    BotPreview.tsx               -- Превью бота в стиле Telegram
  rfm/
    RFMDashboard.tsx             -- Обзор RFM-сегментов
    RFMSegmentCard.tsx           -- Карточка сегмента
    RFMTrendChart.tsx            -- График динамики сегментов
  clients/
    OrderHistory.tsx             -- Вкладка истории POS-заказов
    OrderItemsTable.tsx          -- Таблица позиций заказа
    ClientOrderStats.tsx         -- Топ-позиции клиента
```

### 9.3. API hooks (TanStack Query)

```typescript
// frontend/src/api/onboarding.ts
export const useOnboarding = () => useQuery({ queryKey: ['onboarding'], queryFn: fetchOnboarding });
export const useUpdateOnboarding = () => useMutation({ mutationFn: updateOnboarding });
export const useCompleteOnboarding = () => useMutation({ mutationFn: completeOnboarding });

// frontend/src/api/rfm.ts
export const useRFMDashboard = () => useQuery({ queryKey: ['rfm-dashboard'], queryFn: fetchRFMDashboard });
export const useRFMRecalculate = () => useMutation({ mutationFn: triggerRFMRecalculate });
export const useRFMConfig = () => useQuery({ queryKey: ['rfm-config'], queryFn: fetchRFMConfig });

// frontend/src/api/menus.ts
export const useMenus = () => useQuery({ queryKey: ['menus'], queryFn: fetchMenus });
export const useMenu = (id: number) => useQuery({ queryKey: ['menus', id], queryFn: () => fetchMenu(id) });

// frontend/src/api/clients.ts (расширение)
export const useClientOrderStats = (id: number) =>
    useQuery({ queryKey: ['clients', id, 'order-stats'], queryFn: () => fetchClientOrderStats(id) });
```

### 9.4. Redirect-логика онбординга

```typescript
// В корневом layout или auth-guard:
const { data: user } = useAuth();
const { data: onboarding } = useOnboarding();

useEffect(() => {
    if (user && onboarding && !onboarding.onboarding_completed) {
        navigate({ to: '/dashboard/onboarding' });
    }
}, [user, onboarding]);
```

---

## 10. Стратегия тестирования

### 10.1. Backend unit-тесты

| Пакет | Тесты | Моки |
|---|---|---|
| `usecase/rfm` | Расчёт R/F/M, маппинг сегментов, квинтили, edge cases (0 транзакций, 1 клиент) | `rfmRepo`, `clientsRepo` (struct с function fields) |
| `usecase/onboarding` | Переходы между шагами, skip, complete, reset | `orgRepo` (struct с function fields) |
| `service/pos/rkeeper` | Парсинг ответов r-keeper API, маппинг на внутренние типы | HTTP-мок (httptest) |
| `service/pos/onec` | Парсинг ответов 1C API, маппинг на внутренние типы | HTTP-мок (httptest) |
| `phone.Normalize` | Форматы: +7, 8, 10 цифр, с пробелами, скобками, невалидные | Нет моков |

### 10.2. Integration-тесты

```go
//go:build integration

// tests/integration/rfm_test.go
// - Наполнить БД тестовыми клиентами и транзакциями
// - Запустить RFM-расчёт
// - Проверить корректность сегментации
// - Проверить запись в rfm_history

// tests/integration/onboarding_test.go
// - Создать организацию
// - Пройти все шаги онбординга через API
// - Проверить создание сущностей (program, bot, POS)
// - Проверить onboarding_completed = true

// tests/integration/menu_test.go
// - Создать меню, категории, позиции через API
// - Проверить каскадное удаление
// - Проверить обновление из POS-импорта
```

### 10.3. Frontend unit-тесты (Vitest + MSW)

| Компонент | Тесты |
|---|---|
| `OnboardingWizard` | Навигация по шагам, skip, complete, прогресс-бар |
| `ButtonConstructor` | Добавление/удаление/сортировка кнопок, валидация |
| `RegistrationFormBuilder` | Добавление/удаление полей, required toggle |
| `RFMDashboard` | Отображение сегментов, клик по карточке, пустое состояние |
| `BotPreview` | Рендеринг welcome message, кнопок, формы |

### 10.4. E2E-тесты (Playwright)

```
e2e/tests/
  onboarding.spec.ts         -- Полный flow: регистрация -> онбординг -> dashboard
  bot-constructor.spec.ts    -- Настройка бота: welcome, кнопки, форма, превью
  rfm-dashboard.spec.ts      -- Просмотр RFM-дашборда, переход к клиентам
  client-orders.spec.ts      -- Просмотр истории заказов клиента
```

---

## 11. Definition of Done

### 3A: Integrations v2

- [ ] Заказы из POS автоматически привязываются к клиентам по телефону
- [ ] r-keeper коннектор реализует полный интерфейс `posService` и проходит TestConnection
- [ ] 1C коннектор реализует полный интерфейс `posService` и проходит TestConnection
- [ ] Меню импортируется из POS и хранится локально в `menus`/`menu_items`
- [ ] На странице клиента отображается история POS-заказов с позициями
- [ ] Поэлементная аналитика: топ-позиции клиента, частота заказов
- [ ] Тегирование позиций меню работает, аналитика по тегам отображается
- [ ] Unit-тесты: phone.Normalize, rkeeper adapter, onec adapter
- [ ] Integration-тесты: привязка заказов, меню CRUD
- [ ] Миграция `20260401120000_menus.sql` применяется и откатывается без ошибок

### 3B: RFM-сегментация

- [ ] RFM-показатели корректно вычисляются (проверено на тестовых данных)
- [ ] 7 стандартных сегментов создаются автоматически при первом расчёте
- [ ] Cron-задача пересчитывает RFM ежедневно
- [ ] RFM-дашборд отображает сегменты с количеством клиентов
- [ ] Трендовый график показывает динамику за 30/90 дней
- [ ] Клик по сегменту -> отфильтрованный список клиентов
- [ ] RFM-сегменты доступны как фильтр в кампаниях и аналитике
- [ ] Unit-тесты: расчёт квинтилей, маппинг сегментов, edge cases
- [ ] Integration-тесты: полный цикл расчёта на тестовой БД
- [ ] Миграция `20260401120001_rfm_config.sql` применяется и откатывается без ошибок

### 3C: Onboarding Wizard

- [ ] Новый пользователь после регистрации redirect на `/dashboard/onboarding`
- [ ] Все 6 шагов отображаются с прогресс-баром
- [ ] Каждый шаг можно пропустить кнопкой "Создать позже"
- [ ] Шаги 2-5 создают реальные сущности через существующие API
- [ ] После завершения -- redirect на дашборд, `onboarding_completed = true`
- [ ] Из настроек можно вернуться к онбордингу (reset)
- [ ] Unit-тесты: usecase transitions, frontend wizard navigation
- [ ] E2E-тесты: полный flow регистрация -> онбординг -> дашборд
- [ ] Миграция `20260401120002_onboarding.sql` применяется и откатывается без ошибок

### 3D: Bot Constructor UI

- [ ] Страница бота содержит вкладки: Общее, Сообщения, Кнопки, Форма, Модули, Меню, Превью
- [ ] Welcome message редактируется с live-превью в стиле Telegram
- [ ] Кнопки добавляются/удаляются/переупорядочиваются через drag-and-drop
- [ ] Форма регистрации настраивается: поля, порядок, required
- [ ] Модули включаются/выключаются через toggle
- [ ] Меню загружается файлом или вводится вручную
- [ ] Бот привязывается к POS-локациям через multi-select
- [ ] Превью отображает все настройки в реальном времени
- [ ] Все изменения сохраняются через `PATCH /api/v1/bots/:id/settings`
- [ ] Unit-тесты: ButtonConstructor, RegistrationFormBuilder, BotPreview
- [ ] E2E-тесты: полная настройка бота через конструктор

---

## 12. Порядок реализации и коммиты

Рекомендуемый порядок внутри каждой подфазы.

### 3A: Integrations v2

```
feat: add menus schema migration (20260401120000)
feat: add menu entity and repository
feat: add menu CRUD controller and usecase
feat: implement phone normalization utility
feat: add customer phone matching in sync service
feat: implement r-keeper POS connector
feat: implement 1C POS connector
feat: add client order stats endpoint
feat: add order history tab to client detail page
feat: add menu item tagging and analytics
```

### 3B: RFM-сегментация

```
feat: add rfm_config migration (20260401120001)
feat: add RFM calculation engine usecase
feat: register RFM recalculation as scheduler task
feat: auto-create standard RFM segments on first calculation
feat: add RFM dashboard API endpoints
feat: add RFM segment cards and trend chart to frontend
feat: add RFM segment filter to campaigns and analytics
```

### 3C: Onboarding Wizard

```
feat: add onboarding migration (20260401120002)
feat: add onboarding entity and state management
feat: add onboarding controller and usecase
feat: add onboarding redirect logic to auth flow
feat: implement OnboardingWizard component with 6 steps
feat: reuse existing create forms in wizard context
```

### 3D: Bot Constructor UI

```
feat: add bot_pos_locations relation (included in 3A migration)
feat: add tabbed layout to bot detail page
feat: implement welcome message editor with preview
feat: implement button constructor with drag-and-drop
feat: implement registration form builder
feat: implement module toggles
feat: implement menu upload and display
feat: implement POS location selector for bots
feat: implement bot preview component (Telegram style)
```
