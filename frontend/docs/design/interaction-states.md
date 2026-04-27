# Система интерактивных состояний Revisitr

## 1. Навигация (sidebar, header links)

| Состояние | Стиль |
|-----------|-------|
| Default | `text-neutral-600` |
| Hover | `text-neutral-900 scale-[1.02] origin-left` (текст темнеет + микро-увеличение) |
| Active (текущая страница) | `font-bold text-neutral-900` |
| Active sub-item | `font-medium text-neutral-900` |
| Inactive sub-item | `text-neutral-400 hover:text-neutral-700` |

## 2. Акцентные ссылки (красные)

| Состояние | Стиль |
|-----------|-------|
| Default | `text-[#EF3219]` |
| Hover | `text-[#FF5C47]` (ярче/светлее) |
| Active | `text-[#EF3219] underline` |

## 3. Кнопки (primary / CTA)

| Состояние | Стиль |
|-----------|-------|
| Default | `bg-neutral-900 text-white` (сплошная чёрная) |
| Hover | `bg-neutral-700` |
| Active / Pressed | `bg-[#EF3219]` (красная вспышка) |
| Release | плавный возврат в чёрный за 300ms |
| Disabled | `bg-neutral-300 text-white cursor-not-allowed` |

## 4. Выпадающие меню (CustomSelect)

Единый компонент `CustomSelect` из `@/components/common/CustomSelect`.

| Правило | Значение |
|---------|----------|
| Позиционирование | `absolute` внутри `relative` родителя (привязан к странице) |
| При скролле | Остаётся на месте (не fixed, не закрывается) |
| Открытие/закрытие | CSS `opacity + scale-y + pointer-events`, НЕ `{open && ...}` |
| Стиль кнопки | `border border-neutral-900 rounded px-4 py-2 text-sm font-medium bg-white` |
| Стиль меню | `border border-neutral-900 rounded bg-white py-1 z-50` |
| Выбранный пункт | `font-semibold text-neutral-900 bg-neutral-50` + галочка `Check` |
| Остальные пункты | `text-neutral-600 hover:bg-neutral-50 hover:text-neutral-900` |
| Группировка | Разделители `border-t border-neutral-200 my-1 mx-3` |
| ChevronDown | `rotate-180` при открытии, `transition-transform duration-200` |
| Закрытие | По клику вне (mousedown на document) |
| Вариант `light` | `border-neutral-200` — для селектов внутри белых контейнеров с чёрной обводкой |
| Вариант `ghost` | Прозрачный по умолчанию, белая карточка при hover/open — для селектов внутри `bg-neutral-50` блоков |

**Запрещено**: нативные `<select>`, портал/createPortal, position:fixed.

## 5. Поля ввода

| Состояние | Стиль |
|-----------|-------|
| Default (editable) | `border border-neutral-200 bg-white rounded px-4 py-2.5 text-sm` |
| Focus | `ring-2 ring-accent/20 border-accent` |
| Disabled (read-only) | `border-neutral-300 bg-neutral-100 text-neutral-500` |

**Правило вложенности**: инпуты, селекты, textarea и вложенные карточки внутри блоков с чёрной обводкой (`border-neutral-900`) используют серую обводку (`border-neutral-200`). Чёрная обводка — только на внешних контейнерах верхнего уровня.

## 6. Метрики / виджеты (StatCard)

| Правило | Значение |
|---------|----------|
| Стиль | `border border-neutral-900 rounded bg-white p-4/p-5` |
| Hover | **Нет**. Виджеты не кликабельные — без hover-эффектов |
| Label | `text-xs font-medium uppercase tracking-wide text-neutral-400` |
| Value | `text-3xl font-bold text-neutral-900 tracking-tight` |

## 7. Заголовки страниц

| Элемент | Стиль |
|---------|-------|
| Заголовок (h1) | `font-display text-3xl font-bold text-neutral-900 tracking-tight` |
| Подзаголовок (p) | `font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1` |

**Запрещено**: `text-neutral-500`, `text-neutral-600`, `text-muted-foreground` для подзаголовков.

## 8. Hover общие паттерны

