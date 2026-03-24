# Phase 1: Foundation

Фундаментальные изменения, которые разблокируют всю дальнейшую разработку Revisitr.

## Содержание

1. [Обзор и цели](#1-обзор-и-цели)
2. [Предварительные условия](#2-предварительные-условия)
3. [1A. Миграция БД — иерархия сущностей](#3-1a-миграция-бд--иерархия-сущностей)
4. [1B. Loyalty Engine — движок начислений](#4-1b-loyalty-engine--движок-начислений)
5. [1C. Файловое хранилище (MinIO)](#5-1c-файловое-хранилище-minio)
6. [Стратегия тестирования](#6-стратегия-тестирования)
7. [Definition of Done](#7-definition-of-done)

---

## 1. Обзор и цели

### Текущее состояние

Сущности `Bot`, `LoyaltyProgram` и `POSLocation` существуют независимо друг от друга, связанные только через `org_id`. Бонусная система определена на уровне структур (`LoyaltyLevel.reward_percent`), но фактически не используется для расчетов. Файловое хранилище отсутствует.

### Целевое состояние

- Иерархия сущностей: `LoyaltyProgram (1) -> Bot (N) -> POS (M)`
- Работающий движок начисления бонусов с прогрессией и понижением уровней
- Идентификация клиентов по телефону и QR-коду
- Файловое хранилище для медиа-файлов (MinIO)

### Почему это первая фаза

Все последующие фичи (транзакции с бонусами, Telegram-бот с реальной логикой, рассылки с медиа) зависят от этих трех компонентов. Без иерархии сущностей невозможно корректно привязать бонусную программу к боту, без движка начислений бот не сможет начислять бонусы, без хранилища не будет медиа в рассылках.

---

## 2. Предварительные условия

- PostgreSQL 16 запущен и все миграции до `00013_integration_mock_type.sql` применены
- Redis 7 запущен
- Go 1.23+, Node.js 20+
- Docker Compose (для MinIO)
- `goose` установлен: `go install github.com/pressly/goose/v3/cmd/goose@latest`

Проверка текущего состояния миграций:

```bash
cd backend && goose -dir migrations postgres "$DATABASE_URL" status
```

---

## 3. 1A. Миграция БД — иерархия сущностей

### 3.1. Задачи и критерии приемки

| Задача | Критерий приемки |
|--------|-----------------|
| Связь LoyaltyProgram -> Bot | `bots.program_id` ссылается на `loyalty_programs.id`, NULLABLE |
| Связь Bot -> POS | `pos_locations.bot_id` ссылается на `bots.id`, NULLABLE |
| Расширение типов наград | `loyalty_levels.reward_type` и `reward_amount` добавлены |
| Обязательный телефон клиента | `bot_clients.phone` NOT NULL, нормализованный формат |
| QR-код клиента | `bot_clients.qr_code` уникальный, генерируется при регистрации |
| Индексы | Индексы на `program_id`, `bot_id`, `phone_normalized`, `qr_code` созданы |

### 3.2. SQL-миграция

Файл: `backend/migrations/00014_entity_hierarchy.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- === Связь LoyaltyProgram (1) -> Bot (N) ===
ALTER TABLE bots
    ADD COLUMN program_id INT REFERENCES loyalty_programs(id) ON DELETE SET NULL;

CREATE INDEX idx_bots_program_id ON bots(program_id);

-- === Связь Bot (1) -> POS (M) ===
ALTER TABLE pos_locations
    ADD COLUMN bot_id INT REFERENCES bots(id) ON DELETE SET NULL;

CREATE INDEX idx_pos_locations_bot_id ON pos_locations(bot_id);

-- === Расширение системы наград ===
-- reward_type: 'percent' (% от суммы чека) или 'fixed' (фиксированная сумма)
ALTER TABLE loyalty_levels
    ADD COLUMN reward_type VARCHAR(10) NOT NULL DEFAULT 'percent',
    ADD COLUMN reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0;

-- Заполнение reward_amount из существующего reward_percent для совместимости
UPDATE loyalty_levels SET reward_amount = reward_percent WHERE reward_type = 'percent';

ALTER TABLE loyalty_levels
    ADD CONSTRAINT chk_reward_type CHECK (reward_type IN ('percent', 'fixed'));

-- === Обязательный телефон + нормализация + QR ===
-- Сначала добавляем новые колонки
ALTER TABLE bot_clients
    ADD COLUMN phone_normalized VARCHAR(15),
    ADD COLUMN qr_code VARCHAR(64) UNIQUE;

CREATE INDEX idx_bot_clients_phone_normalized ON bot_clients(phone_normalized);
CREATE INDEX idx_bot_clients_qr_code ON bot_clients(qr_code);

-- Делаем phone NOT NULL только после заполнения существующих записей.
-- Для существующих клиентов без телефона ставим плейсхолдер, который
-- потом нужно будет обновить вручную или через бота.
UPDATE bot_clients SET phone = 'unknown' WHERE phone IS NULL;
ALTER TABLE bot_clients ALTER COLUMN phone SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE bot_clients ALTER COLUMN phone DROP NOT NULL;
DROP INDEX IF EXISTS idx_bot_clients_qr_code;
DROP INDEX IF EXISTS idx_bot_clients_phone_normalized;
ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS qr_code,
    DROP COLUMN IF EXISTS phone_normalized;

ALTER TABLE loyalty_levels
    DROP CONSTRAINT IF EXISTS chk_reward_type,
    DROP COLUMN IF EXISTS reward_amount,
    DROP COLUMN IF EXISTS reward_type;

DROP INDEX IF EXISTS idx_pos_locations_bot_id;
ALTER TABLE pos_locations DROP COLUMN IF EXISTS bot_id;

DROP INDEX IF EXISTS idx_bots_program_id;
ALTER TABLE bots DROP COLUMN IF EXISTS program_id;

-- +goose StatementEnd
```

Применение:

```bash
cd backend && goose -dir migrations postgres "$DATABASE_URL" up
```

### 3.3. Backend — изменения в entity

#### `backend/internal/entity/bot.go`

Добавить поле `ProgramID` в структуру `Bot` и обновить request-структуры.

```go
// Bot — существующая структура, добавить поле:
type Bot struct {
    ID        int         `db:"id" json:"id"`
    OrgID     int         `db:"org_id" json:"org_id"`
    ProgramID *int        `db:"program_id" json:"program_id"` // NEW
    Name      string      `db:"name" json:"name"`
    Token     string      `db:"token" json:"-"`
    Username  string      `db:"username" json:"username"`
    Status    string      `db:"status" json:"status"`
    Settings  BotSettings `db:"settings" json:"settings"`
    CreatedAt time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

// CreateBotRequest — добавить опциональный ProgramID:
type CreateBotRequest struct {
    Name      string `json:"name" binding:"required"`
    Token     string `json:"token" binding:"required"`
    ProgramID *int   `json:"program_id"` // NEW
}

// UpdateBotRequest — добавить опциональный ProgramID:
type UpdateBotRequest struct {
    Name      *string `json:"name,omitempty"`
    Status    *string `json:"status,omitempty"`
    ProgramID *int    `json:"program_id,omitempty"` // NEW
}
```

#### `backend/internal/entity/pos.go`

Добавить поле `BotID` в структуру `POSLocation` и request-структуры.

```go
type POSLocation struct {
    ID        int       `db:"id" json:"id"`
    OrgID     int       `db:"org_id" json:"org_id"`
    BotID     *int      `db:"bot_id" json:"bot_id"` // NEW
    Name      string    `db:"name" json:"name"`
    Address   string    `db:"address" json:"address,omitempty"`
    Phone     string    `db:"phone" json:"phone,omitempty"`
    Schedule  Schedule  `db:"schedule" json:"schedule"`
    IsActive  bool      `db:"is_active" json:"is_active"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreatePOSRequest — добавить опциональный BotID:
type CreatePOSRequest struct {
    Name     string   `json:"name" binding:"required"`
    Address  string   `json:"address"`
    Phone    string   `json:"phone"`
    Schedule Schedule `json:"schedule"`
    BotID    *int     `json:"bot_id"` // NEW
}

// UpdatePOSRequest — добавить опциональный BotID:
type UpdatePOSRequest struct {
    Name     *string   `json:"name,omitempty"`
    Address  *string   `json:"address,omitempty"`
    Phone    *string   `json:"phone,omitempty"`
    Schedule *Schedule `json:"schedule,omitempty"`
    IsActive *bool     `json:"is_active,omitempty"`
    BotID    *int      `json:"bot_id,omitempty"` // NEW
}
```

#### `backend/internal/entity/bot_client.go`

Добавить поля `PhoneNormalized` и `QRCode` в структуру `BotClient`.

```go
type BotClient struct {
    ID              int             `db:"id" json:"id"`
    BotID           int             `db:"bot_id" json:"bot_id"`
    TelegramID      int64           `db:"telegram_id" json:"telegram_id"`
    Username        string          `db:"username" json:"username,omitempty"`
    FirstName       string          `db:"first_name" json:"first_name"`
    LastName        string          `db:"last_name" json:"last_name,omitempty"`
    Phone           string          `db:"phone" json:"phone"`                         // CHANGED: was *string
    PhoneNormalized *string         `db:"phone_normalized" json:"phone_normalized"`   // NEW
    QRCode          *string         `db:"qr_code" json:"qr_code,omitempty"`           // NEW
    Gender          *string         `db:"gender" json:"gender,omitempty"`
    BirthDate       *time.Time      `db:"birth_date" json:"birth_date,omitempty"`
    City            *string         `db:"city" json:"city,omitempty"`
    OS              *string         `db:"os" json:"os,omitempty"`
    Tags            Tags            `db:"tags" json:"tags"`
    Data            json.RawMessage `db:"data" json:"-"`
    RegisteredAt    time.Time       `db:"registered_at" json:"registered_at"`
    // RFM fields
    RFMRecency   *int       `db:"rfm_recency"    json:"rfm_recency,omitempty"`
    RFMFrequency *int       `db:"rfm_frequency"  json:"rfm_frequency,omitempty"`
    RFMMonetary  *float64   `db:"rfm_monetary"   json:"rfm_monetary,omitempty"`
    RFMSegment   *string    `db:"rfm_segment"    json:"rfm_segment,omitempty"`
    RFMUpdatedAt *time.Time `db:"rfm_updated_at" json:"rfm_updated_at,omitempty"`
}
```

#### `backend/internal/entity/loyalty.go`

Расширить `LoyaltyLevel` и `CreateLevelRequest`.

```go
type LoyaltyLevel struct {
    ID            int     `db:"id" json:"id"`
    ProgramID     int     `db:"program_id" json:"program_id"`
    Name          string  `db:"name" json:"name"`
    Threshold     int     `db:"threshold" json:"threshold"`
    RewardPercent float64 `db:"reward_percent" json:"reward_percent"`
    RewardType    string  `db:"reward_type" json:"reward_type"`     // NEW: "percent" | "fixed"
    RewardAmount  float64 `db:"reward_amount" json:"reward_amount"` // NEW: для fixed-бонуса
    SortOrder     int     `db:"sort_order" json:"sort_order"`
}

type CreateLevelRequest struct {
    Name          string  `json:"name" binding:"required"`
    Threshold     int     `json:"threshold" binding:"min=0"`
    RewardPercent float64 `json:"reward_percent" binding:"min=0,max=100"`
    RewardType    string  `json:"reward_type"`                        // NEW
    RewardAmount  float64 `json:"reward_amount" binding:"min=0"`      // NEW
    SortOrder     int     `json:"sort_order"`
}
```

### 3.4. Backend — утилита нормализации телефона

Новый файл: `backend/internal/entity/phone.go`

```go
package entity

import "strings"

// NormalizePhone приводит телефон к формату +7XXXXXXXXXX.
// Обрабатывает форматы: +7..., 8..., 7..., без кода (10 цифр).
// Возвращает пустую строку, если формат не распознан.
func NormalizePhone(phone string) string {
    // Убираем все нецифровые символы
    digits := make([]byte, 0, len(phone))
    for i := 0; i < len(phone); i++ {
        if phone[i] >= '0' && phone[i] <= '9' {
            digits = append(digits, phone[i])
        }
    }

    s := string(digits)

    switch {
    case len(s) == 11 && (s[0] == '7' || s[0] == '8'):
        // 7XXXXXXXXXX или 8XXXXXXXXXX -> +7XXXXXXXXXX
        return "+7" + s[1:]
    case len(s) == 10:
        // 10 цифр без кода -> +7XXXXXXXXXX
        return "+7" + s
    case len(s) == 12 && strings.HasPrefix(s, "7"):
        // Лишняя цифра, но начинается с 7 — скорее всего +7 + 10 цифр
        return "+" + s[:1] + s[2:]
    default:
        return ""
    }
}
```

### 3.5. Backend — утилита генерации QR-кода

Новый файл: `backend/internal/entity/qrcode.go`

```go
package entity

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
)

// GenerateQRCode создает уникальный идентификатор для QR-кода клиента.
// Формат: "RVS-XXXXXXXXXXXXXXXX" (префикс + 16 hex-символов).
func GenerateQRCode() (string, error) {
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("GenerateQRCode: %w", err)
    }
    return "RVS-" + hex.EncodeToString(b), nil
}
```

### 3.6. Backend — изменения в repository

#### `backend/internal/repository/postgres/bots.go`

Обновить SQL-запросы: добавить `program_id` во все SELECT, INSERT, UPDATE.

Конкретные изменения:

- **INSERT**: добавить `program_id` в список колонок и значений
- **SELECT**: добавить `program_id` в список полей
- **UPDATE**: обработать `program_id` в SET-клаузе
- **Новый метод** `GetBotsByProgramID(ctx, programID) ([]Bot, error)` — получить ботов по программе лояльности

#### `backend/internal/repository/postgres/pos.go`

Обновить SQL-запросы: добавить `bot_id` во все SELECT, INSERT, UPDATE.

- **INSERT**: добавить `bot_id` в список колонок
- **SELECT**: добавить `bot_id` в список полей
- **UPDATE**: обработать `bot_id` в SET-клаузе
- **Новый метод** `GetPOSByBotID(ctx, botID) ([]POSLocation, error)` — получить точки по боту

#### `backend/internal/repository/postgres/bot_clients.go`

Обновить SQL-запросы: добавить `phone_normalized`, `qr_code`.

- **INSERT**: добавить `phone_normalized`, `qr_code`
- **SELECT**: добавить `phone_normalized`, `qr_code`
- **Новые методы**:
  - `GetClientByPhone(ctx, phoneNormalized, orgID) (*BotClient, error)` — поиск по нормализованному телефону
  - `GetClientByQRCode(ctx, qrCode) (*BotClient, error)` — поиск по QR-коду

#### `backend/internal/repository/postgres/loyalty.go`

Обновить запросы для `loyalty_levels`: добавить `reward_type`, `reward_amount` в SELECT, INSERT, UPDATE.

### 3.7. Backend — изменения в usecase

#### `backend/internal/usecase/bots/` (существующий)

- При создании бота: записывать `program_id` если передан
- При обновлении бота: обрабатывать смену `program_id`
- Валидация: проверять, что `program_id` принадлежит той же организации

#### `backend/internal/usecase/pos/` (существующий)

- При создании POS: записывать `bot_id` если передан
- При обновлении POS: обрабатывать смену `bot_id`
- Валидация: проверять, что `bot_id` принадлежит той же организации

#### `backend/internal/usecase/clients/` (существующий)

- При регистрации клиента: вызывать `NormalizePhone()` и `GenerateQRCode()`
- Заполнять `phone_normalized` и `qr_code` автоматически

### 3.8. Backend — изменения в controller

#### `backend/internal/controller/http/group/bots/`

- Обновить handler создания бота: принимать `program_id`
- Обновить handler обновления бота: принимать `program_id`
- Передавать новые поля в usecase

#### `backend/internal/controller/http/group/pos/`

- Обновить handler создания POS: принимать `bot_id`
- Обновить handler обновления POS: принимать `bot_id`

#### `backend/internal/controller/http/group/clients/`

- Новый endpoint: `GET /api/v1/clients/identify?phone=...&qr_code=...` — идентификация клиента по телефону или QR-коду

### 3.9. Frontend — изменения

#### Типы: обновить API-типы

Файлы:
- `frontend/src/features/bots/types.ts` — добавить `program_id?: number`
- `frontend/src/features/pos/types.ts` — добавить `bot_id?: number`
- `frontend/src/features/clients/types.ts` — добавить `qr_code?: string`, `phone_normalized?: string`
- `frontend/src/features/loyalty/types.ts` — добавить `reward_type: 'percent' | 'fixed'`, `reward_amount: number`

#### Форма создания бота: `frontend/src/features/bots/`

- Добавить опциональный select "Программа лояльности" в форму создания/редактирования бота
- Загрузка списка программ через `useQuery` из `/api/v1/loyalty/programs`

#### Форма создания POS: `frontend/src/features/pos/`

- Добавить опциональный select "Бот" в форму создания/редактирования POS
- Загрузка списка ботов через `useQuery` из `/api/v1/bots`

#### Профиль клиента: `frontend/src/features/clients/`

- Отображать QR-код клиента (использовать библиотеку `qrcode.react` или аналог)
- Показывать нормализованный телефон

#### Уровни лояльности: `frontend/src/features/loyalty/`

- Добавить переключатель типа награды (процент / фиксированная сумма) в форму уровня
- Показывать `reward_amount` для фиксированного типа, `reward_percent` для процентного

### 3.10. Коммиты

```
feat(db): add entity hierarchy migration (program->bot->pos)
feat(entity): add ProgramID to Bot, BotID to POS, QRCode to BotClient
feat(entity): add phone normalization and QR code generation utils
feat(repo): update repositories for entity hierarchy fields
feat(usecase): handle entity hierarchy in bots, pos, clients
feat(api): add client identification endpoint
feat(frontend): add program selector to bot form, bot selector to POS form
feat(frontend): display client QR code and reward type toggle
```

---

## 4. 1B. Loyalty Engine — движок начислений

### 4.1. Задачи и критерии приемки

| Задача | Критерий приемки |
|--------|-----------------|
| Расчет бонусов (процент) | При транзакции на 1000 руб. с уровнем 5% начисляется 50 бонусов |
| Расчет бонусов (фикс) | При транзакции с уровнем "fixed 30" начисляется 30 бонусов |
| Прогрессия уровней | При достижении threshold клиент автоматически повышается |
| Понижение уровней | Cron-задача проверяет threshold и понижает, если total_earned упал |
| Баланс: earn | Начисление увеличивает balance и total_earned |
| Баланс: spend | Списание уменьшает balance, увеличивает total_spent |
| Баланс: reserve | Резервирование суммы перед списанием (для POS-подтверждения) |
| Идентификация клиента | Поиск по phone_normalized или qr_code возвращает профиль |

### 4.2. Backend — сервис расчета бонусов

#### Изменения в `backend/internal/usecase/loyalty/loyalty.go`

Текущий метод `EarnPoints` принимает готовую сумму бонусов. Нужно добавить метод, который принимает сумму чека и сам рассчитывает бонусы на основе уровня клиента.

Новые методы:

```go
// CalculateBonus вычисляет бонус для клиента на основе его текущего уровня.
// checkAmount — сумма чека. Возвращает сумму бонуса к начислению.
func (uc *Usecase) CalculateBonus(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error) {
    cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
    if err != nil {
        return 0, err
    }

    if cl.LevelID == nil {
        return 0, nil // нет уровня — нет бонуса
    }

    levels, err := uc.repo.GetLevelsByProgramID(ctx, programID)
    if err != nil {
        return 0, fmt.Errorf("usecase.CalculateBonus: %w", err)
    }

    for _, level := range levels {
        if level.ID == *cl.LevelID {
            switch level.RewardType {
            case "percent":
                return checkAmount * level.RewardPercent / 100, nil
            case "fixed":
                return level.RewardAmount, nil
            default:
                return checkAmount * level.RewardPercent / 100, nil
            }
        }
    }

    return 0, nil
}

// EarnFromCheck — основной метод начисления бонусов с чека.
// Рассчитывает бонус, начисляет, обновляет уровень.
func (uc *Usecase) EarnFromCheck(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error) {
    bonus, err := uc.CalculateBonus(ctx, clientID, programID, checkAmount)
    if err != nil {
        return nil, err
    }
    if bonus <= 0 {
        return uc.GetBalance(ctx, clientID, programID)
    }

    desc := fmt.Sprintf("Бонус %.2f с чека %.2f", bonus, checkAmount)
    return uc.EarnPoints(ctx, clientID, programID, bonus, desc)
}
```

#### Прогрессия уровней

Текущий метод `determineLevelID` уже определяет уровень по `totalEarned`. Он используется в `EarnPoints`. Поведение корректно для повышения: при каждом начислении пересчитывается уровень.

Метод работает и для понижения при проверке cron-задачей, так как он всегда выбирает максимальный подходящий уровень по threshold.

#### Понижение уровней — cron-задача

Новый файл: `backend/internal/usecase/loyalty/demotion.go`

```go
package loyalty

import (
    "context"
    "fmt"
)

// DemoteClients проверяет всех клиентов с loyalty и понижает уровень,
// если total_earned больше не достигает threshold текущего уровня.
// Вызывается из scheduler раз в сутки.
func (uc *Usecase) DemoteClients(ctx context.Context) error {
    // Получаем всех клиентов, у которых level_id != NULL
    clients, err := uc.repo.GetClientsWithLevels(ctx)
    if err != nil {
        return fmt.Errorf("usecase.DemoteClients: %w", err)
    }

    for _, cl := range clients {
        newLevelID := uc.determineLevelID(ctx, cl.ProgramID, cl.TotalEarned)

        // Если текущий уровень отличается — обновляем
        if !equalIntPtr(cl.LevelID, newLevelID) {
            cl.LevelID = newLevelID
            if err := uc.repo.UpsertClientLoyalty(ctx, &cl); err != nil {
                uc.logger.Error("demotion failed",
                    "client_id", cl.ClientID,
                    "program_id", cl.ProgramID,
                    "error", err,
                )
                continue
            }
            uc.logger.Info("level updated",
                "client_id", cl.ClientID,
                "old_level", cl.LevelID,
                "new_level", newLevelID,
            )
        }
    }
    return nil
}

func equalIntPtr(a, b *int) bool {
    if a == nil && b == nil {
        return true
    }
    if a == nil || b == nil {
        return false
    }
    return *a == *b
}
```

Новый метод в интерфейсе `repository`:

```go
type repository interface {
    // ... существующие методы ...
    GetClientsWithLevels(ctx context.Context) ([]entity.ClientLoyalty, error) // NEW
}
```

#### Резервирование баланса

Новый файл: `backend/internal/usecase/loyalty/reserve.go`

Резервирование нужно для POS-сценария: кассир видит доступные бонусы, клиент хочет списать часть, но подтверждение приходит позже. Резервирование блокирует сумму, чтобы она не была потрачена дважды.

```go
package loyalty

import (
    "context"
    "fmt"
    "time"

    "revisitr/internal/entity"
)

// ReservePoints резервирует бонусы для последующего списания.
// Возвращает ID резерва для подтверждения или отмены.
func (uc *Usecase) ReservePoints(ctx context.Context, clientID, programID int, amount float64) (int, error) {
    cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
    if err != nil {
        return 0, err
    }

    availableBalance := cl.Balance // TODO: минус уже зарезервированные суммы
    if availableBalance < amount {
        return 0, ErrInsufficientPoints
    }

    reserve := &entity.BalanceReserve{
        ClientID:  clientID,
        ProgramID: programID,
        Amount:    amount,
        Status:    "pending",
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }

    if err := uc.repo.CreateReserve(ctx, reserve); err != nil {
        return 0, fmt.Errorf("usecase.ReservePoints: %w", err)
    }

    return reserve.ID, nil
}

// ConfirmReserve подтверждает резерв и списывает бонусы.
func (uc *Usecase) ConfirmReserve(ctx context.Context, reserveID int) (*entity.ClientLoyalty, error) {
    reserve, err := uc.repo.GetReserve(ctx, reserveID)
    if err != nil {
        return nil, fmt.Errorf("usecase.ConfirmReserve: %w", err)
    }

    if reserve.Status != "pending" {
        return nil, fmt.Errorf("reserve is not pending: %s", reserve.Status)
    }
    if time.Now().After(reserve.ExpiresAt) {
        return nil, fmt.Errorf("reserve expired")
    }

    cl, err := uc.SpendPoints(ctx, reserve.ClientID, reserve.ProgramID, reserve.Amount, "Списание по резерву")
    if err != nil {
        return nil, err
    }

    reserve.Status = "confirmed"
    if err := uc.repo.UpdateReserve(ctx, reserve); err != nil {
        return nil, fmt.Errorf("usecase.ConfirmReserve update: %w", err)
    }

    return cl, nil
}

// CancelReserve отменяет резерв.
func (uc *Usecase) CancelReserve(ctx context.Context, reserveID int) error {
    reserve, err := uc.repo.GetReserve(ctx, reserveID)
    if err != nil {
        return fmt.Errorf("usecase.CancelReserve: %w", err)
    }

    reserve.Status = "cancelled"
    return uc.repo.UpdateReserve(ctx, reserve)
}
```

#### Новая entity: `backend/internal/entity/reserve.go`

```go
package entity

import "time"

type BalanceReserve struct {
    ID        int       `db:"id" json:"id"`
    ClientID  int       `db:"client_id" json:"client_id"`
    ProgramID int       `db:"program_id" json:"program_id"`
    Amount    float64   `db:"amount" json:"amount"`
    Status    string    `db:"status" json:"status"` // "pending", "confirmed", "cancelled", "expired"
    ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

### 4.3. Миграция для резервирования

Файл: `backend/migrations/00015_balance_reserves.sql`

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE balance_reserves (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    program_id INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'cancelled', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_balance_reserves_client_program ON balance_reserves(client_id, program_id);
CREATE INDEX idx_balance_reserves_status ON balance_reserves(status) WHERE status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS balance_reserves;

-- +goose StatementEnd
```

### 4.4. Backend — repository для резервирования

Новый файл: `backend/internal/repository/postgres/reserves.go`

Методы:

```go
type ReservesRepo struct { db *sqlx.DB }

func (r *ReservesRepo) CreateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
func (r *ReservesRepo) GetReserve(ctx context.Context, id int) (*entity.BalanceReserve, error)
func (r *ReservesRepo) UpdateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
func (r *ReservesRepo) GetPendingReserves(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error)
func (r *ReservesRepo) ExpireOldReserves(ctx context.Context) (int, error) // для cron
```

Добавить в интерфейс `repository` в `loyalty.go`:

```go
type repository interface {
    // ... существующие методы ...
    GetClientsWithLevels(ctx context.Context) ([]entity.ClientLoyalty, error)
    CreateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
    GetReserve(ctx context.Context, id int) (*entity.BalanceReserve, error)
    UpdateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
    GetPendingReserves(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error)
    ExpireOldReserves(ctx context.Context) (int, error)
}
```

### 4.5. Backend — scheduler-задачи

Регистрация в `backend/cmd/server/main.go` (или в bootstrap):

```go
sched.Register(scheduler.Task{
    Name:     "loyalty-demotion",
    Interval: 24 * time.Hour,
    Fn:       loyaltyUC.DemoteClients,
})

sched.Register(scheduler.Task{
    Name:     "expire-reserves",
    Interval: 5 * time.Minute,
    Fn: func(ctx context.Context) error {
        n, err := loyaltyRepo.ExpireOldReserves(ctx)
        if err != nil {
            return err
        }
        if n > 0 {
            slog.Info("expired reserves", "count", n)
        }
        return nil
    },
})
```

### 4.6. Backend — новые API-эндпоинты

Добавить в `backend/internal/controller/http/group/loyalty/` или `clients/`:

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/loyalty/earn-from-check` | Начисление бонусов с чека |
| `POST` | `/api/v1/loyalty/reserve` | Резервирование бонусов |
| `POST` | `/api/v1/loyalty/reserve/:id/confirm` | Подтверждение резерва |
| `POST` | `/api/v1/loyalty/reserve/:id/cancel` | Отмена резерва |
| `GET` | `/api/v1/clients/identify` | Идентификация клиента по телефону/QR |

#### Тело запроса `POST /api/v1/loyalty/earn-from-check`

```json
{
    "client_id": 42,
    "program_id": 1,
    "check_amount": 1500.00
}
```

#### Тело запроса `POST /api/v1/loyalty/reserve`

```json
{
    "client_id": 42,
    "program_id": 1,
    "amount": 200.00
}
```

#### Параметры `GET /api/v1/clients/identify`

```
?phone=+79991234567    или    ?qr_code=RVS-a1b2c3d4e5f6g7h8
```

### 4.7. Коммиты

```
feat(loyalty): add bonus calculation service with percent and fixed types
feat(loyalty): add level demotion cron task
feat(db): add balance_reserves table migration
feat(loyalty): add reserve/confirm/cancel flow for POS
feat(api): add earn-from-check and client identification endpoints
feat(scheduler): register demotion and reserve expiry tasks
```

---

## 5. 1C. Файловое хранилище (MinIO)

### 5.1. Задачи и критерии приемки

| Задача | Критерий приемки |
|--------|-----------------|
| MinIO в docker-compose | Контейнер запускается, доступен на порту 9000/9001 |
| Upload endpoint | `POST /api/v1/files/upload` возвращает URL файла |
| Download/proxy | Файлы доступны по URL через API |
| Лимит размера | Файлы больше 50MB отклоняются |
| Абстракция | Интерфейс Storage позволяет подменить MinIO на S3 |
| Продакшен | MinIO добавлен в prod docker-compose |

### 5.2. Docker Compose — dev

Файл: `infra/docker-compose.yml` — добавить сервис:

```yaml
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: revisitr
      MINIO_ROOT_PASSWORD: devpassword
    ports:
      - "9000:9000"   # API
      - "9001:9001"   # Console
    volumes:
      - miniodata:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 3s
      retries: 5
```

Добавить volume `miniodata` в секцию `volumes`.

### 5.3. Docker Compose — prod

Файл: `infra/docker-compose.prod.yml` — добавить аналогичный сервис с production-паролем из переменных окружения.

```yaml
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data
    restart: unless-stopped
```

### 5.4. Backend — интерфейс Storage

Новый файл: `backend/internal/usecase/storage/storage.go`

```go
package storage

import (
    "context"
    "io"
)

// FileInfo — метаданные загруженного файла.
type FileInfo struct {
    Key         string `json:"key"`
    URL         string `json:"url"`
    ContentType string `json:"content_type"`
    Size        int64  `json:"size"`
}

// Storage — абстракция файлового хранилища.
// Реализации: MinIO (сейчас), S3 (будущее).
type Storage interface {
    Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*FileInfo, error)
    Delete(ctx context.Context, bucket, key string) error
    GetURL(ctx context.Context, bucket, key string) (string, error)
}
```

### 5.5. Backend — реализация MinIO

Новый файл: `backend/internal/repository/minio/minio.go`

```go
package minio

import (
    "context"
    "fmt"
    "io"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"

    "revisitr/internal/usecase/storage"
)

type Client struct {
    client   *minio.Client
    endpoint string
}

func New(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
    mc, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("minio.New: %w", err)
    }
    return &Client{client: mc, endpoint: endpoint}, nil
}

func (c *Client) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
    // Создать бакет если не существует
    exists, err := c.client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, fmt.Errorf("minio.Upload check bucket: %w", err)
    }
    if !exists {
        if err := c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, fmt.Errorf("minio.Upload create bucket: %w", err)
        }
    }

    _, err = c.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
        ContentType: contentType,
    })
    if err != nil {
        return nil, fmt.Errorf("minio.Upload: %w", err)
    }

    return &storage.FileInfo{
        Key:         key,
        URL:         fmt.Sprintf("/%s/%s", bucket, key),
        ContentType: contentType,
        Size:        size,
    }, nil
}

func (c *Client) Delete(ctx context.Context, bucket, key string) error {
    return c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (c *Client) GetURL(_ context.Context, bucket, key string) (string, error) {
    return fmt.Sprintf("/%s/%s", bucket, key), nil
}
```

### 5.6. Backend — конфигурация

Добавить в `backend/internal/application/config.go` (или аналогичный файл конфигурации):

```go
type MinIOConfig struct {
    Endpoint  string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
    AccessKey string `env:"MINIO_ACCESS_KEY" envDefault:"revisitr"`
    SecretKey string `env:"MINIO_SECRET_KEY" envDefault:"devpassword"`
    UseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
    Bucket    string `env:"MINIO_BUCKET" envDefault:"revisitr"`
}
```

### 5.7. Backend — HTTP handler для загрузки

Новый файл: `backend/internal/controller/http/group/files/files.go`

```go
package files

import (
    "fmt"
    "net/http"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"

    "revisitr/internal/usecase/storage"
)

const maxFileSize = 50 << 20 // 50MB

type Handler struct {
    storage storage.Storage
    bucket  string
}

func New(s storage.Storage, bucket string) *Handler {
    return &Handler{storage: s, bucket: bucket}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
    files := rg.Group("/files")
    files.POST("/upload", h.Upload)
}

func (h *Handler) Upload(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
        return
    }
    defer file.Close()

    if header.Size > maxFileSize {
        c.JSON(http.StatusRequestEntityTooLarge, gin.H{
            "error": fmt.Sprintf("file too large, max %d MB", maxFileSize>>20),
        })
        return
    }

    ext := filepath.Ext(header.Filename)
    key := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

    info, err := h.storage.Upload(
        c.Request.Context(),
        h.bucket,
        key,
        file,
        header.Size,
        header.Header.Get("Content-Type"),
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
        return
    }

    c.JSON(http.StatusOK, info)
}
```

### 5.8. Порты

Дополнение к таблице портов проекта:

| Service | Local dev | Production |
|---------|-----------|------------|
| MinIO API | 9000 | 9000 |
| MinIO Console | 9001 | 9001 |

### 5.9. Зависимость Go

```bash
cd backend && go get github.com/minio/minio-go/v7
```

### 5.10. Коммиты

```
feat(infra): add MinIO to docker-compose (dev + prod)
feat(storage): add Storage interface and MinIO implementation
feat(api): add file upload endpoint with 50MB limit
feat(config): add MinIO configuration
```

---

## 6. Стратегия тестирования

### 6.1. Unit-тесты (без инфраструктуры)

| Что тестируем | Файл | Паттерн |
|---------------|------|---------|
| `NormalizePhone` | `backend/internal/entity/phone_test.go` | Table-driven тесты: +7, 8, 10 цифр, невалидные |
| `GenerateQRCode` | `backend/internal/entity/qrcode_test.go` | Проверка формата, уникальности |
| `CalculateBonus` (percent) | `backend/internal/usecase/loyalty/loyalty_test.go` | Мок repository, проверка % |
| `CalculateBonus` (fixed) | `backend/internal/usecase/loyalty/loyalty_test.go` | Мок repository, проверка фикс. суммы |
| `determineLevelID` | `backend/internal/usecase/loyalty/loyalty_test.go` | Граничные значения threshold |
| `DemoteClients` | `backend/internal/usecase/loyalty/demotion_test.go` | Мок с клиентами выше/ниже threshold |
| `ReservePoints` | `backend/internal/usecase/loyalty/reserve_test.go` | Достаточный/недостаточный баланс |

Паттерн мока (установленный в проекте — struct с function fields):

```go
type mockRepo struct {
    GetLevelsByProgramIDFn func(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error)
    GetClientLoyaltyFn     func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
    // ... остальные поля
}

func (m *mockRepo) GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error) {
    return m.GetLevelsByProgramIDFn(ctx, programID)
}
```

### 6.2. Integration-тесты (требуют Docker)

| Что тестируем | Файл | Проверка |
|---------------|------|----------|
| Миграция 00014 | `backend/tests/integration/migration_test.go` | Миграция up/down без ошибок |
| Миграция 00015 | `backend/tests/integration/migration_test.go` | Миграция up/down без ошибок |
| Earn from check | `backend/tests/integration/loyalty_test.go` | Полный цикл: чек -> бонус -> баланс |
| Reserve flow | `backend/tests/integration/loyalty_test.go` | Резерв -> подтверждение -> списание |
| Client identify | `backend/tests/integration/clients_test.go` | Поиск по телефону и QR |
| File upload | `backend/tests/integration/files_test.go` | Upload + проверка в MinIO |

Build tag: `//go:build integration`

Запуск:

```bash
cd backend && go test -race -tags=integration ./tests/integration/...
```

### 6.3. Frontend-тесты

| Что тестируем | Файл | Проверка |
|---------------|------|----------|
| Bot form with program select | `frontend/src/features/bots/__tests__/` | Рендер селектора, отправка program_id |
| POS form with bot select | `frontend/src/features/pos/__tests__/` | Рендер селектора, отправка bot_id |
| Reward type toggle | `frontend/src/features/loyalty/__tests__/` | Переключение типа, валидация полей |

Запуск:

```bash
cd frontend && npx vitest run
```

### 6.4. Тестовые данные

Расширить seed-данные в `backend/migrations/00008_seed_data.sql` или создать отдельный тестовый seed:

```sql
-- Привязать бота к программе лояльности
UPDATE bots SET program_id = 1 WHERE id = 1;

-- Привязать POS к боту
UPDATE pos_locations SET bot_id = 1 WHERE id = 1;

-- Добавить уровни с разными типами наград
INSERT INTO loyalty_levels (program_id, name, threshold, reward_percent, reward_type, reward_amount, sort_order)
VALUES
    (1, 'Бронзовый', 0, 3, 'percent', 0, 1),
    (1, 'Серебряный', 5000, 5, 'percent', 0, 2),
    (1, 'Золотой', 15000, 7, 'percent', 0, 3),
    (1, 'Платиновый', 50000, 0, 'fixed', 100, 4);
```

---

## 7. Definition of Done

Phase 1 считается завершенной, когда выполнены все пункты:

### 1A. Иерархия сущностей

- [ ] Миграция `00014_entity_hierarchy.sql` применена без ошибок
- [ ] Миграция откатывается без ошибок (`goose down`)
- [ ] `Bot` создается с опциональным `program_id`
- [ ] `POSLocation` создается с опциональным `bot_id`
- [ ] `BotClient` автоматически получает `phone_normalized` и `qr_code` при регистрации
- [ ] `NormalizePhone` обрабатывает форматы: `+79991234567`, `89991234567`, `9991234567`
- [ ] `LoyaltyLevel` поддерживает `reward_type` = `percent` и `fixed`
- [ ] Frontend: формы создания бота и POS содержат опциональные селекторы
- [ ] Frontend: профиль клиента показывает QR-код
- [ ] Unit-тесты на `NormalizePhone` и `GenerateQRCode` проходят
- [ ] Integration-тесты на новые repository-методы проходят

### 1B. Loyalty Engine

- [ ] Миграция `00015_balance_reserves.sql` применена без ошибок
- [ ] `CalculateBonus` корректно считает % и фикс. бонусы
- [ ] `EarnFromCheck` начисляет бонусы и обновляет уровень
- [ ] `DemoteClients` понижает уровень, если threshold не достигнут
- [ ] Cron-задача `loyalty-demotion` зарегистрирована (интервал 24 часа)
- [ ] `ReservePoints` / `ConfirmReserve` / `CancelReserve` работают
- [ ] Cron-задача `expire-reserves` зарегистрирована (интервал 5 минут)
- [ ] Endpoint `GET /api/v1/clients/identify` находит клиента по телефону и QR
- [ ] Endpoint `POST /api/v1/loyalty/earn-from-check` начисляет бонусы
- [ ] Unit-тесты на `CalculateBonus`, `DemoteClients`, `ReservePoints` проходят
- [ ] Integration-тесты на полный цикл earn/spend/reserve проходят

### 1C. Файловое хранилище

- [ ] MinIO запускается в docker-compose (dev и prod)
- [ ] `POST /api/v1/files/upload` загружает файл и возвращает URL
- [ ] Файлы больше 50MB отклоняются с ошибкой 413
- [ ] Интерфейс `Storage` определен, реализация через MinIO
- [ ] Зависимость `minio-go/v7` добавлена в `go.mod`
- [ ] Integration-тест на upload проходит

### Общее

- [ ] `go vet ./...` без ошибок
- [ ] `go build ./cmd/server` и `go build ./cmd/bot` собираются
- [ ] `npm run build` во frontend собирается
- [ ] `npm run lint` без ошибок
- [ ] Все миграции применяются последовательно на чистой БД
- [ ] Документация CLAUDE.md обновлена (порты MinIO, новые миграции)
