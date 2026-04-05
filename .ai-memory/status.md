# Development Status

Last updated: 2026-04-05

## What's Deployed to Production

- Phases 1-4 (Foundation -> Expansion): all deployed
- Migrations 00001-00029: applied to prod
- Frontend with mock API enabled (`VITE_MOCK_API=true`)
- Admin bot CI/CD workflow active

## Committed to Main (Not Yet Deployed)

- **RFM v2**: migration 00030 (templates), 5 phases complete, 75 frontend unit tests
- **Backlog v2**: Phases A-E complete (frontend only, backend endpoints needed)
- **PDF Corrections**: 5 phases complete
- **Campaigns Restructuring**: MessageContentEditor + TelegramPreview replaces textarea+file
- **Bot Architecture v2**: all 5 phases complete (event bus, composite messages, Telegram sender, bug fixes, UI integration)
- **Frontend Redesign v2**: merged 2026-04-02, design system overhaul (75 files changed)
- Migrations 00030-00032: pending deploy

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

## Migration Count

32 migrations total (00001-00032). 00001-00029 on prod, 00030-00032 pending.
