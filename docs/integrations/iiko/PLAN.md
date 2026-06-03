# iiko Integration Plan

## Product Decision

Revisitr is the source of truth for loyalty:
- clients;
- balances;
- reward/spend rules;
- bonus transactions;
- client-to-order links;
- visit and menu-item history.

iiko is treated as a POS data source first:
- organizations;
- menu, categories, modifiers, prices;
- stop lists;
- orders/checks;
- order items;
- payments, discounts, timestamps;
- POS customers only when available.

Do not assume iiko knows the guest. Many restaurants close checks without a customer
attached. Revisitr must be able to link a guest to a check using its own identity
flow.

## Target Architecture

### Layer 1 — Read-only POS sync

Primary API: iikoCloud Transport API (`https://api-ru.iiko.services/api/1`)

Auth: `apiLogin` token. This is separate from RMS/demo credentials.

OpenAPI reference: `https://api-ru.iiko.services/docs`
(`Redoc.init('/api-docs/docs', ...)`).

Used for:
- org discovery;
- menu sync;
- stop-list sync;
- customer import when available;
- closed order/check sync;
- OLAP/report data.

### Layer 2 — Loyalty engine in Revisitr

Revisitr owns:
- loyalty accounts;
- earn/spend rules;
- pending redemptions;
- reserve/confirm/cancel flow;
- order matching;
- final bonus transactions;
- customer purchase history by item/category.

### Layer 3 — Redemption at POS

Ideal path: iikoFront plugin.

Fallback path: manual code flow through Revisitr cashier UI.

Read-only API alone is not enough for full POS-side redemption. It can prove that
a check happened, but it cannot reliably apply a discount or bonus payment inside
an open iikoFront check.

## Integration Capability Matrix

Every POS provider should expose capabilities explicitly:

| Capability | Meaning | iiko target |
| --- | --- | --- |
| `read_organizations` | fetch restaurants/locations | Transport API |
| `read_menu` | fetch categories, dishes, modifiers, prices | Transport API |
| `read_stoplists` | fetch unavailable dishes/modifiers | Transport API |
| `read_orders` | fetch closed checks/orders | Transport API / Resto fallback |
| `read_order_items` | fetch item-level purchase data | Transport API / OLAP |
| `read_customers` | fetch POS customer records if configured | Transport API |
| `attach_customer_to_order` | POS can store customer on check | plugin/manual flow dependent |
| `apply_discount` | apply discount/bonus in open check | iikoFront plugin |
| `create_order` | send order into POS | deferred |
| `webhooks` | push events instead of polling | provider dependent |

MVP must work with only read capabilities plus Revisitr-owned matching.

## Order Matching Model

Because iiko may not know the customer, order matching must not rely on POS
customer id alone.

Signals:
- Revisitr one-time QR/code;
- phone number;
- POS customer id if present;
- organization/location;
- terminal/cash register;
- order/check id;
- time window;
- total amount;
- item composition;
- payment/discount marker if plugin/manual flow writes one.

Match confidence:
- `exact`: explicit code, plugin session id, or POS check id callback;
- `high`: phone + time + amount/location;
- `medium`: time + amount + location, no identity signal;
- `manual_review`: ambiguous candidates.

Required entities or equivalent:
- `redemption_sessions`;
- `pending_visits`;
- `order_customer_links`;
- `pos_orders`;
- `pos_order_items`;
- `bonus_transactions`;
- `balance_reserves`.

Current schema note: `external_orders.items` is JSONB and now stores item
`external_id`, `name`, `quantity`, `price`, and `category` snapshots. Before
production, decide whether to normalize this into `pos_order_items` for easier
analytics and indexing.

## Redemption Flows

### Flow A — iikoFront plugin (ideal)

Goal: full bonus redemption inside an open iikoFront check.

Expected cashier flow:
1. Cashier opens check in iikoFront.
2. Guest shows QR/code or says phone.
3. Plugin calls Revisitr: find guest, get balance and available spend amount.
4. Cashier chooses spend amount.
5. Revisitr creates/reserves pending spend.
6. Plugin applies discount/payment marker in iikoFront.
7. Check closes.
8. Plugin or sync sends final check id/order id to Revisitr.
9. Revisitr confirms reserve and links order to client.
10. Revisitr stores order items for history/segmentation.

Plugin must handle:
- network failure;
- insufficient balance;
- cashier cancel;
- check cancel;
- refund;
- partial payment;
- duplicate close events;
- reserve expiry;
- idempotent confirm/cancel.

Open questions:
- discount vs payment type for bonus spend;
- required iikoFront license/options;
- deployment model for plugin updates;
- whether plugin can persist Revisitr session id into check metadata/comment;
- exact receipt/fiscal constraints for bonus spend.

### Flow B — Manual code flow (fallback)

