# Анализ Phase 4: Полнота реализации, допущения и заглушки

## Общая картина

| Подфаза | Спецификация | Реализовано | Заглушки | Оценка |
|---------|-------------|-------------|----------|--------|
| **4D Биллинг** | Тарифы, платежи, feature gating | Тарифы, подписки, инфраструктура | Платёжные провайдеры, вебхуки, gating не подключён | **~40%** |
| **4A Админ-бот** | Отдельный сервис, рассылки, аналитика | Линковка, базовые команды, статистика | Рассылки с медиа, авто-действия, уведомления, scheduler | **~45%** |
| **4E Рассылки+** | A/B-тесты, мультиканальность, шаблоны | Варианты, шаблоны, базовая аналитика | Стат. значимость, автопобедитель, мультиканал, воронки | **~30%** |
| **4B Wallet** | .pkpass, Google Wallet, push-обновления | Конфиг, CRUD пассов, push-регистрация | Генерация пассов, push-отправка, интеграция с loyalty | **~30%** |
| **4C Маркетплейс** | 3 сценария, бот, loyalty, Telegram Payments | Админ-панель товаров и заказов | Бот-интерфейс, SpendPoints, платежи | **~35%** |
| **4F Сегментация+** | Предикции, поведение, автоматизация | Схема, CRUD правил, UI прогнозов | ComputePredictions — заглушка, правила не применяются | **~40%** |

---

## Детальный разбор по подфазам

### 4D Биллинг — критические пробелы

**Реализовано полностью:**
- Тарифные планы (Trial/Basic/Pro/Enterprise) с seed-данными
- CRUD подписок и счетов
- Логика `HasFeature()` и `CheckLimit()`
- Scheduler `HandleExpiredSubscriptions` (active → past_due → expired)
- Frontend: выбор тарифа, список счетов

**Заглушки и допущения:**

1. **`ProcessPayment()` — платёж захардкожен как `"succeeded"`**
   - Нет вызова YooKassa/CloudPayments API
   - `ProviderPaymentID` просто принимается из запроса без верификации
   - Нет вебхук-эндпоинта для провайдера (спек требует "отдельный эндпоинт без JWT, с подписью провайдера")

2. **Feature gating НЕ подключён** — `HasFeature()` и `CheckLimit()` **нигде не вызываются** за пределами биллинг-пакета. Пользователь на Trial может использовать Pro-функции.

3. **Нет авто-продления** — подписки создаются, но при истечении нет логики перевыпуска счёта и списания. Только переход active → past_due → expired.

4. **Нет handler `ProcessPayment`** в HTTP-контроллере — метод в usecase есть, но нет эндпоинта для его вызова.

5. **Нет транзакционной безопасности** — создание подписки и счёта не атомарно. При ошибке создания счёта подписка остаётся без него (только логируется).

**Почему**: Платёжные провайдеры требуют аккаунт, договор, ККТ (54-ФЗ). Инфраструктура подготовлена для быстрого подключения, когда юридическая часть будет готова.

**Отсутствующие компоненты:**
- Таблица `payment_provider_config` (API-ключи, секреты)
- Таблица `webhook_logs` (для отладки)
- Таблица `billing_events` (аудит)
- Email-уведомления (напоминание об оплате, просрочка, продление)
- Генерация PDF-счетов
- Управление платёжными методами (сохранённые карты)

---

### 4A Админ-бот — читающий, не управляющий

**Реализовано:**
- Отдельный сервис `cmd/admin-bot` с Docker Compose
- Линковка аккаунта (одноразовый код, 10 мин TTL)
- Команды: `/stats` (7-дневная статистика), `/campaigns` (список), `/promotions`, `/promo CODE`
- Разграничение ролей owner/manager
- Reply-клавиатура с emoji-кнопками

**Заглушки:**

1. **Рассылки с rich media** — спек выделяет это как уникальную ценность ("кружки, голосовые, стикеры"). **Полностью отсутствует** — бот только читает кампании, не создаёт и не отправляет.

2. **Авто-действия (старт/стоп)** — ни одной ссылки на auto-actions в коде админ-бота. Нет repo-интерфейса для авто-действий.

3. **Уведомления (дневная/недельная сводка)** — analyticsRepo передаётся как `nil`. Нет scheduler для периодических сообщений. Комментарий в коде: `nil, // analyticsRepo — placeholder for future use`.

4. **Frontend** для генерации кода линковки — API есть (`POST /api/v1/admin-bot/link-code`), UI нет. Нет React-компонента в settings.

5. **Rate limiting** — нет ограничений на частоту действий через бота.

