# Bot Architecture v2 — Design Document

> Дата: 2026-04-01
> Статус: Draft
> Область: Telegram-боты, UI Preview, Runtime-обновления, Составные сообщения

---

## 1. Обзор

Текущая архитектура ботов имеет три критических ограничения:

1. **Нет runtime-обновлений**: API-сервер и бот-сервис — изолированные процессы. Изменение настроек бота (кнопки, welcome message) не доходит до работающего бота до рестарта.
2. **Только текстовые сообщения**: Welcome message — строка. Campaign sender вызывает только `SendMessage()`, игнорируя `MediaURL`. Нет поддержки стикеров, видео, документов.
3. **Нет live-превью**: UI настройки ботов — формы без визуализации результата. Пользователь не видит, как сообщение будет выглядеть в Telegram.

Этот документ описывает целевую архитектуру, модели данных, паттерны и план реализации.

---

## 2. Архитектура системы

### 2.1. Текущее состояние

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  API Server  │────▶│  PostgreSQL   │
│  (React SPA) │     │  cmd/server  │     └──────────────┘
└─────────────┘     └──────────────┘
                                          ┌──────────────┐
                    ┌──────────────┐      │    Redis      │
                    │  Bot Service │─────▶│  (sessions,   │
                    │   cmd/bot    │      │   queue)      │
                    └──────────────┘      └──────────────┘
                    ┌──────────────┐
                    │  Admin Bot   │      ⚠ Нет связи между
                    │ cmd/admin-bot│         API и Bot Service
                    └──────────────┘
```

**Проблема**: API Server пишет в PostgreSQL, Bot Service читает из PostgreSQL только при старте. Изменения не доходят.

### 2.2. Целевая архитектура

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  API Server  │────▶│  PostgreSQL   │
│  (React SPA) │     │  cmd/server  │     └──────────────┘
└─────────────┘     └──────┬───────┘
                           │ Redis Publish
                           ▼
                    ┌──────────────┐     ┌──────────────┐
                    │    Redis     │     │  Bot Service  │
                    │  Pub/Sub +   │────▶│  cmd/bot      │
                    │  Queue       │     │  (subscribes) │
                    └──────────────┘     └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  Admin Bot   │
                    │ cmd/admin-bot│
                    └──────────────┘
```

**Ключевое изменение**: Redis Pub/Sub как event bus между процессами.

---

## 3. Redis Event Bus

### 3.1. Каналы

| Канал | Формат payload | Подписчик |
|-------|---------------|-----------|
| `revisitr:bot:reload` | `{"bot_id": 1}` | Bot Service (Manager) |
| `revisitr:bot:stop` | `{"bot_id": 1}` | Bot Service (Manager) |
| `revisitr:bot:start` | `{"bot_id": 1}` | Bot Service (Manager) |
| `revisitr:bot:settings` | `{"bot_id": 1, "field": "welcome"}` | Bot Service (Manager) |
| `revisitr:campaign:send` | `{"campaign_id": 5}` | Bot Service (Sender) |

### 3.2. Backend: Publisher (API Server)

```go
// internal/service/eventbus/eventbus.go

package eventbus

type EventBus struct {
    rds *goredis.Client
}

type BotEvent struct {
    BotID int    `json:"bot_id"`
    Field string `json:"field,omitempty"` // "welcome", "buttons", "modules", ""
}

func (eb *EventBus) PublishBotReload(ctx context.Context, botID int) error {
    return eb.publish(ctx, "revisitr:bot:reload", BotEvent{BotID: botID})
}

func (eb *EventBus) PublishBotSettings(ctx context.Context, botID int, field string) error {
    return eb.publish(ctx, "revisitr:bot:settings", BotEvent{BotID: botID, Field: field})
}
```

### 3.3. Backend: Subscriber (Bot Service)

```go
// internal/service/eventbus/subscriber.go

type Subscriber struct {
    rds     *goredis.Client
    logger  *slog.Logger
}

type BotEventHandler interface {
    OnBotReload(ctx context.Context, botID int) error
    OnBotStop(ctx context.Context, botID int) error
    OnBotStart(ctx context.Context, botID int) error
    OnBotSettingsChanged(ctx context.Context, botID int, field string) error
}

func (s *Subscriber) Listen(ctx context.Context, handler BotEventHandler) {
    pubsub := s.rds.Subscribe(ctx,
        "revisitr:bot:reload",
        "revisitr:bot:stop",
        "revisitr:bot:start",
        "revisitr:bot:settings",
    )
    defer pubsub.Close()

    ch := pubsub.Channel()
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-ch:
            s.dispatch(ctx, msg, handler)
        }
    }
}
```

### 3.4. Интеграция с Manager

Manager уже имеет `AddBot()`, `RemoveBot()`, `ReloadBot()` — нужно только подключить к event bus:

