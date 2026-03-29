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

## 4. Интерактивные блоки (фильтры, дропдауны, карточки)

| Состояние | Стиль |
|-----------|-------|
| Default | `border border-neutral-900 bg-white text-neutral-900` |
| Hover | `bg-neutral-50` |
| Active / Selected | `bg-neutral-900 text-white` (инверсия) |
| Disabled | `bg-neutral-300 cursor-not-allowed` |

## 5. Поля ввода

| Состояние | Стиль |
|-----------|-------|
| Default (editable) | `border border-neutral-900 bg-white` |
| Focus | `ring-2 ring-neutral-900/10` |
| Disabled (read-only) | `border-neutral-300 bg-neutral-100 text-neutral-500` |

## 6. Hover общие паттерны

| Элемент | Hover-эффект |
|---------|-------------|
| Навигация (sidebar) | `hover:scale-[1.02] hover:text-neutral-900 origin-left` |
| Профиль (header) | `hover:scale-[1.03]` |
| Красные элементы | `hover:text-[#FF5C47]` (светлее) |
| Чёрные элементы | `hover:bg-neutral-700` (чуть светлее) |
| Дропдауны/карточки | `hover:bg-neutral-50` |

## Серые оттенки (палитра)

| Назначение | Цвет |
|------------|------|
| Disabled поля ввода | `bg-neutral-100` (#F5F5F5) |
| Disabled кнопки | `bg-neutral-300` (#D4D4D4) |
| Разделители внутри блоков | `border-neutral-200` (#E5E5E5) |
| Обводки блоков | `border-neutral-900` (#171717) |

## Общие правила

- Переходы: `transition-all duration-150` (стандарт), `duration-300` (кнопки — для вспышки)
- Жирность: заголовки `font-bold` (700), активные подпункты `font-medium` (500), остальное normal (400)
- Фокус (keyboard): `focus:outline-none focus:ring-2 focus:ring-neutral-900/10`
- Скругления: `rounded` (4px) — минимальное, рублёное
- Обводки: 1px `border-neutral-900`
- Все контентные блоки: `bg-white` (без шума)
- Все дропдауны: кастомные (не нативные), `border border-neutral-900 rounded bg-white`

## Фон

- Body: белый + шум (sparse dot noise через SVG background-image)
- Все виджеты/блоки/меню: `bg-white` — перекрывают шум
- Header: прозрачный — шум виден сквозь него
