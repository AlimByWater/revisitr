# Wallet: Карты лояльности (Apple / Google)

## Статус

| Платформа | Статус | Примечание |
|-----------|--------|------------|
| Apple Wallet | ~95% | pkpass генерация + web service + device registration + APNs push-обновления + admin download. Проверено на реальном продакшн-`.p12` (подпись верифицируется до Apple Root). Осталось открыть пасс на iPhone |
| Google Wallet | ~85% | REST API создание класса/объекта + JWT save URL + admin save button готово. Нужен Google Cloud service account + Google Pay & Wallet console |

---

## Что реализовано (Apple Wallet)

### Бэкенд

- **`internal/usecase/wallet/passgen.go`** — генератор `.pkpass`:
  - pass.json (store card, balance → primaryFields, level → secondaryFields, QR-code → barcode)
  - QR-код (go-qrcode, данные из `bot_clients.qr_code`)
  - Загрузка .p12 сертификата (base64 из credentials, декод через `software.sslmate.com/src/go-pkcs12` — поддерживает современные `.p12` с SHA-256 MAC)
  - Apple WWDR G4 промежуточный сертификат вшит в `wwdr.go` и автоматически добавляется в цепочку подписи (в `.p12` обычно только leaf)
  - manifest.json (SHA-1 хеши)
  - PKCS#7 detached signature (encoding/asn1, без внешних зависимостей)
  - Zip-сборка (icon.png, logo.png опционально из logo_url, signature)
- **`internal/entity/wallet.go`** — расширен `WalletDesign`:
  - `OrganizationName` — отображается на карте
  - `WebServiceURL` — переопределение URL вэб-сервиса (автоопределяется из request)
- **Apple Web Service эндпоинты** (public, без JWT):
  - `GET /v1/passes/:passTypeId/:serial` — отдаёт pkpass (auth: `ApplePass <token>`, поддержка `If-Modified-Since` → 304)
  - `POST /v1/devices/:deviceLibraryId/registrations/:passTypeId/:serial` — регистрация устройства (сохраняет push-токен, auth: `ApplePass <token>`)
  - `GET /v1/devices/:deviceLibraryId/registrations/:passTypeId?passesUpdatedSince=<tag>` — список серийников обновлённых пассов
  - `DELETE /v1/devices/:deviceLibraryId/registrations/:passTypeId/:serial` — отмена регистрации
  - `POST /v1/passes/:passTypeId/:serial/log`, `POST /v1/log` — заглушки логирования
- **Admin download** (JWT):
  - `GET /passes/:serial/download` — для админки
- **Device registration** — таблица `wallet_device_registrations` + репозиторий (upsert/delete/list — реализовано)
- **APNs push-обновления** — `internal/usecase/wallet/apns.go`: token-based (.p8, ES256 JWT) отправка пустого пуша на `api.push.apple.com`. При изменении баланса (`RefreshPassBalance`) всем зарегистрированным устройствам пасса шлётся пуш → устройство подтягивает обновлённый pkpass. Конфиг: `apns_key` (base64/PEM .p8) + `apns_key_id` в credentials Apple-конфига (team_id/pass_type_id переиспользуются)
- **Интеграция с loyalty** — `EarnPoints` / `SpendPoints` вызывают `RefreshPassBalance` (обновляет кешированный баланс в `wallet_passes`)

### Фронтенд

- **`/dashboard/loyalty/wallet`** — страница админки:
  - Статистика (всего/активных/Apple/Google карт)
  - Настройка платформы (вкл/выкл, credentials, дизайн)
  - Конструктор: название организации, цвета (фон/текст/метки), загрузка лого
  - Таблица выданных пассов с возможностью скачать `.pkpass` и отозвать
- **Загрузка лого** — через существующий `/api/v1/files/upload`
- **Кнопка скачивания** — для активных Apple Passes в таблице

### Миграции

- `00046_wallet_device_registrations.sql` — таблица для привязки устройств к пассам

---

## Что реализовано (Google Wallet)

### Бэкенд

