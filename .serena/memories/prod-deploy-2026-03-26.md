# Production Deploy Session — 2026-03-26

## Context
First production deploy of Phases 3-4 to elysium.fm/revisitr.

## Issues Found & Fixed (8 total)

### Backend crashes (3 Gin route param conflicts)
- `menus/:menuId` conflicted with `menus/:id` — changed to `:id`
- `wallet/passes/:serial` conflicted with `wallet/passes/:id` — changed to `:id`
- `loyalty/programs/:programId` conflicted with `loyalty/programs/:id` — changed to `:id`
- **Rule**: All handlers under same Gin group must use identical param name at same path level

### Onboarding infinite loop (steps: null)
- Go nil map serializes to JSON `null`, frontend couldn't iterate
- Fix: Initialize `Steps = make(map[string]OnboardingStep)` in `Scan()` when nil

### Auth redirect loop (window.location vs request.url)
- React Router loader runs before URL updates in browser
- Fix: Use `new URL(request.url)` from loader params instead of `window.location.pathname`

### Frontend/backend type mismatch
- `UpdateOnboardingRequest.Step` was `int` in Go, frontend sent `string` ("info", "bot")
- Fix: Changed Go type to `string`

### MarketplaceStats/WalletStats sqlx failure
- Missing `db:"column_name"` struct tags (only had `json` tags)
- sqlx requires `db` tags for column mapping

### Onboarding crash at completion (array OOB)
- `current_step: 6` but ONBOARDING_STEPS array is 0-5
- Fix: `Math.min(data.current_step, ONBOARDING_STEPS.length - 1)`

### Migrations FK errors
- Migrations 00027-00029 referenced `clients(id)` — table doesn't exist
- Fix: Changed FK to `bot_clients(id)`

## Improvements Added

1. **Router boot test** (`backend/internal/controller/http/router_test.go`)
   - Registers all 20 groups (154 routes) with nil deps
   - Catches route param conflicts at CI time without DB

2. **Auto-migrations on deploy**
   - `deploy.sh`: runs `goose up` after backend container restart
   - `backend.yml`: runs `migrate.sh` after deploy step in CI

3. **Health check** — already existed (30s polling on `/healthz`)

## Lessons Learned
- Always test with real DB before first prod deploy
- Gin route params are group-scoped, not handler-scoped
- Go nil maps → JSON null, always initialize in Scan()
- React Router loaders: use request.url, not window.location
- sqlx needs `db` tags, `json` tags are ignored for SQL mapping
