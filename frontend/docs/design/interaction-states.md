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
| Заголовок (h1) | `font-serif text-3xl font-bold text-neutral-900 tracking-tight` |
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

## Серые оттенки (палитра)

| Назначение | Цвет |
|------------|------|
| Подзаголовки страниц | `font-mono text-xs text-neutral-300 uppercase tracking-wider` |
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

## Шрифты

| Контекст | Шрифт |
|----------|-------|
| Body (default theme) | `Outfit` — будет заменён позже |
| Body (Aurora theme) | `Inter` |
| Mono | `JetBrains Mono` |

**Примечание**: Шрифты будут унифицированы в отдельном проходе.

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