- **`internal/usecase/wallet/googlerest.go`** — REST API клиент для Google Wallet:
  - `EnsureClass()` — создаёт/обновляет LoyaltyClass через `walletobjects.googleapis.com`
  - `CreateObject()` — создаёт/обновляет LoyaltyObject через REST API (баланс, QR, цвета)
  - `GenerateSaveURL()` — подписывает JWT (RS256) со ссылкой на существующий объект
- **`internal/usecase/wallet/googlesave.go`** — (legacy) генератор JWT с встроенным классом + объектом
- **`internal/usecase/wallet/wallet.go`** — `GenerateGoogleSaveURL()`:
  - Проверяет платформу, статус, credentials
  - Возвращает ссылку `https://pay.google.com/gp/v/save/<jwt>`
- **Интеграция**: при сохранении Google Wallet конфига → `EnsureClass()`, при выпуске пасса → `CreateObject()`
- **Эндпоинт** (JWT):
  - `GET /passes/:serial/google-save` — возвращает `{"url": "..."}` для админки

### Фронтенд

- **Кнопка "Save to Wallet"** — для активных Google Passes в таблице
  - GET /wallet/passes/:serial/google-save → открывает ссылку в новом окне

---

## Google Wallet: как настроить

### Без Google Cloud ($0)

```bash
cd backend && go test ./internal/usecase/wallet/... -v -count=1 -run TestGenerateGoogle
```

### С Google Cloud (бесплатно, нужен аккаунт)

1. Создать проект в [Google Cloud Console](https://console.cloud.google.com)
2. Включить [Google Wallet API](https://console.cloud.google.com/apis/library/walletobjects.googleapis.com)
3. Создать service account, скачать JSON-ключ
4. В [Google Pay & Wallet Console](https://pay.google.com/business/console/) добавить service account как пользователя (Developer)
5. Получить Issuer ID из настроек аккаунта
6. В админке Revisitr: заполнить Issuer ID и содержимое JSON-ключа Service Account
7. Включить, сохранить
8. Выпустить пасс клиенту → кнопка "Save to Wallet" в таблице

---

## Как тестировать

### Без Apple Developer ($0)

```bash
cd backend && go test ./internal/usecase/wallet/... -v -count=1
```

### С Apple Developer ($99/год)

1. Создать Pass Type ID в Apple Developer Portal
2. Выпустить сертификат, скачать `.p12`
3. `base64 -i certificate.p12 | pbcopy`
4. В админке: заполнить Pass Type ID, Team ID, base64 сертификата
5. (Опционально, для push-обновлений баланса) создать APNs-ключ `.p8`, заполнить APNs Key ID + содержимое `.p8`
6. Включить, сохранить
7. Через API выпустить пасс клиенту
8. Скачать `.pkpass` из админки → открыть на iPhone

> WWDR-промежуточный сертификат добавлять не нужно — он вшит в код и подмешивается в подпись автоматически.

---

## Что не реализовано (deferred)

### Google Wallet (partial — ~60%)

- ✅ JWT-link генерация через service account key
- ✅ Эндпоинт `/passes/:serial/google-save` + кнопка в админке
- ❌ REST API обновление баланса (сейчас только JWT при первом сохранении)
- ❌ Google Pay для создания классов/объектов через HTTP API (не нужно для JWT-link подхода)

### Полноценный конструктор карт

- Сейчас: только цвета + лого (минимальный набор)
- В будущем: загрузка макетов (background, strip, thumbnail), кастомные поля

### Выпуск пасса через бота

- Сейчас: только через API или админку
- В будущем: бот отправляет ссылку на скачивание / кнопку "Add to Wallet"

---

## Архитектура

```
wallet_configs (per-org платформа)
  ├── credentials: pass_type_id, team_id, certificate (base64 .p12)
  ├── design: colors, logo_url, organization_name, description
  └── is_enabled

wallet_passes (per-client пасс)
  ├── serial_number (32 hex)
  ├── auth_token (64 hex, для ApplePass авторизации)
  ├── push_token (для APNs)
  ├── last_balance, last_level (кеш для быстрой отдачи)
  └── status: active | suspended | revoked

wallet_device_registrations (device → pass)
  ├── device_library_id (от Apple)
  ├── pass_type_id
  ├── serial_number
  └── push_token
```