**Почему**: Бот создан как каркас для быстрого расширения. Чтение данных работает; создание/отправка требует интеграции с существующим bot sender, что увеличивает scope. Rich media (кружки, голосовые) — уникальная возможность Telegram, но требует отдельной entity для медиа и file_id хранения.

**Отсутствующие файлы:**
- `internal/service/adminbot/campaigns.go` — создание/отправка кампаний
- `internal/service/adminbot/actions.go` — управление авто-действиями
- `internal/service/adminbot/scheduler.go` — cron для сводок
- `internal/service/adminbot/analytics.go` — аналитические уведомления
- `frontend/src/features/adminbot/` — UI для настроек админ-бота

---

### 4E Рассылки+ — каркас A/B без статистики

**Реализовано:**
- `campaign_variants` таблица с `audience_pct`, `stats` (JSONB), `is_winner`
- CRUD вариантов, ручной выбор победителя (`PickWinner`)
- Шаблоны кампаний с категориями (welcome, promo, holiday, reactivation)
- 4 системных шаблона в seed
- HTTP-эндпоинты для A/B-тестов и шаблонов
- Frontend: страница шаблонов

**Заглушки:**

1. **Статистическая значимость** — нет chi-square, p-value, confidence intervals. Расчёт: `click_rate = clicked / sent` без оценки значимости. Нет порога минимальной выборки.

2. **Автоматический выбор победителя** — только ручной `PickWinner()`. Нет автоматики по порогу статистической значимости.

3. **Мультиканальность** — только Telegram. Нет push через Wallet, нет SMS. Нет каскадной отправки (Telegram → push → SMS). `campaign_messages` ссылается только на `telegram_id`.

4. **Воронки и heat maps** — нет tracking opens (`opened_at` отсутствует в `campaign_messages`), нет `campaign_results` таблицы, нет временных рядов для heat maps.

5. **Frontend A/B UI** — типы определены, но UI для создания/просмотра A/B-тестов отсутствует. Страница деталей кампании показывает только базовую статистику.

6. **Шаблоны по типу заведения** — категории есть (general, welcome, promo, holiday, reactivation), но нет привязки к типу заведения (бар, кафе, ресторан).

**Недостающие поля в БД:**
- `campaign_messages.opened_at` — для tracking открытий
- `campaign_messages.channel` — для мультиканальности (telegram, push, sms)
- `campaign_variants.p_value` — для статистической значимости
- `campaign_variants.min_sample_reached` — для валидации выборки
- Отсутствует таблица `campaign_results` для event tracking

**Почему**: A/B-тесты полезны только при достаточном объёме аудитории (>1000 получателей для статистической значимости). Инфраструктура готова, статистику можно добавить когда будет реальный трафик.

---

### 4B Wallet — только DB-записи, нет реальных пассов

**Реализовано:**
- Конфигурация Apple/Google платформ (credentials + design в JSONB)
- Выпуск pass-записей (serial 16 байт hex, auth_token 32 байта hex) в БД
- Регистрация push-токенов (публичный endpoint без JWT)
- Отзыв пассов
- Admin UI: настройка платформ, список пассов, цветовые пикеры для дизайна, статистика
- 14 unit-тестов

**Заглушки (критические):**

1. **Генерация `.pkpass`** — нет библиотеки (go-apple-wallet, apple-pass-kit), нет файла, нет endpoint для скачивания. `IssuePass()` только пишет запись в БД. Спек указывает `GET /api/v1/wallet/:serial — скачивание pass`, но этого endpoint нет.

2. **Google Wallet API** — нет вызова API. Нет `google.golang.org/api/walletobjects` в go.mod. Нет JWT-подписи для Google.

3. **APNs/FCM push** — `GetPassesWithPushToken()` существует, но **нигде не вызывается**. Нет библиотек: `apns2` (Apple), `firebase.google.com/go` (Google) отсутствуют в go.mod.

4. **`RefreshPassBalance()`** — метод существует, но **не вызывается из loyalty usecase**. При изменении баланса лояльности пассы не обновляются. Поиск по всему usecase/loyalty/ — ноль ссылок на wallet.

5. **Нет download endpoint** — спек указывает `GET /api/v1/wallet/:serial`, endpoint отсутствует.

6. **QR-код** — serial_number генерируется, но нет библиотеки для генерации QR и встраивания в pass.

7. **Credentials не валидируются** — сертификат Apple хранится as-is, нет проверки формата .p12 или срока действия.

**Что работает end-to-end:**
- ✅ Админ может настроить платформу (credentials + design)
- ✅ Backend создаёт запись пасса в БД
- ✅ Клиент может зарегистрировать push-токен
- ✅ Админ видит статистику

