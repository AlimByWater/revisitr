# Revisitr iiko Loyalty Plugin

Плагин iikoFront (Front.Net, `.dll`), который добавляет тип оплаты **«Revisitr
Бонусы»**: касса идентифицирует гостя по одноразовому слову-коду (ввод вручную
или скан QR), списывает бонусы как оплату и начисляет баллы на денежную часть
чека. Вся бизнес-логика — на бэкенде Revisitr; плагин лишь UI + HTTP.

- Контракт API: `../docs/integrations/iiko/PLUGIN_CONTRACT.md`
- Дизайн: `../docs/integrations/iiko/PLUGIN_DESIGN.md`
- Плейбук раскатки: `../docs/integrations/iiko/PLUGIN_SAMPLE_PLAYBOOK.md`

> Собирается и запускается **только на Windows** с установленным iikoFront.
> Этот проект не собирается на macOS/Linux (нужны Windows-референсы и
> NuGet-контракт iiko).

## Целевая версия API

Собран под **Resto.Front.Api.V8** (`8.7.6032`). V8 obsolete, но поддерживается
iikoFront вплоть до 10.3 — покрывает тестовый хост 9.4 и широкую полосу
клиентских касс. Один бинарь работает на всём диапазоне версий, поддерживающих
V8 (см. `PLUGIN_DESIGN.md` §1.1 про универсальность).

Если у клиента касса вне полосы V8 — тот же исходник пересобирается под другую
версию (V6/V9), меняется только `PackageReference` на контракт. Это
конфигурация сборки, не отдельная кодовая база.

## Лицензия

`[PluginLicenseModuleId(21016318)]` — бесплатный ModuleID **iikoAPIPayment**
(категория «системы лояльности»), выдан поддержкой iiko. Ничего покупать/ждать
портал не нужно.

## Предусловия для сборки

- Windows + Visual Studio 2022 (или Build Tools) + .NET Framework **4.7.2**
  Developer Pack.
- Доступ к NuGet (пакеты `Resto.Front.Api.V8`, `Resto.Front.Api.PluginPackaging`
  из профиля iiko публичные).

## Сборка

```powershell
cd iiko-plugin\RevisitrPlugin
dotnet build -c Release
# или msbuild RevisitrPlugin.csproj /p:Configuration=Release
```

Результат — `bin\Release\Resto.Front.Api.RevisitrPlugin.dll` (+
`revisitr.plugin.config.example.json`).

⚠️ **Контракт-DLL не поставляется** с плагином — iikoFront даёт его в рантайме.
За это отвечает пакет `Resto.Front.Api.PluginPackaging`; для V8 исключение
контракта из вывода происходит автоматически. Если положить
`Resto.Front.Api.V8.dll` рядом с плагином вручную — будет
`TypeLoadException` при загрузке. Не делай этого.

## Установка на кассу

1. Создай папку рядом с `iikoFront.exe`:
   ```
   C:\Program Files\iiko\iikoRMS\Front.Net\Plugins\Resto.Front.Api.RevisitrPlugin\
   ```
   и выдай пользователю iikoFront права на запись в неё.
2. Скопируй туда содержимое `bin\Release\` (как минимум `.dll` и всё, что
   положил packaging, **кроме** контракт-DLL).
3. Переименуй `revisitr.plugin.config.example.json` → `revisitr.plugin.config.json`
   и впиши:
   - `baseUrl` — адрес бэкенда Revisitr (только HTTPS в проде);
   - `apiKey` — ключ заведения из веб-кабинета Revisitr (`rvk_…`, показывается
     один раз);
   - `timeoutSeconds` — таймаут HTTP (по умолчанию 6).
4. В iikoOffice/backoffice → **Типы оплат** → Добавить внешний тип оплаты,
   связанный с плагином (после старта плагина появится тип с ключом `revisitr`).
5. Перезапусти iikoFront. Проверь лог:
   ```
   %appdata%\Roaming\iiko\CashServer\Logs\plugin-Resto.Front.Api.RevisitrPlugin.log
   ```
   Должна быть строка `Revisitr payment system 'revisitr' registered`.

## Как работает на кассе

1. Кассир в экране оплаты выбирает **«Revisitr Бонусы»**.
2. Открывается диалог: гость называет слово (кассир вводит) **или** гость
   показывает QR (кассир сканирует) — один диалог, оба канала.
3. Плагин зовёт `identify` → показывает имя, баланс и «доступно к списанию».
4. Сумма списания автоматически ограничивается доступным (нельзя списать больше
   баланса/чека). Кассир подтверждает сумму (или ставит 0 — только начисление).
5. При закрытии оплаты плагин зовёт `redeem` (если сумма > 0), затем `accrue` на
   денежную часть чека.
6. Гостю прилетает уведомление в Telegram-боте (на стороне Revisitr).

## Поведение при сбоях

- **Плагин не блокирует кассу.** `identify`/`redeem` недоступны → показывается
  понятная ошибка, заказ можно закрыть обычным способом без бонусов.
- `redeem`/`accrue` идемпотентны по `order_id` — ретраи при таймауте безопасны,
  двойного списания нет.
- Ошибка `accrue` не валит уже прошедшую оплату (логируется, глотается).

## Известные ограничения (MVP)

- **Возврат/сторно НЕ восстанавливает списанные бонусы.** При отмене закрытого
  заказа `ReturnPayment`/`EmergencyCancelPayment` только пишут WARN в лог.
  Полноценный возврат требует серверного `cancel`-эндпоинта или схемы
  reserve/confirm (заложено в дорожную карту `PLUGIN_DESIGN.md` §7).
- Начисление считается от денежной части (`ResultSum − сумма_бонусов`). Если
  программа должна начислять на полный чек — правится политика на бэкенде.
- API-ключ хранится в конфиге в открытом виде рядом с DLL. Доступ к папке
  плагина = доступ к ключу; ограничивай правами ОС.

## Структура

| Файл | Назначение |
|------|-----------|
| `RevisitrPlugin.cs` | точка входа `IFrontPlugin`, ModuleID, регистрация типа оплаты |
| `RevisitrPaymentProcessor.cs` | `IPaymentProcessor`: CollectData → Pay (redeem+accrue) |
| `RevisitrApiClient.cs` | HTTP-клиент к бэкенду (identify/redeem/accrue) |
| `PluginSettings.cs` | чтение `revisitr.plugin.config.json` |
| `SessionData.cs` | данные сессии между CollectData и Pay (rollback data) |
