# iiko Integration — Live Findings (Test Stand)

Last updated: 2026-06-18

## Phase status (DONE 2026-06-18)

- **422 window fix**: `GetOrders` now chunks from..to into 24h windows
  (`fetchDeliveriesChunked` / `fetchDeliveries`) and merges. Verified live: 30-day
  window no longer 422s.
- **Phase 6 (aggregates)**: `GetDailyAggregates` computes per-day
  revenue/avg_check/tx_count/guest_count/phones from /deliveries (OLAP not
  available — see below). `SyncAll` now calls `SyncAggregates` after `Sync`.
- **Phase 4 (customers)**: `GetCustomers` resolves ONE customer by phone via
  `/loyalty/iiko/customer/info` (Cloud API has no bulk list). Degrades to empty
  (not error) on 400/401/403/404 via `iikoLoyaltyUnavailable` — so missing
  iikoCard doesn't break sync.
- Tests: `internal/service/pos/iiko_phases_test.go`. Full suite green (27 ok, 0 fail).
- Live smoke verified all paths against the real stand (menu returns real data,
  orders/aggregates/customers return empty without errors — stand is empty + loyalty off).

## Onboarding rework (DONE 2026-06-18)

- **Backend discovery**: `POST /api/v1/integrations/discover` with
  `{type, config:{api_key, api_url}}` → `{organizations:[{id,name}], external_menus:[{id,name}]}`.
  No saved integration needed. Provider methods `ListOrganizations` +
  `ListExternalMenus` (iiko.go); `SyncService.Discover` builds a provider with a
  placeholder org_id ("discovery") since org_id is what's being chosen. Menu
  listing degrades to nil if not permitted. Verified live: api_key alone returns
  "Мой ресторан" + "Revisitr Demo Menu".
- **Frontend**: `CreateIntegrationModal` iiko path is now 2-step — enter api_key →
  discover → pick organization (+optional external menu from dropdowns) → create.
  Other types (mock/rkeeper) keep the simple form. New: `integrationsApi.discover`,
  `useDiscoverIntegrationMutation`, `POSDiscovery` type, `external_menu_id` in
  config, mock route for `/integrations/discover`. tsc + vitest + build all green.

## Two different iiko APIs (critical distinction)

The project code (`backend/internal/service/pos/iiko.go`) targets **iiko Cloud /
Transport API**, NOT the on-premise RMS resto API. Do not confuse them.

| | Cloud / Transport (our code) | RMS resto Server API |
|---|---|---|
| Base | `https://api-ru.iiko.services` | `https://<stand>.iiko.it/resto` |
| Auth | POST `/api/1/access_token` `{apiLogin}` → Bearer token (TTL 1h) | GET `/resto/api/auth?login=&pass=sha1(pass)` → token |
| Format | JSON | XML (mostly) |
| Our provider | YES | no |

## Test stand credentials (Cloud API)

- `IIKO_API_LOGIN=a0e85e6f92cd4d46abd955b025e7d492`
- `IIKO_ORG_ID=22fc8cb3-0e70-4c9a-b195-8d2301ee0c43` (org name: "Мой ресторан", isCloud=true)
- `IIKO_EXTERNAL_MENU_ID=82279` (name: "Revisitr Demo Menu")
- Stored in `.env.local` (not committed, outside what `find` surfaced).

## RMS stand (separate, for back-office config via browser)

- iikoWeb: `https://260-347-461.iikoweb.ru/navigator/ru-RU/index.html#/main`
- resto: `https://260-347-461.iiko.it/resto/`
- Login: `user` / `user#test`, PIN `1111`, CrmID `9345251`
- resto auth verified live: `GET /resto/api/auth?login=user&pass=sha1('user#test')` → 200 token
- resto departments: corporation → jurperson "ООО Мое торговое предприятие" → department "Мой ресторан" (id `7a35c826-7314-9d1c-019e-5eb21b0a0012`)

## Live endpoint status (Cloud API, verified 2026-06-17)

| Endpoint | Status | Notes |
|---|---|---|
| `/api/1/access_token` | ✅ 200 | token present |
| `/api/1/organizations` | ✅ 200 | "Мой ресторан", isCloud=true |
| `/api/2/menu` (list) | ✅ 200 | 1 external menu: id 82279 "Revisitr Demo Menu" |
| `/api/2/menu/by_id` (82279) | ✅ 200 | 1 category "Coffee Demo", 3 items (e.g. "Revisitr Cookie" sku 00005) |
| `/api/1/nomenclature` | ✅ 200 but EMPTY | groups=0 products=0 revision=0 — external menu replaces it |
| `/api/1/deliveries/by_delivery_date_and_status` 1-day | ✅ 200 | 0 orders (terminal disabled) |
| same, 7/30/60-day window | ❌ 422 | `TOO_MANY_DATA_REQUESTED` — must chunk window per-day |
| `/api/1/terminal_groups` includeDisabled=false | 200 | 0 items |
| `/api/1/terminal_groups` includeDisabled=true | 200 | 1 item, active=0 (disabled → cannot create orders) |
| `/api/1/reports/olap` | ❌ 401 | `Right api/1/reports/olap is not allowed for this ApiLogin` |
| `/api/1/loyalty/iiko/customer/info` (& category, program) | ❌ 400 | `Transport_WrongCrmId` / `Organization 9345251 not found` — iikoCard CRM not bound to this Cloud login |

