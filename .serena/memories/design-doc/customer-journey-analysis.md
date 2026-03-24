# Customer Journey Analysis — Design Doc vs Code (2026-03-24)

## Session Summary
Analyzed "Путь клиента" section from updated design document (docs/userdocs/Дизайн-документ revisitr.docx).
Cross-referenced with current codebase. Identified gaps and architectural decisions.

## Key Findings
1. **Auth**: Works (register+login), but missing phone verification via Telegram
2. **Onboarding**: Completely absent — biggest gap. 6-step wizard needed post-registration
3. **Entity hierarchy**: Doc requires LoyaltyProgram → Bot → POS. Code has flat org_id relationships
4. **All CRUD modules exist** as separate pages but no guided flow connecting them
5. **Business logic missing**: loyalty calculations, auto-actions, campaign sending, tracking

## Decided Architecture
- `program_id` in bots: nullable (bot can exist without program)
- `bot_id` in pos: nullable
- Multiple loyalty programs per org: allowed
- Bonus: both % and fixed amount (need reward_type field)
- Levels can go down (yearly reset possible)
- QR code: static, one per client
- POS scanning: in our system (not iiko/rkeeper)
- Campaigns: scheduled sending supported, Redis queue with interface for Kafka
- Admin bot: separate from client bot (third service)
- Phone normalization needed for client matching

## Implementation Phases
1. Foundation: DB migration (hierarchy), loyalty engine, file storage (MinIO)
2. Core: integrations v1 (iiko aggregate), campaigns, promo codes, auto-actions engine
3. Deepening: integrations v2 (client linking), RFM segmentation, onboarding, bot constructor
4. Post-MVP: admin bot, wallet, marketplace, billing

## Open Questions
- Bot modules list (what toggles exist in constructor)
- iiko API version (Cloud v2 or legacy)
- Polling vs webhook for integrations
