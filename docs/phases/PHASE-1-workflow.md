# Фаза 1: Ядро — Пошаговый Workflow

> Детальный порядок реализации с зависимостями, параллелизацией и критериями проверки.

---

## Обзор потоков

```
Step 1: Auth (Backend) ──────────────────────────────────────────────────────┐
        Auth (Frontend)                                                       │
                                                                              │
Step 2: ┌──────────────────┐  ┌─────────────────────┐  ┌──────────────────┐  │
        │ Telegram Bot     │  │ Программа лояльности │  │ Точки продаж     │  │
        │ (Backend + Admin)│  │ (Backend + Admin)    │  │ (Backend + Admin)│  │
        └────────┬─────────┘  └──────────┬──────────┘  └────────┬─────────┘  │
                 │                        │                      │            │
                 ▼                        ▼                      ▼            │
Step 3: ┌─────────────────────────────────────────────────────────────────┐   │
        │        Интеграция: Bot ↔ Loyalty ↔ Admin panel                 │   │
        └────────────────────────┬────────────────────────────────────────┘   │
                                 │                                            │
Step 4: ┌────────────────────────┴────────────────────────────────────────┐   │
        │        Dashboard Layout + Responsive + Polish                    │   │
        └────────────────────────┬────────────────────────────────────────┘   │
                                 │                                            │
Step 5: ┌────────────────────────┴────────────────────────────────────────┐   │
        │        E2E Testing + Deploy + Verification                       │   │
        └─────────────────────────────────────────────────────────────────┘   │
```

---

## Step 1: Auth (аутентификация)

**Зависимости**: Phase 0 завершена (PG + Redis работают, скелет развернут)
**Это блокирующий шаг** — все остальные модули требуют аутентификации.

---

### Step 1A: Auth Backend

**Порядок**: последовательный (каждый слой зависит от предыдущего)

#### 1A.1 — Миграции

Таблицы `users` и `organizations` уже созданы в `00001_init.sql`.
Нужно добавить поля для refresh-токенов.

```
backend/migrations/00002_auth_sessions.sql
```

Таблица `auth_sessions`:
- `id` UUID PK
- `user_id` INT REFERENCES users
- `refresh_token` VARCHAR(500) UNIQUE
- `expires_at` TIMESTAMPTZ
- `created_at` TIMESTAMPTZ DEFAULT NOW()

**Проверка**: `goose -dir migrations postgres "$DATABASE_URL" up` → успешно

#### 1A.2 — Entity

Дополнить существующие entity + новые:

```
backend/internal/entity/user.go          — добавить методы, валидацию
backend/internal/entity/organization.go  — без изменений
backend/internal/entity/auth.go          — TokenPair, RegisterRequest, LoginRequest
```

#### 1A.3 — Repository: Users

```
backend/internal/repository/postgres/users.go
```

Методы:
- `Create(ctx, user) (User, error)` — INSERT с RETURNING
- `GetByID(ctx, id) (User, error)`
- `GetByEmail(ctx, email) (User, error)`
- `GetByPhone(ctx, phone) (User, error)`
- `Update(ctx, user) error`
- `CreateOrganization(ctx, org) (Organization, error)`

Паттерн: отдельный файл для каждого домена, работает с тем же `*sqlx.DB`.

#### 1A.4 — Repository: Auth Sessions (Redis)

```
backend/internal/repository/redis/sessions.go
```

Методы:
- `StoreRefreshToken(ctx, userID, token, ttl) error`
- `GetRefreshToken(ctx, token) (userID, error)`
- `DeleteRefreshToken(ctx, token) error`
- `DeleteUserSessions(ctx, userID) error`

#### 1A.5 — Usecase: Auth

```
backend/internal/usecase/auth/auth.go
```

Зависимости: users repo, sessions repo, config (JWT secret, token TTL)

