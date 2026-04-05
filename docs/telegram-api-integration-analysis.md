# Telegram Bot API -> Revisitr: Анализ интеграции

> Дата анализа: 2026-04-05
> Источники: core.telegram.org/bots/features, core.telegram.org/bots/api, core.telegram.org/bots/api-changelog
> Покрытие: Bot API 8.0 (Nov 2024) — 9.6 (Apr 3, 2026)

---

## Контекст

Telegram выпустил серию обновлений (Bot API 8.0–9.6), кульминацией которых стал **Bot API 9.6 от 3 апреля 2026** — буквально 2 дня назад. Самое важное для нас — **Managed Bots**, но есть и другие фичи, которые могут кардинально изменить продукт.

---

## 1. MANAGED BOTS (Bot API 9.6) — GAME CHANGER

### Что это
Бот может **создавать и управлять другими ботами** от имени пользователя. Пользователь переходит по ссылке `t.me/newbot/{manager}/{username}`, подтверждает — и управляющий бот получает `managed_bot` update с токеном.

**API:**
- `ManagedBotUpdated` — update при создании/смене токена
- `getManagedBotToken(user_id)` — получить токен managed бота
- `replaceManagedBotToken(user_id)` — перегенерировать токен
- `KeyboardButtonRequestManagedBot` — запрос создания бота через кнопку
- `savePreparedKeyboardButton()` — запрос из Mini App
- Ссылка: `https://t.me/newbot/{manager_bot_username}/{suggested_bot_username}?name={suggested_bot_name}`

### Текущая проблема в Revisitr
Сейчас ресторатор должен:
1. Зайти в @BotFather -> `/newbot` -> придумать username
2. Скопировать токен
3. Вставить в админку Revisitr

Это **главный барьер входа** — нетехнические пользователи путаются с BotFather.

### Очевидные идеи

**A. Автоматическое создание ботов из онбординга**
- В шаге "Создать бота" — кнопка "Создать бот в один клик"
- Открывается Telegram-ссылка с предзаполненным именем и username
- Revisitr получает `ManagedBotUpdated` -> вызывает `getManagedBotToken()` -> автоматически сохраняет в БД
- Пользователь никогда не видит BotFather

**B. Управление профилем бота из админки**
- Через токен managed бота: менять фото, описание, about, команды
- `setMyProfilePhoto()`, `setMyDescription()`, `setMyShortDescription()` — всё через API
- Онбординг step: "Загрузите логотип" -> автоматически ставится как фото бота

### Неочевидные идеи

**C. Белая метка как сервис**
- Revisitr создаёт ботов **полностью от своего имени** — ресторатор не знает, что есть "управляющий бот"
- Каждый ресторан получает уникального бота с их брендом, а вся инфраструктура — Revisitr
- Это превращает Revisitr из "конструктора ботов" в **полностью managed loyalty platform**

**D. Bot-as-a-Product: маркетплейс шаблонов**
- Шаблоны ботов: "Кофейня", "Бар", "Ресторан", "Фастфуд", "Доставка"
- Каждый шаблон — преднастроенные кнопки, welcome-сообщения, автосценарии, программа лояльности
- Ресторатор выбирает шаблон -> managed bot создаётся с полной настройкой за 30 секунд

**E. Multi-bot стратегии для одного заведения**
- Основной бот: лояльность + меню
- Eventbot: отдельный бот для мероприятий/банкетов
- Delivery-бот: заказ и доставка
- Все управляются из одной админки через managed bots

**F. Безопасная миграция клиентов**
- `replaceManagedBotToken()` — если ресторатор уходит от конкурента к Revisitr, можно безболезненно перехватить управление ботом
- Или наоборот: если клиент уходит, можно передать ему полный контроль

### Сложность реализации
- **Требуется**: включить Bot Management Mode в BotFather для "мастер-бота" Revisitr
- **Миграция**: переписать flow создания бота в онбординге — вместо ввода токена -> ссылка в Telegram
- **Затронутые файлы**: `internal/usecase/bots/`, `internal/controller/http/group/bots/`, фронт — онбординг

---

## 2. BUSINESS CONNECTIONS (Bot API 9.0) — CRM В TELEGRAM

### Что это
Владельцы Telegram Business могут подключить бота к своему аккаунту. Бот видит все входящие сообщения от клиентов и может отвечать **от имени бизнеса**.

