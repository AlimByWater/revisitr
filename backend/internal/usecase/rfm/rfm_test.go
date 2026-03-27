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
	getConfigFn          func(ctx context.Context, orgID int) (*entity.RFMConfig, error)
	upsertConfigFn       func(ctx context.Context, cfg *entity.RFMConfig) error
	updateCalcStatsFn    func(ctx context.Context, orgID, clientsProcessed int) error
	insertHistoryFn      func(ctx context.Context, h *entity.RFMHistory) error
	getHistoryFn         func(ctx context.Context, orgID int, from, to time.Time) ([]entity.RFMHistory, error)
	getSegmentSummaryFn  func(ctx context.Context, orgID int) ([]entity.RFMSegmentSummary, error)
	getSegmentClientsFn  func(ctx context.Context, orgID int, segment string, sortCol, order string, limit, offset int) ([]entity.SegmentClientRow, int, error)
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
func (m *mockRFMRepo) GetSegmentClients(ctx context.Context, orgID int, segment string, sortCol, order string, limit, offset int) ([]entity.SegmentClientRow, int, error) {
	if m.getSegmentClientsFn != nil {
		return m.getSegmentClientsFn(ctx, orgID, segment, sortCol, order, limit, offset)
	}
	return nil, 0, nil
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
		{Segment: entity.RFMSegmentVIP, ClientCount: 10, Percentage: 25.0},
		{Segment: entity.RFMSegmentLost, ClientCount: 30, Percentage: 75.0},
	}
	trends := []entity.RFMHistory{
		{OrgID: 1, Segment: entity.RFMSegmentVIP, ClientCount: 8},
		{OrgID: 1, Segment: entity.RFMSegmentLost, ClientCount: 25},
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
	// All 7 v2 RFM segments already exist
	segNew := entity.RFMSegmentNew
	segPromising := entity.RFMSegmentPromising
	segRegular := entity.RFMSegmentRegular
	segVIP := entity.RFMSegmentVIP
	segRareValue := entity.RFMSegmentRareValue
	segChurnRisk := entity.RFMSegmentChurnRisk
	segLost := entity.RFMSegmentLost

	existing := []entity.Segment{
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segNew}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segPromising}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segRegular}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segVIP}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segRareValue}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segChurnRisk}},
		{OrgID: 1, Type: "rfm", Filter: entity.SegmentFilter{RFMCategory: &segLost}},
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

func TestListTemplates(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})
	templates := uc.ListTemplates()
	if len(templates) != 4 {
		t.Fatalf("expected 4 templates, got %d", len(templates))
	}
	keys := map[string]bool{}
	for _, tpl := range templates {
		keys[tpl.Key] = true
	}
	for _, k := range []string{"coffeegng", "qsr", "tsr", "bar"} {
		if !keys[k] {
			t.Errorf("missing template key: %s", k)
		}
	}
}

func TestGetActiveTemplate_Default(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	cfg, tpl, err := uc.GetActiveTemplate(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ActiveTemplateType != "standard" {
		t.Errorf("expected standard, got %s", cfg.ActiveTemplateType)
	}
	if cfg.ActiveTemplateKey != "tsr" {
		t.Errorf("expected tsr, got %s", cfg.ActiveTemplateKey)
	}
	if tpl.Key != "tsr" {
		t.Errorf("expected tsr template, got %s", tpl.Key)
	}
}

func TestGetActiveTemplate_Configured(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return &entity.RFMConfig{
				OrgID:              1,
				ActiveTemplateType: "standard",
				ActiveTemplateKey:  "bar",
			}, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	_, tpl, err := uc.GetActiveTemplate(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Key != "bar" {
		t.Errorf("expected bar template, got %s", tpl.Key)
	}
}

func TestSetTemplate_Standard(t *testing.T) {
	var upserted *entity.RFMConfig
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil
		},
		upsertConfigFn: func(_ context.Context, cfg *entity.RFMConfig) error {
			upserted = cfg
			return nil
		},
	}
	segRepo := &mockSegmentsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Segment, error) {
			return nil, nil
		},
	}
	rfmSvc := &mockRFMService{
		recalculateAllFn: func(_ context.Context, _ int) error { return nil },
	}
	uc := newTestUC(rfmRepo, segRepo, rfmSvc)

	tpl, err := uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "standard",
		TemplateKey:  "coffeegng",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Key != "coffeegng" {
		t.Errorf("expected coffeegng, got %s", tpl.Key)
	}
	if upserted == nil {
		t.Fatal("expected UpsertConfig to be called")
	}
	if upserted.ActiveTemplateType != "standard" {
		t.Errorf("expected standard, got %s", upserted.ActiveTemplateType)
	}
	if upserted.ActiveTemplateKey != "coffeegng" {
		t.Errorf("expected coffeegng, got %s", upserted.ActiveTemplateKey)
	}
}

