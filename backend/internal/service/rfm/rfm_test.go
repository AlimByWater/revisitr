package rfm

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── ScoreRecency (template-driven) ──────────────────────────────────────────

func TestScoreRecency_Coffeegng(t *testing.T) {
	th := entity.StandardTemplates["coffeegng"].RThresholds // [3, 7, 14, 30]
	tests := []struct {
		days int
		want int
	}{
		{0, 5}, {3, 5},   // 0-3 → 5
		{4, 4}, {7, 4},   // 4-7 → 4
		{8, 3}, {14, 3},  // 8-14 → 3
		{15, 2}, {30, 2}, // 15-30 → 2
		{31, 1}, {100, 1}, // 31+ → 1
	}
	for _, tt := range tests {
		got := ScoreRecency(tt.days, th)
		if got != tt.want {
			t.Errorf("coffeegng ScoreRecency(%d): want %d, got %d", tt.days, tt.want, got)
		}
	}
}

func TestScoreRecency_TSR(t *testing.T) {
	th := entity.StandardTemplates["tsr"].RThresholds // [10, 21, 45, 90]
	tests := []struct {
		days int
		want int
	}{
		{0, 5}, {10, 5},
		{11, 4}, {21, 4},
		{22, 3}, {45, 3},
		{46, 2}, {90, 2},
		{91, 1}, {365, 1},
	}
	for _, tt := range tests {
		got := ScoreRecency(tt.days, th)
		if got != tt.want {
			t.Errorf("tsr ScoreRecency(%d): want %d, got %d", tt.days, tt.want, got)
		}
	}
}

func TestScoreRecency_AllTemplates(t *testing.T) {
	for key, tmpl := range entity.StandardTemplates {
		// Score=5 at day 0
		if got := ScoreRecency(0, tmpl.RThresholds); got != 5 {
			t.Errorf("%s: day 0 should be score 5, got %d", key, got)
		}
		// Score=1 at very large days
		if got := ScoreRecency(9999, tmpl.RThresholds); got != 1 {
			t.Errorf("%s: day 9999 should be score 1, got %d", key, got)
		}
	}
}

// ── ScoreFrequency (template-driven) ────────────────────────────────────────

func TestScoreFrequency_Coffeegng(t *testing.T) {
	th := entity.StandardTemplates["coffeegng"].FThresholds // [12, 8, 4, 2]
	tests := []struct {
		count int
		want  int
	}{
		{12, 5}, {20, 5}, // 12+ → 5
		{8, 4}, {11, 4},  // 8-11 → 4
		{4, 3}, {7, 3},   // 4-7 → 3
		{2, 2}, {3, 2},   // 2-3 → 2
		{1, 1}, {0, 1},   // 0-1 → 1
	}
	for _, tt := range tests {
		got := ScoreFrequency(tt.count, th)
		if got != tt.want {
			t.Errorf("coffeegng ScoreFrequency(%d): want %d, got %d", tt.count, tt.want, got)
		}
	}
}

func TestScoreFrequency_Bar(t *testing.T) {
	th := entity.StandardTemplates["bar"].FThresholds // [8, 5, 3, 2]
	tests := []struct {
		count int
		want  int
	}{
		{8, 5}, {15, 5},
		{5, 4}, {7, 4},
		{3, 3}, {4, 3},
		{2, 2},
		{1, 1}, {0, 1},
	}
	for _, tt := range tests {
		got := ScoreFrequency(tt.count, th)
		if got != tt.want {
			t.Errorf("bar ScoreFrequency(%d): want %d, got %d", tt.count, tt.want, got)
		}
	}
}

// ── ScoreMonetary (percentile, unchanged) ───────────────────────────────────

func TestScoreMonetary(t *testing.T) {
	sorted := []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
	tests := []struct {
		amount float64
		want   int
	}{
		{50, 1},
		{300, 2},
		{600, 4},
		{1000, 5},
	}
	for _, tt := range tests {
		got := ScoreMonetary(tt.amount, sorted)
		if got != tt.want {
			t.Errorf("ScoreMonetary(%.0f): want %d, got %d", tt.amount, tt.want, got)
		}
	}
}

func TestScoreMonetary_Empty(t *testing.T) {
	if got := ScoreMonetary(500, nil); got != 1 {
		t.Errorf("expected 1 for empty, got %d", got)
	}
}

func TestScoreMonetary_SingleElement(t *testing.T) {
	if got := ScoreMonetary(100, []float64{100}); got != 5 {
		t.Errorf("single element should be top quintile (5), got %d", got)
	}
}

// ── ClassifySegment (7 segments with priority) ──────────────────────────────

