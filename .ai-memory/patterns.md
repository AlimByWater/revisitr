# Code Patterns

Established patterns across the codebase. Follow these when adding new code.

## Backend (Go)

### Repository: Lazy Init via Module

Repos take `*Module` (not raw `*sqlx.DB`) because DB connection is nil before `app.Run()` calls `Init()`. Access DB via `r.pg.DB()`:

```go
type Users struct { pg *Module }
func NewUsers(pg *Module) *Users { return &Users{pg: pg} }
func (r *Users) Create(ctx context.Context, user *entity.User) error {
    rows, err := r.pg.DB().NamedQueryContext(ctx, query, user)
}
```

All repos in `backend/internal/repository/postgres/` follow this pattern.

### JSONB: Scanner + Valuer with nil guard

Every JSONB entity type implements `sql.Scanner` + `driver.Valuer`. Always handle `nil` → initialize empty value:

```go
func (s *Schedule) Scan(src interface{}) error {
    if src == nil { *s = make(Schedule); return nil }
    // json.Unmarshal ...
}
```

~10 entity files use this pattern. Forgetting nil guard → nil map serializes to JSON `null` instead of `{}`.

### Controller Auth

- Controllers receive `jwtSecret string`, apply `middleware.Auth(g.jwtSecret)` to route groups
- Handlers extract org: `orgID, _ := c.Get("org_id")`
- Usecases verify ownership: `entity.OrgID == orgID`

### Error Handling

- Usecases define sentinel errors: `ErrNotFound`, `ErrNotOwner` (see `usecase/menus/menus.go` for example)
- Controllers map sentinels: not found -> 404, not owner -> 403, bind error -> 400, other -> 500
- Repos wrap errors: `fmt.Errorf("method.Name: %w", err)`

### Usecase Init

All usecases with a logger **must** be in the Init loop in `application/`. Missing Init -> nil logger -> panic on first log call. Same applies to `setup_test.go` in integration tests.

### DI: Explicit, No Frameworks

All dependencies are wired explicitly in `application/app.go`. No DI containers.

## Frontend (React + TanStack)

### Feature Module Structure

Every domain follows the same layout:
```
features/{domain}/
  types.ts    — interfaces matching backend JSON responses
  api.ts      — axios calls with typed responses
  queries.ts  — TanStack Query hooks with cache invalidation
```

19 feature modules currently follow this pattern.

### Route Pattern

```
routes/dashboard/{domain}/
  index.tsx       — list page with empty state
  $entityId.tsx   — detail/edit page
components/{domain}/
  Create{Entity}Modal.tsx — creation modal
```

### Auth Flow

- **State**: Zustand store
- **Persistence**: localStorage (not cookies)
- **API interceptor**: reads tokens from localStorage directly (avoids circular imports with store)
- **Refresh**: queue pattern prevents concurrent refresh calls
- **Route guard**: dashboard `beforeLoad` checks localStorage

### Mock API

`src/lib/mock-api.ts` provides full mock data for all endpoints. Enabled via `VITE_MOCK_API=true` build arg. Currently active in production build for demo purposes.