| Элемент | Hover-эффект |
|---------|-------------|
| Навигация (sidebar) | `hover:scale-[1.02] hover:text-neutral-900` |
| Профиль (header) | `hover:scale-[1.03]` |
| Красные элементы | `hover:text-[#FF5C47]` (светлее) |
| Чёрные кнопки | `hover:bg-neutral-700` (чуть светлее) |
| Дропдауны | `hover:bg-neutral-50` |
| Метрики/виджеты | Без hover (не кликабельные) |
| Кликабельные карточки-плитки | `hover:scale-[1.02] transition-transform duration-150` |
| Аватар/icon-placeholder | Без ховер-эффекта — нейтральный серый `bg-neutral-100` фиксирован |

**Исключение (аватар-плейсхолдер)**: кнопка-иконка бота (`bots/index.tsx`, `w-10 h-10 rounded bg-neutral-100`) — это аватарный placeholder, не обычная иконка. Серый фон фиксирован, НЕ применять `group-hover:bg-accent/10` и т.п.

## Серые оттенки (палитра)

| Назначение | Цвет |
|------------|------|
| Подзаголовки страниц | `font-mono text-xs text-neutral-400 uppercase tracking-wider` |
| Disabled поля ввода | `bg-neutral-100` (#F5F5F5) |
| Disabled кнопки | `bg-neutral-300` (#D4D4D4) |
| Разделители внутри блоков | `border-neutral-200` (#E5E5E5) |
| Обводки блоков | `border-neutral-900` (#171717) |

## 9. Графики и диаграммы

### Правило цвета

| Тип графика | Цвет |
|-------------|------|
| Одноцветный (bar, area, line) | `#EF3219` (акцентный красно-оранжевый) |
| Многоцветный (pie, RFM, сегменты) | Семантическая палитра (см. ниже) |
| Aurora тема | `#8B5CF6` (фиолетовый) |

### Палитра для pie-диаграмм

`['#EF3219', '#171717', '#525252', '#a3a3a3', '#d4d4d4', '#f5f5f5']`

Первый сегмент — акцентный, остальные — градация серого.

### Палитра RFM-сегментов (функциональная)

| Сегмент | Цвет | Hex | Иконка (Lucide) |
|---------|------|-----|-----------------|
| new | emerald | `#10b981` | Sprout |
| promising | blue | `#3b82f6` | TrendingUp |
| regular | violet | `#8b5cf6` | UserCheck |
| vip | amber | `#f59e0b` | Crown |
| rare_valuable | purple | `#a855f7` | Gem |
| churn_risk | orange | `#f97316` | AlertTriangle |
| lost | red | `#ef4444` | UserX |

### Статусная палитра (функциональная)

| Статус | Стиль |
|--------|-------|
| Активен / Успех | `bg-green-50 text-green-600` / `bg-green-500 text-white` |
| Неактивен / Ошибка | `bg-red-50 text-red-600` / `bg-red-500 text-white` |
| Ожидание / Предупреждение | `bg-amber-50 text-amber-600` |
| Нейтральный / Черновик | `bg-neutral-100 text-neutral-500` |

### Скругление баров

| Контекст | Radius |
|----------|--------|
| Вертикальные бары (столбцы) | `radius={[1, 1, 0, 0]}` |
| Горизонтальные бары (воронки) | `radius={[0, 2, 2, 0]}` |

### Tooltip

`contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 13 }}`

## § Шрифты

| Контекст | Класс / Переменная |
|----------|-------------------|
| Заголовки (h1, h2 и крупные акцентные тексты) | `font-display` → CSS var `--font-display` (сейчас Inter) |
| Числа, метрики, mono-подписи | `font-mono` (JetBrains Mono) |
| Body / всё остальное | Inter (default sans, задан в `index.css` через `font-family`) |

**Правила**:
- `font-display` — для всех заголовков страниц вместо устаревшего `font-serif`.
- `font-mono` — для числовых значений, tabular-nums, подписей меток (eyebrow).
- Смена шрифта в будущем делается одним изменением `--font-display` без правки кода.

**Запрещено**: `font-serif` — устарел и заменён на `font-display` глобально.

## Общие правила

- Переходы: `transition-all duration-150` (стандарт), `duration-300` (кнопки — для вспышки)
- Жирность: заголовки `font-bold` (700), активные подпункты `font-medium` (500), остальное normal (400)
- Фокус (keyboard): `focus:outline-none focus:ring-2 focus:ring-neutral-900/10`
- Скругления: `rounded` (4px) — минимальное, рублёное
- Обводки: 1px `border-neutral-900`
- Все контентные блоки: `bg-white` (без шума)
- Все выпадающие меню: только `CustomSelect`, кастомные
- Кнопки «Сохранить»: слева в секции, статус справа

## Фон

- Body: белый + шум (sparse dot noise через SVG background-image)
- Все виджеты/блоки/меню: `bg-white` — перекрывают шум
- Header: прозрачный — шум виден сквозь него
- Footer: `bg-neutral-900` + белый dot noise (инверсия)

## 10. Карточки / контейнеры

| Элемент | Стиль |
|---------|-------|
| Внешний контейнер (карточка, таблица) | `bg-white rounded border border-neutral-900` |
| Внутренние разделители | `border-neutral-200` с отступами `mx-4` или `mx-6` (не на всю ширину) |
| Строки таблицы (разделитель) | `<div className="mx-6 border-t border-neutral-200" />` между строками |
| Вложенные контейнеры (секции внутри карточки) | `border border-neutral-200 rounded` |

**Запрещено**: `shadow-sm` на карточках, `border-surface-border`, `rounded-2xl`, `rounded-lg` на контейнерах.

## 11. Попапы / модалки

- Overlay: `bg-black/30`
- Окно: `bg-white border border-neutral-900 rounded p-6 shadow-lg`
- Через `createPortal` в `document.body`
- Закрытие: по клику на overlay

## 12. Данные / Mock API

Аналитика считается из базы ~4500 транзакций (180 дней), сгенерированных с детерминированным seed.
- Транзакции: `store.transactions` — привязаны к клиентам и ботам
- Клиенты: `store.clients` (360) — регистрации с экспоненциальным распределением (больше новых)
- Все analytics endpoints (`/analytics/sales`, `/analytics/loyalty`, `/dashboard/widgets`) **вычисляют** метрики из транзакций
- Фильтры по периоду и боту реально фильтруют данные
- Числа консистентны: сегодня < неделя < месяц
- Включается через `VITE_MOCK_API=true` в `.env`, отключается удалением переменной

---

## § Уведомления (Notice)

Все notice-баннеры используют оранжевое семейство. Исключение — только success (зелёный).

### Классы по интенту

| Интент | Обёртка | Иконка | Заголовок | Текст |
|--------|---------|--------|-----------|-------|
| info | `bg-orange-50 border border-accent/30 rounded p-4` | `Info` `text-accent` | `text-orange-900 font-semibold` | `text-orange-800/90` |
| important | `bg-orange-50 border border-accent/60 rounded p-4` | `AlertTriangle` `text-accent` | `text-orange-900 font-semibold` | `text-orange-800/90` |
| action | `bg-orange-50 border border-accent/30 rounded p-4` | `AlertCircle` `text-accent` | `text-orange-900 font-semibold` | `text-orange-800/90` |
| tip | `bg-orange-50 border border-accent/30 rounded p-4` | `Lightbulb` `text-accent` | `text-orange-900 font-semibold` | `text-orange-800/90` |
| success | `bg-emerald-50 border border-emerald-500/25 rounded p-4` | `CheckCircle2` `text-emerald-600` | `text-emerald-900 font-semibold` | `text-emerald-800/90` |

**Запрещено**: `bg-amber-50`, `bg-blue-50`, `bg-yellow-50` — использовать только оранжевое семейство (кроме success).

Иконка располагается слева `shrink-0 mt-0.5`, текстовый блок — справа с `min-w-0`.

---

## § Теги и статусы

Два стандарта в зависимости от назначения.

### V1 — Статусные теги (outline mono)

Используется для: active/inactive/pending/error/vip статусов сущностей.

```
font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border
```

| Тон | Классы |
|-----|--------|
| emerald (active) | `bg-emerald-500/10 text-emerald-700 border-emerald-500/30` |
| neutral (inactive) | `bg-neutral-100 text-neutral-600 border-neutral-300` |
| amber (pending) | `bg-amber-500/10 text-amber-700 border-amber-500/30` |
| red (error) | `bg-red-500/10 text-red-700 border-red-500/30` |
| accent (vip/featured) | `bg-accent/10 text-accent border-accent/30` |

### V2 — Категориальные теги (solid filled)

Используется для: RFM-сегментов, категорий, типов.

```
font-mono text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded
```

Заливка с белым текстом или тематическими цветами. Для RFM используется `RFM_SEGMENT_COLORS`.

**Запрещено**: `rounded-full` на статусных тегах — только `rounded`.

---

## § Кнопки

Компонент `Button` из `@/components/common/Button`.

| Вариант | Стиль |
|---------|-------|
| `primary` | `bg-accent text-white hover:bg-accent-hover` — основные CTA |
| `secondary` | `bg-white border border-neutral-200 text-neutral-700 hover:bg-neutral-50` — отмена, CSV, назад |
| `dark` | `bg-neutral-900 text-white hover:bg-neutral-700` — сохранить черновик, публикация, удалить |

| Размер | Padding |
|--------|---------|
| `sm` | `py-1.5 px-3 text-xs` |
| `md` (default) | `px-4 py-2.5 text-sm` |

Props: `variant`, `size`, `leftIcon`, `asChild`. Базовые классы: `inline-flex items-center justify-center gap-2 rounded text-sm font-medium transition-colors`.

**Запрещено**: инлайн-стили `bg-accent text-white` на `<button>` — использовать `<Button variant="primary">`.

---

## § Таблицы

| Правило | Значение |
|---------|----------|
| Выравнивание колонок | `text-center` на всех `<th>` и `<td>` кроме первой колонки с названием (`text-left`) |
| Padding ячеек | `px-4 py-3` |
| Заголовки | `font-mono text-[11px] uppercase tracking-wider text-neutral-400` |
| Строки-разделители | `divide-y divide-neutral-200` на `<tbody>` или inline `<div className="mx-4 border-t border-neutral-200" />` |
| Пагинация | ChevronsLeft / ChevronLeft / page numbers / ChevronRight / ChevronsRight — паттерн из `clients/index.tsx` и `rfm/segments/$segment.tsx` |
| Hover строки | `hover:bg-neutral-50 transition-colors` |
| Строки-ссылки | `cursor-pointer` + `onClick={() => navigate(...)}` |

Референс: `clients/index.tsx`, `rfm/segments/$segment.tsx`.

---

## § Виджеты (StatCard)

Иконки — инлайн `text-neutral-400 w-4 h-4`, **без бокса** (исключение — bot avatar placeholder).

```tsx
<div className="border border-neutral-900 rounded bg-white p-5">
  <div className="flex items-center gap-2 text-neutral-400 mb-3">
    <Icon className="w-4 h-4" />
    <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
  </div>
  <p className="font-mono text-3xl font-bold text-neutral-900 tracking-tight tabular-nums">{value}</p>
</div>
```

**Запрещено**: `bg-amber-50`, `bg-green-50`, `bg-blue-50` вокруг иконок в виджетах.

---

## § Цифровые бейджи

| Контекст | Стиль |
|----------|-------|
| Большие (таблицы, детали) | V1 outline mono: `font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border` |
| Маленькие (sidebar счётчики) | `bg-neutral-100 text-neutral-600 border border-neutral-300 rounded-sm text-[10px] font-mono px-1.5 py-0.5` |

**Запрещено**: `rounded-full` на статусных тегах — только `rounded`.

---

## § Empty states

Иконка inline `text-neutral-400 w-8 h-8`, **без бокса**:

```tsx
<Icon className="w-8 h-8 text-neutral-400 mb-4" />
```

**Исключение**: bot avatar placeholder (`bots/index.tsx`) — `w-10 h-10 rounded bg-neutral-100` остаётся с боксом.

---

## § Spacing

| Правило | Значение |
|---------|----------|
| Между top-level секциями страницы | `space-y-6` |
| Внутри блоков (карточки, формы) | `space-y-4` или `space-y-3` |

**Запрещено**: `space-y-8` между верхнеуровневыми секциями страницы.

---

## § Карточки-плитки (list card pattern)

Стандарт для кликабельных карточек в сетке (боты, POS, лояльность и т.д.).

| Элемент | Стиль |
|---------|-------|
| Внешняя обёртка | `bg-white rounded border border-neutral-900 p-5` или `p-6` |
| Ховер (если Link/кликабельная) | `hover:scale-[1.02] transition-transform duration-150` |
| Аватар-слот (если есть) | `w-10 h-10 rounded bg-neutral-100` с иконкой `text-neutral-500` — **без ховер-реколора** |
| Статусный тег | V1 outline mono, в правом верхнем углу |
| Footer-строка (опционально) | `mt-4 pt-4 border-t border-neutral-200` с mono micro-stats |
| Максимум в строке | 2 карточки на `md+` (если нет явной причины больше) |

**Запрещено**: `shadow-sm`, `rounded-lg`, `rounded-xl` на карточках. Ховер-эффекты на аватарах.

---

## § Подблоки (sub-section внутри карточки)

Светло-серая «вкладка» внутри основного белого блока — для группировок: настроек, привязок, форм.

| Элемент | Стиль |
|---------|-------|
| Обёртка | `rounded border border-neutral-200 bg-neutral-50/70 p-3` или `p-4` |
| Заголовок (mono eyebrow) | `text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 mb-2` |
| Внутренние ряды (item) | `bg-white px-3 py-2.5 rounded` (или то же `border-neutral-200`, если нужно подчеркнуть) |

**Используется в**: меню > привязка к точкам продаж, меню-категории > редактирование, бот-настройки > программа лояльности.

**Запрещено**: пустые белые подблоки внутри белой карточки без эйбрового заголовка (нет иерархии). Серая обводка — обязательна для всех «вложенных групп».

---

## § Пагинация (унифицированная)

Все таблицы используют общий компонент `@/components/common/Pagination`.

| Элемент | Стиль |
|---------|-------|
| Левая часть | `text-sm text-neutral-500` — «Всего N записей» (если задан `total` + `itemsLabel`), иначе «Стр. P из N» |
| Правая часть | 5 кнопок: `‹‹` (первая), `‹` (предыдущая), `P / N` лейбл, `›` (следующая), `››` (последняя) |
| Кнопка | `p-2 rounded text-neutral-500 hover:bg-neutral-100`, disabled = `opacity-40 cursor-not-allowed` |
| Лейбл «P / N» | `px-3 py-2 text-sm text-neutral-700 tabular-nums` |

**Запрещено**: дублировать инлайн-пагинацию (нумерованные кнопки 1/2/3/4/5) — использовать только общий компонент. Несогласованная пагинация запрещена.

---

## § Тосты / алерты (inline notifications)

Цветные блоки уведомлений строятся из палитры. См. демо: `/dashboard/_design`.

| Вариант | bg | border | icon |
|---------|----|--------|------|
| success | `bg-emerald-50` | `border-emerald-500/25` | `text-emerald-600` |
| error | `bg-red-50` | `border-red-500/25` | `text-red-600` |
| warning | `bg-orange-50` | `border-orange-400/50` | `text-orange-500` |
| info | `bg-orange-50` | `border-accent/30` | `text-accent` |
| neutral | `bg-neutral-50/70` | `border-neutral-200` | `text-neutral-500` |

Текст для цветных вариантов: title `text-{color}-900 font-semibold`, body `text-{color}-800/90`.
Для `neutral`: title `text-neutral-900 font-semibold`, body `text-neutral-600`.

**Используется в**:
- success / error / warning / info — `WarningBanner` (бот-страница), inline-уведомления о сохранении, ошибках валидации.
- neutral — пояснительные подблоки (привязка POS на странице меню, контекстные подсказки внутри белых карточек) и места, где цветной тост был бы перегрузом.

**Запрещено**: «уведомления» с белым фоном и `border-neutral-900` (это стиль карточек, не алертов). Все inline-блоки строятся из этих 5 вариантов — не выдумывать новые палитры.
