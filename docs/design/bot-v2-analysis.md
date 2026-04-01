# Bot Architecture v2 — Анализ допущений и упрощений

> Дата: 2026-04-01
> Контекст: анализ реализации 5 фаз Bot Architecture v2 (Redis Event Bus, Composite Messages, Telegram Preview, Message Content Editor, UI Integration)

---

## 🔴 КРИТИЧНЫЕ (влияют на корректность в production)

### 1. Handler не получает обновлённые settings при hot reload

**Файлы**: `backend/internal/service/botmanager/manager.go:255`, `handler.go:22`

`OnBotSettingsChanged` обновляет `inst.info`, но `handler` создаётся в `startBot()` с **собственной копией** `entity.Bot`:

```go
handler := newHandler(m, tBot, b)  // handler.info = copy of b
// ...
inst.info = *bot  // обновляет inst, НЕ handler.info
```

Handler продолжает использовать **старые settings** (welcome_content, buttons, registration_form) до полного рестарта бота. Hot reload фактически не работает для handler logic.

**Исправление**: handler должен хранить pointer на `inst.info` или получать settings через Manager.

**Приоритет**: P0 — пользователь меняет welcome message → видит "сохранено" → бот шлёт старое.

---

### 2. Redis Pub/Sub — fire-and-forget без гарантий доставки

**Файлы**: `backend/internal/service/eventbus/eventbus.go:57`, `backend/internal/usecase/bots/bots.go`

В usecase publish-ошибки подавляются (`_ = eb.Publish...`). Redis Pub/Sub не гарантирует доставку:
- Если bot-сервис перезагружается в момент отправки — событие теряется
- Нет retry / dead letter queue
- Нет подтверждения получения

**Допущение**: потеря отдельных событий приемлема в MVP. Для production нужна очередь (Redis Streams / NATS) или retry mechanism.

**Приоритет**: P1 — редкий сценарий, но при рестарте бота все события между stop/start теряются.

---

### 3. Миграция предполагает все media = photo

**Файл**: `backend/migrations/00032_composite_messages.sql:19-26`

```sql
WHEN media_url IS NOT NULL AND media_url != '' THEN
    jsonb_build_object('type', 'photo', ...)
```

Все legacy-записи с `media_url` мигрируют как `type: 'photo'`. Если в базе есть видео, документы или GIF — тип будет некорректным. Telegram попытается отправить видео как фото и получит ошибку.

**Исправление**: определять тип по расширению файла в миграции:
```sql
CASE
  WHEN media_url ~* '\.(mp4|mov|avi)$' THEN 'video'
  WHEN media_url ~* '\.(gif)$' THEN 'animation'
  WHEN media_url ~* '\.(pdf|doc|docx|xls|xlsx)$' THEN 'document'
  ELSE 'photo'
END
```

**Приоритет**: P0 — data corruption при миграции.

---

### 4. Нет rate limiting между частями composite message

**Файл**: `backend/internal/service/telegram/sender.go:28-43`

`SendContent` отправляет parts в tight loop без пауз. Для message из 5 частей — это 5 API-вызовов подряд.

Telegram rate limit: ~30 msgs/sec per bot. Campaign sender даёт 35ms между *клиентами*, но не между частями одного сообщения.

**Расчёт**: при рассылке 5-part message на 1000 клиентов = 5000 вызовов. С 35ms между клиентами (но без пауз внутри) — burst внутри одного клиента может триггерить 429 Too Many Requests.

**Исправление**: добавить `time.Sleep(50 * time.Millisecond)` между частями в `SendContent()`.

**Приоритет**: P1 — проявляется при массовой рассылке composite messages.

---

## 🟡 СУЩЕСТВЕННЫЕ (влияют на UX или полноту)

### 5. Кампании и авто-сценарии не обновлены в UI

Wizard создания/редактирования кампаний и auto-scenarios всё ещё используют старые поля `message` + `media_url`. Новый `MessageContentEditor` интегрирован **только** в:
- Настройки бота → welcome message
- Promotions wizard (через существующую интеграцию)