Методы:
- `Register(ctx, req RegisterRequest) (TokenPair, error)`
  1. Валидация email/phone
  2. Хеш пароля (bcrypt)
  3. Создать Organization
  4. Создать User (role=owner, org_id)
  5. Генерация JWT access + refresh
  6. Сохранить refresh в Redis
  7. Вернуть TokenPair

- `Login(ctx, email, password) (TokenPair, error)`
  1. Найти пользователя по email
  2. Проверить bcrypt
  3. Генерация токенов
  4. Вернуть TokenPair

- `Refresh(ctx, refreshToken) (TokenPair, error)`
  1. Проверить refresh token в Redis
  2. Удалить старый
  3. Генерация новых токенов
  4. Вернуть TokenPair

- `Logout(ctx, refreshToken) error`
  1. Удалить refresh token из Redis

Паттерн: usecase получает интерфейсы repo, не конкретные типы.

#### 1A.6 — Controller: Auth Group

```
backend/internal/controller/http/group/auth/auth.go
```

Реализует `group` interface:
- `Path()` → `"/auth"`
- `Auth()` → `nil` (публичные эндпоинты)
- `Handlers()`:
  - `POST /register` → `handleRegister`
  - `POST /login` → `handleLogin`
  - `POST /refresh` → `handleRefresh`
  - `POST /logout` → `handleLogout`

#### 1A.7 — Wiring в main.go

Обновить `cmd/server/main.go`:
- Создать `usersRepo`
- Создать `sessionsRepo`
- Создать `authUsecase`
- Создать `authGroup`
- Добавить `authGroup` в HTTP module
- Добавить `authUsecase` в Application

**Проверка 1A**:
```bash
# Компиляция
go build ./cmd/server

# Ручное тестирование
curl -X POST localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret123","name":"Test","organization":"My Cafe"}'
# → 201 {access_token, refresh_token}

curl -X POST localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret123"}'
# → 200 {access_token, refresh_token}

curl localhost:8080/healthz
# → 200 (всё ещё работает)
```

---

### Step 1B: Auth Frontend

**Зависимости**: Step 1A (backend auth endpoints работают)
**Может начинаться параллельно**: можно делать UI без backend, потом интегрировать

#### 1B.1 — Auth Store (Zustand)

```
frontend/src/stores/auth.ts
```

Состояние:
- `accessToken: string | null`
- `user: User | null`
- `isAuthenticated: boolean`

Действия:
- `login(email, password) → Promise<void>`
- `register(data) → Promise<void>`
- `logout() → void`
- `refreshToken() → Promise<void>`

Хранение: `localStorage` для refresh token, in-memory для access token.

#### 1B.2 — Auth API

```
frontend/src/features/auth/api.ts
frontend/src/features/auth/types.ts
```

- `authApi.login(email, password)`
- `authApi.register(data)`
- `authApi.refresh(refreshToken)`
- `authApi.logout(refreshToken)`

#### 1B.3 — Страница логина (обновить)

```
frontend/src/routes/auth/login.tsx     — обновить (подключить к API)
```

- Подключить к auth store
- Валидация формы
- Обработка ошибок (неверный пароль, пользователь не найден)
- Редирект на /dashboard после успешного логина

#### 1B.4 — Страница регистрации

```
frontend/src/routes/auth/register.tsx
```

Поля: email, пароль, имя, название организации
Из Figma: node `3:73`

#### 1B.5 — SMS-верификация (заглушка)

```
frontend/src/routes/auth/verify.tsx
```

- Ввод 4-значного кода
- Для MVP: принимаем любой код или код из логов сервера
- Из Figma: node `27:562`

#### 1B.6 — Protected Routes

```
frontend/src/routes/dashboard/route.tsx  — обновить (добавить guard)
```

- `beforeLoad`: проверка auth store → redirect если нет токена
- Обновить API interceptor: автоматический refresh при 401

#### 1B.7 — API Interceptor (обновить)

```
frontend/src/lib/api.ts  — обновить
```

- Request: добавить access token
- Response 401: попытка refresh → retry original request → если нет → logout + redirect

