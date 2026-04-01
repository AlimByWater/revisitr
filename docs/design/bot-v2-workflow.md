# Bot Architecture v2 — Implementation Workflow

> Источник: `docs/design/bot-architecture-v2.md`
> Дата: 2026-04-01
> Стратегия: Systematic (backend-first → frontend → integration)

---

## Обзор

5 фаз, 37 задач. Backend-first подход: сначала инфраструктура (event bus), потом модели и отправка, потом UI.

```
Фаза 1: Event Bus ──────────┐
Фаза 2: Составные сообщения ─┤── Backend (параллельно с Фазой 3)
Фаза 3: Telegram Preview ────┤── Frontend (параллельно с Фазой 2)
Фаза 4: Content Editor ──────┤── Frontend (после Фазы 3)
Фаза 5: Интеграция ──────────┘── Full-stack (после 2 + 4)
```

Зависимости между фазами:
- Фаза 2 зависит от 1 (event bus нужен для propagation настроек)
- Фаза 3 не зависит от backend (работает на mock data)
- Фаза 4 зависит от 3 (использует TelegramPreview)
- Фаза 5 зависит от 2 + 4 (связывает backend API с frontend editor)

---

## Фаза 1: Event Bus + Hot Reload

**Цель**: Redis Pub/Sub между API Server и Bot Service. Горячее обновление настроек бота без перезапуска.

**Предусловия**: Docker compose up (PostgreSQL + Redis)

### 1.1 Entity: MessageContent model
- **Файл**: `backend/internal/entity/message.go` (новый)
- **Что**:
  - `MessagePartType` — string enum (text, photo, video, document, sticker, animation, audio, voice)
  - `MessagePart` — struct (Type, Text, MediaURL, MediaID, ParseMode)
  - `InlineButton` — struct (Text, URL, Data)
  - `MessageContent` — struct (Parts []MessagePart, Buttons [][]InlineButton)
  - `Scan()` / `Value()` — JSONB serialization (паттерн из `TriggerConfig`)
  - `Validate()` — правила: min 1 part, max 5 parts, sticker без caption, media требует URL
- **Зависимости**: нет
- **Проверка**: `go build ./internal/entity/...`

### 1.2 EventBus: Publisher
- **Файл**: `backend/internal/service/eventbus/eventbus.go` (новый)
- **Что**:
  - `EventBus` struct — оборачивает `*goredis.Client`
  - `BotEvent` struct — `{BotID int, Field string}`
  - `New(client *goredis.Client) *EventBus`
  - `PublishBotReload(ctx, botID)` → `PUBLISH revisitr:bot:reload {json}`
  - `PublishBotStop(ctx, botID)` → `PUBLISH revisitr:bot:stop {json}`
  - `PublishBotStart(ctx, botID)` → `PUBLISH revisitr:bot:start {json}`
  - `PublishBotSettings(ctx, botID, field)` → `PUBLISH revisitr:bot:settings {json}`
  - Внутренний `publish(ctx, channel, payload)` — marshal + Publish
- **Зависимости**: `go-redis/v9`
- **Проверка**: unit test

### 1.3 EventBus: Subscriber
- **Файл**: `backend/internal/service/eventbus/subscriber.go` (новый)
- **Что**:
  - `BotEventHandler` interface:
    ```go
    OnBotReload(ctx context.Context, botID int) error
    OnBotStop(ctx context.Context, botID int) error
    OnBotStart(ctx context.Context, botID int) error
    OnBotSettingsChanged(ctx context.Context, botID int, field string) error
    ```
  - `Subscriber` struct — `{rds *goredis.Client, logger *slog.Logger}`
  - `NewSubscriber(client, logger) *Subscriber`
  - `Listen(ctx, handler BotEventHandler)` — блокирующий цикл:
    - Subscribe на 4 канала
    - Dispatch по channel name → метод handler
    - Recover от panic в handler
    - Логирование ошибок
  - `dispatch(ctx, msg, handler)` — switch по msg.Channel, unmarshal BotEvent, вызов handler
- **Зависимости**: 1.2 (BotEvent struct), go-redis
- **Проверка**: unit test с mock handler

