package rfm

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"revisitr/internal/entity"
)

type clientsRepo interface {
	UpdateRFMScores(ctx context.Context, p entity.RFMUpdateParams) error
}

type txRepo interface {
	GetRFMStats(ctx context.Context, orgID int) ([]entity.ClientRFMStats, error)
}

type configRepo interface {
	GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error)
}

type Service struct {
	clients clientsRepo
	txs     txRepo
	configs configRepo
	logger  *slog.Logger
}

func New(clients clientsRepo, txs txRepo, configs configRepo, logger *slog.Logger) *Service {
	return &Service{clients: clients, txs: txs, configs: configs, logger: logger}
}

// RecalculateAll recalculates RFM scores for all clients in the org
// using the org's active template thresholds.
func (s *Service) RecalculateAll(ctx context.Context, orgID int) error {
	// 1. Load template
	tmpl, err := s.resolveTemplate(ctx, orgID)
	if err != nil {
		return err
	}

	// 2. Load client stats (F=90d window, M=180d window)
	stats, err := s.txs.GetRFMStats(ctx, orgID)
	if err != nil {
		return err
	}
	if len(stats) == 0 {
		return nil
	}

	// 3. Collect monetary amounts for percentile scoring
	amounts := make([]float64, 0, len(stats))
	for _, st := range stats {
		amounts = append(amounts, st.MonetarySum)
	}
	sort.Float64s(amounts)

	now := time.Now()

	// 4. Score each client
	for _, st := range stats {
		recencyDays := int(now.Sub(st.LastVisitAt).Hours() / 24)
		r := ScoreRecency(recencyDays, tmpl.RThresholds)
		f := ScoreFrequency(st.FrequencyCount, tmpl.FThresholds)
		m := ScoreMonetary(st.MonetarySum, amounts)
		seg := ClassifySegment(r, f, m, st.TotalVisitsLifetime)

		p := entity.RFMUpdateParams{
			ClientID:        st.ClientID,
			RScore:          r,
			FScore:          f,
			MScore:          m,
			RecencyDays:     recencyDays,
			FrequencyCount:  st.FrequencyCount,
			MonetarySum:     st.MonetarySum,
			TotalVisitsLife: st.TotalVisitsLifetime,
			LastVisitDate:   st.LastVisitAt,
			Segment:         seg,
		}

		if err := s.clients.UpdateRFMScores(ctx, p); err != nil {
			s.logger.Error("rfm: update client", "client_id", st.ClientID, "error", err)
		}
	}

	return nil
}

// resolveTemplate loads the active template for the org, falling back to "tsr" default.
func (s *Service) resolveTemplate(ctx context.Context, orgID int) (entity.RFMTemplate, error) {
	cfg, err := s.configs.GetConfig(ctx, orgID)
	if err != nil {
		return entity.RFMTemplate{}, err
	}

	if cfg != nil {
		if tmpl, ok := cfg.ActiveTemplate(); ok {
			return tmpl, nil
		}
	}

	// Fallback: default tsr template
	return entity.StandardTemplates["tsr"], nil
}

// ── Scoring functions (exported for testing) ────────────────────────────────

// ScoreRecency scores recency days against template thresholds.
// RThresholds: [R5_max, R4_max, R3_max, R2_max]
func ScoreRecency(days int, th [4]int) int {
	switch {
	case days <= th[0]:
		return 5
	case days <= th[1]:
		return 4
	case days <= th[2]:
		return 3
	case days <= th[3]:
		return 2
	default:
		return 1
	}
}

// ScoreFrequency scores visit count against template thresholds.
// FThresholds: [F5_min, F4_min, F3_min, F2_min]
func ScoreFrequency(count int, th [4]int) int {
	switch {
	case count >= th[0]:
		return 5
	case count >= th[1]:
		return 4
	case count >= th[2]:
		return 3
	case count >= th[3]:
		return 2
	default:
		return 1
	}
}

// ScoreMonetary scores monetary amount by quintile position.
// sorted must be ascending. Unchanged from v1 — always relative percentile.
func ScoreMonetary(amount float64, sorted []float64) int {
	n := len(sorted)
	if n == 0 {
		return 1
	}

	p80 := sorted[int(float64(n-1)*0.2)]
	p60 := sorted[int(float64(n-1)*0.4)]
	p40 := sorted[int(float64(n-1)*0.6)]
	p20 := sorted[int(float64(n-1)*0.8)]

	switch {
	case amount >= p20:
		return 5
	case amount >= p40:
		return 4
	case amount >= p60:
		return 3
	case amount >= p80:
		return 2
	default:
		return 1
	}
}

// ClassifySegment assigns one of 7 segments based on R/F/M scores.
// Priority order (first match wins):
//  1. New         — first visit, still fresh
//  2. Lost        — not seen in a long time
//  3. Churn Risk  — cooling off
//  4. VIP         — frequent, high-value, recent
//  5. Regular     — frequent and recent
//  6. Rare Value  — infrequent but high-value
//  7. Promising   — returning, building habit
func ClassifySegment(r, f, m, totalVisits int) string {
	// 1. New: first visit and still fresh
	if totalVisits == 1 && r >= 4 {
		return entity.RFMSegmentNew
	}
	// 2. Lost
	if r == 1 {
		return entity.RFMSegmentLost
	}
	// 3. Churn risk
	if r == 2 {
		return entity.RFMSegmentChurnRisk
	}
	// 4. VIP / Core
	if r >= 3 && f >= 4 && m >= 4 {
		return entity.RFMSegmentVIP
	}
	// 5. Regular
	if r >= 3 && f >= 4 {
		return entity.RFMSegmentRegular
	}
	// 6. Rare but valuable
	if r >= 3 && f <= 2 && m >= 4 {
		return entity.RFMSegmentRareValue
	}
	// 7. Promising (fallback for r >= 3)
	return entity.RFMSegmentPromising
}