**Проверка 1B**:
```bash
npm run dev
# Открыть localhost:5173
# → Redirect на /auth/login
# Заполнить форму → регистрация → redirect на /dashboard
# Refresh страницу → остаёмся на dashboard (token persistence)
```

---

## Step 2: Три параллельных модуля

> После завершения Auth — эти три модуля **независимы** и могут разрабатываться параллельно.

---

### Step 2A: Telegram-бот

**Зависимости**: Step 1 (auth), чтобы admin мог создавать ботов
**Сложность**: высокая (hot-reload, multi-bot management)

#### 2A.1 — Миграции

```
backend/migrations/00003_bots.sql
```

Таблицы:
- `bots` (id, org_id, name, token_encrypted, username, status, settings JSONB, created_at, updated_at)
  - status: `active`, `inactive`, `error`
  - settings: `{modules: [...], buttons: [...], registration_form: {...}}`
- `bot_clients` (id, bot_id, telegram_id, username, first_name, last_name, phone, data JSONB, registered_at)
  - UNIQUE(bot_id, telegram_id)

#### 2A.2 — Entity

```
backend/internal/entity/bot.go
backend/internal/entity/bot_client.go
```

Bot:
- ID, OrgID, Name, Token, Username, Status, Settings (typed JSONB struct), CreatedAt, UpdatedAt
- Settings: `BotSettings{Modules []string, Buttons []BotButton, RegistrationForm []FormField}`

BotClient:
- ID, BotID, TelegramID, Username, FirstName, LastName, Phone, Data, RegisteredAt

#### 2A.3 — Repository: Bots

```
backend/internal/repository/postgres/bots.go
backend/internal/repository/postgres/bot_clients.go
```

Bots:
- `Create(ctx, bot) (Bot, error)`
- `GetByID(ctx, id) (Bot, error)`
- `GetByOrgID(ctx, orgID) ([]Bot, error)`
- `Update(ctx, bot) error`
- `Delete(ctx, id) error`
- `GetAllActive(ctx) ([]Bot, error)` — для загрузки при старте bot-сервиса

BotClients:
- `Create(ctx, client) (BotClient, error)`
- `GetByTelegramID(ctx, botID, telegramID) (BotClient, error)`
- `GetByBotID(ctx, botID, pagination) ([]BotClient, int, error)`

#### 2A.4 — Usecase: Bots (Admin)

```
backend/internal/usecase/bots/bots.go
```

CRUD операции + валидация:
- `Create(ctx, orgID, req CreateBotRequest) (Bot, error)` — проверить token через Telegram API
- `GetByOrgID(ctx, orgID) ([]Bot, error)`
- `Update(ctx, id, orgID, req UpdateBotRequest) (Bot, error)`
- `Delete(ctx, id, orgID) error`
- `GetSettings(ctx, id, orgID) (BotSettings, error)`
- `UpdateSettings(ctx, id, orgID, settings) error`

#### 2A.5 — Controller: Bots Group (Admin API)

```
backend/internal/controller/http/group/bots/bots.go
```

- `Path()` → `"/bots"`
- `Auth()` → JWT middleware
- Handlers:
  - `POST /` → create bot
  - `GET /` → list bots
  - `GET /:id` → get bot
  - `PATCH /:id` → update bot
  - `DELETE /:id` → delete bot
  - `GET /:id/settings` → get settings
  - `PATCH /:id/settings` → update settings

Важно: все handlers извлекают `org_id` из JWT user → проверяют принадлежность.

#### 2A.6 — Bot Service (Multi-Bot Manager)

```
backend/internal/service/botmanager/manager.go
backend/internal/service/botmanager/handler.go
```

BotManager:
- `Start(ctx) error` — загрузить все active боты из БД, запустить каждый
- `AddBot(bot Bot) error` — hot-reload: запустить нового бота
- `RemoveBot(botID) error` — остановить бота
- `ReloadBot(bot Bot) error` — перезапустить (stop + start)
- `Shutdown() error` — остановить все

