# Frontend Redesign Changelog

Финальные изменения визуальной части. Только фронтенд, бизнес-логика не затронута.

## Общие изменения

### Шрифт
- **Файл**: `frontend/src/index.css`
- Основной шрифт сайта заменён с Outfit на **Inter**

### Акцентный цвет
- **Файл**: `frontend/src/index.css`
- `--color-accent`: `#E85D3A` → `#EF3219`
- Hover красных ссылок: `#FF5C47` (ярче/светлее)

### Шум на фоне
- **Файл**: `frontend/src/index.css`
- Noise texture заменён на sparse dot noise (SVG-тайл 40x40px, редкие чёрные точки 1px)
- Шум на `body` через `background-image`, все контентные блоки с `bg-white` перекрывают его
- Header прозрачный — шум виден сквозь него

### Чекбоксы
- **Файл**: `frontend/src/index.css`
- Глобальный `accent-color: #171717` — чёрные вместо синих

### Логотип
- **Файл**: `frontend/public/logo.png`
- PNG логотип `revsitr` (со слэшем), используется как `<img>` в header

---

## Layout панели управления

### Структура (router.tsx)
- **Файл**: `frontend/src/router.tsx`
- Header теперь сверху на всю ширину, под ним sidebar + content в flex-row
- Корневой `div` без `bg-white` (чтобы шум был виден)
- Auth-проверка временно отключена (TODO: вернуть)
- Адаптивные отступы: `px-4 sm:px-8 lg:px-16`

### Header
- **Файл**: `frontend/src/components/layout/Header.tsx`
- Логотип PNG, центрирован над sidebar (`lg:w-[220px] lg:justify-center`)
- Навигация: Панель управления (bold, без подчёркивания), Тарифы, Поддержка, Маркетинг «под ключ» (красный, без подчёркивания, hover ярче), Контакты
- Nav скрыт на < lg, бургер-меню с обводкой 1px
- Профиль: аватарка + имя + PRO badge
- Разделитель: 1px `border-neutral-900` линия под header
- Адаптивность: на мобиле бургер + логотип + аватарка, имя скрыто на < sm

### Sidebar
- **Файл**: `frontend/src/components/layout/Sidebar.tsx`
- Белый фон `bg-white`, обводка 1px `border-neutral-900`, скругление `rounded`
- Автовысота по контенту, не sticky на весь экран
- Порядок: Дашборд | разделитель | Аналитика, Клиенты, Лояльность, Рассылки, Акции | разделитель | Мои боты, Точки продаж, Интеграции
- Разделители: `border-neutral-200`, от левого края иконок
- Все подменю свёрнуты по умолчанию, авто-раскрытие при попадании на страницу
- Активный заголовок: `font-bold text-neutral-900`
- Активный подпункт: `font-medium text-neutral-900` (тоньше заголовка)
- Неактивные подпункты: `text-neutral-400 hover:text-neutral-700`
- Отступ подпунктов: сразу под заголовком, `mb-1` после последнего подпункта

---

## Страница логина

### Редизайн
- **Файл**: `frontend/src/routes/auth/login.tsx`
- Центрированная форма на белом фоне, без левой брендовой панели
- Логотип `rev/sitr` текстом по центру
- Инпуты: underline-стиль (нижняя граница)
- Toggle видимости пароля (Eye/EyeOff)
- Кнопка: `bg-neutral-900`, `rounded-xl`, hover `bg-neutral-700`, active `bg-[#EF3219]` (красная вспышка, плавный возврат 300ms)
- Чекбокс «Запомнить»
- Ссылки: «Зарегистрироваться», «Восстановить пароль» — `text-[#EF3219] underline`

---

## Страница Аналитика > Продажи

### Визуал
- **Файл**: `frontend/src/routes/dashboard/analytics/sales.tsx`
- Все карточки/виджеты: `border border-neutral-900 rounded bg-white`
- Графики: stroke/fill `#171717` вместо акцентного цвета
- Микроанимации: staggered `fadeSlideIn` на заголовок → фильтры → метрики → графики

### PeriodFilter (новый компонент)
- **Файл**: `frontend/src/components/common/PeriodFilter.tsx`
- Кастомный дропдаун периодов (200px)
- Группировка: Сегодня, Вчера, 7/30 дней, Год | Эта неделя, месяц, год | Произвольный
- Кнопка дат с иконкой календаря, автоподстановка дат при выборе пресета
- Кастомный range calendar: выбор диапазона одним календарём
- Запрет выбора будущих дат
- Автоопределение пресета при ручном выборе дат
- Адаптивность: на мобиле столбик, короткий формат дат, full-width

---

## Дизайн-система

### Файл
- `frontend/docs/design/interaction-states.md`
- Зафиксированы состояния: навигация, акцентные ссылки, кнопки (primary), интерактивные блоки
- Общие правила: переходы, жирность, фокус, скругления, обводки

---

## Файлы, которые нужно перенести при мёрдже

| Файл | Тип изменения |
|------|---------------|
| `frontend/src/index.css` | Модификация (шрифт, цвета, шум, чекбоксы) |
| `frontend/src/router.tsx` | Модификация (layout, убрать TODO auth) |
| `frontend/src/components/layout/Header.tsx` | Полная перезапись |
| `frontend/src/components/layout/Sidebar.tsx` | Полная перезапись |
| `frontend/src/components/common/PeriodFilter.tsx` | Новый файл |
| `frontend/src/routes/auth/login.tsx` | Полная перезапись |
| `frontend/src/routes/dashboard/analytics/sales.tsx` | Модификация (визуал + PeriodFilter) |
| `frontend/public/logo.png` | Новый файл |
| `frontend/docs/design/interaction-states.md` | Новый файл |
| `docs/design/CHANGELOG.md` | Новый файл (этот) |
