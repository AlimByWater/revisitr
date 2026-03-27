# E2E Tests — План тестирования

Детальное описание каждого теста, ожидаемого поведения, и матрица покрытия.

---

## Seed Data (через API)

### globalSetup последовательность

```
1. POST /api/v1/auth/register
   → email: "e2e@test.revisitr.local"
   → password: "E2eTestPass123!"
   → name: "E2E Tester"
   → organization: "E2E Test Org"
   → Результат: access_token, refresh_token, user_id, org_id

2. PATCH /api/v1/billing/subscription
   → { "tariff_slug": "pro" }
   → Результат: подписка Pro, все feature gates открыты

3. POST /api/v1/bots
   → { "name": "E2E Bot", "token": "e2e-fake-token-123456" }
   → Результат: bot_id

4. POST /api/v1/loyalty/programs
   → { "name": "E2E Бонусная", "type": "bonus", ... }
   → Результат: program_id

5. POST /api/v1/pos
   → { "name": "E2E Кафе", "address": "ул. Тестовая, 1" }
   → Результат: pos_id

6. PATCH /api/v1/onboarding
   → Mark steps complete

7. Сохранить storageState в e2e/.auth/user.json
   → Все последующие тесты используют сохранённую сессию
```

### Дополнительный seed (в beforeAll конкретных тестов)

- **Campaigns**: создать 2-3 кампании (draft, sent)
- **Promotions**: создать 2 акции (активная, истекшая) + промокод
- **Menus**: создать меню с категориями и позициями
- **RFM config**: установить шаблон через API

---

## Тест-кейсы по модулям

### 1. AUTH (`auth.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| A1 | Login — valid | Ввести email + пароль → Submit | Редирект на `/dashboard` |
| A2 | Login — invalid password | Ввести email + неверный пароль → Submit | Ошибка "Неверный email или пароль" на форме |
| A3 | Login — empty fields | Submit без заполнения | Валидация: обязательные поля |
| A4 | Register — valid | Заполнить все поля → Submit | Редирект на dashboard/onboarding |
| A5 | Register — duplicate email | Email существующего пользователя | Ошибка "Email уже зарегистрирован" |
| A6 | Logout | Header → dropdown → Выйти | Редирект на `/auth/login` |
| A7 | Protected route redirect | Открыть `/dashboard` без токена | Редирект на `/auth/login` |

---

### 2. NAVIGATION SMOKE (`navigation.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| N1 | Dashboard loads | Открыть `/dashboard` | Sidebar + контент рендерятся, нет console.error |
| N2-N30 | Each page renders | Для каждого из ~30 маршрутов: navigate → проверка | Нет "Ошибка загрузки", нет unhandled JS errors |

**Проверки для каждой страницы:**
- `page.waitForLoadState('networkidle')` или `page.waitForSelector` на ключевой элемент
- `page.on('console', ...)` — отловить `console.error`
- Нет текста "Ошибка загрузки" или "Не удалось загрузить"
- Для frontend-only страниц (account, custom-segments): каркас рендерится, API-ошибки допустимы

---

### 3. SIDEBAR (`sidebar.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| S1 | Sidebar renders | Загрузить dashboard | Все пункты меню видны |
| S2 | Submenu toggle | Клик на "Аналитика" | Подменю разворачивается |
| S3 | Active state | Перейти на `/dashboard/bots` | Пункт "Список ботов" подсвечен |
| S4 | Navigation | Клик "Клиенты" → "Клиенты" | URL = `/dashboard/clients` |
| S5 | Deep link active | Открыть `/dashboard/rfm` | "RFM-сегментация" под "Аналитика" активен |

---

### 4. BOTS (`bots.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| B1 | List bots | Открыть `/dashboard/bots` | Таблица/карточки ботов видны, "E2E Bot" в списке |
| B2 | Create bot | Клик "Создать бот" → заполнить имя + токен → Submit | Бот появился в списке |
| B3 | View bot | Клик на бота → переход на `/dashboard/bots/:id` | Детали бота отображаются |
| B4 | Edit bot | Изменить имя → Сохранить | Имя обновлено |
| B5 | Delete bot | Удалить → подтвердить | Бот удалён из списка |
| B6 | Bot settings | Перейти в настройки бота | Настройки рендерятся (welcome message, модули) |

---