### 1.4 Bot Service: Redis init + Subscriber
- **Файл**: `backend/cmd/bot/main.go` (изменение)
- **Что**:
  - Добавить `redisRepo "revisitr/internal/repository/redis"` import
  - Инициализировать Redis module: `rds := redisRepo.New(&redisConfig{cfg: cfg})`
  - `rds.Init(ctx, logger)`
  - Создать subscriber: `sub := eventbus.NewSubscriber(rds.Client(), logger)`
  - Запустить в goroutine: `go sub.Listen(ctx, mgr)` (Manager как handler)
  - Добавить `redisConfig` struct (аналогично postgresConfig)
  - `defer rds.Close()`
- **Зависимости**: 1.3
- **Проверка**: `go build ./cmd/bot`

### 1.5 API Server: EventBus в DI
- **Файл**: `backend/cmd/server/main.go` (изменение)
- **Что**:
  - Создать `eb := eventbus.New(redisModule.Client())`
  - Передать `eb` в bots usecase: `botsUC := bots.New(botsRepo, eb, logger)`
  - (Redis уже инициализирован в server — проверить, что `Client()` доступен)
- **Зависимости**: 1.2
- **Проверка**: `go build ./cmd/server`

### 1.6 Manager: implement BotEventHandler
- **Файл**: `backend/internal/service/botmanager/manager.go` (изменение)
- **Что**:
  - Добавить метод `OnBotReload(ctx, botID)` → вызов существующего `ReloadBot(ctx, botID)`
  - Добавить метод `OnBotStop(ctx, botID)` → вызов существующего `RemoveBot(botID)`
  - Добавить метод `OnBotStart(ctx, botID)` → вызов существующего `AddBot(ctx, botID)`
  - Добавить метод `OnBotSettingsChanged(ctx, botID, field)`:
    ```go
    m.mu.Lock()
    defer m.mu.Unlock()
    inst, ok := m.instances[botID]
    if !ok { return nil }
    bot, err := m.botsRepo.GetByID(ctx, botID)
    if err != nil { return err }
    inst.info = *bot
    m.logger.Info("bot settings hot-reloaded", "bot_id", botID, "field", field)
    return nil
    ```
  - Убедиться что Manager удовлетворяет `BotEventHandler` interface
- **Зависимости**: 1.3 (interface)
- **Проверка**: `go vet ./internal/service/botmanager/...`

### 1.7 Bots Usecase: publish events
- **Файл**: `backend/internal/usecase/bots/bots.go` (изменение)
- **Что**:
  - Добавить `eventBus` field в Usecase struct
  - Обновить `New()` constructor — принять eventBus
  - В `UpdateSettings()`: после успешного repo.UpdateSettings добавить:
    ```go
    if uc.eventBus != nil {
        uc.eventBus.PublishBotSettings(ctx, botID, "")
    }
    ```
  - В `Update()`: если менялся status — PublishBotStart или PublishBotStop
  - В `Delete()`: PublishBotStop
  - В `Create()` (если auto-start): PublishBotStart
  - EventBus ошибки логировать, не возвращать (fire-and-forget)
- **Зависимости**: 1.2, 1.5
- **Проверка**: unit test — проверить что eventBus вызывается

### 1.8 Тесты Phase 1
- **Файлы**:
  - `backend/internal/entity/message_test.go` — Validate() cases
  - `backend/internal/service/eventbus/eventbus_test.go` — Publisher unit
  - `backend/internal/service/eventbus/subscriber_test.go` — Subscriber dispatch
  - `backend/internal/service/botmanager/manager_test.go` — OnBotSettingsChanged
  - `backend/internal/usecase/bots/bots_test.go` — UpdateSettings publishes event
- **Проверка**: `go test ./internal/entity/... ./internal/service/eventbus/... ./internal/service/botmanager/... ./internal/usecase/bots/...`

### Валидация Фазы 1
```bash
# Все тесты проходят
cd backend && go test -race ./internal/entity/... ./internal/service/eventbus/... ./internal/service/botmanager/... ./internal/usecase/bots/...

# Оба бинарника собираются
go build ./cmd/server && go build ./cmd/bot

# Lint
go vet ./...
```