## endpointURL path handling (already correct in code)

`iiko.go endpointURL()` (lines 233-240): if path starts with `/api/`, it strips
the `/api/1` suffix from baseURL and joins at domain root. So `/api/2/menu/by_id`
resolves to `https://api-ru.iiko.services/api/2/menu/by_id` correctly. A curl
probe that naively does `$BASE/api/2/...` (where $BASE already ends `/api/1`)
produces a WRONG `/api/1/api/2/...` → 401. That 401 is a probe artifact, NOT a
real permission block. The code is fine.

## Bugs / gaps found

1. **422 on multi-day windows (REAL prod bug).** `GetOrders` and `SyncAggregates`
   request a 30-day window in one call. iiko caps deliveries window (>1 day → 422
   on this stand). Must chunk the from..to range (per-day or small windows) and
   merge. Independent of phases 4/6.
2. **OLAP closed** for this apiLogin (401). Natural source for daily aggregates
   (full revenue incl. dine-in). Either enable `api/1/reports/olap` right for the
   apiLogin in iiko integration settings, or compute aggregates from /deliveries
   (delivery-only, incomplete revenue).
3. **Loyalty/customers closed** (`WrongCrmId`). iikoCard CRM not configured on
   this Cloud login → Phase 4 (GetCustomers) cannot be live-verified here.

## What CAN be live-tested end-to-end

auth, organizations, **external menu (real data!)**, deliveries (empty result but
real call + window chunking). Menu sync is fully live-testable. OLAP + loyalty are
the only truly blocked areas on this stand.

## iikoWeb back-office findings (browser, 2026-06-17)

Logged into iikoWeb (`user`/`user#test`) → **"Настройки Cloud API"**
(`/integration-management`). This is where the apiLogin is created/managed.

Existing integration "**revisitr-dev**":
- API key: `a0e***492` (our IIKO_API_LOGIN)
- Status: active until 27.05.2028
- **Права (permission template): "Все права" (ALL rights)** ← yet OLAP still 401
- Внешнее меню: "Revisitr Demo Menu" (82279); Источник цен: Внешнее меню
- Email: dev@revisitr.local; "ПЕРЕСОЗДАТЬ" button regenerates the key

**KEY CONCLUSION on OLAP:** apiLogin has "Все права" and `/reports/olap` is STILL
401. So OLAP is NOT gated by the permission template — it is simply not part of
the Cloud/Transport API surface for this stand/tariff. Enabling a right won't fix
it. Daily aggregates for restaurants must come from another path (deliveries, or a
different reporting integration). Don't promise OLAP to the user as a quick toggle.

## How REAL clients onboard iiko (vs our test stand)

Our test stand and a real client follow the SAME path — there is nothing special
about our stand except the data is empty:

1. Client logs into THEIR iikoWeb/iikoOffice back-office.
2. Goes to "Настройки Cloud API" → "Добавить интеграцию".
3. Creates an apiLogin: names it, sets permission template, binds an external menu
   + price source, sets expiry. iiko generates the **API key (apiLogin GUID)**.
4. Client copies that apiLogin into the Revisitr admin panel integration form.
5. Revisitr resolves `org_id` automatically via `/organizations` (already coded in
   TestConnection) — the client does NOT need to find the org GUID manually.
6. `external_menu_id` comes from the bound external menu (listable via `/api/2/menu`).

So the Revisitr onboarding form really only needs to collect the **apiLogin**;
org_id and external menu can be discovered via API and offered as a pick-list.
This is a concrete improvement over the current config (which asks for org_id and
external_menu_id manually).

Caveat: features like OLAP reports and iikoCard loyalty depend on the client's iiko
tariff/modules, not just the apiLogin permission template. Capabilities vary
per client → integration must degrade gracefully when an endpoint returns 401/400.

## Stand limitation (matches prior project memory)

Terminal group is disabled (active=0) → cannot create test orders → deliveries
always empty. To get live order/aggregate data, terminal group must be enabled in
iikoOffice/BackOffice (RMS side), or use mock data per iiko docs.

## 2026-06 update — terminal activated, orders flowing, mechanics solved

The "terminal disabled" / "deliveries always empty" limitation above is RESOLVED.
All the non-obvious mechanics we worked out (terminal activation via the iikoWeb
`isEnabledForTransport` flag, the create→close validation chain, the
`DeliveryByClient`-closes-but-`DeliveryByCourier`-doesn't trick, blocked endpoints,
where real prices/payment types live, the deliveries-window timezone bug + fix,
and the seed tooling) are written up in **`docs/integrations/iiko/PLAYBOOK.md`** —
read that first. Seed scripts: `seed_demo_orders.sh` (live Cloud orders),
`seed_history.py` (3 months of DB history for analytics/dynamics).
