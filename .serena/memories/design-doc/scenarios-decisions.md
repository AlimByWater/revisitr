# Main Scenarios — Business Decisions (2026-03-24)

## Scenario 1: Loyalty Points
- Accrual: % of check (per level) OR fixed amount — user configures freely
- No templates for thresholds — price segments vary across clients
- Redemption at POS via integration (waiter asks, scans QR or enters phone)
- In-bot redemption NOT for MVP (marketplace, special items, payments — later)
- Client always has phone; QR is static, generated on registration

## Scenario 2: Campaigns
- Two systems: web UI (simplified: text+media) + admin Telegram bot (rich: circles, voice, stickers)
- Redis queue → repository interface → future Kafka
- Throttling: 30 msg/sec for large audiences
- Tracking: inline buttons + UTM links, switchable/combinable
- Scheduled/deferred sending supported
- Audience: by segment, loyalty level, or manual selection

## Scenario 3: Analytics + Integrations
- Priority: aggregate data → link to clients → menu items (confirmed)
- Client matching: by phone number (normalization required)
- Sources: iiko, r-keeper, 1C
- Compare loyalty vs non-loyalty customers (key metric)

## Scenario 4: Promotions
- Promo code = UTM analog for channel tracking (SMM, targeting, Yandex Maps)
- Format: auto-generated + user-defined (e.g. "BIRTHDAY2024")
- Auto-actions engine: trigger → timing → condition → actions[]
- Actions: bonus accrual, campaign/push, promo code creation
- Birthday example: 7 days before/after, condition=order, actions=[bonus+campaign+promo]
- Templates provided + fully custom creation

## Scenario 5: Segmentation
- RFM first (standard segments)
- 70-80% will use pre-built templates
- Custom segments: later phase

## Onboarding Order (confirmed)
1. Information → 2. Loyalty (simpler, higher-level) → 3. Bot (complex, needs loyalty) → 4. POS → 5. Integrations → 6. Next steps
- Each step skippable ("Create later")
- Tariff selection in step 6 or later, NOT in MVP