```go
// Manager implements BotEventHandler
func (m *Manager) OnBotReload(ctx context.Context, botID int) error {
    return m.ReloadBot(ctx, botID) // уже существует
}

func (m *Manager) OnBotSettingsChanged(ctx context.Context, botID int, field string) error {
    // Лёгкое обновление: перечитать настройки без перезапуска long polling
    m.mu.Lock()
    defer m.mu.Unlock()

    inst, ok := m.instances[botID]
    if !ok {
        return nil // бот не запущен
    }

    bot, err := m.botsRepo.GetByID(ctx, botID)
    if err != nil {
        return err
    }

    inst.info = *bot // обновить frozen config
    return nil
}
```

### 3.5. Горячее обновление vs полный перезапуск

| Изменение | Действие | Почему |
|-----------|----------|--------|
| Кнопки, welcome message, модули | Hot update (`OnBotSettingsChanged`) | Не требует переподключения к Telegram |
| Токен бота | Full reload (`OnBotReload`) | Нужен новый `telego.Bot` инстанс |
| Статус: active → inactive | Stop (`OnBotStop`) | Прекратить long polling |
| Статус: inactive → active | Start (`OnBotStart`) | Запустить long polling |
| Удаление бота | Stop + cleanup | Остановить и удалить инстанс |

### 3.6. Потоки данных при обновлении настроек

```
Пользователь меняет кнопки бота в UI
    │
    ▼
Frontend: PATCH /api/v1/bots/1/settings
    │
    ▼
API Server:
  1. botsUsecase.UpdateSettings(id, newSettings)
  2. botsRepo.UpdateSettings(id, newSettings)  → PostgreSQL
  3. eventBus.PublishBotSettings(ctx, id, "buttons")  → Redis
    │
    ▼
Redis Pub/Sub: "revisitr:bot:settings" → {"bot_id": 1, "field": "buttons"}
    │
    ▼
Bot Service (subscriber goroutine):
  1. dispatch → handler.OnBotSettingsChanged(ctx, 1, "buttons")
  2. Manager: перечитать bot из БД → обновить inst.info
  3. Следующее сообщение пользователю уже использует новые кнопки
```

---

## 4. Модель составного сообщения (MessageContent)

### 4.1. Структура данных

```go
// internal/entity/message.go

// MessagePartType определяет тип части сообщения.
type MessagePartType string

const (
    PartText      MessagePartType = "text"
    PartPhoto     MessagePartType = "photo"
    PartVideo     MessagePartType = "video"
    PartDocument  MessagePartType = "document"
    PartAnimation MessagePartType = "animation" // GIF
    PartSticker   MessagePartType = "sticker"
    PartAudio     MessagePartType = "audio"
    PartVoice     MessagePartType = "voice"
)

// MessagePart — одна часть составного сообщения.
// Каждый part = один вызов Telegram API.
type MessagePart struct {
    Type      MessagePartType `json:"type"`
    Text      string          `json:"text,omitempty"`      // Текст или caption
    MediaURL  string          `json:"media_url,omitempty"` // URL файла (MinIO)
    MediaID   string          `json:"media_id,omitempty"`  // Telegram file_id (кеш)
    ParseMode string          `json:"parse_mode,omitempty"` // "Markdown", "HTML", ""
}

// InlineButton — кнопка под сообщением.
type InlineButton struct {
    Text string `json:"text"`
    URL  string `json:"url,omitempty"`
    Data string `json:"data,omitempty"` // callback_data
}

// MessageContent — полное описание составного сообщения.
// Хранится как JSONB в PostgreSQL.
type MessageContent struct {
    Parts   []MessagePart    `json:"parts"`
    Buttons [][]InlineButton `json:"buttons,omitempty"` // ряды кнопок
}
```

### 4.2. SQL/JSONB Scan/Value

```go
func (mc *MessageContent) Scan(src interface{}) error {
    // Стандартный паттерн из проекта (как TriggerConfig, BotSettings)
    switch v := src.(type) {
    case []byte:
        return json.Unmarshal(v, mc)
    case string:
        return json.Unmarshal([]byte(v), mc)
    case nil:
        *mc = MessageContent{}
        return nil
    default:
        return fmt.Errorf("MessageContent.Scan: unsupported type %T", src)
    }
}

func (mc MessageContent) Value() (driver.Value, error) {
    b, err := json.Marshal(mc)
    if err != nil {
        return nil, fmt.Errorf("MessageContent.Value: %w", err)
    }
    return b, nil
}
```

### 4.3. Примеры данных