---

## Фаза 2: Составные сообщения (Backend)

**Цель**: Миграция БД, telegram.Sender, обновление campaign/welcome/scheduler.

**Предусловия**: Фаза 1 завершена.

### 2.1 Миграция: composite_messages
- **Файл**: `backend/migrations/00032_composite_messages.sql` (новый)
- **Что**:
  - `ALTER TABLE campaigns ADD COLUMN content JSONB`
  - `ALTER TABLE campaign_templates ADD COLUMN content JSONB`
  - `ALTER TABLE campaign_variants ADD COLUMN content JSONB`
  - `ALTER TABLE auto_scenarios ADD COLUMN content JSONB`
  - UPDATE существующих данных: конвертация message+media_url → content JSONB
  - Goose Down: DROP COLUMN content
- **Зависимости**: нет
- **Проверка**: `goose -dir migrations postgres "$DATABASE_URL" up && goose status`

### 2.2 Entity: BotSettings + WelcomeContent
- **Файл**: `backend/internal/entity/bot.go` (изменение)
- **Что**:
  - Добавить в `BotSettings`:
    ```go
    WelcomeContent *MessageContent `json:"welcome_content,omitempty"`
    ```
  - Импорт не нужен (тот же пакет entity)
- **Зависимости**: 1.1
- **Проверка**: `go build ./internal/entity/...`

### 2.3 Entity: Campaign + Content
- **Файл**: `backend/internal/entity/campaign.go` (изменение)
- **Что**:
  - Добавить в `Campaign`:
    ```go
    Content *MessageContent `db:"content" json:"content,omitempty"`
    ```
  - Добавить в `CampaignTemplate`:
    ```go
    Content *MessageContent `db:"content" json:"content,omitempty"`
    ```
  - Добавить в `CampaignVariant`:
    ```go
    Content *MessageContent `db:"content" json:"content,omitempty"`
    ```
  - Добавить в `AutoScenario`:
    ```go
    Content *MessageContent `db:"content" json:"content,omitempty"`
    ```
  - Добавить helper:
    ```go
    func (c *Campaign) GetContent() MessageContent {
        if c.Content != nil { return *c.Content }
        // Fallback: конвертация legacy полей
        part := MessagePart{Type: PartText, Text: c.Message, ParseMode: "Markdown"}
        if c.MediaURL != nil && *c.MediaURL != "" {
            part.Type = PartPhoto
            part.MediaURL = *c.MediaURL
        }
        return MessageContent{Parts: []MessagePart{part}, Buttons: /* convert c.Buttons */}
    }
    ```
- **Зависимости**: 1.1, 2.1 (миграция)
- **Проверка**: `go build ./internal/entity/...`

### 2.4 Telegram Sender: SendContent
- **Файл**: `backend/internal/service/telegram/sender.go` (новый)
- **Что**:
  - `Sender` struct — `{baseURL string, logger *slog.Logger}`
  - `New(baseURL string, logger *slog.Logger) *Sender`
  - `SendContent(ctx, bot *telego.Bot, chatID int64, content MessageContent) error`
    - Итерация по parts
    - Кнопки на последний part
    - Задержка 50ms между parts (Telegram rate limit)
  - `sendPart(bot, chatID, part, markup)` — switch по part.Type:
    - `PartText` → `bot.SendMessage()`
    - `PartPhoto` → `bot.SendPhoto()`
    - `PartVideo` → `bot.SendVideo()`
    - `PartDocument` → `bot.SendDocument()`
    - `PartSticker` → `bot.SendSticker()`
    - `PartAnimation` → `bot.SendAnimation()`
    - `PartAudio` → `bot.SendAudio()`
    - `PartVoice` → `bot.SendVoice()`
  - `mediaInput(part)` → `FileFromID` или `FileFromURL`
  - `buildInlineKeyboard(buttons)` → `*telego.InlineKeyboardMarkup`
- **Зависимости**: 1.1 (entity), telego
- **Проверка**: unit test с mock telego.Bot