func TestClassifySegment(t *testing.T) {
	tests := []struct {
		name        string
		r, f, m     int
		totalVisits int
		want        string
	}{
		// 1. New — first visit, recent
		{"new_fresh", 5, 1, 2, 1, entity.RFMSegmentNew},
		{"new_r4", 4, 1, 1, 1, entity.RFMSegmentNew},
		// New but old visit → not New (r<4)
		{"new_but_old", 2, 1, 1, 1, entity.RFMSegmentChurnRisk},

		// 2. Lost — r=1
		{"lost", 1, 5, 5, 10, entity.RFMSegmentLost},
		{"lost_low", 1, 1, 1, 5, entity.RFMSegmentLost},

		// 3. Churn risk — r=2
		{"churn_risk", 2, 5, 5, 20, entity.RFMSegmentChurnRisk},
		{"churn_risk_low", 2, 1, 1, 3, entity.RFMSegmentChurnRisk},

		// 4. VIP — r≥3, f≥4, m≥4
		{"vip", 5, 5, 5, 18, entity.RFMSegmentVIP},
		{"vip_boundary", 3, 4, 4, 10, entity.RFMSegmentVIP},

		// 5. Regular — r≥3, f≥4, m<4
		{"regular", 5, 5, 3, 15, entity.RFMSegmentRegular},
		{"regular_boundary", 3, 4, 1, 10, entity.RFMSegmentRegular},

		// 6. Rare but valuable — r≥3, f≤2, m≥4
		{"rare_valuable", 4, 2, 5, 5, entity.RFMSegmentRareValue},
		{"rare_valuable_f1", 5, 1, 4, 3, entity.RFMSegmentRareValue},

		// 7. Promising — r≥3, f=2..3
		{"promising", 3, 3, 2, 5, entity.RFMSegmentPromising},
		{"promising_f2", 4, 2, 2, 4, entity.RFMSegmentPromising},

		// Edge: r≥3, f=3, m≥4 → Promising (f_score not ≥4, not ≤2)
		{"promising_high_m", 3, 3, 5, 8, entity.RFMSegmentPromising},

		// Edge: r≥3, f=1, m<4 → Promising (fallback)
		{"fallback_promising", 3, 1, 2, 3, entity.RFMSegmentPromising},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifySegment(tt.r, tt.f, tt.m, tt.totalVisits)
			if got != tt.want {
				t.Errorf("ClassifySegment(r=%d,f=%d,m=%d,visits=%d): want %q, got %q",
					tt.r, tt.f, tt.m, tt.totalVisits, tt.want, got)
			}
		})
	}
}

// ── Design doc examples ─────────────────────────────────────────────────────

func TestDesignDocExamples(t *testing.T) {
	// Example 1: New customer (coffeegng template)
	coffeegng := entity.StandardTemplates["coffeegng"]
	r := ScoreRecency(2, coffeegng.RThresholds)   // 2 days → R=5
	f := ScoreFrequency(1, coffeegng.FThresholds)  // 1 visit → F=1
	seg := ClassifySegment(r, f, 2, 1)             // total_visits=1
	if seg != entity.RFMSegmentNew {
		t.Errorf("example 1: want New, got %s (r=%d,f=%d)", seg, r, f)
	}

	// Example 2: Regular customer (coffeegng template)
	r = ScoreRecency(5, coffeegng.RThresholds)    // 5 days → R=4
	f = ScoreFrequency(15, coffeegng.FThresholds)  // 15 visits → F=5
	seg = ClassifySegment(r, f, 3, 15)
	if seg != entity.RFMSegmentRegular {
		t.Errorf("example 2: want Regular, got %s (r=%d,f=%d)", seg, r, f)
	}

	// Example 3: VIP (qsr template)
	qsr := entity.StandardTemplates["qsr"]
	r = ScoreRecency(1, qsr.RThresholds)    // 1 day → R=5
	f = ScoreFrequency(18, qsr.FThresholds)  // 18 visits → F=5
	seg = ClassifySegment(r, f, 5, 18)
	if seg != entity.RFMSegmentVIP {
		t.Errorf("example 3: want VIP, got %s (r=%d,f=%d)", seg, r, f)
	}

	// Example 4: Churn risk (qsr template)
	r = ScoreRecency(35, qsr.RThresholds) // 35 days → R=2
	seg = ClassifySegment(r, 4, 3, 8)
	if seg != entity.RFMSegmentChurnRisk {
		t.Errorf("example 4: want ChurnRisk, got %s (r=%d)", seg, r)
	}

	// Example 5: Rare but valuable (tsr template)
	tsr := entity.StandardTemplates["tsr"]
	r = ScoreRecency(12, tsr.RThresholds)   // 12 days → R=4
	f = ScoreFrequency(2, tsr.FThresholds)   // 2 visits → F=2
	seg = ClassifySegment(r, f, 5, 5)
	if seg != entity.RFMSegmentRareValue {
		t.Errorf("example 5: want RareValue, got %s (r=%d,f=%d)", seg, r, f)
	}
}