**API:**
- `BusinessConnection` — объект подключения (connect/edit/disconnect)
- `BusinessBotRights` — права бота (заменяет `can_reply`)
- `business_message`, `edited_business_message`, `deleted_business_messages` — update types
- `business_connection_id` — параметр в send-методах
- `readBusinessMessage()` — отметить как прочитанное
- `deleteBusinessMessages()` — удалить сообщения
- `setBusinessAccountName()`, `setBusinessAccountBio()`, `setBusinessAccountProfilePhoto()` — управление профилем
- `postStory()`, `editStory()`, `deleteStory()` — публикация Stories
- `getBusinessAccountStarBalance()`, `transferBusinessAccountStars()` — Stars баланс
- `getBusinessAccountGifts()`, `convertGiftToStars()`, `upgradeGift()`, `transferGift()` — подарки
- `repostStory()` — репост историй

### Очевидные идеи

**A. Авто-ответчик на частые вопросы**
- "Какой режим работы?" -> автоматический ответ
- "Где вы находитесь?" -> адрес + карта
- "Есть ли столик на вечер?" -> ссылка на бронирование

**B. Автоматическая идентификация клиента**
- Клиент пишет в бизнес-чат -> бот ищет по `telegram_id` в `bot_clients`
- Если нашёл: "Здравствуйте, Иван! У вас 340 бонусов. Могу помочь?"
- Если не нашёл: предлагает присоединиться к программе лояльности

### Неочевидные идеи

**C. Сквозная CRM-аналитика**
- Все диалоги бизнес-аккаунта проходят через бота
- Автоматический тегинг: "бронирование", "жалоба", "комплимент", "доставка"
- Корреляция с RFM-сегментацией: "VIP-клиент с жалобой" -> приоритетный ответ

**D. Business Account Management**
- `setBusinessAccountName()`, `setBusinessAccountBio()`, `setBusinessAccountProfilePhoto()`
- **Revisitr управляет профилем бизнес-аккаунта ресторатора** из одной админки
- `postStory()` — публикация Stories от имени бизнеса! Акции, новинки меню — прямо из CRM

**E. Лид-генерация из чатов**
- Deep link `/start bizChat<user_chat_id>` — бот получает контекст конкретного чата
- Менеджер нажимает "Manage Bot" -> видит полную карточку клиента с историей покупок
- Прямо из чата: начислить бонусы, отправить промокод, пометить как VIP

**F. `readBusinessMessage()` для аналитики**
- Отслеживание скорости ответа менеджеров
- Метрика: среднее время первого ответа по каждому заведению
- Дашборд "Quality of Service" в админке

---

## 3. TELEGRAM STARS & PAYMENTS — МОНЕТИЗАЦИЯ

### API
- `sendInvoice()`, `createInvoiceLink()` — создание счетов
- `answerPreCheckoutQuery()` — подтверждение оплаты
- `subscription_period` — подписки с автопродлением
- `editUserStarSubscription()` — управление подпиской
- `StarTransaction`, `StarTransactions` — история транзакций
- `getMyStarBalance()` — баланс Stars бота
- `transferBusinessAccountStars()` — перевод Stars
- `TransactionPartnerAffiliateProgram` — партнёрские комиссии
- `PaidMediaInfo` с `star_count` — платный контент
- `RefundedPayment` — возвраты
- `is_recurring`, `is_first_recurring` — рекуррентные платежи
- Валюта: `XTR` (Telegram Stars) для цифровых товаров

### Очевидные идеи

**A. Подписка на VIP-программу через Stars**
- `createInvoiceLink(subscription_period)` — автопродление
- VIP-клиент платит N Stars/месяц -> получает x2 начисление бонусов, эксклюзивные акции
- `editUserStarSubscription()` — управление подпиской

**B. Продажа подарочных сертификатов**
- Digital goods через Stars — не требует платёжного провайдера
- Сертификат на 1000р = N Stars -> привязывается к QR-коду -> погашается на кассе

### Неочевидные идеи

**C. Stars как внутренняя валюта экосистемы**
- Промо: "Пополни Stars -> получи +20% бонусов"
- `getMyStarBalance()` — бот отслеживает свой баланс Stars
- `transferBusinessAccountStars()` — перевод Stars между ботом и бизнес-аккаунтом
- Ресторатор получает Stars от клиентов -> конвертирует в Toncoin -> реальный доход

**D. Paid Media для эксклюзивного контента**
- `PaidMediaInfo` с `star_count` — платные фото/видео
- "Мастер-класс шефа" — платное видео за Stars
- "Секретное меню" — доступ за Stars (геймификация)