### 2.5 Campaign Sender: использовать telegram.Sender
- **Файл**: `backend/internal/service/campaign/sender.go` (изменение)
- **Что**:
  - Добавить `tgSender *telegram.Sender` в struct
  - Обновить `NewSender()` — принять `tgSender`
  - В `SendCampaign()`:
    ```go
    // Было:
    tgMsg := tu.Message(tu.ID(messages[i].TelegramID), campaign.Message)
    _, err := tBot.SendMessage(tgMsg)

    // Стало:
    content := campaign.GetContent()
    err := s.tgSender.SendContent(ctx, tBot, messages[i].TelegramID, content)
    ```
  - Удалить прямой импорт `tu "github.com/mymmrac/telego/telegoutil"` (если больше не используется)
- **Зависимости**: 2.3, 2.4
- **Проверка**: `go build ./internal/service/campaign/...`

### 2.6 Bot Handler: welcome через SendContent
- **Файл**: `backend/internal/service/botmanager/handler.go` (изменение)
- **Что**:
  - Добавить `tgSender *telegram.Sender` в handler struct
  - Обновить `newHandler()` — принять tgSender
  - Добавить метод `getWelcomeContent()`:
    ```go
    func (h *handler) getWelcomeContent() *entity.MessageContent {
        s := h.info.Settings
        if s.WelcomeContent != nil && len(s.WelcomeContent.Parts) > 0 {
            return s.WelcomeContent
        }
        if s.WelcomeMessage != "" {
            return &entity.MessageContent{
                Parts: []entity.MessagePart{
                    {Type: entity.PartText, Text: s.WelcomeMessage, ParseMode: "Markdown"},
                },
            }
        }
        return &entity.MessageContent{
            Parts: []entity.MessagePart{
                {Type: entity.PartText, Text: fmt.Sprintf("Добро пожаловать в %s!", h.info.Name)},
            },
        }
    }
    ```
  - В `handleStart()`: заменить `h.sendText(chatID, welcome)` на:
    ```go
    content := h.getWelcomeContent()
    if err := h.tgSender.SendContent(ctx, h.bot, chatID, *content); err != nil {
        h.logger.Error("send welcome", "error", err)
    }
    ```
  - Обновить `manager.go` → `startBot()` — передать tgSender в newHandler
- **Зависимости**: 2.4, 2.2
- **Проверка**: `go build ./internal/service/botmanager/...`

### 2.7 Scheduler: auto-scenarios через SendContent
- **Файл**: `backend/internal/service/campaign/scheduler.go` (изменение)
- **Что**:
  - Добавить `tgSender *telegram.Sender` в Scheduler struct
  - Обновить `NewScheduler()` — принять tgSender
  - В `evaluateBirthday()` и других evaluate-методах:
    ```go
    // Вместо прямого SendMessage использовать:
    content := scenario.GetContent() // или конвертировать scenario.Message
    tgSender.SendContent(ctx, tBot, client.TelegramID, content)
    ```
- **Зависимости**: 2.4, 2.3
- **Проверка**: `go build ./internal/service/campaign/...`

### 2.8 Redis Queue: обновить QueueMessage
- **Файл**: `backend/internal/repository/redis/campaign_queue.go` (изменение)
- **Что**:
  - Заменить отдельные поля `Text`, `MediaURL`, `MediaType`, `Buttons` на:
    ```go
    Content entity.MessageContent `json:"content"`
    ```
  - Обратная совместимость: если в очереди старый формат — корректно десериализовать
- **Зависимости**: 1.1
- **Проверка**: unit test

### 2.9 Тесты Phase 2
- **Файлы**:
  - `backend/internal/service/telegram/sender_test.go` — все типы parts, кнопки, fallback
  - `backend/internal/entity/campaign_test.go` — GetContent() с legacy и новым форматом
  - `backend/internal/service/campaign/sender_test.go` — SendCampaign с content
  - Integration test: миграция + отправка
- **Проверка**: `go test -race ./internal/service/telegram/... ./internal/service/campaign/... ./internal/entity/...`

