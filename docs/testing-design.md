# Test Design: Phase 2 Unit & Integration Tests

## Overview

Coverage gaps identified:
- **Unit tests missing**: `bots`, `pos` usecases + `autoaction` service (executor, scheduler, hook) + `pos/sync` service
- **Integration tests missing**: All Phase 2 API endpoints (campaigns v2, promotions v2, integrations v2, loyalty v2, dashboard v2, clients v2)
- **No tests needed**: `storage` (interface only), `pos/provider.go` (interface only)

Established patterns to follow:
- Manual mock structs with function fields (no external libraries)
- `errors.Is()` for sentinel error checking
- `//go:build integration` tag for integration tests
- `httptest.NewServer` + real Gin router for integration
- `t.Cleanup()` + direct SQL for test data cleanup

---

## Part 1: Unit Tests

### 1A. `backend/internal/usecase/bots/bots_test.go`

**Package**: `bots` (same package)

**Mocks needed**:
```go
type mockBotsRepo struct {
    createFn      func(ctx context.Context, bot *entity.Bot) error
    getByIDFn     func(ctx context.Context, id int) (*entity.Bot, error)
    getByOrgIDFn  func(ctx context.Context, orgID int) ([]entity.Bot, error)
    updateFn      func(ctx context.Context, bot *entity.Bot) error
    updateSettFn  func(ctx context.Context, id int, s entity.BotSettings) error
    deleteFn      func(ctx context.Context, id int) error
}

type mockBotClientsRepo struct {
    countByBotIDFn func(ctx context.Context, botID int) (int, error)
}
```

