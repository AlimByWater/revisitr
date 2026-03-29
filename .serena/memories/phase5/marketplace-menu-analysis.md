# Phase 5: Move Marketplace + Menu to Bot Settings — Analysis

## Requirement Summary (from Правки 29.03.pdf)

### Section 3: Marketplace
- **Move from**: `/dashboard/marketplace` (standalone page)
- **Move to**: Bot settings (each bot can have its own marketplace configuration)
- **Scope**: Per-bot configuration

### Section 4: Menu
- **Move from**: `/dashboard/menus` (standalone section)
- **Move to**: 
  1. Bot settings (for menu display in specific bot)
  2. POS locations (for point of sale menu association)
- **Menu Creation Options in Bot Settings**:
  - Create manually
  - Upload file (pdf, png, jpg)
  - Use from Point of Sale
  - Get from POS system

### Section 5.4: Bot Settings Restructuring
- Remove Menu subsection from dedicated tab
- Move Menu into Modules section

### Section 5.5: Module Configuration
- Each module gets "Настроить" (Configure) button
- Clicking opens page for that module
- Menu module configuration includes the 4 creation options above

---

## CURRENT STRUCTURE

### Frontend Routes
```
/dashboard/marketplace          → MarketplacePage (list products + orders)
/dashboard/menus                → MenusPage (list menus)
/dashboard/menus/$menuId        → MenuDetailPage (edit menu categories/items)
/dashboard/bots/$botId          → BotDetailPage (with tabs: connection, general, modules, preview)
/dashboard/pos/$posId           → POSDetailPage (basic info + schedule)
```

### Feature Data Models

#### Menus (`features/menus/`)
```typescript
interface Menu {
  id: number
  org_id: number
  integration_id?: number
  name: string
  source: 'manual' | 'pos_import'
  last_synced_at?: string
  created_at: string
  updated_at: string
  categories?: MenuCategory[]
}

interface MenuCategory {
  id: number
  menu_id: number
  name: string
  sort_order: number
  created_at: string
  items?: MenuItem[]
}

interface MenuItem {
  id: number
  category_id: number
  name: string
  description?: string
  price: number
  image_url?: string
  tags: string[]
  external_id?: string
  is_available: boolean
  sort_order: number
  created_at: string
  updated_at: string
}
```

**API Endpoints**:
- `GET /menus` — list all menus
- `GET /menus/:id` — get menu with categories and items
- `POST /menus` — create menu
- `PATCH /menus/:id` — update menu
- `DELETE /menus/:id` — delete menu
- `POST /menus/:menuId/categories` — add category
- `POST /menus/:menuId/categories/:categoryId/items` — add item
- `PATCH /menus/items/:itemId` — update item
- `GET /bots/:botId/pos-locations` — get bot's linked POS locations
- `PUT /bots/:botId/pos-locations` — set bot's POS locations

#### Marketplace (`features/marketplace/`)
```typescript
interface MarketplaceProduct {
  id: number
  org_id: number
  name: string
  description: string
  image_url: string
  price_points: number
  stock: number | null
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

interface MarketplaceOrder {
  id: number
  org_id: number
  client_id: number
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled'
  total_points: number
  items: MarketplaceOrderItem[]
  note: string
  created_at: string
  updated_at: string
}

interface MarketplaceStats {
  total_products: number
  active_products: number
  total_orders: number
  total_spent_points: number
}
```

**API Endpoints**:
- `GET /marketplace/products` — list products
- `GET /marketplace/products/:id` — get product
- `POST /marketplace/products` — create product
- `PATCH /marketplace/products/:id` — update product
- `DELETE /marketplace/products/:id` — delete product
- `GET /marketplace/orders` — list orders
- `GET /marketplace/orders/:id` — get order
- `POST /marketplace/orders` — place order
- `PATCH /marketplace/orders/:id/status` — update order status
- `GET /marketplace/stats` — get marketplace stats

#### Bots (`features/bots/`)
```typescript
interface Bot {
  id: number
  org_id: number
  name: string
  username: string
  token_masked?: string
  status: 'active' | 'inactive' | 'error'
  settings: BotSettings
  created_at: string
  updated_at: string
  client_count?: number
  program_id?: number
}

interface BotSettings {
  modules: string[]  // e.g., ['loyalty', 'menu', 'marketplace', 'feedback', 'booking']
  buttons: BotButton[]
  registration_form: FormField[]
  welcome_message: string
}
```

