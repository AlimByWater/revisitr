# E2E Tests — План реализации

Поэтапный план создания E2E тестов (Playwright) для Revisitr.

---

## Фаза 0: Инфраструктура

**Цель**: рабочий Playwright с одним проходящим smoke-тестом.

### Задачи

1. **Инициализация проекта**
   - `e2e/` директория с `package.json`, `playwright.config.ts`, `tsconfig.json`
   - Зависимости: `@playwright/test`, `playwright`
   - Скрипты: `test`, `test:headed`, `test:debug`

2. **Playwright config**
   ```
   baseURL: http://localhost:5173/revisitr
   testDir: ./tests
   fullyParallel: false (sequential — тесты зависят от состояния)
   retries: 1
   timeout: 30_000
   projects: [{ name: 'chromium' }]
   webServer: нет (backend + frontend запускаются отдельно)
   ```

3. **API-хелпер** (`e2e/helpers/api.ts`)
   - `apiClient(baseURL)` — HTTP-клиент для прямых API-вызовов
   - `register(email, password, name, org)` → `AuthResponse`
   - `login(email, password)` → `AuthResponse`
   - `upgradeToProPlan(token)` → обновление подписки до Pro
   - `createBot(token, name)` → `Bot`
   - `createLoyaltyProgram(token, ...)` → `Program`
   - `createCampaign(token, ...)` → `Campaign`
   - `createPromotion(token, ...)` → `Promotion`

4. **Auth fixture** (`e2e/helpers/auth.ts`)
   - `globalSetup`: регистрация тестового пользователя, апгрейд до Pro, сохранение `storageState`
   - `globalTeardown`: очистка (опционально)
   - Переиспользование `storageState` во всех тестах (один логин на весь запуск)

5. **Smoke-тест** (`e2e/tests/smoke.spec.ts`)
   - Открыть `/dashboard` → проверить что загрузилось без ошибок
   - Валидация: sidebar видна, метрики отрендерены или показано пустое состояние

### Результат фазы
- `npx playwright test` проходит с 1 зелёным тестом
- Авторизация через API + `storageState` работает
- Pro-подписка активна — feature gates не блокируют

---

## Фаза 1: Auth + Navigation Smoke

**Цель**: покрытие аутентификации и навигации по всем страницам.

### Задачи

1. **Auth flow** (`e2e/tests/auth.spec.ts`)
   - Логин с валидными креденшалами → редирект на `/dashboard`
   - Логин с невалидным паролем → ошибка на форме
   - Регистрация нового пользователя → редирект на onboarding или dashboard
   - Logout → редирект на `/auth/login`
   - Доступ к `/dashboard` без токена → редирект на `/auth/login`

2. **Navigation smoke** (`e2e/tests/navigation.spec.ts`)
   - Проход по ВСЕМ страницам sidebar → каждая рендерится без ошибки
   - Список страниц (30+ маршрутов):
     ```
     /dashboard
     /dashboard/bots
     /dashboard/clients
     /dashboard/clients/segments
     /dashboard/clients/custom-segments
     /dashboard/clients/predictions
     /dashboard/loyalty
     /dashboard/loyalty/wallet
     /dashboard/campaigns
     /dashboard/campaigns/create
     /dashboard/campaigns/templates
     /dashboard/campaigns/scenarios
     /dashboard/promotions
     /dashboard/promotions/codes
     /dashboard/promotions/archive
     /dashboard/rfm
     /dashboard/rfm/onboarding
     /dashboard/rfm/template
     /dashboard/analytics/sales
     /dashboard/analytics/loyalty
     /dashboard/analytics/mailings
     /dashboard/pos
     /dashboard/menus
     /dashboard/integrations
     /dashboard/marketplace
     /dashboard/billing
     /dashboard/billing/invoices
     /dashboard/account
     /dashboard/onboarding
     ```
   - Для каждой страницы: нет `console.error`, нет "Ошибка загрузки", рендер за <5s
   - **Frontend-only страницы** (account, custom-segments): допускаем 404/ошибку API, но UI-каркас должен рендериться

3. **Sidebar navigation** (`e2e/tests/sidebar.spec.ts`)
   - Клик по каждому пункту sidebar → URL меняется корректно
   - Submenu открывается/закрывается
   - Активный пункт подсвечивается

### Результат фазы
- ~15-20 тестов
- Уверенность что все страницы рендерятся, навигация работает
- Обнаружены страницы с ошибками загрузки (feature gates, отсутствующие endpoints)

---

## Фаза 2: Seed Data + CRUD Flows

**Цель**: тесты на создание/редактирование/удаление основных сущностей.

### Seed Data Strategy

Все данные создаются через API в `globalSetup` или в `beforeAll` тестовых файлов:

