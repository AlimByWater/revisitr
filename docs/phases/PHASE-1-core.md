# Фаза 1: Ядро

> Цель: минимальный рабочий продукт — клиент может зарегистрироваться через бота, получать бонусы, владелец управляет через админку.

## Статус: ⏳ Не начата

## Зависимости: Фаза 0 завершена

---

## Задачи

### 1.1 Auth (аутентификация владельца)

#### Backend
- [ ] Миграция: таблица `users` (id, email, phone, password_hash, name, role, created_at)
- [ ] Миграция: таблица `organizations` (id, name, owner_id, created_at)
- [ ] Entity: `User`, `Organization`
- [ ] Repository: `users` CRUD, `organizations` CRUD
- [ ] Usecase: `auth` — регистрация, логин (email+пароль), JWT (access + refresh)
- [ ] Controller: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`
- [ ] Middleware: JWT auth — проверка токена, извлечение user_id
- [ ] SMS-верификация (заглушка для MVP — просто отправляем код в лог)

#### Frontend
- [ ] Страница логина (из Figma — node `3:7`)
- [ ] Страница регистрации (из Figma — node `3:73`)
- [ ] SMS-верификация (из Figma — node `27:562`)
- [ ] Auth store (Zustand) — token storage, redirect logic
- [ ] Protected routes — редирект на /auth/login если нет токена
- [ ] API interceptor — автообновление access token

### 1.2 Telegram-бот (базовый)

#### Backend
- [ ] Миграция: таблица `bots` (id, org_id, name, token, username, status, created_at)
- [ ] Миграция: таблица `bot_clients` (id, bot_id, telegram_id, username, firstname, phone, registered_at)
- [ ] Entity: `Bot`, `BotClient`
- [ ] Bot service: подключение к Telegram API (telego), обработка /start
- [ ] Bot: форма регистрации (имя, пол, возраст, телефон — настраиваемая)
- [ ] Bot: главное меню с кнопками (настраиваемые через админку)
- [ ] Bot: показ баланса бонусов
- [ ] Hot-reload: при добавлении/изменении бота через админку — перезапуск бота без рестарта сервиса

#### Админка (Backend API)
- [ ] `POST /bots` — создание бота (name, token, привязка к точкам)
- [ ] `GET /bots` — список ботов организации
- [ ] `PATCH /bots/:id` — обновление настроек
- [ ] `DELETE /bots/:id` — удаление бота
- [ ] `GET /bots/:id/settings` — настройки бота (модули, кнопки, анкета)
- [ ] `PATCH /bots/:id/settings` — обновление настроек

#### Админка (Frontend)
- [ ] Страница «Мои боты» — empty state (из Figma — node `21:9`)
- [ ] Модал создания бота (из Figma — node `27:476`)
- [ ] Список ботов (из Figma — node `26:304`)
- [ ] Настройка бота (из Figma — node `22:110`)
  - [ ] Подключение модулей (лояльность, бронирование, меню, обратная связь)
  - [ ] Настройка кнопок
  - [ ] Настройка анкеты регистрации

### 1.3 Программа лояльности (базовая бонусная)

#### Backend
- [ ] Миграция: `loyalty_programs` (id, org_id, name, type[bonus|discount], config JSONB)
- [ ] Миграция: `loyalty_levels` (id, program_id, name, threshold, reward_percent)
- [ ] Миграция: `client_loyalty` (client_id, program_id, level_id, balance, total_earned, total_spent)
- [ ] Миграция: `transactions` (id, client_id, program_id, type[earn|spend], amount, description, created_at)
- [ ] Entity: `LoyaltyProgram`, `LoyaltyLevel`, `ClientLoyalty`, `Transaction`
- [ ] Usecase: `loyalty` — создание программы, начисление/списание бонусов, смена уровня
- [ ] Controller: CRUD для программ и уровней
- [ ] Bot integration: при транзакции — начисление бонусов через бота

#### Админка (Frontend)
- [ ] Страница «Мои программы» — список карточками
- [ ] Создание программы (бонусная/дисконтная)
- [ ] Настройка уровней (таблица: название, порог, % бонусов/скидки)

### 1.4 Точки продаж

#### Backend
- [ ] Миграция: `pos_locations` (id, org_id, name, address, schedule JSONB)
- [ ] Entity: `POSLocation`
- [ ] Repository + Usecase: CRUD
- [ ] Controller: `GET/POST/PATCH/DELETE /pos`

#### Админка (Frontend)
- [ ] Список точек (карточки/список + переключение вида)
- [ ] Создание точки (название, адрес, график по дням недели)
- [ ] Редактирование

### 1.5 Dashboard Layout

#### Frontend (общий layout)
- [ ] Sidebar с навигацией (из Figma — node `3:4`)
  - [ ] Collapsible, 3 уровня вложенности
  - [ ] Иконки для каждого раздела
  - [ ] Активный пункт подсвечен
- [ ] Header (логотип + навигация + профиль)
- [ ] Контекстное отображение ботов в sidebar
- [ ] Responsive (мобильная версия — бургер-меню)

---

## Критерии завершения

- [ ] Владелец может зарегистрироваться и войти в админку
- [ ] Владелец может создать Telegram-бота и подключить его
- [ ] Бот отвечает на /start, показывает меню, регистрирует клиентов
- [ ] Владелец может создать бонусную программу лояльности с уровнями
- [ ] Владелец может добавить точки продаж
- [ ] Клиент в боте видит свой баланс бонусов
- [ ] Sidebar и header работают как в Figma