**Current Modules**: loyalty, menu, marketplace, feedback, booking
(seen in MODULE_DEFS in $botId.tsx)

---

## NAVIGATION STRUCTURE

### Sidebar (`Sidebar.tsx` and `AuroraSidebar.tsx`)
Current structure:
```
├─ Дашборд
├─ Аналитика
├─ Клиенты
├─ Лояльность
├─ Рассылки
├─ Акции
├─ Мои боты
│  └─ Список ботов
├─ Маркетплейс                    ← WILL BE REMOVED
├─ Точки продаж
├─ Меню                            ← WILL BE REMOVED
├─ Интеграции
└─ Биллинг
```

After Phase 5, Marketplace and Menu will be removed from top-level navigation and accessible only through Bot Settings.

---

## IMPLEMENTATION ROADMAP

### BACKEND REQUIREMENTS (speculative based on frontend needs)

1. **Menu-Bot Association** (NEW)
   - Add `menu_id` field to `bots.settings` or create `bot_menus` junction table
   - Add `menu_source` configuration (manual / upload / pos_location / pos_system)
   - For uploads: handle file storage (pdf, png, jpg)
   - For POS location: reference existing POS menu

2. **Marketplace-Bot Association** (NEW)
   - Add `marketplace_config` to `bots.settings`
   - Enable/disable per-bot marketplace
   - Marketplace products remain org-level (shared across bots)

3. **POS-Menu Association** (NEW)
   - Add `menu_id` field to POS locations
   - Allow POS to link to menu for display

### FRONTEND IMPLEMENTATION

#### Phase 5.1: Update BotSettings Type
```typescript
// Add to BotSettings interface
menu_config?: {
  menu_id?: number
  source: 'manual' | 'upload' | 'pos_location' | 'pos_system'
  pos_location_id?: number
  uploaded_file_url?: string
}

marketplace_config?: {
  enabled: boolean
}
```

#### Phase 5.2: Remove Top-Level Routes
- Remove `/dashboard/marketplace` route (but keep feature API/queries)
- Remove `/dashboard/menus` route (but keep feature API/queries)
- Both become accessible only through bot settings

#### Phase 5.3: Update Sidebar Navigation
- Remove "Маркетплейс" top-level item from navigation arrays
- Remove "Меню" top-level item from navigation arrays
- Both remain in Sidebar.tsx and AuroraSidebar.tsx navigation arrays (just commented/removed)

#### Phase 5.4: Modules Tab Expansion
Add "Configure" buttons to module cards:
```
[Module Card]
- Icon + Label + Description
- Toggle Enable/Disable
- [Настроить] button  ← NEW
  └─ Opens dedicated config page in modal or new window
```

#### Phase 5.5: Module Configuration Pages

**Menu Module Configuration** (`/dashboard/bots/$botId/menu-config`):
1. Selection section:
   - Radio button: "Create manually"
   - Radio button: "Upload file (pdf, png, jpg)"
   - Radio button: "Use from Point of Sale"
   - Radio button: "Get from POS system"

2. Conditional forms based on selection:
   - **Manual**: Show menu editor similar to current `/dashboard/menus/$menuId`
   - **Upload**: File input for pdf/png/jpg
   - **POS Location**: Dropdown with list of linked POS locations + their menus
   - **POS System**: Configuration for system integration

3. Back button to return to modules tab

**Marketplace Module Configuration** (`/dashboard/bots/$botId/marketplace-config`):
1. Selection section:
   - Radio button: "Create marketplace" (show current marketplace UI)
   - Products tab with create/edit/delete
   - Orders tab with status management

2. Back button to return to modules tab

#### Phase 5.6: POS Location Updates
- Update POSDetailPage to include Menu section
- Allow selecting/creating menu for POS location
- Menu can be associated with multiple POS locations

---

## FILES AFFECTED

### Routes (will be removed or converted)
```
frontend/src/routes/dashboard/marketplace/index.tsx    ← REMOVE/ARCHIVE
frontend/src/routes/dashboard/menus/index.tsx           ← REMOVE/ARCHIVE
frontend/src/routes/dashboard/menus/$menuId.tsx         ← CONVERT to module config
```

### Routes (will be created)
```
frontend/src/routes/dashboard/bots/$botId/menu-config.tsx        ← NEW (or dialog)
frontend/src/routes/dashboard/bots/$botId/marketplace-config.tsx  ← NEW (or dialog)
```