**Что НЕ работает end-to-end:**
- ❌ Пользователь не может получить реальный файл .pkpass
- ❌ Пользователь не может добавить карту в Google Wallet
- ❌ При изменении баланса push не отправляется
- ❌ Нет бот-команды для получения карты
- ❌ Нет клиентского UI для скачивания

**Почему**: Требуется Apple Developer Account ($99/год) и Google Cloud проект с Wallet API. Инфраструктура готова для подключения, когда аккаунты будут настроены. Credentials хранятся, но не используются.

---

### 4C Маркетплейс — только Сценарий 1, только админка

**Реализовано:**
- Полный CRUD товаров с управлением стоком (nullable stock = безлимит)
- Система заказов со статусным lifecycle (pending → confirmed → completed → cancelled)
- Атомарный декремент стока (`WHERE stock > 0`)
- Admin UI: карточки товаров, таблица заказов, статистика
- 16 unit-тестов

**Заглушки:**

1. **`SpendPoints` НЕ вызывается** — заказ создаётся, `total_points` записывается, но **баланс клиента не меняется**. Это data inconsistency:
   ```
   order: "150 points spent" ← записано
   client_loyalty.balance: unchanged ← НЕ списано
   ```
   В usecase нет зависимости на loyalty:
   ```go
   type Usecase struct {
       logger   *slog.Logger
       products productsRepo
       orders   ordersRepo
       // ← Отсутствует: loyalty interface
   }
   ```

2. **Нет бот-интерфейса** — спек требует "inline-каталог с кнопками, подтверждение заказа, списание баллов". Отсутствует полностью. Нет Telegram bot handler для маркетплейса.

3. **Сценарий 2 (уникальные позиции)** — нет концепции "обед дня", триггерных товаров, эксклюзивного меню бота. Нет time-based активации.

4. **Сценарий 3 (оплата через бота)** — нет Telegram Payments API, нет POS-интеграции для оплаты, нет поддержки смешанных платежей (баллы + деньги).

5. **Нет клиентского UI** — только админ-панель. "Мои заказы" для пользователя отсутствует.

**Почему**: Реализован только Scenario 1 MVP (каталог за баллы) как фундамент. Бот-интеграция требует inline keyboards и state machine для пользовательского flow. **Loyalty-интеграция пропущена — это баг, а не допущение.**

---

### 4F Сегментация+ — схема без вычислений

**Реализовано:**
- `segment_rules` и `client_predictions` таблицы (миграция 00029)
- CRUD правил (field/operator/value с JSONB value)
- Запрос предикций, саммари, high-churn фильтр
- Option pattern в usecase: `WithRules()`, `WithPredictions()`
- Frontend: страница прогнозов с таблицей рисков и карточками саммари
- 25 unit-тестов

**Заглушки:**

1. **`ComputePredictions()` — полная заглушка:**
   ```go
   func (uc *Usecase) ComputePredictions(ctx context.Context, orgID int) error {
       if uc.predictions == nil {
           return fmt.Errorf("predictions not configured")
       }
       // Heuristic prediction will be implemented when POS/loyalty data is available.
       // For now, this is a placeholder that can be connected to the scheduler.
       uc.logger.Info("compute predictions called", "org_id", orgID)
       return nil
   }
   ```
   Только логирует и возвращает nil. Никаких вычислений.

2. **Правила не применяются** — `CreateRule`/`GetRules`/`DeleteRule` работают (CRUD), но **нет evaluation engine**. Правила хранятся в БД и **никогда не используются для фильтрации клиентов**. Нет кода, который интерпретирует `field=days_since_visit, operator=gt, value=30`.

3. **`auto_assign` — флаг без логики**. Поле существует в Segment entity, API его принимает, но **нигде не обрабатывается**. Нет кода для автоматического назначения офферов/скидок по сегменту.

4. **Нет scheduler task** — `ComputePredictions` **не зарегистрирован в scheduler** (проверены все задачи в cmd/server/main.go:213-257). В отличие от rfm_recalculate (зарегистрирован), предикции не пересчитываются.

5. **Поведенческие метрики** — `PredictionFactors` определены (days_since_last_visit, visit_trend, spend_trend, avg_check, total_orders, loyalty_level), но **никогда не вычисляются**. Нет интеграции с POS-данными для поведенческого анализа.

6. **Frontend показывает пустые данные** — страница прогнозов корректно вызывает API, но backend возвращает пустые массивы (нет computed predictions в БД).

**Почему**: Спек прямо указывает: "приоритет: Низкий — зависит от объёма данных; имеет смысл при достаточной клиентской базе". Предикции имеют смысл при наличии POS-данных и истории транзакций. Схема готова для подключения эвристик.

---

## Сводка допущений