**Test matrix** (23 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| Create | success | ID assigned, OrgID set, default settings/modules/buttons |
| Create | repo_error | error propagated |
| GetByOrgID | success | returns list |
| GetByOrgID | repo_error | error propagated |
| GetByID | success | returns bot |
| GetByID | not_found | `ErrBotNotFound` |
| GetByID | wrong_org | `ErrNotBotOwner` |
| Update | success_name | only name changed |
| Update | success_status | only status changed |
| Update | success_both | both changed |
| Update | not_found | `ErrBotNotFound` |
| Update | wrong_org | `ErrNotBotOwner` |
| Delete | success | repo.Delete called |
| Delete | not_found | `ErrBotNotFound` |
| Delete | wrong_org | `ErrNotBotOwner` |
| GetSettings | success | returns settings |
| GetSettings | not_found | `ErrBotNotFound` |
| GetSettings | wrong_org | `ErrNotBotOwner` |
| UpdateSettings | success_modules | only modules patched |
| UpdateSettings | success_buttons | only buttons patched |
| UpdateSettings | success_form | only form patched |
| UpdateSettings | success_welcome | only welcome_message patched |
| UpdateSettings | not_found | `ErrBotNotFound` |

**Helper**:
```go
func newUC(bots botsRepo, clients botClientsRepo) *Usecase {
    uc := New(bots, clients)
    uc.Init(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
    return uc
}

func testBot(orgID int) *entity.Bot {
    return &entity.Bot{ID: 1, OrgID: orgID, Name: "Test Bot", Status: "active"}
}
```

---

### 1B. `backend/internal/usecase/pos/pos_test.go`

**Package**: `pos` (same package)

**Mock needed**:
```go
type mockPOSRepo struct {
    createFn     func(ctx context.Context, pos *entity.POSLocation) error
    getByIDFn    func(ctx context.Context, id int) (*entity.POSLocation, error)
    getByOrgIDFn func(ctx context.Context, orgID int) ([]entity.POSLocation, error)
    updateFn     func(ctx context.Context, pos *entity.POSLocation) error
    deleteFn     func(ctx context.Context, id int) error
}
```

**Test matrix** (17 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| Create | success | ID assigned, empty schedule defaulted |
| Create | nil_schedule | defaults to empty `entity.POSSchedule{}` |
| Create | repo_error | error propagated |
| GetByOrgID | success | returns list |
| GetByOrgID | repo_error | error propagated |
| GetByID | success | returns location |
| GetByID | not_found | `ErrPOSNotFound` (sql.ErrNoRows mapped) |
| GetByID | wrong_org | `ErrNotPOSOwner` |
| GetByID | repo_error | error propagated (non-ErrNoRows) |
| Update | success_name | only name changed |
| Update | success_schedule | only schedule changed |
| Update | success_is_active | only is_active changed |
| Update | not_found | `ErrPOSNotFound` |
| Update | wrong_org | `ErrNotPOSOwner` |
| Update | refetch_after | verifies GetByID called after Update |
| Delete | success | repo.Delete called |
| Delete | not_found | `ErrPOSNotFound` |

---

### 1C. `backend/internal/service/autoaction/executor_test.go`

**Package**: `autoaction` (same package)

**Mock needed**:
```go
type mockScenariosRepo struct {
    createActionLogFn     func(ctx context.Context, log *entity.AutoActionLog) error
    checkDedupFn          func(ctx context.Context, scenarioID, clientID int, key string) (bool, error)
    createDedupFn         func(ctx context.Context, scenarioID, clientID int, key string) error
    getActiveDateBasedFn  func(ctx context.Context) ([]entity.AutoScenario, error)
    getActiveByTriggerFn  func(ctx context.Context, triggerType string) ([]entity.AutoScenario, error)
}
```

**Test matrix** (14 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| Execute | success_multiple_actions | each action logged |
| Execute | message_fallback | no actions + message → wraps as campaign action |
| Execute | empty_actions_no_message | no-op |
| Execute | action_types | bonus/campaign/promo_code/level_change each handled |
| Execute | unknown_action_type | result="skipped" |
| Execute | log_creation_error | logged but doesn't fail |
| ExecuteWithDedup | first_execution | dedup check false → execute → create dedup |
| ExecuteWithDedup | duplicate | dedup check true → skip execution |
| ExecuteWithDedup | check_dedup_error | error propagated |
| ExecuteWithDedup | create_dedup_error | logged but doesn't fail main execution |
| logAction | success | log entry created with correct fields |
| logAction | marshal_error | graceful handling |
| executeAction | bonus | returns success result |
| executeAction | unknown | returns skipped with error |

---

### 1D. `backend/internal/service/autoaction/scheduler_test.go`

**Package**: `autoaction` (same package)

**Additional mock**:
```go
type mockBotClientsRepo struct {
    getByBotIDFn func(ctx context.Context, botID, limit, offset int) ([]entity.BotClient, int, error)
}
```

**Test matrix** (16 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| Evaluate | success | processes all scenarios |
| Evaluate | empty_scenarios | no-op |
| Evaluate | repo_error | error returned |
| evaluateBirthday | match_today | client with today's birthday triggered |
| evaluateBirthday | match_days_before | triggered N days before |
| evaluateBirthday | match_days_after | triggered N days after |
| evaluateBirthday | no_birthdate | skipped |
| evaluateBirthday | no_match | birthday outside window |
| evaluateBirthday | dedup_key_format | "birthday:YYYY" |
| evaluateHoliday | exact_match | today matches month/day |
| evaluateHoliday | no_match | different date |
| evaluateHoliday | missing_config | returns early (no month/day) |
| evaluateInactivity | inactive_client | registered > N days ago |
| evaluateInactivity | active_client | registered recently, skipped |
| evaluateInactivity | missing_days | returns early (days=0) |
| getAllClients | pagination | multiple batches fetched correctly |

---

### 1E. `backend/internal/service/autoaction/hook_test.go`

**Package**: `autoaction` (same package)

**Test matrix** (8 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| OnEvent | matching_scenario | scenario found, client matched, executed |
| OnEvent | no_scenarios | early return |
| OnEvent | client_not_found | skipped for that scenario |
| OnEvent | multiple_scenarios | all processed |
| OnEvent | execute_error | logged, continues |
| OnEvent | repo_error | error returned |
| findClient | found_first_batch | returns client |
| findClient | found_later_batch | pagination works |
| findClient | not_found | returns nil |

---

### 1F. `backend/internal/service/pos/sync_test.go`

**Package**: `pos` (same package)

**Mocks needed**:
```go
type mockIntegrationsRepo struct {
    updateLastSyncFn func(ctx context.Context, id int, status string) error
    upsertOrderFn    func(ctx context.Context, order *entity.ExternalOrder) error
    getByIDFn        func(ctx context.Context, id int) (*entity.Integration, error)
    getActiveFn      func(ctx context.Context) ([]entity.Integration, error)
    upsertAggFn      func(ctx context.Context, agg *entity.IntegrationAggregate) error
    upsertClientMapFn func(ctx context.Context, m *entity.IntegrationClientMap) error
    matchClientsFn   func(ctx context.Context, integrationID int) (int, error)
}

type mockClientsRepo struct {
    getByPhoneFn func(ctx context.Context, orgID int, phone string) (*entity.BotClient, error)
}
```

**Note**: `NewProvider()` is a package-level function creating concrete providers. Tests need either:
- A `mock` provider type (already exists in `pos/mock.go`)
- Or test integrations with `provider_type: "mock"`

**Test matrix** (18 test functions):

| Method | Test | Asserts |
|--------|------|---------|
| TestConnection | success | no error |
| TestConnection | provider_error | error propagated |
| Sync | success_with_orders | orders upserted, status "active" |
| Sync | connection_fail | status updated to "error" |
| Sync | first_sync_default_30d | since = now - 30 days |
| Sync | subsequent_sync_uses_lastsyncat | since = LastSyncAt |
| Sync | order_with_phone_match | client_id populated |
| Sync | order_without_phone | client_id stays nil |
| Sync | clients_repo_nil | graceful (no phone lookup) |
| Sync | get_orders_error | status "error", error returned |
| Sync | upsert_order_partial_fail | continues processing remaining |
| SyncAggregates | success | aggregates upserted, clients matched |
| SyncAggregates | get_aggregates_error | error returned |
| SyncAggregates | upsert_aggregate_error | continues |
| SyncAggregates | match_clients_error | logged, continues |
| SyncAll | multiple_integrations | each synced |
| SyncAll | partial_failure | continues on error |
| SyncAll | empty_list | no-op |

---

## Part 2: Integration Tests

All integration tests go in `backend/tests/integration/` with `//go:build integration` tag.

### Prerequisite: Extend Setup

**File**: `backend/tests/integration/setup_test.go`

Add Phase 2 usecases and controller groups to TestMain:

```go
// New repos needed:
campaignClicksRepo   // for campaign analytics
promoCodesRepo       // for promo code operations
integrationAggRepo   // for aggregates
loyaltyReserveRepo   // for point reserves

// New usecases:
campaignsUC   // with Schedule, Analytics, Click
promotionsUC  // with PromoCode validation/activation
integrationsUC // with Aggregates
loyaltyUC     // with Reserve, EarnFromCheck, BatchUpdateLevels
dashboardUC   // with WithSalesUsecase

// New controller groups registered:
campaignsGroup, promotionsGroup, integrationsGroup,
loyaltyGroup, dashboardGroup, clientsGroup
```

### Helpers to Add

**File**: `backend/tests/integration/helpers_test.go`

```go
// Domain-specific helpers (following existing mustCreatePromotion pattern):
func mustCreateBot(t *testing.T, token, name string) botResp
func mustCreateCampaign(t *testing.T, token, name string, botID int) campaignResp
func mustCreateLoyaltyProgram(t *testing.T, token, name string) programResp
func mustCreateLoyaltyLevel(t *testing.T, token string, programID int, name string, threshold int) levelResp
func mustCreateIntegration(t *testing.T, token, name, providerType string, botID int) integrationResp
func mustCreatePOS(t *testing.T, token, name string, botID int) posResp
func mustCreatePromoCode(t *testing.T, token, code string, promoID int) promoCodeResp
func mustCreateClient(t *testing.T, botID int) clientResp  // Direct DB insert
func mustCreateScenario(t *testing.T, token string, botID int, name string) scenarioResp
```

---

### 2A. `backend/tests/integration/campaigns_v2_test.go`

**Test functions** (12 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestCampaign_CRUD | create → get → list → update → delete | 201, 200, fields match |
| TestCampaign_Send | create draft → send | status changes to "sent" |
| TestCampaign_Send_AlreadySent | send sent campaign | 409 |
| TestCampaign_Schedule | create → schedule → get | scheduled_at set, status "scheduled" |
| TestCampaign_CancelSchedule | schedule → cancel | status back to "draft" |
| TestCampaign_CancelSchedule_NotScheduled | cancel non-scheduled | 409 |
| TestCampaign_Analytics | send → record clicks → get analytics | click counts, button breakdown |
| TestCampaign_RecordClick | send → click with button_idx | 201, click persisted |
| TestCampaign_PreviewAudience | create audience filter → preview | returns count |
| TestCampaign_WrongOrg | create as org1, access as org2 | 403 for get/update/delete |
| TestScenario_CRUD | create → list → update → delete | all ops work |
| TestScenario_Templates | list templates → clone | template cloned with new bot_id |

---

### 2B. `backend/tests/integration/promotions_v2_test.go`

**Test functions** (10 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestPromoCode_Create | create promotion → create promo code | 201, code persisted |
| TestPromoCode_Generate | generate unique code | returns random code string |
| TestPromoCode_Validate_Success | create code → validate | validation result with discount |
| TestPromoCode_Validate_Expired | create expired code → validate | 409 expired |
| TestPromoCode_Validate_LimitReached | exhaust usage limit → validate | 409 limit |
| TestPromoCode_Validate_PerUserLimit | exhaust per-user limit → validate | 409 per-user limit |
| TestPromoCode_Validate_MinAmount | validate with low amount | 409 min amount |
| TestPromoCode_Activate | validate → activate | usage count incremented |
| TestPromoCode_Deactivate | create → deactivate → validate | 409 inactive |
| TestPromoCode_ChannelAnalytics | create codes with channels → analytics | per-channel breakdown |
| TestPromoCode_ByPromotion | create promo + codes → get by promotion | lists codes for promotion |

---

### 2C. `backend/tests/integration/integrations_v2_test.go`

**Note**: Existing `integrations_test.go` covers basic CRUD. New file extends with Phase 2 endpoints.

**Test functions** (7 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestIntegration_TestConnection_Mock | create mock integration → test | 200 success |
| TestIntegration_Sync_Mock | create mock → sync | orders created, lastSyncAt updated |
| TestIntegration_GetOrders | sync → get orders | paginated list with limit/offset |
| TestIntegration_GetCustomers | sync → get customers | customer list with search |
| TestIntegration_GetMenu | mock → get menu | menu structure returned |
| TestIntegration_GetStats | sync → get stats | order count, revenue |
| TestIntegration_GetAggregates | sync aggregates → get | date-range filtered aggregates |

---

### 2D. `backend/tests/integration/loyalty_v2_test.go`

**Test functions** (10 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestLoyalty_Program_CRUD | create → get → list → update | all ops work |
| TestLoyalty_Level_Create | program → create level | 201, threshold/percent set |
| TestLoyalty_Level_BatchUpdate | program → create levels → batch update | all levels updated atomically |
| TestLoyalty_Level_Delete | create → delete | 200, level removed |
| TestLoyalty_EarnFromCheck | program + level + client → earn | points added, level evaluated |
| TestLoyalty_EarnFromCheck_LevelUp | earn enough to cross threshold | level changes |
| TestLoyalty_Reserve_Success | earn → reserve → confirm | balance reduced, reserve confirmed |
| TestLoyalty_Reserve_Cancel | earn → reserve → cancel | balance restored |
| TestLoyalty_Reserve_InsufficientPoints | reserve more than balance | 400 insufficient |
| TestLoyalty_WrongOrg | create as org1, access as org2 | 403 |

---

### 2E. `backend/tests/integration/dashboard_v2_test.go`

**Test functions** (3 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestDashboard_Widgets | create bot + clients → widgets | client count, bot count |
| TestDashboard_Charts | create clients with dates → charts | time-series data |
| TestDashboard_Sales | sync integration aggregates → sales | revenue/check/transaction aggregates |

---

### 2F. `backend/tests/integration/clients_v2_test.go`

**Test functions** (6 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestClients_List | create bot + clients → list | paginated, filter works |
| TestClients_Get | create client → get profile | full profile with loyalty |
| TestClients_Update | create → update tags/notes | fields changed |
| TestClients_Stats | create clients → stats | total, active, new counts |
| TestClients_Count | create clients with filter → count | filtered count correct |
| TestClients_Identify | create client with phone → identify | found by phone |

---

## Implementation Order

Priority by business value and dependency chain:

### Wave 1: Unit Tests (no infrastructure needed)
1. **`bots_test.go`** — simple CRUD, quick win (23 tests)
2. **`pos_test.go`** — simple CRUD, quick win (17 tests)
3. **`executor_test.go`** — core auto-action logic (14 tests)

### Wave 2: Unit Tests (more complex)
4. **`scheduler_test.go`** — date logic, time-sensitive (16 tests)
5. **`hook_test.go`** — event-driven (8 tests)
6. **`sync_test.go`** — orchestration, multiple repos (18 tests)

### Wave 3: Integration Test Infrastructure
7. **Extend `setup_test.go`** — add Phase 2 usecases/groups
8. **Extend `helpers_test.go`** — add domain helpers

### Wave 4: Integration Tests (require Docker)
9. **`campaigns_v2_test.go`** — most complex, 12 tests
10. **`promotions_v2_test.go`** — promo code flows, 10 tests
11. **`loyalty_v2_test.go`** — point operations, 10 tests
12. **`integrations_v2_test.go`** — sync flows, 7 tests
13. **`clients_v2_test.go`** — profile/identify, 6 tests
14. **`dashboard_v2_test.go`** — aggregate queries, 3 tests

---

### 2G. `backend/tests/integration/storage_test.go`

**Requires**: MinIO в docker-compose (уже есть в prod, добавить в dev).

**Test functions** (5 tests):

| Test | Flow | Assertions |
|------|------|------------|
| TestStorage_Upload | upload file → check result | FileInfo returned, key/URL correct |
| TestStorage_Upload_CreatesBucket | upload to new bucket | bucket auto-created |
| TestStorage_Upload_ExistingBucket | upload twice to same bucket | no error on second |
| TestStorage_Delete | upload → delete | no error, object removed |
| TestStorage_GetURL | get URL for key | correct format `/{bucket}/{key}` |

---

## Totals

| Category | Files | Tests |
|----------|-------|-------|
| Unit: Usecases | 2 | 40 |
| Unit: Services | 4 | 56 |
| Integration: Setup | 2 | — |
| Integration: API Tests | 6 | 48 |
| Integration: Storage | 1 | 5 |
| **Total** | **15** | **149** |

## Run Commands

```bash
# Unit tests only (fast, no infra)
cd backend && go test -race -count=1 ./internal/usecase/bots/ ./internal/usecase/pos/ ./internal/service/autoaction/ ./internal/service/pos/

# Integration tests (requires docker compose up)
cd backend && go test -race -tags=integration ./tests/integration/...

# All tests
cd backend && go test -race ./internal/usecase/... ./internal/service/... && \
  go test -race -tags=integration ./tests/integration/...
```