### Components (will be updated)
```
frontend/src/components/layout/Sidebar.tsx                  ← Remove marketplace/menu items
frontend/src/components/layout/AuroraSidebar.tsx            ← Remove marketplace/menu items
frontend/src/routes/dashboard/bots/$botId.tsx               ← Update ModulesTab with config buttons
frontend/src/routes/dashboard/pos/$posId.tsx                ← Add Menu section
```

### Features (unchanged but will be reused)
```
frontend/src/features/menus/
  - types.ts     ← Use as-is
  - api.ts       ← Use as-is, add menu-bot/menu-pos associations
  - queries.ts   ← Use as-is, add new queries for associations

frontend/src/features/marketplace/
  - types.ts     ← Use as-is, add marketplace-bot associations
  - api.ts       ← Use as-is, add new endpoints for bot config
  - queries.ts   ← Use as-is, add new queries for bot config
```

---

## DATA FLOW CHANGES

### Current Flow
```
Organization
├─ Menus (global, org-level)
├─ Marketplace Products (global, org-level)
├─ Bots (reference loyalty program but not menus/marketplace)
└─ POS Locations (standalone)
```

### Target Flow
```
Organization
├─ Menus (global, org-level)
│  └─ Associated with Bot(s) via bot_id in menu config
│  └─ Associated with POS Location(s) via pos_id in menu config
├─ Marketplace Products (global, org-level)
│  └─ Referenced by Bot(s) via marketplace_config in bot settings
├─ Bots (now have menu_config and marketplace_config)
│  ├─ menu_config: source + reference (menu_id or pos_location_id or file)
│  └─ marketplace_config: enabled flag
└─ POS Locations
   └─ menu_id: reference to menu for display
```

---

## UI MOCKUP: BOT MODULES TAB (UPDATED)

```
[Modules Tab Content]

Section: Включенные модули

[Module Card - Loyalty]
├─ 💜 Loyalty Icon
├─ Лояльность | Начисление и списание бонусов
├─ [Toggle: enabled]
└─ (no Configure button, settings in separate Loyalty section)

[Module Card - Menu]
├─ 🍽️ Menu Icon
├─ Меню | Показ меню заведения в боте
├─ [Toggle: enabled]
└─ [Настроить] button
   └─ Opens Modal/Page with 4 menu source options

[Module Card - Marketplace]
├─ 🛍️ Marketplace Icon
├─ Маркетплейс | Каталог товаров для заказа
├─ [Toggle: enabled]
└─ [Настроить] button
   └─ Opens Modal/Page with products list

[Module Card - Feedback]
├─ 💬 Feedback Icon
├─ Обратная связь | Сбор отзывов от клиентов
├─ [Toggle: enabled]
└─ [Настроить] button

[Module Card - Booking]
├─ 📅 Booking Icon
├─ Бронирование | Бронирование столиков
├─ [Toggle: enabled]
└─ [Настроить] button
```

---

## TECHNICAL NOTES

1. **Menu Editor Component**: Extract `MenuDetailPage` content into reusable component that can be embedded in bot settings modal or used in new route

2. **Marketplace Editor Component**: Extract marketplace product list into reusable component

3. **File Upload for Menu**: Need file upload handling (pdf, png, jpg) → likely backend stores file, returns URL

4. **POS Integration**: Menu can be pulled from linked POS locations (existing `menusApi.getBotPOSLocations`)

5. **Shared Components**: Both menu and marketplace editors already exist, just need to be integrated into bot settings workflow

6. **Modal vs New Page**: Could use Modal dialog for configuration or dedicated route. Modal keeps user in bot settings context, new route is simpler UX-wise

---

## SUMMARY OF CHANGES

| Item | Current | Target | Status |
|------|---------|--------|--------|
| Marketplace Route | `/dashboard/marketplace` | Via bot settings module | TODO |
| Menu Route | `/dashboard/menus` + `$menuId` | Via bot settings module + POS settings | TODO |
| Sidebar Items | Top-level items | Removed | TODO |
| Bot Settings | Basic settings only | + menu config + marketplace config | TODO |
| POS Settings | Basic info + schedule | + menu selection | TODO |
| Menu Editor | Standalone page | Embedded in bot/POS settings | TODO |
| Marketplace Editor | Standalone page | Embedded in bot settings | TODO |