```
1. Register user → access_token
2. Upgrade to Pro → feature gates разблокированы
3. Create bot "Test Bot" → bot_id
4. Create loyalty program "Бонусная" → program_id
5. Create POS location "Кафе Центр" → pos_id
6. Complete onboarding → onboarding done
```

### Задачи

1. **Bot CRUD** (`e2e/tests/bots.spec.ts`)
   - Список ботов: таблица/карточки видны
   - Создание бота: модал → заполнение → сохранение → появление в списке
   - Редактирование бота: переход → изменение имени → сохранение
   - Удаление бота: подтверждение → исчезновение из списка

2. **Loyalty CRUD** (`e2e/tests/loyalty.spec.ts`)
   - Список программ: карточки
   - Создание программы: модал → заполнение → сохранение
   - Настройка уровней: добавление уровня → имя + порог → сохранение
   - Toggle active: переключение статуса

3. **Promotions CRUD** (`e2e/tests/promotions.spec.ts`)
   - Список акций: таблица
   - Создание акции: тип, скидка, период, лимит → сохранение
   - Редактирование: изменение параметров
   - Удаление: подтверждение → удаление из списка
   - Промокоды: создание, генерация, просмотр

4. **POS CRUD** (`e2e/tests/pos.spec.ts`)
   - Создание точки продаж: название, адрес
   - Список точек
   - Редактирование, удаление

5. **Campaigns CRUD** (`e2e/tests/campaigns.spec.ts`)
   - Список кампаний: таблица со статусами
   - Создание: выбор бота → имя → текст → аудитория → превью → сохранение
   - Редактирование черновика
   - Шаблоны: создание, редактирование, применение
   - Авто-сценарии: список, создание (если UI позволяет)

### Результат фазы
- ~25-35 тестов
- Покрытие основных CRUD-операций
- Seed data создаётся и переиспользуется через API

---

## Фаза 3: RFM + Аналитика

**Цель**: покрытие RFM-модуля и аналитических страниц.

### Предусловия
- Pro-подписка активна (из globalSetup)
- Бот и клиенты с транзакциями нужны для RFM-данных
  - Если нет данных — тестируем пустые состояния и UI-элементы

### Задачи

1. **RFM Onboarding** (`e2e/tests/rfm-onboarding.spec.ts`)
   - 3-шаговый квиз: ответы на вопросы → рекомендация шаблона
   - Принять рекомендацию → переход на dashboard
   - Выбрать другой → переход на template page
   - Кастомные пороги → валидация → сохранение

2. **RFM Dashboard** (`e2e/tests/rfm-dashboard.spec.ts`)
   - Таблица 7 сегментов: иконки, count, %, avg check, total
   - Информация о текущем шаблоне
   - Кнопка "Пересчитать" → запуск → обратная связь
   - Переход к детализации сегмента

3. **RFM Template** (`e2e/tests/rfm-template.spec.ts`)
   - 4 стандартных шаблона: выбор → подтверждение двойным кликом → сохранение
   - Кастомный шаблон: ввод порогов (4×2) → валидация → сохранение
   - Текущий шаблон отмечен

4. **RFM Segment Detail** (`e2e/tests/rfm-segment.spec.ts`)
   - Таблица клиентов сегмента: 7 колонок, сортировка
   - ScoreBadge (1-5 цветовое кодирование)
   - Пагинация
   - CSV export → скачивание файла

5. **Analytics pages** (`e2e/tests/analytics.spec.ts`)
   - Продажи: графики/метрики рендерятся (или пустое состояние)
   - Лояльность: аналогично
   - Рассылки: аналогично
   - Переключение периодов

### Результат фазы
- ~15-20 тестов
- Полное покрытие RFM-модуля
- Аналитика проверена на рендеринг

---

## Фаза 4: Onboarding + Account + Billing

**Цель**: покрытие вспомогательных страниц и user journeys.

### Задачи

1. **Onboarding wizard** (`e2e/tests/onboarding.spec.ts`)
   - Прогресс через все шаги (4 шага)
   - Disabled "Далее" на action steps (пока не создана сущность)
   - Completion → редирект на dashboard
   - Onboarding progress в sidebar

2. **Account settings** (`e2e/tests/account.spec.ts`)
   - **Frontend-only**: backend endpoints ещё не реализованы
   - Рендеринг страницы: профиль, безопасность, реквизиты
   - Переключение типа реквизитов (Самозанятый → ИП → ООО)
   - Формы заполняются без JS-ошибок
   - Header dropdown: Settings + Logout

3. **Billing** (`e2e/tests/billing.spec.ts`)
   - Страница тарифов: карточки тарифов видны
   - Текущая подписка: статус, дата продления
   - Переключение тарифа (если API позволяет)
   - Счета: список инвойсов

4. **Clients** (`e2e/tests/clients.spec.ts`)
   - Таблица клиентов: колонки, сортировка, поиск
   - Карточка клиента: профиль, транзакции, заказы
   - Теги: добавление/удаление
   - Custom segments: filter builder (16 атрибутов), превью, создание (frontend-only)