Каждый бот запускается в отдельной горутине с telego long polling.

Handler (обработчик для одного бота):
- `/start` → приветствие + анкета регистрации (если не зарегистрирован)
- Главное меню → кнопки из настроек
- Баланс → показать бонусы

#### 2A.7 — Интеграция BotManager с Admin API

Когда admin создает/обновляет/удаляет бота через API:
- `POST /bots` → usecase создает в БД → notifies BotManager.AddBot()
- `PATCH /bots/:id` → usecase обновляет → notifies BotManager.ReloadBot()
- `DELETE /bots/:id` → usecase удаляет → notifies BotManager.RemoveBot()

Паттерн: usecase вызывает callback/channel, не знает о BotManager напрямую.

#### 2A.8 — Wiring в cmd/bot/main.go

Полностью переписать `cmd/bot/main.go`:
- Загрузить конфигурацию
- Подключить PostgreSQL (для чтения ботов и клиентов)
- Создать BotManager
- Запустить HTTP-сервер для healthcheck + webhook для hot-reload
- Graceful shutdown

#### 2A.9 — Frontend: Боты (Admin Panel)

```
frontend/src/features/bots/api.ts
frontend/src/features/bots/types.ts
frontend/src/features/bots/queries.ts
frontend/src/routes/dashboard/bots/index.tsx       — список ботов
frontend/src/routes/dashboard/bots/$botId.tsx       — настройка бота
frontend/src/components/bots/BotCard.tsx            — карточка бота
frontend/src/components/bots/CreateBotModal.tsx     — модал создания
frontend/src/components/bots/BotSettingsForm.tsx    — форма настроек
```

Страницы:
- **Список ботов** (Figma node `26:304`):
  - Empty state (Figma node `21:9`) → кнопка "Создать бота"
  - Карточки ботов: имя, username, статус (active/inactive), кол-во клиентов
- **Модал создания** (Figma node `27:476`):
  - Поля: имя бота, токен (от @BotFather)
  - Привязка к точкам продаж (если есть)
- **Настройка бота** (Figma node `22:110`):
  - Модули: toggle лояльности, бронирования, меню, обратной связи
  - Кнопки главного меню: drag-n-drop порядок
  - Анкета регистрации: настраиваемые поля

**Проверка 2A**:
```bash
# Backend
curl -X POST localhost:8080/api/v1/bots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Bot","token":"123456:ABC..."}'
# → 201 {id, name, username, status}

# Бот отвечает в Telegram на /start
# Создание через Admin Panel → бот появляется и работает
```

---

### Step 2B: Программа лояльности

**Зависимости**: Step 1 (auth)
**Сложность**: средняя

#### 2B.1 — Миграции

```
backend/migrations/00004_loyalty.sql
```

Таблицы:
- `loyalty_programs` (id, org_id, name, type ENUM('bonus','discount'), config JSONB, is_active, created_at, updated_at)
  - config: `{welcome_bonus: 100, currency_name: "баллы", ...}`
- `loyalty_levels` (id, program_id, name, threshold INT, reward_percent DECIMAL, sort_order INT)
- `client_loyalty` (id, client_id, program_id, level_id, balance DECIMAL, total_earned DECIMAL, total_spent DECIMAL, updated_at)
  - UNIQUE(client_id, program_id)
- `loyalty_transactions` (id, client_id, program_id, type ENUM('earn','spend','adjust'), amount DECIMAL, balance_after DECIMAL, description TEXT, created_by INT, created_at)

#### 2B.2 — Entity

```
backend/internal/entity/loyalty.go
backend/internal/entity/transaction.go
```

LoyaltyProgram, LoyaltyLevel, ClientLoyalty, LoyaltyTransaction

#### 2B.3 — Repository: Loyalty

```
backend/internal/repository/postgres/loyalty.go
backend/internal/repository/postgres/transactions.go
```