**Не обновлены**:
- Создание/редактирование кампании
- Редактирование авто-сценариев
- A/B test variant editor

**Последствие**: основной use case composite messages (рассылки) недоступен через UI.

**Приоритет**: P0 для product completeness.

---

### 6. Нет рендеринга Markdown в TelegramPreview

**Файл**: `frontend/src/features/telegram-preview/components/MessageBubble.tsx`

Preview показывает raw text через `whitespace-pre-wrap`. Telegram рендерит:
- `*bold*` → **bold**
- `_italic_` → _italic_
- `` `code` `` → `code`
- ```code blocks```

Пользователь видит markdown-разметку вместо форматированного текста в preview.

**Допущение**: MVP показывает текст as-is. Для WYSIWYG нужен парсер Telegram Markdown → HTML.

**Приоритет**: P2 — косметический, но влияет на доверие к preview.

---

### 7. DOM-манипуляции вместо React state в MediaMessage

**Файл**: `frontend/src/features/telegram-preview/components/MediaMessage.tsx:40-42`

```tsx
onError={(e) => {
  e.currentTarget.style.display = 'none'
  e.currentTarget.nextElementSibling?.classList.remove('hidden')
}}
```

Императивные DOM-операции вместо React state. Может вызвать рассинхрон с virtual DOM при re-render, особенно при изменении `mediaUrl` prop.

**Исправление**: использовать `useState<boolean>` для `hasError`, условный рендеринг.

**Приоритет**: P3 — работает, но anti-pattern.

---

### 8. MessageContentEditor жёстко привязан к campaignsApi

**Файл**: `frontend/src/features/telegram-preview/components/MessageContentEditor.tsx:35`

```tsx
import { campaignsApi } from '@/features/campaigns/api'
```

Upload файлов вызывает `campaignsApi.uploadFile()`. Компонент нельзя переиспользовать без этой зависимости.

**Исправление**: принимать `onUpload: (file: File) => Promise<string>` как prop.

**Приоритет**: P2 — блокирует чистое переиспользование в auto-scenarios и других контекстах.

---

### 9. `voice` и `animation` типы есть в backend, но не в editor UI

**Backend**: `entity/message.go:21` — определяет `PartVoice` и `PartAnimation`
**Sender**: `telegram/sender.go:108-139` — обрабатывает отправку
**Frontend editor**: `MessageContentEditor.tsx:43-50` — `PART_TYPE_OPTIONS` не включает `voice` и `animation`

Пользователь не может создать voice или GIF-сообщение через UI.

**Приоритет**: P2 — расширение функционала, не баг.

---

### 10. `media_id` (Telegram file_id caching) не реализован

**Файлы**: `entity/message.go:30`, `telegram/sender.go:148-149`

`MessagePart.MediaID` определён и sender проверяет его первым, но **нигде не заполняется** после успешной отправки. Каждая отправка в рассылке загружает файл по URL заново.

**Расчёт**: кампания с фото на 1000 человек = 1000 скачиваний одного файла с MinIO.

**Правильный подход**: после первой отправки сохранять `file_id` из ответа Telegram API, обновлять MessagePart и использовать для последующих отправок.

**Приоритет**: P1 — значительный overhead при масштабных рассылках.

---

## 🟢 MINOR (технический долг)

### 11. Нет unit-тестов для нового кода

Ни один из новых файлов не покрыт тестами:
- `entity/message.go` — Validate(), Scan/Value, TextContent()
- `eventbus/eventbus.go` — publish logic
- `eventbus/subscriber.go` — dispatch, channel routing
- `telegram/sender.go` — SendContent, sendPart, mediaInput, buildInlineKeyboard
- `botmanager/handler.go` — sendWelcomeContent flow
- `campaign/sender.go` — GetContent() fallback logic

**Приоритет**: P1 — блокирует уверенный рефакторинг.

---

### 12. Buttons в миграции variant'ов потеряны

**Файл**: `migrations/00032_composite_messages.sql:65-66`