**Welcome message: стикер → фото с текстом → кнопки**
```json
{
  "parts": [
    {
      "type": "sticker",
      "media_url": "/revisitr/storage/stickers/welcome-wave.webp"
    },
    {
      "type": "photo",
      "text": "Добро пожаловать в *Кофейня Sunrise*! ☕\n\nТеперь вы участник нашей программы лояльности.",
      "media_url": "/revisitr/storage/bots/1/welcome-hero.jpg",
      "parse_mode": "Markdown"
    }
  ],
  "buttons": [
    [
      {"text": "📋 Наше меню", "url": "https://sunrise-cafe.ru/menu"},
      {"text": "📍 Как нас найти", "url": "https://maps.google.com/..."}
    ]
  ]
}
```

**Рассылка: текст с фото**
```json
{
  "parts": [
    {
      "type": "photo",
      "text": "🔥 Только сегодня: двойные баллы за любой заказ!\n\nПокажи это сообщение на кассе.",
      "media_url": "/revisitr/storage/campaigns/55/promo-banner.jpg",
      "parse_mode": "Markdown"
    }
  ],
  "buttons": [
    [{"text": "Подробности", "url": "https://..."}]
  ]
}
```

**Простое текстовое сообщение (обратная совместимость)**
```json
{
  "parts": [
    {
      "type": "text",
      "text": "Спасибо за визит! Вам начислено 50 баллов.",
      "parse_mode": "Markdown"
    }
  ]
}
```

### 4.4. Правила валидации

```go
func (mc MessageContent) Validate() error {
    if len(mc.Parts) == 0 {
        return errors.New("message must have at least one part")
    }
    if len(mc.Parts) > 5 {
        return errors.New("message cannot have more than 5 parts")
    }

    for i, p := range mc.Parts {
        switch p.Type {
        case PartText:
            if p.Text == "" {
                return fmt.Errorf("part %d: text part must have text", i)
            }
            if p.MediaURL != "" {
                return fmt.Errorf("part %d: text part cannot have media", i)
            }
        case PartPhoto, PartVideo, PartDocument, PartAnimation, PartAudio:
            if p.MediaURL == "" && p.MediaID == "" {
                return fmt.Errorf("part %d: media part must have media_url or media_id", i)
            }
            // caption (text) опционален
        case PartSticker:
            if p.MediaURL == "" && p.MediaID == "" {
                return fmt.Errorf("part %d: sticker must have media_url or media_id", i)
            }
            if p.Text != "" {
                return fmt.Errorf("part %d: stickers cannot have captions", i)
            }
        default:
            return fmt.Errorf("part %d: unknown type %q", i, p.Type)
        }
    }

    // Кнопки допускаются только после последнего part
    for _, row := range mc.Buttons {
        if len(row) > 8 {
            return errors.New("button row cannot have more than 8 buttons")
        }
        for _, btn := range row {
            if btn.Text == "" {
                return errors.New("button must have text")
            }
        }
    }

    return nil
}
```

---

## 5. Telegram Message Sender

### 5.1. Универсальный отправщик

