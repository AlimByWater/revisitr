# POS Integrations & Dev Mode — Design Document

## Context

Revisitr has integration scaffolding (entities, repos, controllers, migration) but the sync service is a stub. Frontend has no `/dashboard/integrations` page. No dev mode exists.

**Goal**: Real iiko/r-keeper API clients + admin UI for integrations + dev mode toggle for testing with mock data.

---

## 1. iiko & r-keeper — How to Get Test Access

### iiko (Cloud API)

| Item | Details |
|------|---------|
| **Type** | Cloud-only SaaS API |
| **Base URL** | `https://api-ru.iiko.services/api/1/` |
| **Docs** | https://api-ru.iiko.services/swagger |
| **Auth** | POST `/access_token` with `{ "apiLogin": "<api_key>" }` → Bearer token (1h TTL) |
| **Test account** | Demo credentials exist (server `iiko.biz:9900`, login `demoDelivery`). For your own org: iikoCloud → "Cloud API Settings" → create key. Email `api@iiko.ru` for dev access. |
| **Sandbox** | None separate — demo org on prod API serves as sandbox |
| **Go SDK** | [iiko-go/iiko-go](https://github.com/iiko-go/iiko-go) (community, auto-refreshes token) |
| **Local/Docker** | Not possible — cloud only |

**Key endpoints** (all POST with JSON body):
- `/access_token` — auth
- `/organizations` — list orgs
- `/nomenclature` — menu
- `/loyalty/iiko/customer/info` — customer by phone/card
- `/loyalty/iiko/customer/create_or_update` — upsert customer
- `/deliveries/by_delivery_date_and_status` — orders by date
- `/order/create` — create order
- `/discounts` — discount programs

### r-keeper 7

| Item | Details |
|------|---------|
| **Type** | On-premise (XML over HTTPS) |
| **Endpoint** | `https://<ip>:<port>/rk7api/v0/xmlinterface.xml` |
| **Docs** | https://docs.rkeeper.ru/rk7/latest/ru/xml-interfejs-r_keeper-7-19605640.html |
| **Auth** | HTTP Basic Auth (employee name + password from manager station) |
| **Test access** | Request demo license from UCS dealer, or email `integrations@rkeeper.ru` |
| **Sandbox** | No cloud sandbox. Need Windows machine/VM with r-keeper installed. |
| **Go SDK** | None — build own XML client |
| **Local/Docker** | Need Windows + r-keeper license. No Docker image. |

**Key XML commands**:
- `GetRefData` (MENUITEMS) — menu
- `GetOrder` / `GetOrderList` / `SaveOrder` — orders
- `PayOrder` — payments
- CRM module (separate API) — customers/loyalty

### Practical Reality

- **iiko**: Easy to start. Get API key → hit cloud API. Best for first integration.
- **r-keeper**: Harder. Need Windows + license or dealer relationship. XML protocol. Build custom Go client.
- **Neither has Docker or emulators** — hence the need for dev mode with mock providers.

---

## 2. Architecture — POS Provider Interface

### Current State
```
SyncService (stub) → integrations.UpdateLastSync() → done
```

### Target State
```
SyncService
  ├── POSProvider interface
  │     ├── IikoProvider (real API client)
  │     ├── RKeeperProvider (real XML client)
  │     └── MockProvider (dev mode — generates fake data)
  ├── integrations repo (existing)
  └── clients repo (existing, for phone matching)
```

### POSProvider Interface

```go
// backend/internal/service/pos/provider.go

type POSProvider interface {
    // Auth/connection
    TestConnection(ctx context.Context) error

    // Customers
    GetCustomer(ctx context.Context, phone string) (*POSCustomer, error)
    ListCustomers(ctx context.Context, opts CustomerListOpts) ([]POSCustomer, error)
    UpsertCustomer(ctx context.Context, c *POSCustomer) error

    // Orders/Checks
    GetOrders(ctx context.Context, from, to time.Time) ([]POSOrder, error)
    GetOrderByID(ctx context.Context, externalID string) (*POSOrder, error)

    // Menu
    GetMenu(ctx context.Context) (*POSMenu, error)
}

type POSCustomer struct {
    ExternalID string
    Phone      string
    Name       string
    Email      string
    Birthday   *time.Time
    Balance    float64  // loyalty points/balance
    CardNumber string
}

type POSOrder struct {
    ExternalID string
    CustomerPhone string
    Items      []POSOrderItem
    Total      float64
    Discount   float64
    OrderedAt  time.Time
    Status     string // "open", "closed", "cancelled"
    TableNum   string
    WaiterName string
}

type POSOrderItem struct {
    Name     string
    Quantity int
    Price    float64
    Category string
}

type POSMenu struct {
    Categories []MenuCategory
}

type MenuCategory struct {
    Name  string
    Items []MenuItem
}

type MenuItem struct {
    ExternalID  string
    Name        string
    Price       float64
    Description string
}

type CustomerListOpts struct {
    Limit  int
    Offset int
    Search string // phone or name substring
}
```

### Provider Factory

```go
// backend/internal/service/pos/factory.go

func NewProvider(integration *entity.Integration) (POSProvider, error) {
    switch integration.Type {
    case "iiko":
        return NewIikoProvider(integration.Config)
    case "rkeeper":
        return NewRKeeperProvider(integration.Config)
    case "mock":
        return NewMockProvider(integration.Config)
    default:
        return nil, fmt.Errorf("unknown integration type: %s", integration.Type)
    }
}
```

---

## 3. iiko Provider

```go
// backend/internal/service/pos/iiko.go

type IikoProvider struct {
    baseURL   string // https://api-ru.iiko.services/api/1
    apiLogin  string
    orgID     string
    token     string
    tokenExp  time.Time
    client    *http.Client
    mu        sync.Mutex
}

func NewIikoProvider(cfg entity.IntegrationConfig) (*IikoProvider, error) {
    // cfg.APIKey = iiko API Login
    // cfg.APIURL = base URL (default: https://api-ru.iiko.services/api/1)
    // cfg.OrgID stored in extended config (see section 5 below)
}
```

Token auto-refresh: if `time.Now().After(tokenExp - 5min)` → re-auth.

Maps iiko's POST-based JSON API to `POSProvider` interface.

---

## 4. r-keeper Provider

```go
// backend/internal/service/pos/rkeeper.go

type RKeeperProvider struct {
    baseURL  string // https://ip:port/rk7api/v0/xmlinterface.xml
    username string
    password string
    client   *http.Client
}
```

Builds XML requests, sends via HTTPS + Basic Auth, parses XML responses. Each method constructs the appropriate XML command (`GetRefData`, `GetOrderList`, etc.).

---

## 5. Extended IntegrationConfig

Current `IntegrationConfig` only has `api_url` and `api_key`. Need to extend for provider-specific fields:

```go
type IntegrationConfig struct {
    APIURL   string `json:"api_url,omitempty"`
    APIKey   string `json:"api_key,omitempty"`

    // iiko-specific
    OrgID    string `json:"org_id,omitempty"`    // iiko organization ID

    // r-keeper-specific
    Username string `json:"username,omitempty"`  // HTTP Basic Auth
    Password string `json:"password,omitempty"`  // HTTP Basic Auth

    // Sync settings
    SyncInterval int `json:"sync_interval,omitempty"` // minutes, 0 = manual only
}
```

Since config is JSONB, this is backward-compatible. No migration needed.

---

## 6. Mock Provider (Dev Mode)

```go
// backend/internal/service/pos/mock.go

type MockProvider struct {
    customers []POSCustomer
    orders    []POSOrder
    menu      *POSMenu
    mu        sync.RWMutex
}

func NewMockProvider(_ entity.IntegrationConfig) (*MockProvider, error) {
    p := &MockProvider{}
    p.seedData() // Generate ~20 customers, ~50 orders, menu with 3 categories
    return p, nil
}
```

Mock provider holds data in-memory and allows full CRUD through the interface. Generates realistic Russian HoReCa data (names, phone numbers, menu items like "Маргарита", "Американо", etc.).

### Dev Mode Activation

**No global toggle needed.** Instead:
- Add `"mock"` as valid integration type (alongside `iiko`, `rkeeper`, `1c`)
- When user creates integration with `type: "mock"`, it uses `MockProvider`
- Mock integration behaves identically to real ones in the UI
- Available in all environments (dev, staging, prod) — useful for demos too

**Database change**: Add `'mock'` to the CHECK constraint:
```sql
ALTER TABLE integrations
DROP CONSTRAINT integrations_type_check,
ADD CONSTRAINT integrations_type_check
CHECK (type IN ('iiko', 'rkeeper', '1c', 'mock'));
```

---

## 7. Updated SyncService

```go
// backend/internal/service/pos/sync.go

type SyncService struct {
    integrations integrationsRepo
    clients      clientsRepo
    logger       *slog.Logger
}

func (s *SyncService) Sync(ctx context.Context, integration *entity.Integration) error {
    provider, err := NewProvider(integration)
    if err != nil {
        return fmt.Errorf("create provider: %w", err)
    }

    // 1. Test connection
    if err := provider.TestConnection(ctx); err != nil {
        s.integrations.UpdateLastSync(ctx, integration.ID, "error")
        return fmt.Errorf("test connection: %w", err)
    }

    // 2. Sync orders (last 24h or since last_sync_at)
    since := time.Now().Add(-24 * time.Hour)
    if integration.LastSyncAt != nil {
        since = *integration.LastSyncAt
    }

    orders, err := provider.GetOrders(ctx, since, time.Now())
    if err != nil {
        s.integrations.UpdateLastSync(ctx, integration.ID, "error")
        return fmt.Errorf("get orders: %w", err)
    }

    // 3. Upsert orders, match clients by phone
    for _, order := range orders {
        extOrder := &entity.ExternalOrder{
            IntegrationID: integration.ID,
            ExternalID:    order.ExternalID,
            Items:         toEntityItems(order.Items),
            Total:         order.Total,
            OrderedAt:     &order.OrderedAt,
        }

        // Match client by phone
        if order.CustomerPhone != "" {
            if client, err := s.clients.GetByPhone(ctx, integration.OrgID, order.CustomerPhone); err == nil {
                extOrder.ClientID = &client.ID
            }
        }

        if err := s.integrations.UpsertOrder(ctx, extOrder); err != nil {
            s.logger.Error("upsert order", "error", err, "external_id", order.ExternalID)
        }
    }

    // 4. Mark success
    s.integrations.UpdateLastSync(ctx, integration.ID, "active")
    return nil
}
```

---

## 8. New API Endpoints

Add to existing `/api/v1/integrations`:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/:id/orders` | List synced orders for integration |
| GET | `/:id/orders/:orderId` | Get single order detail |
| GET | `/:id/customers` | List POS customers (live from provider) |
| POST | `/:id/customers` | Create/update POS customer (push to provider) |
| GET | `/:id/menu` | Get POS menu (live from provider) |
| POST | `/:id/test` | Test connection to POS |
| GET | `/:id/stats` | Sync stats (total orders, last sync, matched clients %) |

New endpoint group for external orders:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/orders` | List all external orders across integrations |
| GET | `/api/v1/orders/:id` | Get order detail |
| PATCH | `/api/v1/orders/:id` | Edit order (dev mode: mock provider only) |
| DELETE | `/api/v1/orders/:id` | Delete order (dev mode: mock provider only) |

---

## 9. Frontend — Integrations Page

### Route: `/dashboard/integrations`

**List view**:
- Cards for each integration (type icon, name/url, status badge, last sync time)
- "Add Integration" button → modal with type selector (iiko / r-keeper / Mock)
- Each card: "Sync Now", "Settings", "Delete" actions

**Detail view** (`/dashboard/integrations/:id`):

Tabs:
1. **Overview** — connection status, config, sync stats, "Test Connection" button
2. **Orders** — paginated table of synced orders (date, customer, items, total, status)
   - Click order → side panel with full order detail
   - For mock type: inline edit + add/delete orders
3. **Customers** — customers from POS (phone, name, balance, matched/unmatched status)
   - For mock type: inline edit + add/delete customers
4. **Menu** — POS menu tree (categories → items with prices)
5. **Settings** — edit API credentials, sync interval, test connection
6. **Logs** — sync history with timestamps and error details

### Integration Setup Flow (per type)

**iiko**:
1. Enter API Login (key)
2. Click "Test Connection" → fetches organizations
3. Select organization from dropdown
4. Save → status goes to "active"

**r-keeper**:
1. Enter server URL (`https://ip:port`)
2. Enter username + password
3. Click "Test Connection"
4. Save

**Mock**:
1. Just click "Create" — no config needed
2. Immediately active with seed data
3. UI shows "Dev Mode" badge on the integration card

### Dev Mode UI Elements

When integration type is `mock`:
- Orange "DEV" badge on integration card and detail header
- Orders/Customers tables show "Edit" and "Delete" buttons (hidden for real integrations)
- "Generate Data" button — creates N random orders/customers
- "Reset Data" button — wipes mock data back to seed state
- Inline editing on all data fields

### Mock Data Controls (additional API for mock type)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/:id/mock/generate` | Generate N random orders/customers |
| POST | `/:id/mock/reset` | Reset to seed data |
| PATCH | `/:id/mock/customers/:cid` | Edit mock customer |
| DELETE | `/:id/mock/customers/:cid` | Delete mock customer |
| POST | `/:id/mock/customers` | Add mock customer |
| PATCH | `/:id/mock/orders/:oid` | Edit mock order |
| DELETE | `/:id/mock/orders/:oid` | Delete mock order |
| POST | `/:id/mock/orders` | Add mock order |

These endpoints only work when integration type is `mock` — return 403 for real integrations.

---

## 10. File Structure

```
backend/
  internal/
    service/pos/
      provider.go       — POSProvider interface + types
      factory.go        — NewProvider factory
      iiko.go           — IikoProvider implementation
      rkeeper.go        — RKeeperProvider implementation
      mock.go           — MockProvider with seed data
      mock_seed.go      — Russian HoReCa seed data generator
      sync.go           — Updated SyncService (replace stub)
    controller/http/group/integrations/
      integrations.go   — existing (extend with new handlers)
      orders.go         — order listing/detail handlers
      customers.go      — POS customer handlers
      mock.go           — mock data control handlers
    controller/http/group/orders/
      orders.go         — cross-integration order listing
    entity/
      integration.go    — extend IntegrationConfig

frontend/
  src/
    routes/dashboard/integrations/
      index.tsx         — integrations list
      $integrationId.tsx — integration detail (tabs)
    features/integrations/
      api.ts            — API calls
      queries.ts        — TanStack Query hooks
      types.ts          — TypeScript interfaces
    components/integrations/
      IntegrationCard.tsx
      CreateIntegrationModal.tsx
      ConnectionTestButton.tsx
      OrdersTable.tsx
      CustomersTable.tsx
      MenuTree.tsx
      MockControls.tsx  — generate/reset/edit buttons
      SyncLogTable.tsx
```

---

## 11. Implementation Order

### Phase A: Mock Provider + Admin UI (1-2 weeks)
1. `POSProvider` interface + `MockProvider` with seed data
2. Migration: add `mock` to type check
3. Update `SyncService` to use provider factory
4. New API endpoints (orders, customers, menu, mock controls)
5. Frontend: integrations list + detail page with all tabs
6. Tests: unit for mock provider, integration for new endpoints

### Phase B: iiko Provider (1 week)
1. `IikoProvider` with token management
2. Map iiko API responses to `POSProvider` types
3. Integration setup flow (org selection)
4. E2E test with iiko demo account

### Phase C: r-keeper Provider (1-2 weeks)
1. XML client builder/parser
2. `RKeeperProvider` mapping XML to `POSProvider` types
3. Integration setup flow
4. Test with r-keeper demo (requires Windows VM or dealer access)

### Phase D: Advanced Features
1. Scheduled sync (cron-based via existing scheduler)
2. Webhook support (iiko supports webhooks for order events)
3. Sync conflict resolution
4. Customer auto-matching improvements (fuzzy phone matching)
5. Analytics: revenue by POS, order trends

---

## 12. Setup for Testing

### iiko
1. Email `api@iiko.ru` requesting developer API login
2. Or: use demo credentials `demoDelivery` on `iiko.biz:9900` (verify if still active)
3. Get API login → store in integration config
4. No local setup needed — hits cloud API directly

### r-keeper
**Option A** (recommended for now): Skip until Phase C, use mock provider
**Option B**: Get demo from UCS dealer
- Need Windows machine (VM on your Mac via UTM/Parallels)
- Install r-keeper 7 demo
- Configure XML interface on a port
- Point integration config to `https://localhost:<port>/rk7api/v0/xmlinterface.xml`

**Option C**: Ask a partner restaurant that uses r-keeper for test credentials to their server (read-only access)

### Mock (immediate, no setup)
Just create integration with type `mock` → instant data to work with.

---

## 13. Security Notes

- API keys stored in JSONB `config` field — already in DB, not in code
- Consider encrypting `config.api_key` and `config.password` at rest (AES-GCM with app-level key)
- r-keeper Basic Auth credentials should be treated as secrets
- Mock endpoints restricted to `type=mock` integrations only
- Rate limiting on sync endpoints (prevent abuse of external POS APIs)