### Валидация Фазы 2
```bash
# Миграция
cd backend && goose -dir migrations postgres "$DATABASE_URL" up

# Все тесты
go test -race ./...

# Оба бинарника
go build ./cmd/server && go build ./cmd/bot

# Lint
go vet ./... && staticcheck ./...
```

---

## Фаза 3: Telegram Preview Component (Frontend)

**Цель**: React-компонент для визуализации Telegram-сообщений в стиле iOS.

**Предусловия**: нет (параллельно с Фазой 2). Работает на mock data.

### 3.1 Типы и структуры
- **Файл**: `frontend/src/features/telegram-preview/types.ts` (новый)
- **Что**:
  - `MessagePartType` — union type ('text' | 'photo' | 'video' | 'document' | 'sticker' | 'animation')
  - `MessagePart` — interface {type, text?, mediaUrl?, mediaId?, parseMode?}
  - `InlineButton` — interface {text, url?, data?}
  - `MessageContent` — interface {parts: MessagePart[], buttons?: InlineButton[][]}
  - `TelegramPreviewProps` — interface для главного компонента
- **Зависимости**: нет
- **Проверка**: TypeScript compilation

### 3.2 Стили: telegram.css
- **Файл**: `frontend/src/features/telegram-preview/styles/telegram.css` (новый)
- **Что**:
  - `.tg-chat-bg` — фон чата (серо-голубой, опциональный паттерн)
  - `.tg-bubble` — пузырь (border-radius 16px, белый, тень, tail через clip-path)
  - `.tg-bubble::before` — SVG tail (left-bottom)
  - `.tg-bubble-media` — медиа в пузыре (overflow hidden, border-radius)
  - `.tg-sticker` — стикер без пузыря (160x160)
  - `.tg-inline-btn` — кнопка под сообщением (iOS blue)
  - `.tg-timestamp` — время в правом нижнем углу пузыря
  - `.tg-caption` — текст под медиа
  - Шрифт: `-apple-system, 'SF Pro Text', 'Helvetica Neue', sans-serif`
  - Размеры как в iOS Telegram: 16px текст, 21px line-height
- **Зависимости**: нет
- **Проверка**: visual review

### 3.3 TelegramHeader
- **Файл**: `frontend/src/features/telegram-preview/components/TelegramHeader.tsx` (новый)
- **Что**:
  - Props: `{botName: string, botAvatar?: string, online?: boolean}`
  - Рендер: аватар (круг 36px, fallback на первую букву) + имя + "bot" badge + "online"/"last seen"
  - Стиль: iOS навбар (blur background, border-bottom)
  - Back arrow (декоративная, не кликабельная)
- **Зависимости**: 3.2 (стили)
- **Проверка**: Storybook/visual

### 3.4 MessageBubble
- **Файл**: `frontend/src/features/telegram-preview/components/MessageBubble.tsx` (новый)
- **Что**:
  - Props: `{text: string, parseMode?: string, timestamp?: string, isFirst?: boolean}`
  - Рендер пузыря с текстом
  - Markdown → HTML (простой: **bold**, *italic*, `code`, [links](url))
  - Timestamp в правом нижнем углу (серый, мелкий)
  - Tail только на isFirst (первое сообщение в группе)
  - Max-width: 85%
- **Зависимости**: 3.2
- **Проверка**: render test

### 3.5 MediaMessage
- **Файл**: `frontend/src/features/telegram-preview/components/MediaMessage.tsx` (новый)
- **Что**:
  - Props: `{part: MessagePart, timestamp?: string, isFirst?: boolean}`
  - Photo: `<img>` с `object-fit: cover`, max-height 300px, border-radius
  - Video: thumbnail placeholder с play button
  - Document: иконка файла + имя + размер
  - Caption текст под медиа (внутри пузыря)
  - Пузырь без padding сверху если медиа (край в край)
- **Зависимости**: 3.2, 3.4 (для caption)
- **Проверка**: render test

### 3.6 StickerMessage
- **Файл**: `frontend/src/features/telegram-preview/components/StickerMessage.tsx` (новый)
- **Что**:
  - Props: `{mediaUrl: string}`
  - Рендер: `<img>` 160x160px, без пузыря, без тени
  - WebP поддержка (нативная в современных браузерах)
  - Fallback: placeholder если изображение не загрузилось