func TestSetTemplate_Custom(t *testing.T) {
	var upserted *entity.RFMConfig
	rfmRepo := &mockRFMRepo{
		getConfigFn: func(_ context.Context, _ int) (*entity.RFMConfig, error) {
			return nil, nil
		},
		upsertConfigFn: func(_ context.Context, cfg *entity.RFMConfig) error {
			upserted = cfg
			return nil
		},
	}
	segRepo := &mockSegmentsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Segment, error) {
			return nil, nil
		},
	}
	rfmSvc := &mockRFMService{
		recalculateAllFn: func(_ context.Context, _ int) error { return nil },
	}
	uc := newTestUC(rfmRepo, segRepo, rfmSvc)

	name := "Мой формат"
	rTh := [4]int{5, 14, 30, 60}
	fTh := [4]int{10, 6, 3, 2}
	tpl, err := uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "custom",
		CustomName:   &name,
		RThresholds:  &rTh,
		FThresholds:  &fTh,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Key != "custom" {
		t.Errorf("expected custom, got %s", tpl.Key)
	}
	if upserted == nil {
		t.Fatal("expected UpsertConfig to be called")
	}
	if upserted.ActiveTemplateType != "custom" {
		t.Errorf("expected custom, got %s", upserted.ActiveTemplateType)
	}
}

func TestSetTemplate_InvalidStandardKey(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	_, err := uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "standard",
		TemplateKey:  "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for invalid template key")
	}
}

func TestSetTemplate_InvalidCustomThresholds(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	// Non-ascending R thresholds
	rTh := [4]int{30, 14, 5, 60}
	fTh := [4]int{10, 6, 3, 2}
	_, err := uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "custom",
		RThresholds:  &rTh,
		FThresholds:  &fTh,
	})
	if err == nil {
		t.Fatal("expected error for invalid r_thresholds")
	}

	// Non-descending F thresholds
	rTh2 := [4]int{5, 14, 30, 60}
	fTh2 := [4]int{2, 3, 6, 10}
	_, err = uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "custom",
		RThresholds:  &rTh2,
		FThresholds:  &fTh2,
	})
	if err == nil {
		t.Fatal("expected error for invalid f_thresholds")
	}

	// Missing thresholds
	_, err = uc.SetTemplate(context.Background(), 1, entity.SetTemplateRequest{
		TemplateType: "custom",
	})
	if err == nil {
		t.Fatal("expected error for missing thresholds")
	}
}

func TestRecommendTemplate_AllSame(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	result, err := uc.RecommendTemplate([]int{1, 1, 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Recommended.Key != "coffeegng" {
		t.Errorf("expected coffeegng, got %s", result.Recommended.Key)
	}
	if result.AllScores["coffeegng"] != 3 {
		t.Errorf("expected score 3, got %d", result.AllScores["coffeegng"])
	}
	if result.Alternative != nil {
		t.Error("expected no alternative when all answers are same")
	}
}

func TestRecommendTemplate_TieBreak(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	// Q1=bar, Q2=coffeegng, Q3=tsr → all score 1, tie-break by Q1 → bar
	result, err := uc.RecommendTemplate([]int{4, 1, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Recommended.Key != "bar" {
		t.Errorf("expected bar (tie-break by Q1), got %s", result.Recommended.Key)
	}
}

func TestRecommendTemplate_InvalidAnswerCount(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	_, err := uc.RecommendTemplate([]int{1, 2})
	if err == nil {
		t.Fatal("expected error for 2 answers")
	}
}

func TestRecommendTemplate_InvalidAnswerID(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	_, err := uc.RecommendTemplate([]int{1, 5, 2})
	if err == nil {
		t.Fatal("expected error for invalid answer ID 5")
	}
}

func TestOnboardingQuestions(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})
	questions := uc.GetOnboardingQuestions()
	if len(questions) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(questions))
	}
	for _, q := range questions {
		if len(q.Answers) != 4 {
			t.Errorf("question %d: expected 4 answers, got %d", q.ID, len(q.Answers))
		}
	}
}