### 5. LOYALTY (`loyalty.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| L1 | List programs | Открыть `/dashboard/loyalty` | Карточки программ видны |
| L2 | Create program | Создать → заполнить → сохранить | Программа в списке |
| L3 | View program | Клик → переход | Детали: уровни, правила |
| L4 | Add level | Добавить уровень → имя + порог → сохранить | Уровень отображается |
| L5 | Toggle active | Переключить статус | Статус обновлён |
| L6 | Wallet page | Открыть `/dashboard/loyalty/wallet` | Страница рендерится |

---

### 6. CAMPAIGNS (`campaigns.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| C1 | List campaigns | Открыть `/dashboard/campaigns` | Таблица с колонками: имя, бот, тип, статус |
| C2 | Create campaign | Создать → выбрать бот → имя → текст → сохранить | Кампания в списке (draft) |
| C3 | Preview audience | На форме создания → превью | Показано кол-во получателей |
| C4 | Edit draft | Открыть черновик → изменить текст → сохранить | Текст обновлён |
| C5 | Templates page | Открыть `/dashboard/campaigns/templates` | Список шаблонов |
| C6 | Create template | Создать шаблон → имя + текст → сохранить | Шаблон в списке |
| C7 | Scenarios page | Открыть `/dashboard/campaigns/scenarios` | Список авто-сценариев |

---

### 7. PROMOTIONS (`promotions.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| P1 | List promotions | Открыть `/dashboard/promotions` | Таблица акций |
| P2 | Create promotion | Создать → тип + скидка + период → сохранить | Акция в списке (статус "Активна") |
| P3 | Edit promotion | Изменить скидку → сохранить | Значение обновлено |
| P4 | Delete promotion | Удалить → подтвердить | Удалена из списка |
| P5 | Promo codes page | Открыть `/dashboard/promotions/codes` | Список кодов |
| P6 | Generate code | Создать промокод → заполнить → сохранить | Код в списке |
| P7 | Archive page | Открыть `/dashboard/promotions/archive` | Страница рендерится |

---

### 8. RFM ONBOARDING (`rfm-onboarding.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| R1 | Load onboarding | Открыть `/dashboard/rfm/onboarding` | 3 вопроса отображены |
| R2 | Answer questions | Выбрать ответы на все 3 вопроса | Кнопка рекомендации активна |
| R3 | Get recommendation | Submit → получить рекомендацию | Показан рекомендуемый шаблон |
| R4 | Accept recommendation | Клик "Использовать" | Редирект на RFM dashboard |
| R5 | Choose other | Клик "Выбрать другой" | Редирект на template page |

---

### 9. RFM DASHBOARD (`rfm-dashboard.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| RD1 | Dashboard loads | Открыть `/dashboard/rfm` | Таблица 7 сегментов |
| RD2 | Segment info | Проверить таблицу | Иконки, count, %, avg check, total |
| RD3 | Template info | Посмотреть блок шаблона | Имя текущего шаблона, дата |
| RD4 | Recalculate | Клик "Пересчитать" | Loading → success feedback |
| RD5 | Segment click | Клик на строку сегмента | Переход на `/dashboard/rfm/segments/:segment` |

---

### 10. RFM TEMPLATE (`rfm-template.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| RT1 | Templates list | Открыть `/dashboard/rfm/template` | 4 стандартных шаблона |
| RT2 | Select standard | Двойной клик на шаблон → подтверждение | Шаблон выбран, редирект |
| RT3 | Custom mode | Переключиться на "Кастомный" | Форма с 4×2 полями порогов |
| RT4 | Custom validation | Ввести невалидные пороги | Ошибки валидации |
| RT5 | Custom save | Ввести валидные пороги → сохранить | Success → редирект |
| RT6 | Active marker | Текущий шаблон | Отмечен галочкой |

---

### 11. RFM SEGMENT DETAIL (`rfm-segment.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| RS1 | Segment loads | Открыть `/dashboard/rfm/segments/champions` | Таблица клиентов |
| RS2 | Sort columns | Клик на заголовок колонки | Сортировка меняется |
| RS3 | Score badges | Проверить ScoreBadge | Цветовое кодирование 1-5 |
| RS4 | Pagination | Переключить страницу | Новые данные загружены |
| RS5 | CSV export | Клик "Экспорт CSV" | Файл скачивается |

---

### 12. ANALYTICS (`analytics.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| AN1 | Sales page | Открыть `/dashboard/analytics/sales` | Графики/метрики или пустое состояние |
| AN2 | Loyalty page | Открыть `/dashboard/analytics/loyalty` | Рендеринг без ошибок |
| AN3 | Mailings page | Открыть `/dashboard/analytics/mailings` | Рендеринг без ошибок |
| AN4 | Period filter | Переключить период (7д/30д/90д) | Данные обновляются |

