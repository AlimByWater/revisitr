# Фаза 1: Ядро

> Цель: минимальный рабочий продукт — клиент может зарегистрироваться через бота, получать бонусы, владелец управляет через админку.

## Статус: 🔄 В процессе

## Зависимости: Фаза 0 завершена

---

## Задачи

### 1.1 Auth (аутентификация владельца)

#### Backend
- [x] Миграция: таблица `users` (id, email, phone, password_hash, name, role, created_at)
- [x] Миграция: таблица `organizations` (id, name, owner_id, created_at)
- [x] Миграция: таблица `refresh_tokens` (id, user_id, token_hash, expires_at)
- [x] Entity: `User`, `Organization`, `Auth` (TokenPair, RegisterRequest, LoginRequest)
- [x] Repository: `users` CRUD, `organizations` CRUD
- [x] Repository: `sessions` (Redis — refresh token storage)
- [x] Usecase: `auth` — регистрация, логин (email+пароль), JWT (access + refresh)
- [x] Controller: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- [x] Middleware: JWT auth — проверка токена, извлечение user_id, org_id, role
- [ ] SMS-верификация (заглушка для MVP — просто отправляем код в лог)

#### Frontend
- [x] Страница логина (из Figma — node `3:7`)
- [x] Страница регистрации (из Figma — node `3:73`)
- [ ] SMS-верификация (из Figma — node `27:562`)
- [x] Auth store (Zustand) — token storage, redirect logic
- [x] Protected routes — редирект на /auth/login если нет токена
- [x] API interceptor — автообновление access token с queue pattern

### 1.2 Telegram-бот (базовый)

#### Backend
- [x] Миграция: таблица `bots` (id, org_id, name, token, username, status, settings JSONB)
- [x] Миграция: таблица `bot_clients` (id, bot_id, telegram_id, username, firstname, phone)
- [x] Entity: `Bot`, `BotClient`, `BotSettings`, `BotButton`, `FormField`
- [ ] Bot service: подключение к Telegram API (telego), обработка /start
- [ ] Bot: форма регистрации (имя, пол, возраст, телефон — настраиваемая)
- [ ] Bot: главное меню с кнопками (настраиваемые через админку)
- [ ] Bot: показ баланса бонусов
- [ ] Hot-reload: при добавлении/изменении бота через админку — перезапуск бота без рестарта сервиса

#### Админка (Backend API)
- [x] `POST /bots` — создание бота (name, token)
- [x] `GET /bots` — список ботов организации
- [x] `GET /bots/:id` — получение бота
- [x] `PATCH /bots/:id` — обновление настроек
- [x] `DELETE /bots/:id` — удаление бота
- [x] `GET /bots/:id/settings` — настройки бота (модули, кнопки, анкета)
- [x] `PATCH /bots/:id/settings` — обновление настроек

#### Админка (Frontend)
- [x] Страница «Мои боты» — empty state + список карточками
- [x] Модал создания бота
- [x] Список ботов (карточки с именем, username, статус, кол-во клиентов)
- [x] Настройка бота (страница деталей)
  - [x] Подключение модулей (лояльность, бронирование, меню, обратная связь)
  - [x] Настройка кнопок
  - [x] Настройка анкеты регистрации

### 1.3 Программа лояльности (базовая бонусная)

#### Backend
- [x] Миграция: `loyalty_programs` (id, org_id, name, type[bonus|discount], config JSONB)
- [x] Миграция: `loyalty_levels` (id, program_id, name, threshold, reward_percent)
- [x] Миграция: `client_loyalty` (client_id, program_id, level_id, balance, total_earned, total_spent)
- [x] Миграция: `loyalty_transactions` (id, client_id, program_id, type[earn|spend|adjust], amount, balance_after)
- [x] Entity: `LoyaltyProgram`, `LoyaltyLevel`, `ClientLoyalty`, `LoyaltyTransaction`
- [x] Usecase: `loyalty` — создание программы, начисление/списание бонусов, смена уровня
- [x] Controller: CRUD для программ и уровней
- [ ] Bot integration: при транзакции — начисление бонусов через бота

#### Админка (Frontend)
- [x] Страница «Мои программы» — список карточками
- [x] Создание программы (бонусная/дисконтная) — модал
- [x] Настройка уровней (таблица: название, порог, % бонусов/скидки, inline edit)

### 1.4 Точки продаж

#### Backend
- [x] Миграция: `pos_locations` (id, org_id, name, address, phone, schedule JSONB)
- [x] Entity: `POSLocation`, `Schedule`, `DaySchedule`
- [x] Repository + Usecase: CRUD с проверкой org ownership
- [x] Controller: `GET/POST/PATCH/DELETE /pos`

#### Админка (Frontend)
- [x] Список точек (карточки)
- [x] Создание точки (модал: название, адрес, телефон)
- [x] Редактирование (страница деталей с графиком по дням недели)

### 1.5 Dashboard Layout

#### Frontend (общий layout)
- [x] Sidebar с навигацией (из Figma — node `3:4`)
  - [x] Иконки для каждого раздела
  - [x] Активный пункт подсвечен
- [x] Header (логотип + навигация + профиль)
- [ ] Контекстное отображение ботов в sidebar
- [ ] Responsive (мобильная версия — бургер-меню)

---

## Оставшиеся задачи

### Приоритет 1: Bot Service (BotManager)
- [ ] `internal/service/botmanager/manager.go` — multi-bot manager с горутинами
- [ ] `internal/service/botmanager/handler.go` — обработка /start, регистрация, меню, баланс
- [ ] Переписать `cmd/bot/main.go` — подключить BotManager + PostgreSQL + Redis
- [ ] Hot-reload через callback при CRUD операциях с ботами

### Приоритет 2: Интеграция Bot ↔ Loyalty
- [ ] При регистрации клиента в боте → создать ClientLoyalty запись
- [ ] Welcome bonus при первой регистрации
- [ ] Кнопка «Баланс» в боте → текущий баланс и уровень

### Приоритет 3: Polish
- [ ] SMS-верификация (заглушка)
- [ ] Responsive layout (бургер-меню)
- [ ] Empty states с иллюстрациями
- [ ] Loading/Error states

---

## Критерии завершения

- [x] Владелец может зарегистрироваться и войти в админку
- [ ] Владелец может создать Telegram-бота и подключить его
- [ ] Бот отвечает на /start, показывает меню, регистрирует клиентов
- [x] Владелец может создать бонусную программу лояльности с уровнями
- [x] Владелец может добавить точки продаж
- [ ] Клиент в боте видит свой баланс бонусов
- [x] Sidebar и header работают как в Figma
