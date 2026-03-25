package segments

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockSegmentsRepo struct {
	createFn        func(ctx context.Context, seg *entity.Segment) error
	getByIDFn       func(ctx context.Context, id int) (*entity.Segment, error)
	getByOrgIDFn    func(ctx context.Context, orgID int) ([]entity.Segment, error)
	updateFn        func(ctx context.Context, seg *entity.Segment) error
	deleteFn        func(ctx context.Context, id int) error
	getClientsFn    func(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error)
	syncClientsFn   func(ctx context.Context, segmentID int, clientIDs []int) error
	countByFilterFn func(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error)
}

func (m *mockSegmentsRepo) Create(ctx context.Context, seg *entity.Segment) error {
	return m.createFn(ctx, seg)
}
func (m *mockSegmentsRepo) GetByID(ctx context.Context, id int) (*entity.Segment, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockSegmentsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockSegmentsRepo) Update(ctx context.Context, seg *entity.Segment) error {
	return m.updateFn(ctx, seg)
}
func (m *mockSegmentsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockSegmentsRepo) GetClients(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error) {
	return m.getClientsFn(ctx, segmentID, limit, offset)
}
func (m *mockSegmentsRepo) SyncClients(ctx context.Context, segmentID int, clientIDs []int) error {
	return m.syncClientsFn(ctx, segmentID, clientIDs)
}
func (m *mockSegmentsRepo) CountByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error) {
	return m.countByFilterFn(ctx, orgID, f)
}

type mockClientsRepo struct {
	getIDsByFilterFn func(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error)
}

func (m *mockClientsRepo) GetIDsByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error) {
	return m.getIDsByFilterFn(ctx, orgID, f)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func testSegment(id, orgID int) *entity.Segment {
	return &entity.Segment{
		ID:        id,
		OrgID:     orgID,
		Name:      "test segment",
		Type:      "custom",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	repo := &mockSegmentsRepo{
		createFn: func(_ context.Context, seg *entity.Segment) error {
			seg.ID = 1
			return nil
		},
	}
	uc := New(repo, &mockClientsRepo{})

	seg, err := uc.Create(context.Background(), 10, &entity.CreateSegmentRequest{
		Name: "VIP", Type: "custom",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seg.OrgID != 10 {
		t.Errorf("expected org_id=10, got %d", seg.OrgID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Segment, error) {
			return nil, fmt.Errorf("segments.GetByID: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 99, 1)
	if err != ErrSegmentNotFound {
		t.Errorf("expected ErrSegmentNotFound, got: %v", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 5), nil // owned by org 5
		},
	}
	uc := New(repo, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 1, 99) // requesting as org 99
	if err != ErrNotSegmentOwner {
		t.Errorf("expected ErrNotSegmentOwner, got: %v", err)
	}
}

func TestDelete_OwnershipCheck(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 5), nil // owned by org 5
		},
		deleteFn: func(_ context.Context, _ int) error { return nil },
	}
	uc := New(repo, &mockClientsRepo{})

	if err := uc.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("owner should be able to delete: %v", err)
	}

	if err := uc.Delete(context.Background(), 1, 99); err != ErrNotSegmentOwner {
		t.Errorf("expected ErrNotSegmentOwner for non-owner, got: %v", err)
	}
}

func TestRecalculateCustom_NonCustomSegment(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			seg := testSegment(id, 1)
			seg.Type = "rfm" // not custom
			return seg, nil
		},
	}
	uc := New(repo, &mockClientsRepo{})

	err := uc.RecalculateCustom(context.Background(), 1, 1)
	if err != ErrNotCustomSegment {
		t.Errorf("expected ErrNotCustomSegment, got: %v", err)
	}
}

func TestRecalculateCustom_Success(t *testing.T) {
	synced := false
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 1), nil
		},
		syncClientsFn: func(_ context.Context, _ int, ids []int) error {
			synced = true
			if len(ids) != 2 {
				t.Errorf("expected 2 ids, got %d", len(ids))
			}
			return nil
		},
	}
	clients := &mockClientsRepo{
		getIDsByFilterFn: func(_ context.Context, _ int, _ entity.SegmentFilter) ([]int, error) {
			return []int{10, 20}, nil
		},
	}
	uc := New(repo, clients)

	if err := uc.RecalculateCustom(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !synced {
		t.Error("SyncClients was not called")
	}
}