```sql
-- campaign_variants получают пустой buttons
'buttons', '[]'::jsonb
```

Campaigns используют `COALESCE(buttons, '[]'::jsonb)`, но campaign_variants всегда получают пустой массив. Если у variant были buttons — они потеряются при миграции.

**Приоритет**: P1 — data loss при миграции (если buttons существуют у variants).

---

### 13. Timestamp "12:00" hardcoded в preview

**Файлы**: `MessageBubble.tsx`, `MediaMessage.tsx`

Preview всегда показывает `12:00`. Telegram показывает реальное время отправки. Мелочь, но выглядит как placeholder.

**Приоритет**: P3.

---

### 14. Нет стикерной галереи

Editor позволяет только upload `.webp` файла. Telegram стикеры обычно выбираются из пакетов по `file_id`. Для конечного пользователя workflow "скачай .webp и загрузи" неинтуитивен.

**Приоритет**: P3 — nice-to-have.

---

### 15. `max 5 parts` — произвольное ограничение

**Файлы**: `entity/message.go:76`, `MessageContentEditor.tsx:62`

Telegram API не ограничивает количество последовательных сообщений. Лимит в 5 частей — UX/performance решение, не техническое ограничение.

**Приоритет**: P3 — осознанное ограничение, документировать.

---

### 16. Callback-кнопки не поддержаны в UI

`InlineButton.Data` (callback_data) определён в типах, sender обрабатывает в `buildInlineKeyboard()`, но editor и preview показывают только URL-кнопки. Callback-кнопки нигде не создаются через UI.

**Приоритет**: P3 — для MVP достаточно URL-кнопок.

---

## Сводная таблица

| # | Проблема | Серьёзность | Приоритет | Усилие |
|---|----------|-------------|-----------|--------|
| 1 | Handler не обновляется при hot reload | 🔴 Critical | P0 | 1h |
| 2 | Redis Pub/Sub fire-and-forget | 🔴 Critical | P1 | 4h |
| 3 | Миграция: все media = photo | 🔴 Critical | P0 | 30m |
| 4 | Нет rate limit между частями | 🔴 Critical | P1 | 30m |
| 5 | Кампании не используют новый editor | 🟡 Major | P0 | 8h |
| 6 | Нет Markdown rendering в preview | 🟡 Major | P2 | 4h |
| 7 | DOM-манипуляции в MediaMessage | 🟡 Major | P3 | 30m |
| 8 | Editor привязан к campaignsApi | 🟡 Major | P2 | 1h |
| 9 | voice/animation не в editor | 🟡 Major | P2 | 1h |
| 10 | media_id caching не реализован | 🟡 Major | P1 | 4h |
| 11 | Нет unit-тестов | 🟢 Minor | P1 | 8h |
| 12 | Buttons variant'ов потеряны | 🟢 Minor | P1 | 30m |
| 13 | Timestamp hardcoded | 🟢 Minor | P3 | 15m |
| 14 | Нет стикерной галереи | 🟢 Minor | P3 | 8h |
| 15 | max 5 parts произвольный | 🟢 Minor | P3 | — |
| 16 | Callback-кнопки не в UI | 🟢 Minor | P3 | 2h |

## Рекомендуемый порядок исправления

### Sprint 1 (P0 — блокеры)
1. **#1** — Fix handler hot reload (pointer to inst.info)
2. **#3** — Fix migration media type detection
3. **#12** — Fix variant buttons migration

### Sprint 2 (P1 — production readiness)
4. **#4** — Add rate limiting between parts
5. **#10** — Implement media_id caching after first send
6. **#11** — Write unit tests for new code
7. **#2** — Evaluate Redis Streams vs Pub/Sub

### Sprint 3 (P0 product + P2 polish)
8. **#5** — Integrate MessageContentEditor into campaigns UI
9. **#8** — Extract upload dependency from editor
10. **#9** — Add voice/animation types to editor
11. **#6** — Implement Telegram Markdown → HTML parser

### Backlog (P3)
12. #7, #13, #14, #15, #16