**E. Affiliate Program для виральности**
- `TransactionPartnerAffiliateProgram` (Bot API 8.1)
- Клиент приводит друга -> оба получают Stars/бонусы
- Автоматическое отслеживание комиссий через Bot API
- Интеграция с текущей системой промокодов (`PromoCode` entity)

**F. Subscription tiers для B2B**
- Ресторатор платит за Revisitr подписку через Stars!
- Текущий `billing.go` с `Tariff` -> альтернативный способ оплаты без платёжного провайдера
- `subscription_period` + `is_recurring` — автоматическое продление тарифа

---

## 4. GIFTS — ВИРАЛЬНОСТЬ И УДЕРЖАНИЕ

### API
- `sendGift(user_id, gift_id, text, text_parse_mode)` — отправка подарка
- `Gift`, `GiftInfo`, `UniqueGiftInfo` — объекты подарков
- `UniqueGift` с `rarity`, `is_burned`, `is_premium` — уникальные подарки
- `getUserGifts()`, `getChatGifts()` — получение списка подарков
- `getBusinessAccountGifts()` — подарки бизнес-аккаунта
- `convertGiftToStars()` — конвертация в Stars
- `upgradeGift()` — апгрейд подарка
- `transferGift()` — передача подарка
- `GiftUpgradeSent` — уведомление об апгрейде

### Очевидные идеи

**A. Автоматические подарки по триггерам**
- Интеграция с `auto_scenario.go`: триггер `birthday` -> `sendGift()` от имени ресторана
- Триггер `visit_count >= 10` -> подарок "Верный гость"
- Визуально красивее обычного текстового сообщения

### Неочевидные идеи

**B. Брендированные уникальные подарки**
- `UniqueGift` с `rarity` — создание коллекционных подарков ресторана
- "Золотая карта [Ресторан]" — unique gift, выдаётся при достижении топ-уровня
- Клиенты коллекционируют -> социальное доказательство -> приводят друзей

**C. Gift-to-Friend как реферальный механизм**
- Клиент получает подарок -> может передать другу (`transferGift()`)
- Друг видит подарок от ресторана -> переходит в бот -> регистрируется
- Органический рост базы через gifting viral loop

**D. Gift Marketplace между заведениями**
- Партнёрская сеть: подарки от Ресторана A можно обменять на скидку в Баре B
- `getBusinessAccountGifts()` — управление полученными подарками
- `convertGiftToStars()` — монетизация неиспользованных подарков

---

## 5. MINI APPS — ПОЛНОЦЕННЫЙ DIGITAL EXPERIENCE

### API (Bot API 8.0–9.0)
- Full-screen mode: `requestFullscreen()`, `exitFullscreen()`
- Home screen: `addToHomeScreen()`, `checkHomeScreenStatus()`
- Emoji status: `setUserEmojiStatus()`, `setEmojiStatus()`
- Geolocation: `LocationManager`
- Device motion: `Accelerometer`, `DeviceOrientation`, `Gyroscope`
- Storage: `DeviceStorage` (persistent), `SecureStorage` (sensitive data)
- Media: `shareMessage()`, `downloadFile()`, `shareToStory()`
- Shared context: `chat_instance` для групповых сценариев
- Biometrics: `biometricsManager`
- QR: нативный QR-сканер

### Очевидные идеи

**A. Карта лояльности как Mini App**
- Визуальная штамп-карта (не текст "У вас 5 штампов", а красивый UI)
- Баланс, история транзакций, доступные награды
- Home screen shortcut — "лояльность в одно касание"

### Неочевидные идеи

**B. Self-service kiosk в Mini App**
- Клиент за столиком -> открывает Mini App -> сканирует QR стола
- Видит меню -> заказывает -> оплачивает Stars -> начисляются бонусы
- **Full-screen mode** для иммерсивного меню с фото блюд

**C. Геолокационные триггеры**
- `LocationManager` (Bot API 8.0) — Mini App запрашивает геолокацию
- Клиент в 200м от ресторана -> push: "Ваш любимый капучино ждёт! -20% в ближайший час"
- Интеграция с `auto_scenario` триггерами

**D. Shared Context для групповых заказов**
- `chat_instance` — Mini App в групповом чате
- Друзья в чате -> открывают Mini App -> каждый выбирает блюда -> общий заказ
- Автоматическое разделение счёта и начисление бонусов каждому

