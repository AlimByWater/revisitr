package rfm

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockRFMRepo struct {
	getConfigFn        func(ctx context.Context, orgID int) (*entity.RFMConfig, error)
	upsertConfigFn     func(ctx context.Context, cfg *entity.RFMConfig) error
	updateCalcStatsFn  func(ctx context.Context, orgID, clientsProcessed int) error
	insertHistoryFn    func(ctx context.Context, h *entity.RFMHistory) error
	getHistoryFn       func(ctx context.Context, orgID int, from, to time.Time) ([]entity.RFMHistory, error)
	getSegmentSummaryFn func(ctx context.Context, orgID int) ([]entity.RFMSegmentSummary, error)
}

func (m *mockRFMRepo) GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error) {
	if m.getConfigFn != nil {
		return m.getConfigFn(ctx, orgID)
	}
	return nil, nil
}
func (m *mockRFMRepo) UpsertConfig(ctx context.Context, cfg *entity.RFMConfig) error {
	if m.upsertConfigFn != nil {
		return m.upsertConfigFn(ctx, cfg)
	}
	return nil
}
func (m *mockRFMRepo) UpdateCalcStats(ctx context.Context, orgID, clientsProcessed int) error {
	if m.updateCalcStatsFn != nil {
		return m.updateCalcStatsFn(ctx, orgID, clientsProcessed)
	}
	return nil
}
func (m *mockRFMRepo) InsertHistory(ctx context.Context, h *entity.RFMHistory) error {
	if m.insertHistoryFn != nil {
		return m.insertHistoryFn(ctx, h)
	}
	return nil
}
func (m *mockRFMRepo) GetHistory(ctx context.Context, orgID int, from, to time.Time) ([]entity.RFMHistory, error) {
	if m.getHistoryFn != nil {
		return m.getHistoryFn(ctx, orgID, from, to)
	}
	return nil, nil
}
func (m *mockRFMRepo) GetSegmentSummary(ctx context.Context, orgID int) ([]entity.RFMSegmentSummary, error) {
	if m.getSegmentSummaryFn != nil {
		return m.getSegmentSummaryFn(ctx, orgID)
	}
	return nil, nil
}

type mockSegmentsRepo struct {
	getByOrgIDFn func(ctx context.Context, orgID int) ([]entity.Segment, error)
	createFn     func(ctx context.Context, seg *entity.Segment) error
}

func (m *mockSegmentsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error) {
	if m.getByOrgIDFn != nil {
		return m.getByOrgIDFn(ctx, orgID)
	}
	return nil, nil
}
func (m *mockSegmentsRepo) Create(ctx context.Context, seg *entity.Segment) error {
	if m.createFn != nil {
		return m.createFn(ctx, seg)
	}
	return nil
}

type mockRFMService struct {
	recalculateAllFn func(ctx context.Context, orgID int) error
}

func (m *mockRFMService) RecalculateAll(ctx context.Context, orgID int) error {
	if m.recalculateAllFn != nil {
		return m.recalculateAllFn(ctx, orgID)
	}
	return nil
}

func newTestUC(rfmRepo *mockRFMRepo, segRepo *mockSegmentsRepo, rfmSvc *mockRFMService) *Usecase {
	uc := New(rfmRepo, segRepo, rfmSvc)
	uc.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return uc
}

// --- tests ---

func TestGetDashboard_Empty(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getSegmentSummaryFn: func(_ context.Context, _ int) ([]entity.RFMSegmentSummary, error) {
			return []entity.RFMSegmentSummary{}, nil
		},
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil
		},
		getHistoryFn: func(_ context.Context, _ int, _, _ time.Time) ([]entity.RFMHistory, error) {
			return []entity.RFMHistory{}, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	dash, err := uc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dash.Segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(dash.Segments))
	}
	if len(dash.Trends) != 0 {
		t.Errorf("expected 0 trends, got %d", len(dash.Trends))
	}
}

func TestGetDashboard_WithData(t *testing.T) {
	segments := []entity.RFMSegmentSummary{
		{Segment: entity.RFMChampions, ClientCount: 10, Percentage: 25.0},
		{Segment: entity.RFMLost, ClientCount: 30, Percentage: 75.0},
	}
	trends := []entity.RFMHistory{
		{OrgID: 1, Segment: entity.RFMChampions, ClientCount: 8},
		{OrgID: 1, Segment: entity.RFMLost, ClientCount: 25},
	}
	cfg := &entity.RFMConfig{OrgID: 1, PeriodDays: 365, RecalcInterval: "24h"}

	rfmRepo := &mockRFMRepo{
		getSegmentSummaryFn: func(_ context.Context, _ int) ([]entity.RFMSegmentSummary, error) {
			return segments, nil
		},
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return cfg, nil
		},
		getHistoryFn: func(_ context.Context, _ int, _, _ time.Time) ([]entity.RFMHistory, error) {
			return trends, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	dash, err := uc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dash.Segments) != 2 {
		t.Errorf("expected 2 segments, got %d", len(dash.Segments))
	}
	if len(dash.Trends) != 2 {
		t.Errorf("expected 2 trends, got %d", len(dash.Trends))
	}
	if dash.Config == nil {
		t.Error("expected config to be present")
	}
	if dash.Config.PeriodDays != 365 {
		t.Errorf("expected period_days=365, got %d", dash.Config.PeriodDays)
	}
}

