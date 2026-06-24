# iiko POS — Playbook (non-obvious mechanics)

Hard-won mechanics from the 2026-06 integration tests. Read this before any iiko
POS work. Companions: `PLAN.md` (architecture/phases), the seed scripts in this
dir, and live findings in `.ai-memory/iiko_integration.md`.

Stand credentials live in `.env.local` (`IIKO_API_LOGIN`, `IIKO_ORG_ID`,
`IIKO_EXTERNAL_MENU_ID`) — never commit the raw key. RMS stand:
`260-347-461.iiko.it`, login `user` / `user#test` / PIN `1111`.

## Two distinct APIs — never confuse
- **Cloud / Transport** `https://api-ru.iiko.services` — JSON,
  `POST /api/1/access_token {apiLogin}` → Bearer (TTL 1h). This is what
  `backend/internal/service/pos/iiko.go` uses.
- **RMS Resto** `https://<stand>.iiko.it/resto` — `GET /api/auth?login=&pass=sha1(pass)`
  → token. Used as a data source / back-office reads (products, prices, payment
  types, cooking places).

## Activating a terminal group (the first big blocker)
A group shows only with `terminal_groups{includeDisabled:true}` and
`deliveries/create` fails `TerminalGroupDisabled` until activated. Activation is
**NOT** a Front-UI action and there is **no Cloud method**. It is an iikoWeb
back-office flag `isEnabledForTransport`. Flip it via the hidden iikoWeb endpoint
`POST /api/settings/transport/saveRmsSettings` (`isEnabledForTransport:true`),
then restart `iikoFront.Net.exe` → handshake + `ReportTerminalGroupAlive`.
(Port 5671/AMQPS may be firewalled; transport falls back to HTTPS/443.)

## deliveries/create → close validation chain (four separate walls)
1. **No `customer` object → `AnonymousCustomerDisabled`.** Every order needs
   `customer:{name,surname,type:"regular"}`. This is just name/phone, NOT an
   iikoCard POS customer.
2. **Product without cooking place → `SettingsIssue` "doesn't have cooking place
   type".** Assign one in BackOffice. Two types exist on the stand: `Бар`
   (`22c1becd-…`), `Кухня` (`d17fa3fa-…`). Resto product field is `placeType`.
3. **Close without payment → `PaymentSumNotEnough`.** Add
   `payments:[{paymentTypeKind:"Cash", sum:total, paymentTypeId:…}]` at create.
4. **Close without an open cash shift → `CafeSessionNotFound`.** iikoFront must
   have an open кассовая смена and be online when `deliveries/close` runs.

## Order type decides whether the order can close (KEY trick)
- `orderServiceType:"DeliveryByCourier"` → after confirm+pay+close stays
  `Waiting`; Cloud `close` returns command-Success but never reaches `Closed`
  (courier needs the full lifecycle via `change_delivery_status`).
- **`orderServiceType:"DeliveryByClient"` (pickup) → closes immediately to
  `status:Closed` with `whenClosed`.** No `deliveryPoint`/address needed.
  **Use pickup to produce closeable orders.**

## Endpoints blocked for our apiLogin (don't depend on them)
- `deliveries/change_delivery_status` → "not allowed for this ApiLogin" (can't
  shortcut a courier order to Delivered).
- `/reports/olap` → 401 (OLAP not in the Cloud surface for this tariff, even with
  "Все права"). Derive aggregates from `/deliveries`.
- `loyalty/iiko/customers|customer/*` → 500/400 `WrongCrmId` (iikoCard not
  enabled). No POS-customer import → Revisitr matches by **phone** (it does not
  require a POS customer).

## Data quirks
- External menu (82279) prices are **null** and `iikoItemId` null (fall back to
  `id`). **Real prices live in Resto** (`defaultSalePrice`); pull products from
  `GET /resto/api/v2/entities/products/list` (also gives `placeType`).
- Cloud `/payment_types` is **empty**, but Resto
  `/api/v2/entities/list?rootType=PaymentType` returns `Наличные` (`09322f46-…`)
  / `Банковские карты` — those GUIDs work in Cloud `deliveries/create`.
- **Cannot backdate orders** via Cloud (delivery date ≈ now). For historical /
  time-series demo data, seed Revisitr's DB directly (see `seed_history.py`).
- **Cancelling stuck orders via Cloud is unreliable** (paid orders need a refund;
  bulk cancel is async/flaky). Clear them on the iikoFront terminal. Stuck
  non-Closed orders are invisible to Revisitr anyway.

## Revisitr-side mechanics
- `GetOrders` fetches only `statuses:["Closed","Delivered"]`. `ordered_at` falls
  back whenClosed → whenCreated → now.
- **Timezone:** iiko filters `deliveryDateFrom/To` in the **org's local timezone**
  (MSK) while Revisitr sent UTC → orders from the last few hours were dropped.
  Fixed: `iikoOrderWindowPad` pads the window ±24h before chunking; `UpsertOrder`
  is idempotent on `external_id`. The window is also chunked ≤24h to avoid 422
  `TOO_MANY_DATA_REQUESTED`.
- Scheduler `pos_sync` runs `SyncAll` every 4h (+ once ~10s after startup).
  `GetActive` returns only `status='active'` → an integration stuck in
  `status='error'` is skipped forever. Recover: `UPDATE integrations SET
  status='active'` then restart backend, or `POST /api/v1/integrations/:id/sync`
  (a successful sync flips status back to active).
- Dashboard stats: integration overview = `external_orders` (`GetSyncStats`);
  analytics charts = `integration_aggregates` (daily; `GetAggregates` /
  `GetDashboardAggregates`). `SyncAggregates` recomputes only days in
  `[last_sync, now]` from `/deliveries`, so historically-seeded aggregate days
  are safe.

## Tooling (this dir)
- `seed_demo_deliveries.sh` — original 2-order Cloud delivery fixtures.
- `seed_demo_orders.sh` — richer Cloud seed: real Resto prices, repeat customers,
  inline cash payment, `DeliveryByClient` so orders actually close. Needs an
  active terminal + open cash shift.
- `seed_history.py` — generates ~3 months of realistic café history straight into
  the DB (orders + clients + daily aggregates + RFM). Reproducible, self-cleaning;
  demo rows tagged `external_id LIKE 'demo3m-%'` / `data->>'seed'='demo3m'`.
  Pipe its stdout into `psql` on the target server.

## Prod servers (2026-06 migration)
- OLD **elysium** `elysium.fm` → 80.93.187.209 (path-prefixed `/revisitr/…`).
- NEW **dsru14** `revisitr.ru` → 31.172.64.81 (root paths, HTTPS). The
  **self-hosted GitHub Actions runner is on the NEW server**, so `push to main`
  deploys there. Each server has its own DB. Migration cut over to the new server
  (main commit d150970).