**E. Emoji Status как социальное доказательство**
- `setUserEmojiStatus()` (Bot API 8.0)
- "Постоянный гость Restaurant X" как emoji-статус в Telegram
- Бесплатная реклама ресторана в каждом чате клиента

**F. Device Storage для офлайн-карты**
- `DeviceStorage` и `SecureStorage` (Bot API 9.0)
- QR-код карты лояльности хранится локально
- Работает без интернета — критично для подвалов/зон без связи

**G. Stories из Mini App**
- `shareToStory()` — клиент делится своим статусом лояльности в Stories
- "Я VIP-гость в [Ресторан] — 1200 бонусов!" + ссылка на бот
- Виральный рост без затрат на рекламу

**H. Биометрическая авторизация**
- `biometricsManager` — вход по Face ID / отпечатку для списания бонусов
- Безопасность: никто кроме владельца не спишет бонусы с карты

---

## 6. CHECKLISTS (Bot API 9.1) — ОПЕРАЦИОННАЯ ЭФФЕКТИВНОСТЬ

### API
- `Checklist`, `InputChecklist` — структура чеклиста (1-30 задач)
- `ChecklistTask` — задача с отслеживанием выполнения
- `sendChecklist()`, `editMessageChecklist()` — отправка/редактирование
- `ChecklistTasksDone`, `ChecklistTasksAdded` — service messages
- `others_can_add_tasks`, `others_can_mark_tasks_as_done` — совместная работа
- `completed_by_chat` — кто выполнил

### Неочевидные идеи

**A. Трекинг заказов для доставки/самовывоза**
- `sendChecklist()` с задачами: Принят -> Готовится -> Готов -> В пути
- `ChecklistTasksDone` — автоматическое уведомление клиенту

**B. Чеклисты для менеджеров через Admin Bot**
- "Утренний чеклист": открытие, проверка продуктов, включение оборудования
- Данные идут в аналитику -> можно отслеживать дисциплину персонала

**C. Feedback-чеклист после визита**
- Вместо "Оцените от 1 до 5" -> чеклист: Еда была вкусной / Обслуживание быстрое / Атмосфера хорошая
- `others_can_mark_tasks_as_done: false` — только клиент отмечает

---

## 7. POLLS 2.0 (Bot API 9.6) — ENGAGEMENT

### API (обновлённый)
- Несколько правильных ответов: `correct_option_ids` (массив)
- `allows_multiple_answers` теперь работает для квизов
- `allows_revoting` — возможность переголосовать
- `shuffle_options` — перемешивание вариантов
- `allow_adding_options` — пользователи добавляют свои варианты
- `hide_results_until_closes` — скрытие результатов до закрытия
- `persistent_id` в PollOption — стабильный идентификатор
- `description`, `description_entities` — описание опроса
- Макс. время автозакрытия увеличено до 2,628,000 сек (~30 дней)
- `PollOptionAdded`, `PollOptionDeleted` — service messages

### Идеи

**A. Голосование за новые блюда**
- `allow_adding_options: true` — клиенты сами предлагают варианты
- `allows_revoting: true` — можно передумать
- `hide_results_until_closes: true` — интрига до конца

**B. NPS через квизы**
- `correct_option_ids` (несколько правильных) — квизы о меню
- Правильные ответы -> бонусные баллы (геймификация)

**C. A/B тест акций**
- Два варианта промо -> опрос -> победивший автоматически превращается в кампанию
- `persistent_id` для стабильного трекинга

---

## 8. TAGS (Bot API 9.5) — ВИЗУАЛИЗАЦИЯ ЛОЯЛЬНОСТИ

### API
- `setChatMemberTag(chat_id, user_id, tag)` — назначить тег
- `sender_tag` — тег в сообщениях
- `can_edit_tag`, `can_manage_tags` — права

### Идеи

**A. VIP-теги в групповых чатах**
- `setChatMemberTag()` — в чате ресторана VIP-клиенты видны по тегу
- "Gold Member" рядом с именем -> социальное доказательство

---

## 9. ДОПОЛНИТЕЛЬНЫЕ ФИЧИ

### Direct Messages & Paid Messages (Bot API 9.1–9.2)
- `DirectMessagePriceChanged` — платные личные сообщения в каналах
- `PaidMessagePriceChanged` — платные сообщения в группах
- Идея: премиум-канал ресторана с эксклюзивным контентом за Stars

### Suggested Posts (Bot API 9.2)
- `SuggestedPostParameters` — предложение публикаций
- `approveSuggestedPost()`, `declineSuggestedPost()` — модерация
- Идея: клиенты предлагают отзывы для публикации в канале ресторана

