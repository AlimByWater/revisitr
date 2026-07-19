# iikoFront Plugin — Pipeline Validation Playbook (native sample first)

**Цель:** доказать, что весь путь развёртывания плагина работает на тестовом
iikoFront — **до** того, как писать нашу логику. Раскатываем родной
`SamplePaymentPlugin` из iiko SDK. Если он грузится и виден как тип оплаты на
кассе — значит сборка, лицензия, регистрация и загрузка работают, и можно
писать Revisitr-плагин на той же рельсе.

Все шаги выполняются на **Windows-хосте** с тестовым iikoFront. macOS/этот
репозиторий тут ни при чём — здесь только инструкция.

---

## 0. Предусловия (выяснить ДО кода)

1. **Версия тестового iikoFront.** iikoFront → Справка → о программе. Запиши
   полную версию (напр. `8.x` / `7.x`). От неё зависит, какой пакет
   `Resto.Front.Api.Vх` брать — версия API **должна совпадать** с версией
   iikoFront, иначе плагин не загрузится.
2. **Лицензия — РЕШЕНО, бесплатно.** Поддержка iiko подтвердила: регистрация
   через портал ещё не реализована, поэтому пока используем бесплатные ModuleID
   напрямую в коде. Для нас — лицензия **`iikoAPIPayment`**, ModuleID
   **`21016318`** (явная категория: «Интеграция с системами лояльности»).
   В плагине прописывается как атрибут: `[PluginLicenseModuleId(21016318)]`.
   Альтернативный бесплатный ModuleID: `19011701` (iikoAPIFront, более общий).
   Источник: <https://ru.iiko.help/articles/#!api-documentations/licenzii-api>
3. **Инструменты сборки на Windows:** Visual Studio 2022 (или Build Tools) +
   .NET Framework **4.7.2** Developer Pack. SDK таргетит именно 4.7.2.

---

## 1. Забрать SDK

```powershell
git clone https://github.com/iiko/front.api.sdk.git
cd front.api.sdk
```

Внутри — папка `sample/` с примерами. Нужен `Resto.Front.Api.SamplePaymentPlugin`
(регистрирует внешний тип оплаты — ровно наш кейс).

---

## 2. Выбрать версию API под свой iikoFront

Пакеты лежат на NuGet под профилем `iiko`
(<https://www.nuget.org/profiles/iiko>). Например:

- `Resto.Front.Api.V7` (напр. `7.9.6015`)
- `Resto.Front.Api.V6` (напр. `7.0.6022`)
- `Resto.Front.Api.V8` / `V9Preview` — для новых сборок

Открой проект sample под нужную версию (в `sample/` они разложены по `v6`,
`v7`, …) и убедись, что `PackageReference`/ссылка на `Resto.Front.Api.Vх`
соответствует версии твоего iikoFront (из шага 0.1).

---

## 3. Собрать sample

В Visual Studio: открыть решение sample → выбрать конфигурацию **Release** →
Build. Либо из консоли:

```powershell
msbuild Resto.Front.Api.SamplePaymentPlugin.csproj /p:Configuration=Release
```

Результат — `Resto.Front.Api.SamplePaymentPlugin.dll` (+ зависимости) в `bin\Release`.

---

## 4. Разложить в папку плагинов iikoFront

Каждый плагин — в **своей подпапке** папки `Plugins` рядом с `iikoFront.exe`.
По умолчанию:

```
C:\Program Files\iiko\iikoRMS\Front.Net\Plugins\Resto.Front.Api.SamplePaymentPlugin\
```

Скопируй туда `.dll` (и все зависимости из `bin\Release`). Выдай пользователю,
под которым работает iikoFront, **права на запись** в эту папку (иначе плагин
не сможет писать/логировать).

> iikoFront изолирует каждый плагин в отдельном процессе-контейнере — падение
> одного плагина не роняет кассу.

---

## 5. Добавить тип оплаты в backoffice

Внешний тип оплаты появляется только **после успешного старта плагина**, но
завести его нужно в backoffice:

1. Backoffice → **Типы оплат** → Добавить.
2. Настроить по учётной политике (безналичный, «Api Payment»-подобный тип).
3. После старта плагина в списке появится внешний тип с `paymentSystemKey`
   плагина (у sample — свой ключ; у нашего будущего плагина будет `Revisitr`).

---

## 6. Перезапустить iikoFront и проверить загрузку

1. Полностью перезапусти iikoFront.
2. Открой лог плагина:
   ```
   %appdata%\Roaming\iiko\CashServer\Logs\plugin-Resto.Front.Api.SamplePaymentPlugin.log
   ```
3. Убедись, что плагин **инициализировался без ошибок** (нет исключений,
   есть строки о старте). Если лога нет — плагин не подхватился: проверь
   папку/права/версию API/лицензию.

---

## 7. Проверить на кассе

1. Открой смену.
2. Создай заказ, перейди к оплате.
3. Убедись, что наш **внешний тип оплаты доступен** в списке типов оплат.
   (Sample-плагин на выбор этого типа покажет свой демо-диалог/поведение.)

Если тип оплаты виден и плагин реагирует — **пайплайн доказан**. Дальше можно
писать Revisitr-плагин точно так же: та же сборка, та же папка, тот же
тип оплаты, только внутри — HTTP-вызовы нашего API (`identify`/`redeem`/`accrue`).

---

## 8. Траблшутинг

| Симптом | Вероятная причина | Что делать |
|---|---|---|
| Лога плагина нет | Не подхватился | Проверь имя подпапки, что `.dll` на месте, права на запись |
| В логе ошибка версии/типов | API ≠ версии iikoFront | Пересобрать под правильный `Resto.Front.Api.Vх` |
| Плагин не стартует, лицензия | Неверный ModuleID в атрибуте | Проверь `[PluginLicenseModuleId(21016318)]` в коде |
| Тип оплаты не появился | Плагин не стартовал ИЛИ тип не заведён | Сначала лог (шаг 6), потом backoffice (шаг 5) |

---

## 9. Что это доказывает и что дальше

Пройденный плейбук закрывает главный неизвестный кусок — **как плагин
собирается, лицензируется, ставится и грузится**. После этого:

- **Шаг 2 (следующий):** минимальный Revisitr-плагин — диалог ввода
  слова-кода → `identify` → показ баланса → `redeem` как оплата + `accrue` при
  закрытии чека. Контракт API описан в `PLUGIN_CONTRACT.md`, дизайн плагина —
  в `PLUGIN_DESIGN.md`. Backend уже готов и протестирован
  (`backend/scripts/test_plugin.sh`).
- Плагин будет ходить на наш backend: локально — через tuna-тунель до
  `:9721`, либо на прод (когда там накатят миграцию `00046`).

---

## Источники

- iiko Front API SDK: <https://github.com/iiko/front.api.sdk>
- SamplePaymentPlugin (v6):
  <https://github.com/iiko/front.api.sdk/blob/master/sample/v6/Resto.Front.Api.SamplePaymentPlugin/SamplePaymentPlugin.cs>
- `IFrontPlugin` (v7 reference):
  <https://iiko.github.io/front.api.sdk/v7/html/T_Resto_Front_Api_IFrontPlugin.htm>
- iikoFront API docs (intro): <https://iiko.github.io/front.api.doc/intro.html>
- NuGet (профиль iiko): <https://www.nuget.org/profiles/iiko>