Goal: redemption without iikoFront plugin.

Expected cashier flow:
1. Guest opens Telegram/WebApp and requests spend.
2. Revisitr creates one-time code/QR and reserves points.
3. Cashier enters code in Revisitr cashier UI.
4. Cashier manually applies configured discount/payment in POS.
5. POS sync later imports closed check.
6. Revisitr matches check by code session, time, amount, location, cashier.
7. Revisitr confirms reserve and links order to client.

Tradeoffs:
- works with almost any POS;
- more manual action for cashier;
- higher mismatch/error risk;
- needs manual review queue for ambiguous checks.

## Demo Stand (RMS, sandbox)

- CrmID: `9345251`
- Domain: `https://260-347-461.iiko.it/resto/`
- BackOffice login: `user` / `user#test` / pin `1111`
- iikoWeb: `https://260-347-461.iikoweb.ru/navigator/ru-RU/index.html#/main`
- Front installer: `https://downloads.iiko.online/9.4.9102.0/iiko/RMS/Front/Setup.Front.exe`
- BackOffice installer: `https://downloads.iiko.online/9.4.9102.0/iiko/RMS/BackOffice/Setup.RMS.BackOffice.exe`
- Stand archives if idle >1 month: run "приказ об изменении цен" monthly.

The demo RMS credentials do not replace `apiLogin`. They are useful for Resto API
fallback and for filling demo data.

## Demo Cloud API Access (configured 2026-05-27)

iikoWeb path:
`Настройки Cloud API` -> `Добавить интеграцию` -> `Подключенные точки`.

- API login name: `revisitr-dev`
- API key: created in iikoWeb, stored outside git; visible prefix/suffix:
  `a0e******492`
- A separately created key on `clear.yourmind@yandex.ru` was checked on
  2026-05-28 and returned `is not authorized`; do not use it unless it is
  explicitly connected in iikoWeb Cloud API settings.
- Email used for demo key notifications: `dev@revisitr.local`
- Connected point: `Мой ресторан`
- Organization id: `22fc8cb3-0e70-4c9a-b195-8d2301ee0c43`
- Cloud check: `POST /access_token` returns token.
- Organization check: `POST /organizations` returns `Мой ресторан`,
  `isCloud=true`.
- Menu check: `POST /nomenclature` works but demo menu is empty
  (`groups=0`, `products=0`, `revision=0`).
- External menu check: `POST /api/2/menu` returns `Revisitr Demo Menu`
  (`externalMenuId=82279`).
- External menu payload: `POST /api/2/menu/by_id` with
  `externalMenuId=82279` returns category `Coffee Demo` and items:
  `Revisitr Beans`, `Revisitr Drip`, `Revisitr Cookie`.
- External menu item ids may arrive as `itemId` while `iikoItemId` is null.
  The provider must fall back from `iikoItemId` to `itemId` to keep menu sync
  useful on this stand.
- External menu prices are currently `null`; base price list still needs setup.
- Resto nomenclature fallback was filled with 4 product groups and 15 products:
  - `Revisitr Coffee`: Espresso, Cappuccino, Raf, Cold Brew;
  - `Revisitr Bakery`: Croissant, Cookie, Brownie, Cheesecake;
  - `Revisitr Bowls`: Granola, Salad, Chicken Bowl, Soup;
  - `Revisitr Drinks`: Lemonade, Matcha, Tea.
- Product/category images are out of scope for the integration demo. Demo data
  only needs enough structure for API sync: categories, products, prices,
  stop-list/order checks.
- Delivery-order seed script:
  `docs/integrations/iiko/seed_demo_deliveries.sh`. It uses
  `IIKO_API_LOGIN` from env, creates two Cloud delivery fixtures: one with an
  inline regular customer, one without a POS customer object. Delivery phone is
  still required by iikoCloud.
- Current stand limitation: `POST /terminal_groups` returns the only group only
  with `includeDisabled=true`, and `POST /deliveries/create` fails with
  `TerminalGroupDisabled`. To seed real Cloud orders, iikoFront/terminal group
  must be registered and active, or we need to use another demo stand.
- Loyalty customer write check: `POST /loyalty/iiko/customer/create_or_update`
  returns `Transport_WrongCrmId / Common_OrganizationNotFound`. This means
  iikoBiz/iikoCard loyalty is not enabled for this demo org, so POS-side iiko
  customers cannot be seeded through Cloud loyalty endpoints yet.
- Terminal group check: `POST /terminal_groups` returns 1 terminal group only
  when disabled groups are included.

Do not commit the raw API key. Put it only into integration config/secrets.

## Resto API Fallback (verified live, parked)

Auth flow tested 2026-05-25:

```http
GET /resto/api/auth?login=user&pass=<sha1(password)>
```

