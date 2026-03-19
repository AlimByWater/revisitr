package rfm

import (
	"testing"
	"time"
)

// ── scoreRecency ──────────────────────────────────────────────────────────────

func TestScoreRecency(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		lastTx  time.Time
		want    int
	}{
		{"<7d",   now.AddDate(0, 0, -3),  5},
		{"7d",    now.AddDate(0, 0, -7),  4},
		{"<30d",  now.AddDate(0, 0, -15), 4},
		{"30d",   now.AddDate(0, 0, -30), 3},
		{"<90d",  now.AddDate(0, 0, -45), 3},
		{"90d",   now.AddDate(0, 0, -90), 2},
		{"<180d", now.AddDate(0, 0, -100),2},
		{"180d+", now.AddDate(0, 0, -200),1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreRecency(now, tt.lastTx)
			if got != tt.want {
				t.Errorf("scoreRecency(%s): want %d, got %d", tt.name, tt.want, got)
			}
		})
	}
}

// ── scoreFrequency ────────────────────────────────────────────────────────────

func TestScoreFrequency(t *testing.T) {
	tests := []struct {
		count int
		want  int
	}{
		{0, 1}, {1, 1}, {2, 2}, {3, 2},
		{4, 3}, {6, 3}, {7, 4}, {9, 4},
		{10, 5}, {100, 5},
	}
	for _, tt := range tests {
		got := scoreFrequency(tt.count)
		if got != tt.want {
			t.Errorf("scoreFrequency(%d): want %d, got %d", tt.count, tt.want, got)
		}
	}
}

// ── scoreMonetary ─────────────────────────────────────────────────────────────

func TestScoreMonetary(t *testing.T) {
	sorted := []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}

	tests := []struct {
		amount float64
		want   int
	}{
		{50, 1},   // below p80 threshold
		{300, 2},  // around p60–80
		{600, 4},  // around p20–40
		{1000, 5}, // top 20%
	}

	for _, tt := range tests {
		got := scoreMonetary(tt.amount, sorted)
		if got != tt.want {
			t.Errorf("scoreMonetary(%.0f): want %d, got %d", tt.amount, tt.want, got)
		}
	}
}

func TestScoreMonetary_Empty(t *testing.T) {
	got := scoreMonetary(500, []float64{})
	if got != 1 {
		t.Errorf("expected 1 for empty sorted slice, got %d", got)
	}
}

// ── classifySegment ───────────────────────────────────────────────────────────

func TestClassifySegment(t *testing.T) {
	tests := []struct {
		total int
		want  string
	}{
		{15, "champions"},
		{13, "champions"},
		{12, "loyal"},
		{10, "loyal"},
		{9, "at_risk"},
		{7, "at_risk"},
		{6, "cant_lose"},
		{5, "cant_lose"},
		{4, "hibernating"},
		{3, "hibernating"},
		{2, "lost"},
		{0, "lost"},
	}

	for _, tt := range tests {
		got := classifySegment(tt.total)
		if got != tt.want {
			t.Errorf("classifySegment(%d): want %q, got %q", tt.total, tt.want, got)
		}
	}
}

// ── scheduler unit test ───────────────────────────────────────────────────────

func TestScoreRecency_BoundaryExact(t *testing.T) {
	now := time.Now()
	// Exactly 7 days ago falls into the <30d bucket (score=4)
	exactly7d := now.AddDate(0, 0, -7)
	got := scoreRecency(now, exactly7d)
	if got != 4 {
		t.Errorf("exactly 7 days: want 4, got %d", got)
	}
}