Programs:
- `CreateProgram(ctx, program) (LoyaltyProgram, error)`
- `GetProgramsByOrgID(ctx, orgID) ([]LoyaltyProgram, error)`
- `GetProgramByID(ctx, id) (LoyaltyProgram, error)`
- `UpdateProgram(ctx, program) error`

Levels:
- `CreateLevel(ctx, level) (LoyaltyLevel, error)`
- `GetLevelsByProgramID(ctx, programID) ([]LoyaltyLevel, error)`
- `UpdateLevel(ctx, level) error`
- `DeleteLevel(ctx, id) error`

ClientLoyalty:
- `GetClientLoyalty(ctx, clientID, programID) (ClientLoyalty, error)`
- `UpsertClientLoyalty(ctx, cl) error`

Transactions:
- `CreateTransaction(ctx, tx) (LoyaltyTransaction, error)`
- `GetTransactions(ctx, clientID, programID, pagination) ([]LoyaltyTransaction, int, error)`

#### 2B.4 — Usecase: Loyalty

```
backend/internal/usecase/loyalty/loyalty.go
```

Programs & Levels:
- `CreateProgram(ctx, orgID, req) (LoyaltyProgram, error)`
- `GetPrograms(ctx, orgID) ([]LoyaltyProgram, error)`
- `UpdateProgram(ctx, id, orgID, req) error`
- `CreateLevel(ctx, programID, orgID, req) (LoyaltyLevel, error)`
- `UpdateLevels(ctx, programID, orgID, levels) error`

Операции с баллами:
- `EarnPoints(ctx, clientID, programID, amount, description) (ClientLoyalty, error)`
  1. Получить текущий баланс
  2. Добавить баллы
  3. Проверить переход на новый уровень
  4. Создать транзакцию
  5. Обновить баланс
  6. Вернуть обновленный ClientLoyalty

- `SpendPoints(ctx, clientID, programID, amount, description) (ClientLoyalty, error)`
  1. Проверить достаточность баллов
  2. Списать
  3. Создать транзакцию

- `GetBalance(ctx, clientID, programID) (ClientLoyalty, error)`

#### 2B.5 — Controller: Loyalty Group

```
backend/internal/controller/http/group/loyalty/loyalty.go
```

- `Path()` → `"/loyalty"`
- `Auth()` → JWT middleware
- Handlers:
  - `POST /programs` → create program
  - `GET /programs` → list programs
  - `GET /programs/:id` → get program with levels
  - `PATCH /programs/:id` → update program
  - `POST /programs/:id/levels` → add level
  - `PUT /programs/:id/levels` → batch update levels
  - `DELETE /programs/:id/levels/:levelId` → remove level
  - `POST /programs/:id/earn` → earn points (для будущего POS-интеграции)
  - `POST /programs/:id/spend` → spend points

#### 2B.6 — Frontend: Лояльность (Admin Panel)

```
frontend/src/features/loyalty/api.ts
frontend/src/features/loyalty/types.ts
frontend/src/features/loyalty/queries.ts
frontend/src/routes/dashboard/loyalty/index.tsx        — список программ
frontend/src/routes/dashboard/loyalty/$programId.tsx   — настройка программы
frontend/src/components/loyalty/ProgramCard.tsx         — карточка программы
frontend/src/components/loyalty/CreateProgramModal.tsx  — модал создания
frontend/src/components/loyalty/LevelsTable.tsx         — таблица уровней
```

Страницы:
- **Список программ**: карточки бонусных/дисконтных программ
- **Создание**: тип (bonus/discount), название, настройки
- **Настройка уровней**: таблица (имя, порог, % бонусов), inline-edit, drag-n-drop сортировка

**Проверка 2B**:
```bash
curl -X POST localhost:8080/api/v1/loyalty/programs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Бонусная","type":"bonus","config":{"welcome_bonus":100}}'
# → 201

curl localhost:8080/api/v1/loyalty/programs \
  -H "Authorization: Bearer $TOKEN"
# → 200 [{id, name, type, levels: [...]}]
```

---

### Step 2C: Точки продаж