func TestGetSegmentClients(t *testing.T) {
	score5 := 5
	score4 := 4
	freq := 15
	money := 25000.0

	clients := []entity.SegmentClientRow{
		{ID: 1, FirstName: "Иван", LastName: "Петров", RScore: &score5, FScore: &score5, MScore: &score5, FrequencyCount: &freq, MonetarySum: &money, TotalVisitsLifetime: 45},
		{ID: 2, FirstName: "Мария", LastName: "Иванова", RScore: &score4, FScore: &score5, MScore: &score4, FrequencyCount: &freq, MonetarySum: &money, TotalVisitsLifetime: 30},
	}

	rfmRepo := &mockRFMRepo{
		getSegmentClientsFn: func(_ context.Context, _ int, seg string, _ string, _ string, limit, offset int) ([]entity.SegmentClientRow, int, error) {
			if seg != "vip" {
				t.Errorf("expected segment 'vip', got '%s'", seg)
			}
			if limit != 20 {
				t.Errorf("expected limit 20, got %d", limit)
			}
			if offset != 0 {
				t.Errorf("expected offset 0, got %d", offset)
			}
			return clients, 42, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	result, err := uc.GetSegmentClients(context.Background(), 1, "vip", 1, 20, "monetary_sum", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Segment != "vip" {
		t.Errorf("expected segment 'vip', got '%s'", result.Segment)
	}
	if result.SegmentName != "VIP / Ядро" {
		t.Errorf("expected segment name 'VIP / Ядро', got '%s'", result.SegmentName)
	}
	if result.Total != 42 {
		t.Errorf("expected total 42, got %d", result.Total)
	}
	if len(result.Clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(result.Clients))
	}
	if result.Page != 1 || result.PerPage != 20 {
		t.Errorf("expected page=1/per_page=20, got %d/%d", result.Page, result.PerPage)
	}
}

func TestGetSegmentClients_InvalidSegment(t *testing.T) {
	uc := newTestUC(&mockRFMRepo{}, &mockSegmentsRepo{}, &mockRFMService{})

	_, err := uc.GetSegmentClients(context.Background(), 1, "nonexistent", 1, 20, "monetary_sum", "desc")
	if err == nil {
		t.Fatal("expected error for invalid segment")
	}
}

func TestGetSegmentClients_PaginationClamping(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getSegmentClientsFn: func(_ context.Context, _ int, _ string, _ string, _ string, limit, offset int) ([]entity.SegmentClientRow, int, error) {
			if limit != 100 {
				t.Errorf("expected clamped limit 100, got %d", limit)
			}
			if offset != 200 {
				t.Errorf("expected offset 200, got %d", offset)
			}
			return nil, 0, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	// per_page=999 should be clamped to 100, page=3 → offset=(3-1)*100=200
	result, err := uc.GetSegmentClients(context.Background(), 1, "vip", 3, 999, "monetary_sum", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PerPage != 100 {
		t.Errorf("expected per_page clamped to 100, got %d", result.PerPage)
	}
}

func TestGetSegmentClients_Page0Clamped(t *testing.T) {
	rfmRepo := &mockRFMRepo{
		getSegmentClientsFn: func(_ context.Context, _ int, _ string, _ string, _ string, _, offset int) ([]entity.SegmentClientRow, int, error) {
			if offset != 0 {
				t.Errorf("expected offset 0 for page=0 clamped to 1, got %d", offset)
			}
			return nil, 0, nil
		},
	}
	uc := newTestUC(rfmRepo, &mockSegmentsRepo{}, &mockRFMService{})

	result, err := uc.GetSegmentClients(context.Background(), 1, "lost", 0, 20, "monetary_sum", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page clamped to 1, got %d", result.Page)
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