func TestRecalculate_CreatesSegments(t *testing.T) {
	var createdSegments []entity.Segment
	segRepo := &mockSegmentsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Segment, error) {
			return []entity.Segment{}, nil // no existing segments
		},
		createFn: func(_ context.Context, seg *entity.Segment) error {
			createdSegments = append(createdSegments, *seg)
			return nil
		},
	}
	recalcCalled := false
	rfmSvc := &mockRFMService{
		recalculateAllFn: func(_ context.Context, _ int) error {
			recalcCalled = true
			return nil
		},
	}
	rfmRepo := &mockRFMRepo{
		getSegmentSummaryFn: func(_ context.Context, _ int) ([]entity.RFMSegmentSummary, error) {
			return []entity.RFMSegmentSummary{}, nil
		},
	}
	uc := newTestUC(rfmRepo, segRepo, rfmSvc)

	if err := uc.Recalculate(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !recalcCalled {
		t.Error("expected rfmSvc.RecalculateAll to be called")
	}
	if len(createdSegments) != 7 {
		t.Errorf("expected 7 standard segments created, got %d", len(createdSegments))
	}
	for _, seg := range createdSegments {
		if seg.Type != "rfm" {
			t.Errorf("expected segment type 'rfm', got '%s'", seg.Type)
		}
		if !seg.AutoAssign {
			t.Errorf("expected auto_assign=true for segment '%s'", seg.Name)
		}
		if seg.OrgID != 1 {
			t.Errorf("expected org_id=1, got %d", seg.OrgID)
		}
	}
}

func TestRecalculate_ExistingSegments(t *testing.T) {
	// All 7 RFM segments already exist
	champions := entity.RFMChampions
	loyal := entity.RFMLoyal
	potLoyalist := "potential_loyalist"
	newCust := "new_customers"
	atRisk := entity.RFMAtRisk
	cantLose := entity.RFMCantLose
	lost := entity.RFMLost

	existing := []entity.Segment{
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &champions}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &loyal}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &potLoyalist}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &newCust}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &atRisk}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &cantLose}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &lost}},
	}

	createCalled := 0
	segRepo := &mockSegmentsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Segment, error) {
			return existing, nil
		},
		createFn: func(_ context.Context, _ *entity.Segment) error {
			createCalled++
			return nil
		},
	}
	rfmSvc := &mockRFMService{
		recalculateAllFn: func(_ context.Context, _ int) error { return nil },
	}
	rfmRepo := &mockRFMRepo{
		getSegmentSummaryFn: func(_ context.Context, _ int) ([]entity.RFMSegmentSummary, error) {
			return []entity.RFMSegmentSummary{}, nil
		},
	}
	uc := newTestUC(rfmRepo, segRepo, rfmSvc)

	if err := uc.Recalculate(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createCalled != 0 {
		t.Errorf("expected 0 segments created (all exist), got %d", createCalled)
	}
}

func TestGetConfig(t *testing.T) {
	cfg := &entity.RFMConfig{
		OrgID:          1,
		PeriodDays:     180,
		RecalcInterval: "12h",
	}
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return cfg, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	result, err := uc.GetConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PeriodDays != 180 {
		t.Errorf("expected period_days=180, got %d", result.PeriodDays)
	}
	if result.RecalcInterval != "12h" {
		t.Errorf("expected recalc_interval='12h', got '%s'", result.RecalcInterval)
	}
}

func TestGetConfig_ReturnsDefaults(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	result, err := uc.GetConfig(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OrgID != 5 {
		t.Errorf("expected org_id=5, got %d", result.OrgID)
	}
	if result.PeriodDays != 365 {
		t.Errorf("expected default period_days=365, got %d", result.PeriodDays)
	}
	if result.RecalcInterval != "24h" {
		t.Errorf("expected default recalc_interval='24h', got '%s'", result.RecalcInterval)
	}
}

func TestUpdateConfig(t *testing.T) {
	var upserted *entity.RFMConfig
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil // no existing config
		},
		upsertConfigFn: func(_ context.Context, cfg *entity.RFMConfig) error {
			upserted = cfg
			return nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	days := 90
	interval := "6h"
	result, err := uc.UpdateConfig(context.Background(), 1, entity.UpdateRFMConfigRequest{
		PeriodDays:     &days,
		RecalcInterval: &interval,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PeriodDays != 90 {
		t.Errorf("expected period_days=90, got %d", result.PeriodDays)
	}
	if result.RecalcInterval != "6h" {
		t.Errorf("expected recalc_interval='6h', got '%s'", result.RecalcInterval)
	}
	if upserted == nil {
		t.Fatal("expected UpsertConfig to be called")
	}
}