**Зависимости**: Step 1 (auth)
**Сложность**: низкая (простой CRUD)

#### 2C.1 — Миграция

```
backend/migrations/00005_pos.sql
```

Таблица `pos_locations`:
- id, org_id, name, address, phone, schedule JSONB, is_active, created_at, updated_at
- schedule: `{"mon": {"open": "09:00", "close": "23:00"}, ...}`

#### 2C.2 — Entity

```
backend/internal/entity/pos.go
```

POSLocation с typed Schedule struct.

#### 2C.3 — Repository

```
backend/internal/repository/postgres/pos.go
```

Стандартный CRUD: Create, GetByID, GetByOrgID, Update, Delete.

#### 2C.4 — Usecase

```
backend/internal/usecase/pos/pos.go
```

CRUD с проверкой org ownership.

#### 2C.5 — Controller

```
backend/internal/controller/http/group/pos/pos.go
```

- `Path()` → `"/pos"`
- `Auth()` → JWT middleware
- CRUD: POST, GET, GET/:id, PATCH/:id, DELETE/:id

#### 2C.6 — Frontend: Точки продаж

```
frontend/src/features/pos/api.ts
frontend/src/features/pos/types.ts
frontend/src/features/pos/queries.ts
frontend/src/routes/dashboard/pos/index.tsx
frontend/src/routes/dashboard/pos/$posId.tsx
frontend/src/components/pos/POSCard.tsx
frontend/src/components/pos/CreatePOSModal.tsx
frontend/src/components/pos/ScheduleEditor.tsx
```

Страницы:
- Список точек: карточки/список (переключение вида)
- Создание: название, адрес, телефон, график работы по дням
- Редактирование

**Проверка 2C**:
```bash
curl -X POST localhost:8080/api/v1/pos \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Кафе на Тверской","address":"ул. Тверская, 10","schedule":{"mon":{"open":"09:00","close":"23:00"}}}'
# → 201
```

---

## Step 3: Интеграция

**Зависимости**: Steps 2A + 2B + 2C завершены

### 3.1 — Bot ↔ Loyalty

```
backend/internal/service/botmanager/handler.go — дополнить
```

- При регистрации клиента в боте → создать ClientLoyalty запись
- Кнопка «Баланс» → показать текущий баланс и уровень
- Welcome bonus: при первой регистрации → начислить welcome_bonus баллов
- Уведомления: при начислении/списании → отправить сообщение в бот

### 3.2 — Bot ↔ POS

- Привязка бота к точкам продаж (через bot settings)
- В боте: показать список точек с адресами и графиком

### 3.3 — Admin API: User Context

Убедиться что все CRUD endpoints правильно фильтруют по `org_id` текущего пользователя:
- `GET /bots` → только боты своей организации
- `GET /loyalty/programs` → только свои программы
- `GET /pos` → только свои точки

---

## Step 4: Dashboard Layout + Polish

**Зависимости**: Steps 2-3 (есть данные для отображения)

### 4.1 — Sidebar (обновить)

```
frontend/src/components/layout/Sidebar.tsx — обновить
```

- Показывать количество ботов, программ, точек в badge
- Контекстное отображение ботов (если один → сразу его настройки)
- Подсветка активного раздела
- Figma node `3:4`

### 4.2 — Responsive

```
frontend/src/components/layout/MobileNav.tsx
```

- Бургер-меню для мобильных устройств (< 768px)
- Drawer с sidebar навигацией
- Header адаптируется

### 4.3 — Empty States

Каждый раздел при отсутствии данных:
- Иллюстрация + текст + CTA кнопка
- Figma empty states для ботов (node `21:9`)

### 4.4 — Loading & Error States

```
frontend/src/components/common/LoadingSkeleton.tsx
frontend/src/components/common/ErrorState.tsx
```

- Skeleton loading для списков и карточек
- Error boundary + retry кнопка

---

## Step 5: Тестирование + Deploy

**Зависимости**: Steps 1-4

### 5.1 — Backend Tests