```go
// internal/service/telegram/sender.go

package telegram

// Sender отправляет MessageContent через Telegram Bot API.
// Используется в: campaign sender, welcome message, auto-actions, admin bot.
type Sender struct {
    baseURL string // Base URL для media (например "https://elysium.fm")
    logger  *slog.Logger
}

// SendContent отправляет составное сообщение (все parts последовательно).
func (s *Sender) SendContent(ctx context.Context, bot *telego.Bot, chatID int64, content entity.MessageContent) error {
    for i, part := range content.Parts {
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // Кнопки прикрепляются только к последнему part
        var markup *telego.InlineKeyboardMarkup
        if i == len(content.Parts)-1 && len(content.Buttons) > 0 {
            markup = s.buildInlineKeyboard(content.Buttons)
        }

        if err := s.sendPart(bot, chatID, part, markup); err != nil {
            return fmt.Errorf("send part %d (%s): %w", i, part.Type, err)
        }
    }
    return nil
}

func (s *Sender) sendPart(
    bot *telego.Bot,
    chatID int64,
    part entity.MessagePart,
    markup *telego.InlineKeyboardMarkup,
) error {
    switch part.Type {
    case entity.PartText:
        msg := tu.Message(tu.ID(chatID), part.Text)
        if part.ParseMode != "" {
            msg = msg.WithParseMode(part.ParseMode)
        }
        if markup != nil {
            msg = msg.WithReplyMarkup(markup)
        }
        _, err := bot.SendMessage(msg)
        return err

    case entity.PartPhoto:
        photo := tu.Photo(tu.ID(chatID), s.mediaInput(part))
        if part.Text != "" {
            photo = photo.WithCaption(part.Text)
        }
        if part.ParseMode != "" {
            photo = photo.WithParseMode(part.ParseMode)
        }
        if markup != nil {
            photo = photo.WithReplyMarkup(markup)
        }
        _, err := bot.SendPhoto(photo)
        return err

    case entity.PartVideo:
        video := tu.Video(tu.ID(chatID), s.mediaInput(part))
        if part.Text != "" {
            video = video.WithCaption(part.Text)
        }
        if part.ParseMode != "" {
            video = video.WithParseMode(part.ParseMode)
        }
        if markup != nil {
            video = video.WithReplyMarkup(markup)
        }
        _, err := bot.SendVideo(video)
        return err

    case entity.PartDocument:
        doc := tu.Document(tu.ID(chatID), s.mediaInput(part))
        if part.Text != "" {
            doc = doc.WithCaption(part.Text)
        }
        if markup != nil {
            doc = doc.WithReplyMarkup(markup)
        }
        _, err := bot.SendDocument(doc)
        return err

    case entity.PartSticker:
        sticker := tu.Sticker(tu.ID(chatID), s.mediaInput(part))
        // Стикеры не поддерживают caption и parse_mode
        _, err := bot.SendSticker(sticker)
        return err

    case entity.PartAnimation:
        anim := tu.Animation(tu.ID(chatID), s.mediaInput(part))
        if part.Text != "" {
            anim = anim.WithCaption(part.Text)
        }
        if markup != nil {
            anim = anim.WithReplyMarkup(markup)
        }
        _, err := bot.SendAnimation(anim)
        return err

    default:
        return fmt.Errorf("unsupported part type: %s", part.Type)
    }
}

// mediaInput выбирает FileFromID (если есть кеш) или FileFromURL.
func (s *Sender) mediaInput(part entity.MessagePart) telego.InputFile {
    if part.MediaID != "" {
        return tu.FileFromID(part.MediaID)
    }
    // Превращаем относительный URL в абсолютный
    url := part.MediaURL
    if !strings.HasPrefix(url, "http") {
        url = s.baseURL + url
    }
    return tu.FileFromURL(url)
}

func (s *Sender) buildInlineKeyboard(buttons [][]entity.InlineButton) *telego.InlineKeyboardMarkup {
    var rows [][]telego.InlineKeyboardButton
    for _, row := range buttons {
        var tgRow []telego.InlineKeyboardButton
        for _, btn := range row {
            if btn.URL != "" {
                tgRow = append(tgRow, tu.InlineKeyboardButton(btn.Text).WithURL(btn.URL))
            } else if btn.Data != "" {
                tgRow = append(tgRow, tu.InlineKeyboardButton(btn.Text).WithCallbackData(btn.Data))
            }
        }
        rows = append(rows, tgRow)
    }
    kb := tu.InlineKeyboard(rows...)
    return kb
}
```

### 5.2. Интеграция с существующим кодом

**Campaign Sender** — заменить прямой `SendMessage` на `telegram.Sender.SendContent`:
```go
// Было:
tgMsg := tu.Message(tu.ID(messages[i].TelegramID), campaign.Message)
_, err := tBot.SendMessage(tgMsg)

// Стало:
err := tgSender.SendContent(ctx, tBot, messages[i].TelegramID, campaign.Content)
```

**Welcome Message** — заменить `sendText(chatID, h.info.Settings.WelcomeMessage)`:
```go
// Было:
h.sendText(chatID, h.info.Settings.WelcomeMessage)

// Стало:
if h.info.Settings.WelcomeContent != nil {
    h.tgSender.SendContent(ctx, h.bot, chatID, *h.info.Settings.WelcomeContent)
} else if h.info.Settings.WelcomeMessage != "" {
    // Fallback для старых ботов без миграции
    h.sendText(chatID, h.info.Settings.WelcomeMessage)
}
```

**Auto-Scenario Scheduler** — аналогично кампаниям.

---

## 6. Миграция базы данных

### 6.1. Стратегия: Мягкая миграция

Сохраняем старые поля, добавляем новые. Fallback на старые поля если `content` is NULL.

### 6.2. SQL миграция

```sql
-- 00032_composite_messages.sql

-- +goose Up

-- Добавить JSONB поле для составных сообщений в campaigns
ALTER TABLE campaigns ADD COLUMN content JSONB;

-- Добавить JSONB поле для welcome content в bots.settings
-- (settings уже JSONB, добавляем welcome_content как новый ключ)
-- Миграция данных не нужна — BotSettings расширяется в Go-коде

-- Добавить content в campaign_templates
ALTER TABLE campaign_templates ADD COLUMN content JSONB;

-- Добавить content в campaign_variants (A/B testing)
ALTER TABLE campaign_variants ADD COLUMN content JSONB;

-- Добавить content в auto_scenarios
ALTER TABLE auto_scenarios ADD COLUMN content JSONB;

-- Мигрировать существующие текстовые кампании в новый формат
UPDATE campaigns
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        CASE
            WHEN media_url IS NOT NULL AND media_url != '' THEN
                jsonb_build_object(
                    'type', 'photo',
                    'text', message,
                    'media_url', media_url,
                    'parse_mode', 'Markdown'
                )
            ELSE
                jsonb_build_object(
                    'type', 'text',
                    'text', message,
                    'parse_mode', 'Markdown'
                )
        END
    ),
    'buttons', COALESCE(buttons, '[]'::jsonb)
)
WHERE content IS NULL AND message != '';

-- Аналогично для campaign_templates
UPDATE campaign_templates
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        CASE
            WHEN media_url IS NOT NULL AND media_url != '' THEN
                jsonb_build_object('type', 'photo', 'text', message, 'media_url', media_url, 'parse_mode', 'Markdown')
            ELSE
                jsonb_build_object('type', 'text', 'text', message, 'parse_mode', 'Markdown')
        END
    ),
    'buttons', COALESCE(buttons, '[]'::jsonb)
)
WHERE content IS NULL AND message != '';

-- Для auto_scenarios
UPDATE auto_scenarios
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        jsonb_build_object('type', 'text', 'text', message, 'parse_mode', 'Markdown')
    )
)
WHERE content IS NULL AND message != '';

-- +goose Down
ALTER TABLE campaigns DROP COLUMN IF EXISTS content;
ALTER TABLE campaign_templates DROP COLUMN IF EXISTS content;
ALTER TABLE campaign_variants DROP COLUMN IF EXISTS content;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS content;
```