---

### 13. ONBOARDING (`onboarding.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| O1 | Load onboarding | Открыть `/dashboard/onboarding` | Wizard: шаги, progress bar |
| O2 | Step navigation | "Далее" → следующий шаг | Прогресс обновляется |
| O3 | Disabled next | На action-step без сущности | Кнопка "Далее" disabled |
| O4 | Complete | Пройти все шаги → Complete | Редирект на dashboard |
| O5 | Sidebar progress | На дашборде | OnboardingProgress виджет (N/4) |

---

### 14. ACCOUNT (`account.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| AC1 | Page renders | Открыть `/dashboard/account` | Секции: Профиль, Безопасность, Реквизиты |
| AC2 | Profile form | Поля имя, email, телефон видны | Данные пользователя отображены |
| AC3 | Entity type switch | Выбрать "ИП" → "ООО" | Динамические поля меняются |
| AC4 | Password form | Заполнить текущий + новый пароль | Форма принимает ввод, валидация match |
| AC5 | Header dropdown | Клик на аватар | Dropdown: Настройки + Выйти |

**Примечание**: backend endpoints не реализованы — тестируем UI rendering и client-side логику.

---

### 15. BILLING (`billing.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| BI1 | Tariffs page | Открыть `/dashboard/billing` | Карточки тарифов (Trial/Basic/Pro/Enterprise) |
| BI2 | Current plan | Проверить текущий тариф | "Pro" выделен, статус подписки |
| BI3 | Invoices page | Открыть `/dashboard/billing/invoices` | Таблица инвойсов |

---

### 16. CLIENTS (`clients.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| CL1 | List clients | Открыть `/dashboard/clients` | Таблица клиентов |
| CL2 | Search | Ввести в поиск | Фильтрация работает |
| CL3 | Sort | Клик на заголовок | Сортировка меняется |
| CL4 | Pagination | Переключить страницу | Данные обновлены |
| CL5 | Client detail | Клик на клиента | Профиль: данные, транзакции |
| CL6 | Segments page | Открыть `/dashboard/clients/segments` | RFM-сегменты отображены |
| CL7 | Custom segments | Открыть `/dashboard/clients/custom-segments` | Filter builder рендерится |
| CL8 | Create segment | Добавить фильтр → превью | UI работает (backend может вернуть ошибку) |

---

### 17. POS (`pos.spec.ts`)

| # | Тест | Действия | Ожидание |
|---|------|----------|----------|
| PO1 | List POS | Открыть `/dashboard/pos` | Список точек продаж |
| PO2 | Create POS | Создать → имя + адрес → сохранить | Точка в списке |
| PO3 | Edit POS | Изменить имя → сохранить | Обновлено |
| PO4 | Delete POS | Удалить → подтвердить | Удалена |

---

### 18. OTHER PAGES

| # | Тест | Страница | Ожидание |
|---|------|----------|----------|
| M1 | Menus | `/dashboard/menus` | Рендерится |
| M2 | Integrations | `/dashboard/integrations` | Рендерится |
| M3 | Marketplace | `/dashboard/marketplace` | Рендерится |

---

### 19. USER JOURNEYS (`journeys/`)

#### J1: New User Journey (`new-user.spec.ts`)
```
1. POST /auth/register (через API в beforeAll)
2. Login через UI
3. Onboarding: пройти 4 шага
4. Создать бота через UI
5. Создать программу лояльности
6. Вернуться на dashboard
7. Проверить: dashboard показывает бот, нет ошибок
```

#### J2: Campaign Launch (`campaign-launch.spec.ts`)
```
1. Dashboard → Рассылки → Создать
2. Выбрать бот из dropdown
3. Ввести имя + текст сообщения
4. Настроить аудиторию
5. Превью → проверить count
6. Сохранить как черновик
7. Проверить: кампания в списке со статусом "Черновик"
```

#### J3: RFM Setup (`rfm-setup.spec.ts`)
```
1. RFM Onboarding → ответить на вопросы
2. Принять рекомендованный шаблон
3. RFM Dashboard → проверить сегменты
4. Клик на сегмент → детализация
5. CSV export → проверить скачивание
```

