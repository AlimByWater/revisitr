# Architecture & Business Logic Decisions

Key decisions that shape development. These are NOT obvious from reading code alone.

## Entity Hierarchy

**Loyalty Program -> Bot -> POS** (not flat via org_id)

- `program_id` in `bots`: nullable (bot can exist without program temporarily)
- `bot_id` in `pos`: bot can have 0..N POS locations
- Multiple loyalty programs per org: allowed
- Org isolation for `bot_clients`: via JOIN through `bots.org_id` (no direct `org_id` column)

## Loyalty Engine

- Bonus: both **% of check** AND **fixed amount** (user chooses per level)
- Levels: user defines % and threshold freely (no templates)
- Level progression: automatic based on `total_earned` >= `threshold`
- Levels CAN go down (e.g., yearly reset)
- Welcome bonus: applied in bot handler on registration
- Points redemption at POS: via integration (iiko/rkeeper), not in-bot (not MVP)

## Client Identification

- Phone number: required field, always present
- At POS: QR code OR phone number
- QR code: static, one per client, permanent
- Waiter scans QR in our system (not in POS system)
- Phone normalization: +7 vs 8 vs no country code — must normalize

## Campaigns / Mailings

- Queue: Redis with repository interface (future Kafka swap)
- Web UI: simplified (text + media)
- Admin bot: separate Telegram bot (richer media — circles, voice, stickers)
- Tracking: inline buttons + UTM links (switchable/combinable)
- Scheduled sending: supported
- Rate limiting: 30 msg/sec throttle for 10K+ subscribers
- File storage: MinIO on server, Telegram limit 50MB
- Composite messages: multi-part (text, photo, video, sticker, document, animation, audio, voice + inline buttons)

## Promotions & Promo Codes

- Promo code = UTM analog for channel tracking
- Both auto-generated AND user-defined (e.g., "BIRTHDAY2024")
- Analytics: usage count, orders, conversion

## Auto-Actions Engine

- Structure: **trigger -> timing -> condition -> actions[]**
- Example: birthday trigger, 7 days before, condition=order, actions=[bonus + campaign + promo]
- Actions: bonus accrual, send campaign/push, create promo code
- Templates (NULL org_id/bot_id) + fully custom

## Integrations Priority

1. Aggregate data from iiko/rkeeper/1C
2. Link POS operations to client profiles
3. Menu import + per-item analytics
- Client matching: by phone (format normalization)
- iiko: Cloud API only, no sandbox/emulator
- r-keeper: on-premise XML, Windows license needed
- MockProvider available for dev (type="mock")

## RFM Segmentation

- Metrics: R (days since last visit), F (visits in 90d), M (revenue in 180d)
- Score: 1-5 per metric
- 4 standard templates: coffeegng, qsr, tsr, bar
- 70-80% users use pre-built templates; custom segments later

## Onboarding Order

1. Information (about system, FAQ)
2. Loyalty program setup
3. Bot creation + configuration
4. POS locations
5. Integrations
6. Next steps
- Each step: "Create later" skip button
- Tariff selection: NOT in MVP

## Two Services, One Repo

- `cmd/server` — REST API for admin dashboard
- `cmd/bot` — Telegram bot for end users (telego)
- Communication: Redis Pub/Sub event bus (bot:reload, bot:stop, bot:start, bot:settings, campaign:send)
- Bot hot-reload: handler stores `*entity.Bot` pointer (not value) for live settings updates

## Frontend Architecture

- TanStack Router: file-based routing
- TanStack Query: server state management with cache invalidation
- TanStack Table + Form: data grids and form handling
- shadcn/ui + Tailwind: component library
- Design: minimalist black/white + noise grain, accent red (#EF3219)