### 6.3. Обновление entity

```go
// entity/bot.go — расширение BotSettings
type BotSettings struct {
    Modules          []string        `json:"modules"`
    Buttons          []BotButton     `json:"buttons"`
    RegistrationForm []FormField     `json:"registration_form"`
    WelcomeMessage   string          `json:"welcome_message"`           // Legacy
    WelcomeContent   *MessageContent `json:"welcome_content,omitempty"` // New
}

// entity/campaign.go — расширение Campaign
type Campaign struct {
    // ... existing fields ...
    Message  string          `db:"message" json:"message"`       // Legacy
    MediaURL *string         `db:"media_url" json:"media_url"`   // Legacy
    Content  *MessageContent `db:"content" json:"content"`       // New
}
```

---

## 7. Frontend: Telegram Preview Component

### 7.1. Решение

Готовых npm-пакетов нет. Строим кастомный `<TelegramPreview>` на Tailwind + shadcn/ui.

Референсы:
- iOS Chat Bubbles CSS (samuelkraft.com)
- Telegram Desktop CSS (CodePen: sattellite/pen/bRaBEx)
- TelegramUI (github.com/telegram-mini-apps-dev/TelegramUI) — для иконок/типографики

### 7.2. Компонентная архитектура

```
frontend/src/features/telegram-preview/
├── components/
│   ├── TelegramPreview.tsx      — Контейнер: рамка чата, header, скролл
│   ├── TelegramHeader.tsx       — Аватар + имя бота + статус
│   ├── MessageBubble.tsx        — Пузырь сообщения (text, timestamp, tail)
│   ├── MediaMessage.tsx         — Фото/видео/документ в пузыре
│   ├── StickerMessage.tsx       — Стикер (без пузыря, inline)
│   ├── InlineKeyboard.tsx       — Кнопки под сообщением
│   ├── TypingIndicator.tsx      — "typing..." анимация (опционально)
│   └── PhoneFrame.tsx           — iPhone рамка (опционально)
├── styles/
│   └── telegram.css             — Telegram-специфичные стили
├── hooks/
│   └── useBotInfo.ts            — Загрузка аватара и username бота
└── index.ts                     — Публичный API
```

### 7.3. API компонента

```tsx
interface TelegramPreviewProps {
  botName: string;
  botAvatar?: string;     // URL аватара
  content: MessageContent; // Составное сообщение
  className?: string;
  showFrame?: boolean;     // iPhone рамка
  theme?: 'light' | 'dark';
}

// Использование:
<TelegramPreview
  botName={bot.username}
  botAvatar={bot.avatarUrl}
  content={welcomeContent}
  showFrame
  theme="light"
/>
```

### 7.4. Стили пузырей (iOS Telegram)

```css
/* Telegram-style message bubble (incoming = от бота) */
.tg-bubble {
  position: relative;
  max-width: 85%;
  padding: 8px 12px;
  border-radius: 16px 16px 16px 4px;  /* tail bottom-left */
  background: #FFFFFF;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
  font-family: -apple-system, 'SF Pro Text', 'Helvetica Neue', sans-serif;
  font-size: 16px;
  line-height: 21px;
  color: #000000;
}

/* Tail (SVG pseudo-element) */
.tg-bubble::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: -6px;
  width: 12px;
  height: 16px;
  background: #FFFFFF;
  clip-path: path('M 0 16 Q 0 0, 12 0 L 12 16 Z');
}

/* Media в пузыре */
.tg-bubble-media {
  border-radius: 12px 12px 12px 4px;
  overflow: hidden;
  margin: -8px -12px 4px -12px;
}

.tg-bubble-media img {
  width: 100%;
  max-height: 300px;
  object-fit: cover;
}

/* Стикер (без пузыря) */
.tg-sticker {
  width: 160px;
  height: 160px;
}

.tg-sticker img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

/* Inline buttons */
.tg-inline-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  border-radius: 8px;
  background: rgba(0, 122, 255, 0.08);
  color: #007AFF;
  font-size: 15px;
  font-weight: 500;
  cursor: default; /* Preview, не кликабельно */
}

/* Chat background */
.tg-chat-bg {
  background-color: #C8D1DB;
  background-image: url('/telegram-pattern.svg'); /* Optional wallpaper */
}
```