func TestPreviewCount(t *testing.T) {
	repo := &mockSegmentsRepo{
		countByFilterFn: func(_ context.Context, _ int, _ entity.SegmentFilter) (int, error) {
			return 42, nil
		},
	}
	uc := New(repo, &mockClientsRepo{})

	count, err := uc.PreviewCount(context.Background(), 1, entity.SegmentFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

// ── Mock Rules Repo ─────────────────────────────────────────────────────────

type mockRulesRepo struct {
	createRuleFn           func(ctx context.Context, segmentID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error)
	getRulesFn             func(ctx context.Context, segmentID int) ([]entity.SegmentRule, error)
	deleteRuleFn           func(ctx context.Context, ruleID int) error
	deleteRulesBySegmentFn func(ctx context.Context, segmentID int) error
}

func (m *mockRulesRepo) CreateRule(ctx context.Context, segmentID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error) {
	return m.createRuleFn(ctx, segmentID, req)
}
func (m *mockRulesRepo) GetRules(ctx context.Context, segmentID int) ([]entity.SegmentRule, error) {
	return m.getRulesFn(ctx, segmentID)
}
func (m *mockRulesRepo) DeleteRule(ctx context.Context, ruleID int) error {
	return m.deleteRuleFn(ctx, ruleID)
}
func (m *mockRulesRepo) DeleteRulesBySegment(ctx context.Context, segmentID int) error {
	return m.deleteRulesBySegmentFn(ctx, segmentID)
}

// ── Mock Predictions Repo ───────────────────────────────────────────────────

type mockPredictionsRepo struct {
	upsertPredictionFn      func(ctx context.Context, pred *entity.ClientPrediction) error
	getPredictionsFn        func(ctx context.Context, orgID int, limit, offset int) ([]entity.ClientPrediction, error)
	getPredictionByClientFn func(ctx context.Context, clientID int) (*entity.ClientPrediction, error)
	getHighChurnClientsFn   func(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error)
	getPredictionSummaryFn  func(ctx context.Context, orgID int) (*entity.PredictionSummary, error)
}

func (m *mockPredictionsRepo) UpsertPrediction(ctx context.Context, pred *entity.ClientPrediction) error {
	return m.upsertPredictionFn(ctx, pred)
}
func (m *mockPredictionsRepo) GetPredictions(ctx context.Context, orgID int, limit, offset int) ([]entity.ClientPrediction, error) {
	return m.getPredictionsFn(ctx, orgID, limit, offset)
}
func (m *mockPredictionsRepo) GetPredictionByClient(ctx context.Context, clientID int) (*entity.ClientPrediction, error) {
	return m.getPredictionByClientFn(ctx, clientID)
}
func (m *mockPredictionsRepo) GetHighChurnClients(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error) {
	return m.getHighChurnClientsFn(ctx, orgID, threshold)
}
func (m *mockPredictionsRepo) GetPredictionSummary(ctx context.Context, orgID int) (*entity.PredictionSummary, error) {
	return m.getPredictionSummaryFn(ctx, orgID)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ── Tests: Rules ────────────────────────────────────────────────────────────

func TestAddRule_NoRulesConfigured(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}) // no WithRules

	_, err := uc.AddRule(context.Background(), 1, 10, entity.CreateSegmentRuleRequest{})
	if err == nil {
		t.Fatal("expected error for nil rules repo")
	}
}

func TestAddRule_SegmentNotFound(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Segment, error) {
			return nil, fmt.Errorf("segments.GetByID: %w", sql.ErrNoRows)
		},
	}
	rules := &mockRulesRepo{}
	uc := New(repo, &mockClientsRepo{}, WithRules(rules))

	_, err := uc.AddRule(context.Background(), 1, 10, entity.CreateSegmentRuleRequest{})
	if err != ErrSegmentNotFound {
		t.Errorf("expected ErrSegmentNotFound, got: %v", err)
	}
}

func TestAddRule_Success(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 10), nil
		},
	}
	rules := &mockRulesRepo{
		createRuleFn: func(_ context.Context, segID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error) {
			return &entity.SegmentRule{
				ID:        1,
				SegmentID: segID,
				Field:     req.Field,
				Operator:  req.Operator,
				Value:     req.Value,
			}, nil
		},
	}
	uc := New(repo, &mockClientsRepo{}, WithRules(rules))

	rule, err := uc.AddRule(context.Background(), 1, 10, entity.CreateSegmentRuleRequest{
		Field:    "total_orders",
		Operator: "gte",
		Value:    json.RawMessage(`10`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Field != "total_orders" {
		t.Errorf("expected field=total_orders, got %q", rule.Field)
	}
}

func TestGetRules_NilRepo(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})

	rules, err := uc.GetRules(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rules != nil {
		t.Errorf("expected nil, got %v", rules)
	}
}

func TestGetRules_Success(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 10), nil
		},
	}
	rules := &mockRulesRepo{
		getRulesFn: func(_ context.Context, _ int) ([]entity.SegmentRule, error) {
			return []entity.SegmentRule{{ID: 1}, {ID: 2}}, nil
		},
	}
	uc := New(repo, &mockClientsRepo{}, WithRules(rules))

	result, err := uc.GetRules(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result))
	}
}