```
backend/internal/usecase/auth/auth_test.go
backend/internal/usecase/bots/bots_test.go
backend/internal/usecase/loyalty/loyalty_test.go
backend/internal/usecase/pos/pos_test.go
```

Минимум: unit-тесты для usecases с mock repositories.

### 5.2 — Frontend Lint + Build

```bash
cd frontend && npm run lint && npm run build
```

Убедиться что production build проходит без ошибок.

### 5.3 — Миграции на продакшене

```bash
# Через CI/CD или вручную
scripts/migrate.sh up
```

### 5.4 — Deploy

```bash
git push origin main
# GitHub Actions → build → deploy
```

### 5.5 — Production Verification

- [ ] `elysium.fm/revisitr/` → логин → dashboard
- [ ] Создание бота → бот отвечает в Telegram
- [ ] Создание программы лояльности
- [ ] Создание точки продаж
- [ ] Клиент /start в боте → регистрация → баланс

---

## Граф зависимостей

```
Step 1A (Auth Backend) ─────────────────────┐
Step 1B (Auth Frontend) ←── depends on 1A   │
                                             │
Step 2A (Bots)     ←── depends on Step 1 ───┤
Step 2B (Loyalty)  ←── depends on Step 1 ───┼──▶ Step 3 (Integration) ──▶ Step 4 (Polish) ──▶ Step 5 (Deploy)
Step 2C (POS)      ←── depends on Step 1 ───┘

Параллелизация:
- 1A + 1B.1-1B.2 (UI без backend) — частичная параллелизация
- 2A + 2B + 2C — полная параллелизация
- 2A.9 + 2B.6 + 2C.6 (frontend) — полная параллелизация
```

## Оптимальная стратегия с Claude Code

### Порядок реализации

1. **Step 1A** → `backend-architect` agent → Auth backend (миграции → entity → repo → usecase → controller → wiring)
2. **Step 1B** → `frontend-architect` agent → Auth frontend (store → API → pages → interceptor)
3. **Step 2A + 2B + 2C** → три параллельных агента:
   - `backend-architect` → Bots backend + Bot service
   - `backend-architect` → Loyalty + POS backend
   - `frontend-architect` → Bots + Loyalty + POS admin pages
4. **Step 3** → `backend-architect` → Integration (Bot ↔ Loyalty ↔ POS)
5. **Step 4** → `frontend-architect` → Dashboard polish + responsive
6. **Step 5** → `quality-engineer` → Tests + `devops-architect` → Deploy

### Файлы, которые будут созданы/изменены