Response:
- `200`;
- body = token UUID;
- `Set-Cookie: key=<token>`.

Working endpoints with `key=<token>`:
- `/corporation/departments` 200;
- `/products` 200;
- `/employees` 200;
- `/v2/entities/products/list` 200;
- `/v2/reports/olap` POST only.

Use fallback if:
- `apiLogin` issuance blocks delivery;
- a client has RMS/Resto API but no Cloud API;
- demo verification needs current stand immediately.

Cost: separate `RestoProvider`, different auth, different response shapes,
different mappers.

## Existing Codebase

Already present:
- `entity.Integration` + JSONB config;
- integration repository CRUD;
- order/aggregate/client-map repositories;
- integration usecase: create/list/sync/test/stats;
- `POSProvider` interface + factory;
- mock provider;
- `SyncService` for orders, menu, aggregates, client matching by phone;
- loyalty reserve/confirm/cancel primitives;
- HTTP endpoints under `/integrations`;
- frontend integration routes/components;
- migrations for existing integration/loyalty tables.

Current iiko status:
- HTTP foundation exists in `backend/internal/service/pos/iiko.go`;
- token cache and request wrapper are covered by `iiko_test.go`;
- `TestConnection` and `GetMenu` have real Cloud API implementations;
- customers/orders/aggregates still need real provider implementations.

## Phase Plan

### Phase 0 — Access and demo data

- [x] Create `apiLogin` in iikoWeb. Name: `revisitr-dev`.
- [x] Connect demo point `Мой ресторан` to the API login.
- [x] Verify `POST /access_token`.
- [x] Verify `POST /organizations`.
- [x] Create Resto product group `Revisitr Demo`.
- [x] Create demo products `Revisitr Beans`, `Revisitr Drip`,
  `Revisitr Cookie`.
- [x] Create richer Resto demo groups/products: 4 categories, 15 products.
- [x] Create external menu `Revisitr Demo Menu`.
- [x] Connect external menu `Revisitr Demo Menu` to API login `revisitr-dev`.
- [x] Verify `POST /api/2/menu`.
- [x] Verify `POST /api/2/menu/by_id`.
- [ ] Install BackOffice on Windows/VM.
- [ ] Configure base price list so external menu item prices are non-null.
- [ ] Fill demo data: employees, real dishes with cooking place.
- [ ] Enable iikoBiz/iikoCard or use Resto/customer fallback, then fill demo
  customers with phones.
- [ ] Punch test checks:
  - [ ] with POS customer attached;
  - [ ] without POS customer object;
  - [ ] with delivery data;
  - [ ] with dine-in/front data if available;
  - [ ] with discounts.

### Phase 1 — Transport API read foundation

- [x] Provider constructor validates `apiLogin` and `orgID`.
- [x] Token fetch: `POST /access_token {apiLogin}`.
- [x] Token cache with refresh before TTL.
- [x] Request wrapper with bearer auth.
- [x] Retry once on 401 after token invalidation.
- [x] Unit tests for token, refresh, decode, error handling.

### Phase 2 — Organizations and connection test

- [x] `POST /organizations {organizationIds:null, returnAdditionalInfo:true}`.
- [ ] Map iiko organizations to internal POS organizations/locations.
- [x] `TestConnection`: verify configured `cfg.OrgID` exists.
- [ ] Save useful org metadata for diagnostics.

### Phase 3 — Menu sync

- [x] Add optional `config.external_menu_id` for iiko external menus.
- [x] Implement iiko `GetMenu` external menu read path.
- [x] Implement iiko `GetMenu` nomenclature fallback path.
- [x] `POST /nomenclature {organizationId}` for base RMS menu.
- [ ] `POST /api/2/menu` for external menus and price categories.
- [x] `POST /api/2/menu/by_id {externalMenuId, organizationIds}` for
  web/external menu.
- [x] Map groups/categories.
- [x] Map products/dishes.
- [x] Preserve external ids.
- [x] Preserve item name snapshots.
- [ ] Include modifiers where available.
- [x] Skip `isDeleted=true`.
- [ ] Decide how to handle non-dish product types.

### Phase 4 — Stop lists

- [ ] Find correct iiko endpoint for terminal/group stop lists.
- [ ] Map unavailable products/modifiers to internal stop-list model.
- [ ] Add provider capability flag if not universally available.
- [ ] Sync stop-list deltas often enough for bot/menu UX.

### Phase 5 — Customers import

- [ ] `POST /loyalty/iiko/customer/list {organizationIds, page, pageSize}`.
- [ ] Map `ExternalID`, phone, name, email, birthday, card number.
- [ ] Normalize phones to one canonical format.
- [ ] Do not require POS customer for order matching.

### Phase 6 — Orders and order items