func TestDeleteRule_NoRulesConfigured(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})

	err := uc.DeleteRule(context.Background(), 1, 10, 5)
	if err == nil {
		t.Fatal("expected error for nil rules repo")
	}
}

func TestDeleteRule_Success(t *testing.T) {
	deleted := false
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 10), nil
		},
	}
	rules := &mockRulesRepo{
		deleteRuleFn: func(_ context.Context, _ int) error {
			deleted = true
			return nil
		},
	}
	uc := New(repo, &mockClientsRepo{}, WithRules(rules))

	err := uc.DeleteRule(context.Background(), 1, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("deleteRule was not called")
	}
}

// ── Tests: Predictions ──────────────────────────────────────────────────────

func TestGetPredictions_NilRepo(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})

	preds, err := uc.GetPredictions(context.Background(), 10, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preds != nil {
		t.Errorf("expected nil, got %v", preds)
	}
}

func TestGetPredictions_Success(t *testing.T) {
	preds := &mockPredictionsRepo{
		getPredictionsFn: func(_ context.Context, _ int, _, _ int) ([]entity.ClientPrediction, error) {
			return []entity.ClientPrediction{
				{ID: 1, ChurnRisk: 0.8},
				{ID: 2, ChurnRisk: 0.3},
			}, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	result, err := uc.GetPredictions(context.Background(), 10, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 predictions, got %d", len(result))
	}
}

func TestGetClientPrediction_NilRepo(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})

	_, err := uc.GetClientPrediction(context.Background(), 999)
	if err != ErrPredictionNotFound {
		t.Errorf("expected ErrPredictionNotFound, got: %v", err)
	}
}

func TestGetClientPrediction_NotFound(t *testing.T) {
	preds := &mockPredictionsRepo{
		getPredictionByClientFn: func(_ context.Context, _ int) (*entity.ClientPrediction, error) {
			return nil, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	_, err := uc.GetClientPrediction(context.Background(), 999)
	if err != ErrPredictionNotFound {
		t.Errorf("expected ErrPredictionNotFound, got: %v", err)
	}
}

func TestGetClientPrediction_Success(t *testing.T) {
	preds := &mockPredictionsRepo{
		getPredictionByClientFn: func(_ context.Context, _ int) (*entity.ClientPrediction, error) {
			return &entity.ClientPrediction{ID: 1, ClientID: 42, ChurnRisk: 0.9}, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	pred, err := uc.GetClientPrediction(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pred.ChurnRisk != 0.9 {
		t.Errorf("expected churn_risk=0.9, got %f", pred.ChurnRisk)
	}
}

func TestGetHighChurnClients_DefaultThreshold(t *testing.T) {
	var receivedThreshold float32
	preds := &mockPredictionsRepo{
		getHighChurnClientsFn: func(_ context.Context, _ int, threshold float32) ([]entity.ClientPrediction, error) {
			receivedThreshold = threshold
			return nil, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	_, err := uc.GetHighChurnClients(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedThreshold != 0.7 {
		t.Errorf("expected threshold=0.7, got %f", receivedThreshold)
	}
}

func TestGetHighChurnClients_CustomThreshold(t *testing.T) {
	var receivedThreshold float32
	preds := &mockPredictionsRepo{
		getHighChurnClientsFn: func(_ context.Context, _ int, threshold float32) ([]entity.ClientPrediction, error) {
			receivedThreshold = threshold
			return nil, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	_, err := uc.GetHighChurnClients(context.Background(), 10, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedThreshold != 0.5 {
		t.Errorf("expected threshold=0.5, got %f", receivedThreshold)
	}
}

func TestGetPredictionSummary_NilRepo(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})

	summary, err := uc.GetPredictionSummary(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
}

func TestGetPredictionSummary_Success(t *testing.T) {
	preds := &mockPredictionsRepo{
		getPredictionSummaryFn: func(_ context.Context, _ int) (*entity.PredictionSummary, error) {
			return &entity.PredictionSummary{
				HighChurnCount:  5,
				AvgChurnRisk:    0.45,
				HighUpsellCount: 3,
				TotalPredicted:  100,
			}, nil
		},
	}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))

	summary, err := uc.GetPredictionSummary(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalPredicted != 100 {
		t.Errorf("expected total=100, got %d", summary.TotalPredicted)
	}
}

func TestComputePredictions_NilRepo(t *testing.T) {
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{})
	uc.Init(context.Background(), testLogger())

	err := uc.ComputePredictions(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for nil predictions repo")
	}
}

func TestComputePredictions_Success(t *testing.T) {
	preds := &mockPredictionsRepo{}
	uc := New(&mockSegmentsRepo{}, &mockClientsRepo{}, WithPredictions(preds))
	uc.Init(context.Background(), testLogger())

	err := uc.ComputePredictions(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
