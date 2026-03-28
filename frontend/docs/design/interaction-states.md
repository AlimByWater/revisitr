# Система интерактивных состояний Revisitr

## 1. Навигация (sidebar, header links)

| Состояние | Стиль |
|-----------|-------|
| Default | `text-neutral-500` |
| Hover | `text-neutral-900` |
| Active (текущая страница) | `font-bold text-neutral-900` |
| Active sub-item | `font-medium text-neutral-900` |

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
| Hover | `bg-neutral-700` (~95% серая) |
| Active / Pressed | `bg-[#EF3219]` (красная вспышка) |
| Release | плавный возврат в чёрный за 300ms |
| Disabled | `opacity-50 cursor-not-allowed` |

Паттерн: `transition-colors duration-300` + `active:bg-[#EF3219]` — при нажатии красный мгновенно, при отпускании плавно гаснет.

## 4. Интерактивные блоки (фильтры, дропдауны, карточки)

| Состояние | Стиль |
|-----------|-------|
| Default | `border border-neutral-900 bg-white text-neutral-900` |
| Hover | `bg-neutral-50` |
| Active / Selected | `bg-neutral-900 text-white` (инверсия) |
| Disabled | `opacity-50 cursor-not-allowed` |

## Общие правила

- Переходы: `transition-colors duration-150` (стандарт), `duration-300` (кнопки — для эффекта вспышки)
- Жирность: заголовки `font-bold` (700), активные подпункты `font-medium` (500), остальное normal (400)
- Фокус (keyboard): `focus:ring-2 focus:ring-neutral-900/20 focus:ring-offset-1`
- Скругления: `rounded` (4px) — минимальное, рублёное
- Обводки: 1px `border-neutral-900`
- Все контентные блоки: `bg-white` (без шума)

## Фон

- Body: белый + шум (sparse dot noise через SVG background-image)
- Все виджеты/блоки/меню: `bg-white` — перекрывают шум
- Header: прозрачный — шум виден сквозь него