- [x] Fetch closed delivery orders for period.
- [x] Start with `/deliveries/by_delivery_date_and_status` for delivery orders.
- [x] Add demo seed script for two delivery fixtures.
- [ ] Research dine-in/front check source; do not assume deliveries cover all
      restaurant revenue.
- [ ] Map:
  - [x] external order/check id;
  - [x] organization/location;
  - [ ] terminal/cash register if present;
  - [x] ordered/closed time;
  - [x] total;
  - [x] discounts;
  - [ ] payment types;
- [x] customer phone if present;
- [ ] customer id if present;
- [x] item rows with name, qty, price;
- [x] item product id;
- [ ] item modifiers.
- [ ] Make sync idempotent by external id + organization.

### Phase 7 — OLAP aggregates

- [ ] Build OLAP request helper.
- [ ] Fetch daily revenue, transaction count, avg check.
- [ ] Fetch phone/customer breakdown if available.
- [ ] Use OLAP as aggregate verification against imported checks.
- [ ] Fall back to summing imported orders when OLAP blocks.

### Phase 8 — Revisitr order matching

- [ ] Design explicit `order_customer_links` model.
- [ ] Store match source and confidence.
- [ ] Link by POS customer/phone when present.
- [ ] Link by redemption session/code when present.
- [ ] Add ambiguous-match/manual-review state.
- [ ] Ensure bonus earn runs only after a confirmed link.
- [ ] Keep raw unmatched POS orders for later linking.

### Phase 9 — Manual redemption fallback

- [ ] Cashier UI: enter/scan one-time code.
- [ ] Create reserve using existing loyalty reserve primitives.
- [ ] Bind reserve to org/location/cashier/time window.
- [ ] Show amount cashier must apply manually in POS.
- [ ] Confirm reserve after matching imported check.
- [ ] Cancel/expire reserve when no check matches.
- [ ] Add audit trail for manual correction.

### Phase 10 — iikoFront plugin discovery and prototype

- [ ] Confirm plugin license/deployment requirements.
- [ ] Confirm how to search guest by phone/QR from plugin.
- [ ] Confirm how to apply bonus spend: discount vs payment type.
- [ ] Confirm where plugin can store Revisitr session id/check marker.
- [ ] Build minimal plugin proof:
  - [ ] call Revisitr API;
  - [ ] reserve points;
  - [ ] apply discount/payment marker;
  - [ ] send final check/order id;
  - [ ] cancel reserve on check cancel.

### Phase 11 — End-to-end QA

- [ ] Connect integration in admin UI.
- [ ] Run sync against demo iiko data.
- [ ] Verify org/menu/stop-list/order data in DB.
- [ ] Verify customer-linked check.
- [ ] Verify unlinked check remains importable and matchable later.
- [ ] Verify item-level purchase history.
- [ ] Verify earn after confirmed link.
- [ ] Verify spend reserve confirm/cancel.
- [ ] Verify duplicate sync does not duplicate orders or transactions.

### Phase 12 — Hardening

- [ ] Rate limit guard.
- [ ] Retry/backoff on transient iiko errors.
- [ ] Typed provider errors for auth, permission, config, upstream.
- [ ] Sync logs with last success/error.
- [ ] Secret handling for `config.api_key`.
- [ ] Monitoring for unmatched order rate.
- [ ] Manual review tooling for ambiguous matches.

## Time Map

| Area | Estimate |
| --- | ---: |
| Transport read sync: orgs/menu/customers/orders/OLAP | 18-24h |
| Stop lists | 3-6h |
| Order matching model + earn flow | 8-14h |
| Manual code redemption fallback | 10-18h |
| iikoFront plugin discovery/prototype | 24-40h |
| E2E QA and hardening | 10-18h |

Read-only sync can ship before plugin work. Full POS-side redemption requires
plugin or manual cashier flow.

## Risk Register

1. `apiLogin` issuance delay: use Resto API fallback for demo/read sync.
2. iiko may not expose all dine-in checks through delivery endpoints: research
   front/check source before claiming full order coverage.
3. POS customer often missing: Revisitr identity/matching flow is mandatory.
4. Stop-list endpoint and terminal group behavior may vary by setup; inactive
   terminal groups block Cloud delivery order creation even when menu/org reads
   work.
5. Plugin deployment may require iiko licensing, admin access, Windows install,
   and client IT support.
6. Bonus spend fiscal treatment may differ by country/client: discount vs payment
   type must be confirmed before production rollout.
7. Manual code flow can mismatch checks: needs confidence model and review queue.
8. Duplicate sync/duplicate close events can double-spend or double-earn unless
   all confirm paths are idempotent.
9. Phone formats differ across POS, Telegram, and manual entry: normalize all
   sides.
10. Imported menu names/prices change over time: store item snapshots on orders.