### Результат фазы
- ~15-20 тестов
- Frontend-only страницы проверены на рендеринг и UI-логику
- User journeys покрыты end-to-end

---

## Фаза 5: User Journeys (E2E Scenarios)

**Цель**: сквозные сценарии, имитирующие реальные пользовательские потоки.

### Задачи

1. **Journey: Новый пользователь** (`e2e/tests/journeys/new-user.spec.ts`)
   ```
   Регистрация → Онбординг → Создание бота → Настройка лояльности → Dashboard
   ```

2. **Journey: Запуск кампании** (`e2e/tests/journeys/campaign-launch.spec.ts`)
   ```
   Dashboard → Создать кампанию → Выбрать бот → Написать текст →
   Выбрать аудиторию → Превью → Сохранить (draft)
   ```

3. **Journey: RFM настройка** (`e2e/tests/journeys/rfm-setup.spec.ts`)
   ```
   RFM Onboarding → Выбор шаблона → Dashboard → Детализация сегмента → CSV export
   ```

4. **Journey: Управление акциями** (`e2e/tests/journeys/promotions-flow.spec.ts`)
   ```
   Создать акцию → Создать промокод → Просмотр → Архивация
   ```

5. **Journey: Account management** (`e2e/tests/journeys/account-flow.spec.ts`)
   ```
   Настройки → Изменить имя → Биллинг → Просмотр тарифа → Счета → Logout
   ```

### Результат фазы
- ~5-8 тестов (длинные сценарии)
- Максимальное покрытие реальных user flows
- Все ключевые бизнес-процессы проверены

---

## Итого по фазам

| Фаза | Тестов | Покрытие | Зависимости |
|-------|--------|----------|-------------|
| 0: Инфраструктура | 1 | Smoke | - |
| 1: Auth + Navigation | 15-20 | Все страницы рендерятся | Фаза 0 |
| 2: CRUD Flows | 25-35 | Основные сущности | Фаза 1 |
| 3: RFM + Аналитика | 15-20 | RFM-модуль, графики | Фаза 2 |
| 4: Onboarding + Account + Billing | 15-20 | Вспомогательные страницы | Фаза 1 |
| 5: User Journeys | 5-8 | Сквозные сценарии | Фаза 2-4 |
| **ИТОГО** | **~75-105** | **Максимальное** | - |

---

## Структура файлов

```
e2e/
├── package.json
├── playwright.config.ts
├── tsconfig.json
├── helpers/
│   ├── api.ts              # HTTP-клиент для API-вызовов
│   ├── auth.ts             # globalSetup/teardown, storageState
│   ├── fixtures.ts         # Playwright test fixtures (page + API)
│   └── selectors.ts        # CSS/data-testid селекторы
├── tests/
│   ├── smoke.spec.ts
│   ├── auth.spec.ts
│   ├── navigation.spec.ts
│   ├── sidebar.spec.ts
│   ├── bots.spec.ts
│   ├── loyalty.spec.ts
│   ├── promotions.spec.ts
│   ├── pos.spec.ts
│   ├── campaigns.spec.ts
│   ├── rfm-onboarding.spec.ts
│   ├── rfm-dashboard.spec.ts
│   ├── rfm-template.spec.ts
│   ├── rfm-segment.spec.ts
│   ├── analytics.spec.ts
│   ├── onboarding.spec.ts
│   ├── account.spec.ts
│   ├── billing.spec.ts
│   ├── clients.spec.ts
│   └── journeys/
│       ├── new-user.spec.ts
│       ├── campaign-launch.spec.ts
│       ├── rfm-setup.spec.ts
│       ├── promotions-flow.spec.ts
│       └── account-flow.spec.ts
└── .auth/
    └── user.json           # storageState (gitignored)
```

---

## Запуск

```bash
# Предусловия
cd infra && docker compose up -d          # PostgreSQL + Redis
cd backend && go run ./cmd/server &       # API server на :8080
cd frontend && npm run dev &              # Frontend на :5173

# Запуск тестов
cd e2e && npx playwright test             # Все тесты
cd e2e && npx playwright test auth        # Только auth
cd e2e && npx playwright test --headed    # С браузером
cd e2e && npx playwright test --debug     # Debug mode

# CI
make test-e2e                             # Через Makefile
```

---

## Открытые вопросы

1. **Backend endpoints для account/custom segments** — когда будут реализованы? До тех пор тестируем только frontend rendering.
2. **Тестовые клиенты с транзакциями** — нужен ли API для seed клиентов? Сейчас клиенты создаются через Telegram-бот, прямого API "create client" может не быть.
3. **CI интеграция** — добавлять E2E в GitHub Actions? Нужен full stack в CI (postgres + redis + backend + frontend).
4. **Параллелизация** — начинаем sequential, переходим на parallel после стабилизации?
