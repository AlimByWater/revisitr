# Gotchas & Known Pitfalls

Things that have caused bugs or confusion. Check here before debugging.

## Backend (Go)

### Gin Route Param Conflicts
All handlers under the same router group must use the **same param name** at the same path level. Using `:id` in one handler and `:menuId` in another under the same group causes routing failures.

### Missing `db` Tags on Structs
sqlx requires `db:"col_name"` tags for scanning. `json` tags alone won't work. Stats/aggregate structs are especially prone to this.

### nil Go Maps -> JSON null
Uninitialized Go maps serialize to `null` in JSON, not `{}`. Always initialize maps in `Scan()` methods and constructors.

### Schema vs Entity Mismatches
- `bot_clients.username` is nullable in DB but `entity.BotClient.Username` is `string` (not `*string`). Causes scan error on NULL. Workaround: always provide username when inserting.
- `bot_clients` has **no** `org_id` column. Org isolation is via `bots.org_id` (JOIN through `bot_id`).
- `auto_scenarios.org_id` / `bot_id` are nullable (for templates). Entity fields are `int`. `GetTemplates` uses `COALESCE(org_id, 0)`.

### Usecase Init Panic
If a usecase has `Init(ctx, logger)` but isn't in the Init loop in `application/` -> nil logger -> panic on first log call. Same trap in `setup_test.go` for integration tests.

### Migration Table Names
The actual table is `bot_clients`, not `clients`. Several early migrations referenced the wrong name.

### Loyalty CalculateBonus Chicken-and-Egg
`CalculateBonus` requires `LevelID` to be set. New `client_loyalty` with `LevelID=nil` returns bonus=0. Level is auto-assigned by `determineLevelID` after first `EarnPoints` call. First `EarnFromCheck` returns 0 bonus for brand-new client.

### Auto-Scenario Templates: NULL org_id/bot_id
Templates are global (not org-specific) and have NULL `org_id`/`bot_id`. Migration 00020 drops NOT NULL constraints. Queries use `COALESCE(org_id, 0)`.

### Loyalty Level reward_type Constraint
DB CHECK constraint enforces `reward_type` must be `'percent'` or `'fixed'`. Inserting other values fails at DB level.

## Frontend (React)

### TanStack Router: loader vs window.location
Use `request.url` from loader params, not `window.location.pathname`. The latter causes hydration mismatches.

### Type Mismatches with Backend
Onboarding `step` was `int` in Go but `string` in TS — caused silent failures. Always verify types match between `entity/*.go` and `features/*/types.ts`.

### DOM Manipulation Anti-pattern
Never use direct DOM manipulation (e.g., `document.querySelector` in React components). Use `useState`/`useRef`. Fixed in `MediaMessage` component (bot v2).

### MessageContentEditor Coupling
`MessageContentEditor` was originally coupled to `campaignsApi`. Now uses `onUpload` prop for decoupling. Follow this pattern for reusable components.

## Deployment

### Migrations Are NOT Auto-applied
After adding new migrations, you must run them manually on prod:
```bash
ssh -i ~/.ssh/elysium root@80.93.187.209 'docker exec infra-backend-1 goose -dir /migrations postgres "..." up'
```

### Port Mapping (Don't Confuse Local vs Prod)
| Service    | Local | Prod (host->container) |
|------------|-------|------------------------|
| Backend    | 8080  | 8090->8080             |
| Frontend   | 5173  | 3340->80               |
| PostgreSQL | 5433  | 5433->5432             |
| Redis      | 6380  | 6380->6379             |

### Frontend Mock API in Prod
`VITE_MOCK_API=true` is currently set in CI workflow (`frontend.yml`). To connect to real backend: remove `VITE_MOCK_API=true` from workflow build-args.

## Phase 4 Stubs (Not Fully Implemented)

These features exist as stubs — UI may be present but backend logic is incomplete:
- **Billing**: ~40% done, payment provider integration missing, `ProcessPayment` returns hardcoded "succeeded"
- **Wallet**: ~30% done, `RefreshPassBalance` not called from loyalty flow, pass generation not implemented
- **Marketplace**: ~35% done, bot interface and loyalty integration missing
- **Campaigns+**: ~30% done, A/B statistics incomplete
- **Segmentation+**: ~40% done, `ComputePredictions` is a stub

See `docs/phase4-analysis.md` for full breakdown.