### 7.5. Паттерн "Editor + Live Preview"

```tsx
// features/bots/components/WelcomeMessageEditor.tsx

export function WelcomeMessageEditor({ botId }: { botId: number }) {
  const { data: bot } = useBotQuery(botId);
  const [content, setContent] = useState<MessageContent>(
    bot?.settings.welcome_content ?? { parts: [], buttons: [] }
  );
  const mutation = useUpdateBotSettingsMutation(botId);

  return (
    <div className="grid grid-cols-2 gap-6 h-[calc(100vh-200px)]">
      {/* Left: Editor */}
      <div className="space-y-4 overflow-y-auto">
        <MessageContentEditor
          value={content}
          onChange={setContent}
          allowStickers
          allowMedia
          allowButtons
          maxParts={5}
        />
        <Button
          onClick={() => mutation.mutate({ welcome_content: content })}
          disabled={mutation.isPending}
        >
          Сохранить
        </Button>
      </div>

      {/* Right: Live Preview */}
      <div className="sticky top-0">
        <TelegramPreview
          botName={bot?.username ?? 'Bot'}
          botAvatar={bot?.avatar_url}
          content={content}
          showFrame
        />
      </div>
    </div>
  );
}
```

### 7.6. MessageContentEditor

Компонент редактирования составного сообщения:

```tsx
// features/telegram-preview/components/MessageContentEditor.tsx

interface MessageContentEditorProps {
  value: MessageContent;
  onChange: (content: MessageContent) => void;
  allowStickers?: boolean;
  allowMedia?: boolean;
  allowButtons?: boolean;
  maxParts?: number;
}

// Функциональность:
// - Drag & drop для порядка parts (dnd-kit)
// - "Добавить блок" → выбор типа (текст, фото, видео, стикер, документ)
// - Для текста: textarea с поддержкой Markdown (toolbar: bold, italic, link)
// - Для медиа: dropzone с загрузкой в MinIO + caption textarea
// - Для стикера: file upload (.webp) или выбор из галереи
// - Для кнопок: dynamic form (text + url), drag & drop рядов
// - Удаление блоков
// - Предел: maxParts (дефолт 5)
```

### 7.7. Стикеры: источники

| Источник | Реализация | Приоритет |
|----------|------------|-----------|
| Загрузка .webp файла | File upload → MinIO | MVP |
| Предустановленная галерея | Набор 20-30 популярных стикеров в assets | MVP |
| Telegram file_id | Ввод ID, валидация через Bot API | Post-MVP |
| Поиск по стикерпакам | Telegram Bot API `getStickerSet` | Post-MVP |

---

## 8. Обновление BotSettings

### 8.1. Расширенная модель настроек

```go
type BotSettings struct {
    // Существующие поля
    Modules          []string        `json:"modules"`
    Buttons          []BotButton     `json:"buttons"`
    RegistrationForm []FormField     `json:"registration_form"`
    WelcomeMessage   string          `json:"welcome_message"` // Legacy

    // Новые поля
    WelcomeContent   *MessageContent `json:"welcome_content,omitempty"`

    // Будущее расширение (не для MVP)
    // MenuContent      *MessageContent `json:"menu_content,omitempty"`
    // AboutContent     *MessageContent `json:"about_content,omitempty"`
    // RegistrationDone *MessageContent `json:"registration_done,omitempty"`
}
```

### 8.2. Handler: логика выбора welcome message

```go
func (h *handler) getWelcomeContent() *entity.MessageContent {
    settings := h.info.Settings

    // Приоритет: новый формат → legacy → дефолт
    if settings.WelcomeContent != nil && len(settings.WelcomeContent.Parts) > 0 {
        return settings.WelcomeContent
    }

    if settings.WelcomeMessage != "" {
        return &entity.MessageContent{
            Parts: []entity.MessagePart{
                {Type: entity.PartText, Text: settings.WelcomeMessage, ParseMode: "Markdown"},
            },
        }
    }

    // Дефолтное сообщение
    return &entity.MessageContent{
        Parts: []entity.MessagePart{
            {Type: entity.PartText, Text: fmt.Sprintf("Добро пожаловать в %s!", h.info.Name)},
        },
    }
}
```

---

## 9. Файловая структура изменений

### 9.1. Backend — новые файлы