- **Зависимости**: 3.2
- **Проверка**: render test

### 3.7 InlineKeyboard
- **Файл**: `frontend/src/features/telegram-preview/components/InlineKeyboard.tsx` (новый)
- **Что**:
  - Props: `{buttons: InlineButton[][]}`
  - Рендер: ряды кнопок, каждый ряд — flex row
  - Стиль: iOS inline buttons (голубой текст, прозрачный фон, border-radius 8px)
  - URL кнопки: иконка внешней ссылки
  - Не кликабельные (preview mode)
- **Зависимости**: 3.2
- **Проверка**: render test

### 3.8 PhoneFrame
- **Файл**: `frontend/src/features/telegram-preview/components/PhoneFrame.tsx` (новый)
- **Что**:
  - Props: `{children: ReactNode, className?: string}`
  - CSS рамка iPhone (border-radius, notch, status bar)
  - Aspect ratio ~9:19.5
  - Тёмная рамка, внутри children
  - Status bar: время, wifi, battery (статичные SVG)
  - Responsive: масштабируется по высоте контейнера
- **Зависимости**: нет
- **Проверка**: visual

### 3.9 TelegramPreview (главный компонент)
- **Файл**: `frontend/src/features/telegram-preview/components/TelegramPreview.tsx` (новый)
- **Что**:
  - Props: `TelegramPreviewProps` (botName, botAvatar, content, showFrame, theme)
  - Композиция:
    ```tsx
    <MaybePhoneFrame show={showFrame}>
      <div className="tg-chat-bg flex flex-col h-full">
        <TelegramHeader botName={...} botAvatar={...} />
        <div className="flex-1 overflow-y-auto p-4 space-y-1">
          {content.parts.map((part, i) => (
            <MessagePartRenderer key={i} part={part} isLast={i === parts.length - 1} />
          ))}
          {content.buttons && <InlineKeyboard buttons={content.buttons} />}
        </div>
      </div>
    </MaybePhoneFrame>
    ```
  - `MessagePartRenderer` — switch по part.type → MessageBubble / MediaMessage / StickerMessage
  - Auto-scroll to bottom при изменении content
  - Timestamp: текущее время (для реализма)
- **Зависимости**: 3.3-3.8 (все подкомпоненты)
- **Проверка**: render test, visual review

### 3.10 Export index
- **Файл**: `frontend/src/features/telegram-preview/index.ts` (новый)
- **Что**: export TelegramPreview, types, MessageContentEditor (после Фазы 4)
- **Зависимости**: 3.9

### 3.11 Тесты Phase 3
- **Файлы**:
  - `frontend/src/features/telegram-preview/__tests__/TelegramPreview.test.tsx`
  - Test cases: text-only, photo+caption, sticker, buttons, multi-part, empty content
- **Проверка**: `npx vitest run --reporter=verbose`

### Валидация Фазы 3
```bash
cd frontend && npm run build   # TypeScript компиляция без ошибок
npx vitest run                 # Unit тесты
npm run lint                   # ESLint
```

---

## Фаза 4: Message Content Editor (Frontend)

**Цель**: Визуальный редактор составных сообщений с drag & drop.

**Предусловия**: Фаза 3 завершена (использует TelegramPreview для live preview).

### 4.1 Зависимость: dnd-kit
- **Команда**: `cd frontend && npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`
- **Проверка**: `npm ls @dnd-kit/core`

### 4.2 Part Type Selector
- **Файл**: `frontend/src/features/telegram-preview/components/editor/PartTypeSelector.tsx` (новый)
- **Что**:
  - Dropdown/Popover с типами: Текст, Фото, Видео, Документ, Стикер, GIF
  - Иконки для каждого типа (из lucide-react)
  - `onSelect(type: MessagePartType)` callback
  - Визуально: кнопка "+" → выпадающий список типов
- **Зависимости**: shadcn/ui (Popover, Button)
- **Проверка**: render test