| Допущение | Затронутые модули | Обоснование |
|-----------|------------------|-------------|
| Платёжные провайдеры подключатся позже | 4D Billing | Требуют юр. договор, ККТ (54-ФЗ), аккаунты |
| Apple/Google аккаунты будут настроены отдельно | 4B Wallet | $99/год Apple Developer, Google Cloud Wallet API |
| Бот-интерфейсы — отдельная итерация | 4A, 4C | Inline keyboards + state machine = значительный scope |
| A/B статистика нужна при >1000 получателей | 4E Campaigns | Маленькие выборки не дают significance |
| Предикции нужны при достаточной базе | 4F Segmentation | Спек прямо указывает "Низкий приоритет" |
| Feature gating подключится при запуске биллинга | 4D Billing | Gating бессмысленен без реальных тарифов |
| ProcessPayment всегда "succeeded" | 4D Billing | Временно, до интеграции с провайдером |
| Credentials хранятся без шифрования | 4B Wallet | JSONB в plain text, нужно шифрование для production |

---

## Критические проблемы (требуют исправления)

| # | Проблема | Модуль | Severity | Статус |
|---|----------|--------|----------|--------|
| 1 | ~~SpendPoints не вызывается~~ | 4C Marketplace | ~~BUG~~ | ✅ **ИСПРАВЛЕНО** — loyalty interface добавлен, SpendPoints вызывается при заказе |
| 2 | ~~ProcessPayment хардкодит "succeeded"~~ | 4D Billing | ~~SECURITY~~ | ✅ **ИСПРАВЛЕНО** — платёж создаётся как "pending", подтверждение через webhook |
| 3 | ~~Feature gating не подключён~~ | 4D Billing | ~~HIGH~~ | ✅ **ИСПРАВЛЕНО** — FeatureGate middleware применён к loyalty, campaigns, integrations, analytics, rfm |
| 4 | **RefreshPassBalance не вызывается** | 4B Wallet | **HIGH** | ⏳ Требует доработки |
| 5 | **Подписка и счёт не атомарны** | 4D Billing | **MEDIUM** | ⏳ Требует доработки |
| 6 | **ComputePredictions — заглушка** | 4F Segmentation | **LOW** | ⏳ Ожидаемо по приоритету спека |

---

## Рекомендации по приоритету доработки

### P0 — Исправить баги ✅ DONE
1. ~~Marketplace: добавить loyalty interface в usecase, вызвать SpendPoints при создании заказа~~ — **ИСПРАВЛЕНО**. Добавлен `loyaltySpender` interface, баланс проверяется перед заказом, SpendPoints вызывается. ProgramID добавлен в PlaceOrderRequest. +1 тест (InsufficientPoints).
2. ~~Billing: убрать хардкод "succeeded" из ProcessPayment, добавить проверку~~ — **ИСПРАВЛЕНО**. ProcessPayment создаёт платёж со статусом "pending". Добавлены валидации: проверка суммы (ErrAmountMismatch), проверка повторной оплаты (ErrInvoiceAlreadyPaid). +5 тестов.

### P1 — Для запуска MVP ✅ PARTIALLY DONE
3. ~~Billing: подключить вебхуки~~ — **ИНФРАСТРУКТУРА ГОТОВА**. Добавлен `ConfirmPayment()` метод в usecase, `POST /api/v1/billing/webhook` эндпоинт (без JWT, для провайдера), `POST /api/v1/billing/payments` (с JWT, для клиента). Реализован `GetPaymentByProviderID`/`UpdatePaymentStatus` в repo. **Осталось**: подключить реальный SDK YooKassa/CloudPayments + верификацию подписи вебхука.
4. ~~Billing: активировать feature gating в ключевых usecase-ах~~ — **ИСПРАВЛЕНО**. Создан `FeatureGate` middleware. Применён к 5 группам: loyalty, campaigns, integrations, analytics, rfm. Fail-open стратегия (при ошибке биллинга доступ не блокируется).
5. Admin-bot: добавить создание/отправку рассылок (ключевая ценность) — **НЕ РЕАЛИЗОВАНО** (требует интеграции с bot sender, inline keyboards, state machine)

### P2 — Для полноценного продукта
6. Wallet: подключить Apple/Google библиотеки, реализовать генерацию пассов
7. Wallet: интегрировать с loyalty (RefreshPassBalance hook)
8. Marketplace: создать бот-интерфейс с inline-каталогом
9. Campaigns: добавить статистическую значимость и автовыбор победителя

### P3 — При достаточной базе пользователей
10. Segmentation: реализовать ComputePredictions (эвристики)
11. Segmentation: подключить evaluation engine для segment_rules
12. Campaigns: мультиканальность (push + SMS fallback)
13. Admin-bot: scheduler для дневных/недельных сводок