```
backend/internal/
├── entity/
│   └── message.go              — MessageContent, MessagePart, InlineButton
├── service/
│   ├── eventbus/
│   │   ├── eventbus.go         — Publisher (API → Redis)
│   │   └── subscriber.go       — Subscriber (Redis → Bot Manager)
│   └── telegram/
│       └── sender.go           — Универсальный SendContent
└── migrations/
    └── 00032_composite_messages.sql
```

### 9.2. Backend — изменяемые файлы

```
backend/
├── cmd/bot/main.go             — Добавить Redis, subscriber, telegram.Sender
├── cmd/server/main.go          — Добавить eventbus в DI
├── internal/
│   ├── entity/
│   │   ├── bot.go              — BotSettings: добавить WelcomeContent
│   │   └── campaign.go         — Campaign: добавить Content
│   ├── usecase/bots/bots.go    — UpdateSettings → publish event
│   ├── service/
│   │   ├── botmanager/
│   │   │   ├── manager.go      — Implement BotEventHandler, inject Sender
│   │   │   └── handler.go      — Использовать telegram.Sender для welcome
│   │   └── campaign/
│   │       └── sender.go       — Использовать telegram.Sender
│   └── repository/redis/
│       └── redis.go            — (без изменений, Client() уже экспортирован)
```

### 9.3. Frontend — новые файлы

```
frontend/src/features/
└── telegram-preview/
    ├── components/
    │   ├── TelegramPreview.tsx
    │   ├── TelegramHeader.tsx
    │   ├── MessageBubble.tsx
    │   ├── MediaMessage.tsx
    │   ├── StickerMessage.tsx
    │   ├── InlineKeyboard.tsx
    │   └── PhoneFrame.tsx
    ├── components/
    │   └── MessageContentEditor.tsx
    ├── styles/
    │   └── telegram.css
    ├── hooks/
    │   └── useBotInfo.ts
    ├── types.ts
    └── index.ts
```

### 9.4. Frontend — изменяемые файлы

```
frontend/src/features/
├── bots/
│   ├── types.ts                — BotSettings: добавить welcome_content
│   └── components/
│       └── BotSettingsPage.tsx  — Заменить textarea на WelcomeMessageEditor
└── campaigns/
    └── components/
        └── CampaignEditor.tsx   — Использовать MessageContentEditor + Preview
```

---

## 10. План реализации

### Фаза 1: Event Bus + Hot Reload (Backend)

**Scope**: Redis Pub/Sub, subscriber в bot service, горячее обновление настроек.

| # | Задача | Файлы | Оценка |
|---|--------|-------|--------|
| 1.1 | `entity/message.go` — MessageContent, MessagePart, InlineButton, Validate | новый | S |
| 1.2 | `service/eventbus/eventbus.go` — Publisher | новый | S |
| 1.3 | `service/eventbus/subscriber.go` — Subscriber + BotEventHandler interface | новый | M |
| 1.4 | `cmd/bot/main.go` — Redis init, subscriber goroutine | изменение | S |
| 1.5 | `cmd/server/main.go` — EventBus в DI | изменение | S |
| 1.6 | `service/botmanager/manager.go` — implement BotEventHandler | изменение | M |
| 1.7 | `usecase/bots/bots.go` — publish event после UpdateSettings | изменение | S |
| 1.8 | Тесты: eventbus unit, manager handler tests | новые | M |

### Фаза 2: Составные сообщения (Backend)

**Scope**: Миграция, telegram.Sender, обновление campaign sender и welcome handler.

| # | Задача | Файлы | Оценка |
|---|--------|-------|--------|
| 2.1 | `migrations/00032_composite_messages.sql` | новый | S |
| 2.2 | `entity/bot.go` — WelcomeContent в BotSettings | изменение | S |
| 2.3 | `entity/campaign.go` — Content в Campaign | изменение | S |
| 2.4 | `service/telegram/sender.go` — SendContent | новый | L |
| 2.5 | `service/campaign/sender.go` — использовать telegram.Sender | изменение | M |
| 2.6 | `service/botmanager/handler.go` — welcome через SendContent | изменение | M |
| 2.7 | `service/campaign/scheduler.go` — auto-scenarios через SendContent | изменение | M |
| 2.8 | `repository/redis/campaign_queue.go` — QueueMessage → MessageContent | изменение | S |
| 2.9 | Тесты: telegram.Sender unit, campaign integration | новые | L |

### Фаза 3: Telegram Preview Component (Frontend)

**Scope**: TelegramPreview, стили, базовый рендеринг.

