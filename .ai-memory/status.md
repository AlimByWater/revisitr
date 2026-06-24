# Development Status

Last updated: 2026-06-17

## What's Deployed to Production

- Phases 1-4 (Foundation -> Expansion): all deployed
- All migrations applied to prod through goose version **45** (verified via `goose_db_version`)
- Frontend with mock API enabled (`VITE_MOCK_API=true`)
- Admin bot CI/CD workflow active
- Everything previously listed as "pending deploy" (RFM v2, Backlog v2, PDF
  corrections, campaigns restructuring, Bot Architecture v2, Frontend Redesign
  v2) is now on prod.

## Recent Work (on main)

- **Bot menu module**: menu editor UX overhaul, preset customization, preview
  geometry aligned with iOS/Figma, callbacks kept on edited-message path
- **Bot registration**: refactored into sequential multi-field flow
- **iiko Cloud**: live — orders sync end-to-end (deliveries-window timezone fix);
  mechanics + tooling in `docs/integrations/iiko/PLAYBOOK.md`
- **Bot media**: download media files and upload to telegram-bot-api instead of
  passing URL (commit 95ab67c)
- **Bot debug logging**: menu rendering + callbacks (commit 9743f60)

## Known Incomplete / Stubs

- **Wallet**: RefreshPassBalance not called, pass generation not implemented
- **Billing**: payment provider integration missing
- **Marketplace**: bot interface + loyalty integration missing
- **Campaigns+**: A/B statistics incomplete
- **Segmentation+**: ComputePredictions is stub
- **Subscription+invoice**: not atomic

## Backend Endpoints Still Needed

- Account settings (profile, security, requisites)
- Custom segments CRUD
- Bot settings persistence (new button type, form options)

## Migrations — Two Directories

goose runs over **two** migration sets (see `infra/scripts/migrate.sh`):

- `backend/migrations/` → baked into image as `/migrations` (00001-00032,
  00041-00045). Core schema.
- `migrations/` (repo root) → mounted as `/extra-migrations` (00033-00040).
  master_bot, post_codes, and the `baratie_*` demo content/templates.

Both are git-tracked and both applied on prod. Note `157571d` renamed
`backend/migrations/00033_bot_modules_menu_booking.sql` -> `00041_...` to avoid
colliding with the root 33-40 series.

Latest version on prod: **45**.