// ── RecalculateAll integration (with mocks) ─────────────────────────────────

type mockClientsRepo struct {
	updates []entity.RFMUpdateParams
}

func (m *mockClientsRepo) UpdateRFMScores(_ context.Context, p entity.RFMUpdateParams) error {
	m.updates = append(m.updates, p)
	return nil
}

type mockTxRepo struct {
	stats []entity.ClientRFMStats
}

func (m *mockTxRepo) GetRFMStats(_ context.Context, _ int) ([]entity.ClientRFMStats, error) {
	return m.stats, nil
}

type mockConfigRepo struct {
	cfg *entity.RFMConfig
}

func (m *mockConfigRepo) GetConfig(_ context.Context, _ int) (*entity.RFMConfig, error) {
	return m.cfg, nil
}

func TestRecalculateAll(t *testing.T) {
	now := time.Now()
	clients := &mockClientsRepo{}
	txs := &mockTxRepo{
		stats: []entity.ClientRFMStats{
			{ClientID: 1, LastVisitAt: now.AddDate(0, 0, -2), FrequencyCount: 15, MonetarySum: 12000, TotalVisitsLifetime: 15},
			{ClientID: 2, LastVisitAt: now.AddDate(0, 0, -50), FrequencyCount: 1, MonetarySum: 500, TotalVisitsLifetime: 1},
		},
	}
	configs := &mockConfigRepo{
		cfg: &entity.RFMConfig{
			ActiveTemplateType: entity.TemplateTypeStandard,
			ActiveTemplateKey:  "coffeegng",
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := New(clients, txs, configs, logger)

	if err := svc.RecalculateAll(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients.updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(clients.updates))
	}

	// Client 1: 2 days ago, 15 visits, 12000 revenue → coffeegng: R=5, F=5
	u1 := clients.updates[0]
	if u1.ClientID != 1 {
		t.Errorf("expected client 1, got %d", u1.ClientID)
	}
	if u1.RScore != 5 {
		t.Errorf("client 1: want r=5, got %d", u1.RScore)
	}
	if u1.FScore != 5 {
		t.Errorf("client 1: want f=5, got %d", u1.FScore)
	}
	// With only 2 clients, monetary percentiles: client1=12000 is top → m=5
	if u1.Segment != entity.RFMSegmentVIP {
		t.Errorf("client 1: want VIP, got %s", u1.Segment)
	}

	// Client 2: 50 days ago, 1 visit, 500 revenue → coffeegng: R=1, F=1
	u2 := clients.updates[1]
	if u2.RScore != 1 {
		t.Errorf("client 2: want r=1, got %d", u2.RScore)
	}
	if u2.Segment != entity.RFMSegmentLost {
		t.Errorf("client 2: want Lost, got %s", u2.Segment)
	}
}

func TestRecalculateAll_EmptyStats(t *testing.T) {
	clients := &mockClientsRepo{}
	txs := &mockTxRepo{stats: nil}
	configs := &mockConfigRepo{cfg: nil}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := New(clients, txs, configs, logger)

	if err := svc.RecalculateAll(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients.updates) != 0 {
		t.Errorf("expected 0 updates for empty stats, got %d", len(clients.updates))
	}
}

func TestRecalculateAll_DefaultTemplate(t *testing.T) {
	now := time.Now()
	clients := &mockClientsRepo{}
	txs := &mockTxRepo{
		stats: []entity.ClientRFMStats{
			{ClientID: 1, LastVisitAt: now, FrequencyCount: 10, MonetarySum: 5000, TotalVisitsLifetime: 10},
		},
	}
	// No config → fallback to tsr
	configs := &mockConfigRepo{cfg: nil}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := New(clients, txs, configs, logger)

	if err := svc.RecalculateAll(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients.updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(clients.updates))
	}

	u := clients.updates[0]
	// tsr thresholds: R=[10,21,45,90], F=[6,4,3,2]
	// 0 days → R=5, 10 visits → F=5, single client → M=5
	if u.RScore != 5 {
		t.Errorf("want r=5, got %d", u.RScore)
	}
	if u.FScore != 5 {
		t.Errorf("want f=5, got %d", u.FScore)
	}
	if u.Segment != entity.RFMSegmentVIP {
		t.Errorf("want VIP, got %s", u.Segment)
	}
}