| # | Задача | Файлы | Оценка |
|---|--------|-------|--------|
| 3.1 | `TelegramPreview.tsx` — контейнер с чат-фоном | новый | M |
| 3.2 | `TelegramHeader.tsx` — аватар + имя + статус | новый | S |
| 3.3 | `MessageBubble.tsx` — пузырь с tail, timestamp | новый | M |
| 3.4 | `MediaMessage.tsx` — фото/видео в пузыре | новый | M |
| 3.5 | `StickerMessage.tsx` — стикер без пузыря | новый | S |
| 3.6 | `InlineKeyboard.tsx` — кнопки под сообщением | новый | S |
| 3.7 | `telegram.css` — iOS-стиль пузырей | новый | M |
| 3.8 | `PhoneFrame.tsx` — iPhone рамка | новый | S |
| 3.9 | Тесты: visual snapshot tests (Vitest) | новые | M |

### Фаза 4: Message Content Editor (Frontend)

**Scope**: Редактор составных сообщений, drag & drop, загрузка медиа.

| # | Задача | Файлы | Оценка |
|---|--------|-------|--------|
| 4.1 | `MessageContentEditor.tsx` — основной компонент | новый | L |
| 4.2 | Part type selector (текст/фото/видео/стикер) | новый | M |
| 4.3 | Text part editor (Markdown toolbar) | новый | M |
| 4.4 | Media part editor (dropzone + caption) | новый | M |
| 4.5 | Sticker picker (upload + галерея) | новый | M |
| 4.6 | Button editor (dynamic rows) | новый | M |
| 4.7 | Drag & drop (dnd-kit) для порядка parts | интеграция | M |
| 4.8 | API types + hooks (frontend) | изменение | S |

### Фаза 5: Интеграция в UI

**Scope**: Подключить Preview + Editor в существующие страницы.

| # | Задача | Файлы | Оценка |
|---|--------|-------|--------|
| 5.1 | Bot Settings → WelcomeMessageEditor с превью | изменение | M |
| 5.2 | Campaign Create/Edit → MessageContentEditor с превью | изменение | M |
| 5.3 | Auto-Scenario Editor → MessageContentEditor | изменение | M |
| 5.4 | Admin Bot page в UI (настройки, привязка) | новый | L |
| 5.5 | E2E тесты: Playwright | новые | L |

---

## 11. Зависимости (npm)

Новые frontend-зависимости:

| Пакет | Назначение | Размер |
|-------|-----------|--------|
| `@dnd-kit/core` + `@dnd-kit/sortable` | Drag & drop для parts | ~15KB |
| (нет других) | Всё на Tailwind + shadcn/ui | — |

Backend зависимости: **нет новых** (go-redis уже в проекте).

---

## 12. Риски и решения

| Риск | Вероятность | Митигация |
|------|------------|-----------|
| Redis Pub/Sub message loss | Low | Некритично — worst case бот перечитает при следующем рестарте. Для гарантированной доставки можно добавить periodic DB poll (fallback). |
| Telegram rate limits при отправке составных сообщений | Medium | Добавить задержку между parts (50ms). Для массовых рассылок — parts отправляются как одна "группа" для каждого клиента. |
| Большие медиафайлы (видео) | Medium | Лимит 50MB уже в upload handler. Для Telegram Bot API лимит 50MB для отправки через URL. |
| WebP стикеры не отображаются в старых браузерах | Low | WebP поддерживается во всех современных браузерах. Fallback: показывать placeholder. |
| Обратная совместимость при миграции | Low | Soft migration: старые поля сохраняются, fallback в коде. |

---

## 13. Метрики успеха

- [ ] Изменение настроек бота через UI применяется в работающем боте за < 2 секунды
- [ ] Welcome message поддерживает стикер + фото + текст + кнопки
- [ ] Campaign sender отправляет медиа (фото, видео, документы)
- [ ] Live preview в UI точно отражает результат в Telegram (визуальное соответствие > 90%)
- [ ] Нет downtime при обновлении настроек бота

---

## Приложение A: telego API Reference

```go
// Отправка разных типов контента через telego
bot.SendMessage(msg)      // текст
bot.SendPhoto(photo)      // фото + caption
bot.SendVideo(video)      // видео + caption
bot.SendDocument(doc)     // документ + caption
bot.SendSticker(sticker)  // стикер (без caption)
bot.SendAnimation(anim)   // GIF + caption
bot.SendAudio(audio)      // аудио + caption
bot.SendVoice(voice)      // голосовое + caption

// Создание файловых инпутов
tu.FileFromURL("https://...")     // из URL (MinIO)
tu.FileFromID("AgACAgIAAxkB...")  // из кеша Telegram
tu.File(reader)                   // из io.Reader
```

## Приложение B: Redis Pub/Sub Commands

```
# Публикация (API Server)
PUBLISH revisitr:bot:reload '{"bot_id": 1}'
PUBLISH revisitr:bot:settings '{"bot_id": 1, "field": "buttons"}'

# Подписка (Bot Service)
SUBSCRIBE revisitr:bot:reload revisitr:bot:stop revisitr:bot:start revisitr:bot:settings
```