### 4.3 TextPartEditor
- **Файл**: `frontend/src/features/telegram-preview/components/editor/TextPartEditor.tsx` (новый)
- **Что**:
  - Props: `{value: MessagePart, onChange: (part: MessagePart) => void}`
  - Textarea для текста
  - Markdown toolbar: Bold (**), Italic (*), Code (`), Link ([text](url))
  - Toolbar вставляет обёртку вокруг выделенного текста
  - ParseMode selector (Markdown / HTML / None)
  - Character count (Telegram limit: 4096)
- **Зависимости**: shadcn/ui (Textarea, Toggle, Tooltip)
- **Проверка**: render test

### 4.4 MediaPartEditor
- **Файл**: `frontend/src/features/telegram-preview/components/editor/MediaPartEditor.tsx` (новый)
- **Что**:
  - Props: `{value: MessagePart, onChange, onUpload}`
  - Dropzone для загрузки файла (drag & drop файла + click)
  - Превью загруженного изображения (thumbnail)
  - Caption textarea (опционально)
  - Вызывает `onUpload(file)` → загрузка в MinIO → получение URL → onChange({...part, mediaUrl})
  - Индикатор прогресса загрузки
  - Валидация: max 50MB, разрешённые типы по part.type
- **Зависимости**: shadcn/ui, существующий file upload API
- **Проверка**: render test

### 4.5 StickerPicker
- **Файл**: `frontend/src/features/telegram-preview/components/editor/StickerPicker.tsx` (новый)
- **Что**:
  - Tab 1: Upload .webp (dropzone, preview)
  - Tab 2: Gallery (предустановленные стикеры — 20-30 в assets)
  - Клик по стикеру → onChange({type: 'sticker', mediaUrl: url})
  - Предустановленные стикеры: `/public/stickers/` (wave, heart, fire, party, etc.)
- **Зависимости**: shadcn/ui (Tabs)
- **Проверка**: render test

### 4.6 ButtonEditor
- **Файл**: `frontend/src/features/telegram-preview/components/editor/ButtonEditor.tsx` (новый)
- **Что**:
  - Props: `{value: InlineButton[][], onChange}`
  - Dynamic rows: "Добавить ряд кнопок"
  - Каждый ряд: Input (text) + Input (URL) + удалить
  - Max 8 кнопок в ряд, max 5 рядов
  - Drag & drop рядов (dnd-kit sortable)
- **Зависимости**: dnd-kit, shadcn/ui (Input, Button)
- **Проверка**: render test

### 4.7 MessageContentEditor (главный)
- **Файл**: `frontend/src/features/telegram-preview/components/MessageContentEditor.tsx` (новый)
- **Что**:
  - Props: `MessageContentEditorProps` (value, onChange, allowStickers, allowMedia, allowButtons, maxParts)
  - Layout: список parts (sortable via dnd-kit) + "Добавить блок" + buttons section
  - Каждый part: drag handle + type badge + editor component + delete button
  - Switch по part.type → TextPartEditor / MediaPartEditor / StickerPicker
  - DndContext + SortableContext для reorder
  - Кнопки: ButtonEditor (отдельная секция внизу)
  - Валидация: maxParts, обязательные поля
- **Зависимости**: 4.2-4.6, dnd-kit
- **Проверка**: render test, interaction test

### 4.8 API Types + Hooks
- **Файл**: `frontend/src/features/bots/types.ts` (изменение)
- **Что**:
  - Добавить в `BotSettings`:
    ```ts
    welcome_content?: MessageContent;
    ```
  - Import MessageContent из telegram-preview/types
- **Файл**: `frontend/src/features/campaigns/types.ts` (изменение)
- **Что**:
  - Добавить в Campaign: `content?: MessageContent`
- **Файл**: `frontend/src/features/bots/api.ts` (изменение)
- **Что**:
  - `updateBotSettings()` — уже работает с JSONB, content пройдёт автоматически
- **Зависимости**: 3.1 (types)
- **Проверка**: TypeScript compilation

### Валидация Фазы 4
```bash
cd frontend && npm run build
npx vitest run
npm run lint
```

---

## Фаза 5: Интеграция в UI

**Цель**: Подключить Editor + Preview в существующие страницы.

**Предусловия**: Фазы 2 + 4 завершены.

### 5.1 Bot Settings: WelcomeMessageEditor
- **Файл**: найти текущий BotSettingsPage и добавить WelcomeMessageEditor
- **Что**:
  - Заменить textarea для welcome_message на:
    ```tsx
    <div className="grid grid-cols-2 gap-6">
      <MessageContentEditor value={welcomeContent} onChange={setWelcomeContent} allowStickers allowMedia allowButtons />
      <TelegramPreview botName={bot.username} botAvatar={bot.avatarUrl} content={welcomeContent} showFrame />
    </div>
    ```
  - Save → PATCH /api/v1/bots/:id/settings с welcome_content
  - Backend → Redis event → Bot hot-reload
- **Зависимости**: Фаза 2 (backend API), Фаза 4 (editor)
- **Проверка**: manual test + Playwright

### 5.2 Campaign Editor: content вместо message+media_url
- **Файл**: найти текущий CampaignEditor
- **Что**:
  - Заменить раздельные поля (message textarea + media upload) на MessageContentEditor
  - Рядом — TelegramPreview
  - Save → POST/PUT /api/v1/campaigns с content JSONB
  - Fallback: если campaign.content is null → показать legacy поля
- **Зависимости**: Фаза 2, Фаза 4
- **Проверка**: manual test + Playwright

### 5.3 Auto-Scenario Editor: content
- **Файл**: найти текущий ScenarioEditor
- **Что**:
  - Аналогично 5.2: заменить message textarea на MessageContentEditor + Preview
- **Зависимости**: Фаза 2, Фаза 4
- **Проверка**: manual test

### 5.4 Admin Bot Page
- **Файлы**: новая страница в frontend
- **Что**:
  - Route: `/dashboard/bots/admin`
  - Отображение: статус привязки, telegram_id, org_id
  - Действия: генерация link code, отвязка
  - Использует API: GET /api/v1/admin-bot/status, POST /api/v1/admin-bot/code, DELETE /api/v1/admin-bot/link
- **Зависимости**: существующий API adminbot group
- **Проверка**: manual test

### 5.5 E2E тесты
- **Файлы**: `e2e/tests/bot-settings.spec.ts` (новый)
- **Что**:
  - Test: открыть настройки бота → изменить welcome message → сохранить → проверить preview
  - Test: добавить медиа → проверить что preview обновился
  - Test: drag & drop parts → проверить порядок
  - Test: campaign editor с content
- **Зависимости**: всё остальное
- **Проверка**: `cd e2e && npx playwright test`

### Валидация Фазы 5
```bash
# Full stack проверка
cd frontend && npm run build
cd backend && go test -race ./...
cd e2e && npx playwright test bot-settings
```

---

## Параллелизация

```
Неделя 1:
  ├── [Backend]  Фаза 1: Event Bus (1.1-1.8)
  └── [Frontend] Фаза 3: TelegramPreview (3.1-3.11)  ← параллельно!

Неделя 2:
  ├── [Backend]  Фаза 2: Составные сообщения (2.1-2.9)
  └── [Frontend] Фаза 4: Content Editor (4.1-4.8)     ← параллельно!

Неделя 3:
  └── [Full-stack] Фаза 5: Интеграция (5.1-5.5)
```

---

## Чеклист перед деплоем

- [ ] Все unit тесты зелёные: `go test -race ./...` + `npx vitest run`
- [ ] Lint чистый: `go vet ./...` + `npm run lint`
- [ ] Миграция 00032 протестирована на dev-базе
- [ ] Миграция обратно-совместима (goose down работает)
- [ ] Redis Pub/Sub работает между server и bot процессами
- [ ] Hot reload: изменение кнопок в UI → бот обновился за <2 сек
- [ ] Welcome message: стикер + фото + текст отправляется корректно
- [ ] Campaign sender: фото/видео/документ отправляются
- [ ] TelegramPreview визуально соответствует iOS Telegram
- [ ] E2E тесты bot-settings проходят
- [ ] Production deploy: `infra/scripts/deploy.sh backend bot frontend`