### Stories (Bot API 9.0)
- `postStory()`, `editStory()`, `deleteStory()` — управление Stories
- `StoryArea` с типами: location, link, weather, uniqueGift, suggestedReaction
- Идея: автоматическая публикация Stories с акциями дня из CRM

### Profile Management (Bot API 9.4)
- `setMyProfilePhoto()`, `removeMyProfilePhoto()` — фото бота
- `VideoQuality` — поддержка нескольких качеств видео
- Идея: автоматическое обновление фото бота под сезонные акции

---

## ПРИОРИТЕТЫ ВНЕДРЕНИЯ

| Приоритет | Фича | Ценность для бизнеса | Сложность | ROI |
|-----------|-------|---------------------|-----------|-----|
| **P0** | Managed Bots (автосоздание) | Убирает главный барьер входа | Средняя | MAX |
| **P1** | Mini App (карта лояльности) | Визуальный wow-эффект | Высокая | MAX |
| **P1** | Stars подписка (VIP) | Новый revenue stream | Средняя | Высокий |
| **P2** | Gifts + автосценарии | Виральность + удержание | Низкая | Высокий |
| **P2** | Business Connection (CRM) | Upsell для крупных клиентов | Средняя | Высокий |
| **P3** | Emoji Status / Stories | Бесплатный виральный рост | Низкая | Средний |
| **P3** | Polls 2.0 (engagement) | Увеличение retention | Низкая | Средний |
| **P3** | Checklists (operations) | Ops efficiency | Низкая | Средний |
| **P4** | Affiliate Program | Масштабирование | Средняя | Средний |
| **P4** | Tags (VIP visibility) | Социальное доказательство | Низкая | Низкий |

---

## P0: MANAGED BOTS — ПЛАН РЕАЛИЗАЦИИ

Это первое, что стоит сделать, потому что:
1. Убирает **#1 фрикшн** при онбординге
2. Реализуется относительно быстро
3. Не ломает существующий flow (можно оставить ручной ввод токена как fallback)

**Что нужно:**
1. Зарегистрировать "мастер-бот" Revisitr и включить Bot Management Mode
2. Добавить webhook/long-polling для `managed_bot` updates на мастер-боте
3. API endpoint: `POST /api/v1/bots/create-managed` -> генерирует ссылку `t.me/newbot/revisitr_master/{suggested_username}`
4. При получении `ManagedBotUpdated` -> `getManagedBotToken()` -> сохраняет в `bots` таблицу
5. Фронт: в онбординге кнопка "Создать бот автоматически" -> redirect в Telegram -> polling/webhook на результат

**Затронутые компоненты:**
- `backend/cmd/bot/main.go` — добавить мастер-бот listener
- `backend/internal/service/botmanager/` — обработка managed_bot updates
- `backend/internal/usecase/bots/` — новый метод CreateManaged
- `backend/internal/controller/http/group/bots/` — новый endpoint
- `backend/internal/entity/bot.go` — поле `is_managed`, `managed_by`
- `frontend/` — UI для автосоздания бота в онбординге

---

## CHANGELOG TELEGRAM BOT API (краткий, 8.0–9.6)

| Версия | Дата | Ключевые фичи |
|--------|------|---------------|
| 8.0 | Nov 17, 2024 | Star subscriptions, Mini App full-screen/home screen/emoji status/geolocation/device motion, Gifts |
| 8.1 | Dec 4, 2024 | Affiliate programs, nanostar amounts |
| 8.2 | Jan 1, 2025 | Bot verification, gift upgrades |
| 8.3 | Feb 12, 2025 | Gifts to channels, video covers, reactions on service messages |
| 9.0 | Apr 11, 2025 | Business account management, Stories, gift CRUD, DeviceStorage/SecureStorage, Premium gifting |
| 9.1 | Jul 3, 2025 | Checklists, getMyStarBalance, direct message pricing |
| 9.2 | Aug 15, 2025 | Direct messages in channels, suggested posts, gift publisher |
| 9.3 | Dec 31, 2025 | Topics in private chats, sendMessageDraft, expanded gift features |
| 9.4 | Feb 9, 2026 | Bot profile management, VideoQuality, custom emoji for Premium bot owners |
| 9.5 | Mar 1, 2026 | Member tags, date_time entities, sendMessageDraft for all bots |
| 9.6 | Apr 3, 2026 | **Managed Bots**, Polls 2.0 (multiple answers, revoting, adding options, hiding results) |