#### J4: Promotions Flow (`promotions-flow.spec.ts`)
```
1. Акции → Создать акцию (скидка 10%)
2. Создать промокод для акции
3. Просмотреть акцию в списке
4. Удалить акцию
5. Проверить: акция в архиве или удалена
```

#### J5: Account Flow (`account-flow.spec.ts`)
```
1. Header → Настройки → /dashboard/account
2. Проверить профиль (имя, email)
3. Перейти на Биллинг
4. Проверить текущий тариф
5. Просмотреть счета
6. Logout
7. Проверить: редирект на /auth/login
```

---

## Матрица покрытия

### По фичам

| Модуль | Страниц | Тестов | CRUD | Navigation | Errors |
|--------|---------|--------|------|------------|--------|
| Auth | 2 | 7 | - | ✅ | ✅ |
| Dashboard | 1 | 1 | - | ✅ | ✅ |
| Bots | 2 | 6 | ✅ | ✅ | ✅ |
| Clients | 4 | 8 | Partial | ✅ | ✅ |
| Loyalty | 3 | 6 | ✅ | ✅ | ✅ |
| Campaigns | 4 | 7 | ✅ | ✅ | ✅ |
| Promotions | 3 | 7 | ✅ | ✅ | ✅ |
| RFM | 4 | 16 | ✅ | ✅ | ✅ |
| Analytics | 3 | 4 | - | ✅ | ✅ |
| Onboarding | 1 | 5 | - | ✅ | ✅ |
| Account | 1 | 5 | UI-only | ✅ | ✅ |
| Billing | 2 | 3 | Partial | ✅ | ✅ |
| POS | 2 | 4 | ✅ | ✅ | ✅ |
| Menus | 1 | 1 | - | ✅ | - |
| Integrations | 1 | 1 | - | ✅ | - |
| Marketplace | 1 | 1 | - | ✅ | - |
| Navigation | - | 5 | - | ✅ | - |
| Sidebar | - | 5 | - | ✅ | - |
| Journeys | - | 5 | ✅ | ✅ | ✅ |
| **ИТОГО** | **35** | **~95** | | | |

### По типу проверок

| Тип проверки | Количество | Примеры |
|-------------|-----------|---------|
| Page renders | ~30 | Каждый маршрут загружается |
| CRUD operations | ~25 | Create/Read/Update/Delete сущностей |
| Form validation | ~10 | Обязательные поля, невалидные данные |
| Navigation | ~10 | Sidebar, breadcrumbs, redirects |
| Error handling | ~10 | 403/404/500, пустые состояния |
| User journeys | ~5 | Сквозные бизнес-процессы |
| Data display | ~5 | Таблицы, графики, бейджи |

### По приоритету

| Приоритет | Тесты | Описание |
|-----------|-------|----------|
| P0 (Critical) | Auth, Navigation smoke | Если не работает — всё сломано |
| P1 (High) | Bots, Loyalty, Campaigns CRUD | Core business logic |
| P2 (Medium) | RFM, Promotions, Analytics | Important features |
| P3 (Low) | Account (frontend-only), Menus, Marketplace | Nice to have |

---

## Стратегия для frontend-only страниц

Страницы без backend endpoints (account, custom segments, predictions):

1. **Рендеринг**: проверяем что UI каркас загружается без JS-ошибок
2. **Client-side логика**: переключение форм, валидация полей, dropdown
3. **API ошибки**: ожидаем 404/500 от API, проверяем что UI показывает error state или graceful degradation
4. **Маркер в тесте**: `test.describe('frontend-only')` с комментарием "backend not implemented"

```typescript
test.describe('Account Settings (frontend-only)', () => {
  test('page renders without JS errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', e => errors.push(e.message));
    await page.goto('/dashboard/account');
    // API may return 404, but UI should render
    await expect(page.getByText('Профиль')).toBeVisible();
    expect(errors).toHaveLength(0);
  });
});
```

---

## CI Integration (будущее)

```yaml
# .github/workflows/e2e.yml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e:
    runs-on: self-hosted
    services:
      postgres: ...
      redis: ...
    steps:
      - uses: actions/checkout@v4
      - run: cd backend && go build -o bin/server ./cmd/server
      - run: cd frontend && npm ci && npm run build
      - run: cd backend && ./bin/server &
      - run: cd frontend && npx serve -s dist -l 5173 &
      - run: cd e2e && npm ci && npx playwright install chromium
      - run: cd e2e && npx playwright test
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: e2e/playwright-report/
```
