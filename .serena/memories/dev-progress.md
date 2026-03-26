# Development Progress

## Current State (2026-03-26)
- **Phases 1-4**: All implemented, deployed to production
- **Migrations**: 00001-00029, all applied to prod
- **Last commit**: `4c03748` — router boot test + auto-migrations

## Phase 4 Stubs (documented in docs/phase4-analysis.md)
- Wallet: no .pkpass generation, no APNs/FCM, RefreshPassBalance unused
- Marketplace: no bot interface, no Telegram Payments, only Scenario 1
- Billing: no real payment SDK, no webhook signature verification, no auto-renewal
- Admin Bot: no rich media, no auto-actions, no scheduler notifications
- Advanced Campaigns: manual A/B winner, no multichannel, no opens tracking
- Segmentation: ComputePredictions is placeholder, rules stored but not evaluated

## Remaining Critical Issues
1. Wallet RefreshPassBalance not called from loyalty
2. Subscription+invoice creation not atomic

## CI/CD
- GitHub Actions: lint → test → build → push GHCR → deploy → migrate
- Self-hosted runner on production server
- Manual deploy available: `docker save | ssh docker load`

## Test Coverage
- 122+ unit tests (18 packages)
- 88 integration tests
- Router boot test (154 routes, 20 groups)
- No E2E tests yet (Playwright setup exists but no test files)
