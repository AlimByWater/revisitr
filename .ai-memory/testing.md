# Testing

## Test Levels

| Level | Tool | Scope | Infra needed |
|-------|------|-------|--------------|
| Backend unit | Go `testing` | usecase layer, manual mocks | None |
| Frontend unit | Vitest + MSW | components, stores, API hooks | None |
| Integration | Go `httptest` + real DB | API handlers + PostgreSQL | Docker |
| E2E | Playwright | full user flows via browser | Full stack |
| Bot unit | Go `testing` | bot handlers, mock telego | None |

## Commands

```bash
go test -race ./internal/usecase/...                          # backend unit
npx vitest run                                                # frontend unit
go test -race -tags=integration ./tests/integration/...       # integration (needs docker)
npx playwright test                                           # e2e (needs full stack)
```

## Coverage

### Unit Tests: 122+ tests across 18 usecase packages

Key packages: billing (18), bots (23), segments (25), pos (18), marketplace (15), wallet (14), rfm (7), menus (8), onboarding (5).

### Integration Tests: 93 tests (92 pass)

Key suites: auth (10), bots (12), campaigns_v2 (12), loyalty_v2 (10), segments (7), integrations_v2 (7), promotions (7+6).

Requires: `docker compose up -d postgres redis` + migrations up to 00029.

**1 pre-existing failure**: `TestIntegrations_SyncNow` — iiko integration without API credentials -> 500. This is expected.

### Frontend Unit (RFM): 75 tests (6 files)

## Mock Pattern

All backend mocks: **struct with function fields** (no external mocking libraries):

```go
type MockBotRepo struct {
    CreateFn func(ctx context.Context, bot *entity.Bot) error
    GetFn    func(ctx context.Context, id int) (*entity.Bot, error)
}
func (m *MockBotRepo) Create(ctx context.Context, bot *entity.Bot) error {
    return m.CreateFn(ctx, bot)
}
```

Integration tests use real DB via `pgMod.DB()` with `t.Cleanup()`.

## Build Tags

Integration tests use build tag: `//go:build integration`

## CI

GitHub Actions with self-hosted runner. Workflows:
- `backend.yml` — build + test
- `frontend.yml` — build (with `VITE_MOCK_API=true`)
- `admin-bot.yml` — admin bot deploy