```
# Новые миграции (4 файла)
backend/migrations/00002_auth_sessions.sql
backend/migrations/00003_bots.sql
backend/migrations/00004_loyalty.sql
backend/migrations/00005_pos.sql

# Entity (6 файлов: 2 изменены, 4 новых)
backend/internal/entity/auth.go           NEW
backend/internal/entity/bot.go            NEW
backend/internal/entity/bot_client.go     NEW
backend/internal/entity/loyalty.go        NEW
backend/internal/entity/transaction.go    NEW
backend/internal/entity/pos.go            NEW

# Repository (7 файлов)
backend/internal/repository/postgres/users.go        NEW
backend/internal/repository/postgres/bots.go         NEW
backend/internal/repository/postgres/bot_clients.go  NEW
backend/internal/repository/postgres/loyalty.go      NEW
backend/internal/repository/postgres/transactions.go NEW
backend/internal/repository/postgres/pos.go          NEW
backend/internal/repository/redis/sessions.go        NEW

# Usecase (4 файла)
backend/internal/usecase/auth/auth.go      NEW
backend/internal/usecase/bots/bots.go      NEW
backend/internal/usecase/loyalty/loyalty.go NEW
backend/internal/usecase/pos/pos.go        NEW

# Controller (4 файла)
backend/internal/controller/http/group/auth/auth.go       NEW
backend/internal/controller/http/group/bots/bots.go       NEW
backend/internal/controller/http/group/loyalty/loyalty.go  NEW
backend/internal/controller/http/group/pos/pos.go          NEW

# Bot service (2 файла)
backend/internal/service/botmanager/manager.go   NEW
backend/internal/service/botmanager/handler.go   NEW

# Entrypoints (2 файла изменены)
backend/cmd/server/main.go   MODIFIED
backend/cmd/bot/main.go      MODIFIED

# Frontend Auth (5 файлов)
frontend/src/stores/auth.ts                  NEW
frontend/src/features/auth/api.ts            NEW
frontend/src/features/auth/types.ts          NEW
frontend/src/routes/auth/register.tsx        NEW
frontend/src/routes/auth/verify.tsx          NEW
frontend/src/routes/auth/login.tsx           MODIFIED
frontend/src/routes/dashboard/route.tsx      MODIFIED
frontend/src/lib/api.ts                      MODIFIED

# Frontend Bots (6 файлов)
frontend/src/features/bots/api.ts            NEW
frontend/src/features/bots/types.ts          NEW
frontend/src/features/bots/queries.ts        NEW
frontend/src/routes/dashboard/bots/index.tsx     NEW
frontend/src/routes/dashboard/bots/$botId.tsx    NEW
frontend/src/components/bots/BotCard.tsx          NEW
frontend/src/components/bots/CreateBotModal.tsx   NEW
frontend/src/components/bots/BotSettingsForm.tsx  NEW

# Frontend Loyalty (6 файлов)
frontend/src/features/loyalty/api.ts             NEW
frontend/src/features/loyalty/types.ts           NEW
frontend/src/features/loyalty/queries.ts         NEW
frontend/src/routes/dashboard/loyalty/index.tsx      NEW
frontend/src/routes/dashboard/loyalty/$programId.tsx NEW
frontend/src/components/loyalty/ProgramCard.tsx       NEW
frontend/src/components/loyalty/CreateProgramModal.tsx NEW
frontend/src/components/loyalty/LevelsTable.tsx       NEW

# Frontend POS (5 файлов)
frontend/src/features/pos/api.ts             NEW
frontend/src/features/pos/types.ts           NEW
frontend/src/features/pos/queries.ts         NEW
frontend/src/routes/dashboard/pos/index.tsx      NEW
frontend/src/routes/dashboard/pos/$posId.tsx     NEW
frontend/src/components/pos/POSCard.tsx           NEW
frontend/src/components/pos/CreatePOSModal.tsx    NEW
frontend/src/components/pos/ScheduleEditor.tsx    NEW

# Frontend Common (4 файла)
frontend/src/components/common/LoadingSkeleton.tsx  NEW
frontend/src/components/common/ErrorState.tsx       NEW
frontend/src/components/layout/MobileNav.tsx        NEW
frontend/src/components/layout/Sidebar.tsx          MODIFIED

# Tests (4 файла)
backend/internal/usecase/auth/auth_test.go      NEW
backend/internal/usecase/bots/bots_test.go      NEW
backend/internal/usecase/loyalty/loyalty_test.go NEW
backend/internal/usecase/pos/pos_test.go        NEW

# Итого: ~60 новых файлов, ~6 изменённых
```

---

## Критерии завершения Phase 1

- [ ] Владелец регистрируется (email + пароль) → попадает в dashboard
- [ ] Владелец логинится → видит свою организацию
- [ ] JWT access + refresh tokens работают
- [ ] Владелец создает Telegram-бота через админку
- [ ] Бот появляется в Telegram и отвечает на /start
- [ ] Клиент регистрируется в боте (анкета)
- [ ] Владелец создает бонусную программу лояльности
- [ ] Владелец настраивает уровни (порог + % бонусов)
- [ ] Клиент видит баланс бонусов в боте
- [ ] Владелец добавляет точки продаж
- [ ] Sidebar и header работают как в Figma
- [ ] Responsive layout (мобильная версия)
- [ ] Push → CI/CD → deploy → production работает
