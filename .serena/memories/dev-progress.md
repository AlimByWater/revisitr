# Development Progress

## Current State (2026-03-27)
- **Phases 1-4**: All implemented, deployed to production
- **RFM v2**: Phases 1-4 complete (local, not deployed). Phase 5 (Frontend) TODO.
- **Migrations**: 00001-00029 applied to prod, 00030 committed (rfm_templates, not yet deployed)
- **Last commit**: `52369ea` — fix: add missing SegmentFilter import and use unknown cast in test mocks
- **All committed and pushed**: RFM v2 (5 phases) + Backlog v2 (phases A-E frontend)
- **CI**: All pipelines green (Backend, Bot, Frontend, Infrastructure)

## RFM v2 Implementation Progress

### Phase 1 — Data Model ✅
- Migration 00030: template fields on rfm_configs + v2 score fields on bot_clients
- Dropped legacy rfm_recency/rfm_frequency/rfm_monetary columns (no v1/v2 coexistence)
- entity/rfm.go: 7 new segments, RFMTemplate, StandardTemplates (coffeegng/qsr/tsr/bar), RFMConfig with template fields
- entity/bot_client.go: r_score/f_score/m_score, recency_days/frequency_count/monetary_sum, total_visits_lifetime, last_visit_date
- CHECK constraints on template_type, template_key, score ranges 1-5
- 8 unit tests for entity layer

### Phase 2 — Template Engine + Scoring ✅
- service/rfm/rfm.go: fully rewritten — template-driven ScoreRecency/ScoreFrequency, ClassifySegment with 7 segments priority order
- repo/bot_clients.go: UpdateRFMScores writes all 10 v2 fields
- repo/loyalty.go: GetRFMStats — SQL with F=90d, M=180d windows, total_visits_lifetime
- DI updated in cmd/server/main.go (rfmService.New now takes configRepo)
- 28 test cases: all templates, all segments, design doc examples, RecalculateAll mock integration
- Old v1 constants in segment.go marked Deprecated

### Phase 3 — API (Templates + Onboarding) ✅
- entity/rfm.go: ValidateCustomThresholds, GetOnboardingQuestions (3 questions), RecommendTemplate (scoring + tie-break by Q1)
- usecase/rfm/rfm.go: ListTemplates, GetActiveTemplate, SetTemplate (validate+save+auto-recalc), GetOnboardingQuestions, RecommendTemplate
- controller/http/group/rfm/rfm.go: 5 new handlers — GET /templates, GET /template, PUT /template, GET /onboarding/questions, POST /onboarding/recommend
- repo/postgres/rfm.go: UpsertConfig extended with all template fields (type, key, custom name/thresholds)
- 13 new unit tests (usecase): ListTemplates, GetActiveTemplate (default+configured), SetTemplate (standard+custom+invalid key+invalid thresholds), RecommendTemplate (all same+tie-break+invalid count+invalid ID), OnboardingQuestions
### Phase 4 — Enhanced Dashboard + Segment Detail ✅
- repo/rfm.go: GetSegmentSummary SQL updated with AVG/SUM(monetary_sum) for avg_check/total_check
- repo/rfm.go: GetSegmentClients — paginated, sortable (6 columns), SQL-injection safe via whitelist
- entity/rfm.go: SegmentClientRow, SegmentClientsResponse structs
- usecase/rfm.go: GetSegmentClients with segment validation, pagination clamping (max 100)
- controller/rfm.go: GET /segments/:segment/clients with page/per_page/sort/order query params
- 4 new unit tests: GetSegmentClients, InvalidSegment, PaginationClamping, Page0Clamped
### Phase 5 — Frontend ✅
- types.ts: v2 segments (7), RFMTemplate, SetTemplateRequest, SegmentClientRow, OnboardingQuestion, TemplateRecommendation
- api.ts: 10 endpoints (dashboard, recalculate, config, templates, active template, set template, onboarding questions, recommend, segment clients)
- queries.ts: 8 SWR hooks with cache invalidation (dashboard, config, templates, active template, onboarding, recommend, segment clients, recalculate)
- utils.ts: shared formatMoney, formatDate, pluralClients, escapeCsvField
- Routes (4): /dashboard/rfm, /dashboard/rfm/onboarding, /dashboard/rfm/template, /dashboard/rfm/segments/:segment
- Onboarding page: 3-step quiz, recommendation, use/choose other/custom actions, loading/error states, mutation feedback
- Dashboard page: 7-segment table (icons, count, %, avg check, total), recalculate button, template info, success/error feedback
- Template page: 4 standard template cards (confirm on double-click), custom editor (name + 4×2 thresholds with validation), async save with error handling
- Segment detail: sortable client table (7 cols), RFM ScoreBadge (1-5 color-coded), pagination, CSV export (with field escaping), ScenarioLauncher button
- Sidebar: "RFM-сегментация" added under Аналитика (both default and Aurora themes)
- clients/segments.tsx: updated from v1 to v2 types (distribution→segment/client_count, recency_days→period_days)
- Improvements applied: mutation error handling (try/catch + feedback), loading spinners, disabled buttons during pending, no unused imports
- Unit tests: Vitest tests written for onboarding, template, segment detail pages
- NOT done: E2E tests (Playwright)


## Phase 4 Stubs (pre-existing, documented in docs/phase4-analysis.md)
- Wallet: no .pkpass generation, no APNs/FCM, RefreshPassBalance unused
- Marketplace: no bot interface, no Telegram Payments, only Scenario 1
- Billing: no real payment SDK, no webhook signature verification, no auto-renewal
- Admin Bot: no rich media, no auto-actions, no scheduler notifications
- Advanced Campaigns: manual A/B winner, no multichannel, no opens tracking
- Segmentation: ComputePredictions is placeholder, rules stored but not evaluated

## CI/CD
- GitHub Actions: lint → test → build → push GHCR → deploy → migrate
- Self-hosted runner on production server

## Test Coverage
- 122+ unit tests (18 packages) + 28 RFM service tests + 8 RFM entity tests + 17 RFM usecase tests (Phase 3+4)
- 88 integration tests
- Router boot test (154 routes, 20 groups)
- No E2E tests yet (Playwright setup exists but no test files)
