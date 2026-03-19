package rfm

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"revisitr/internal/entity"
)

type clientsRepo interface {
	GetAllByOrgID(ctx context.Context, orgID int) ([]entity.BotClient, error)
	UpdateRFM(ctx context.Context, clientID, recency, frequency int, monetary float64, segment string) error
}

type txRepo interface {
	GetTxStatsPerClient(ctx context.Context, orgID int) ([]entity.ClientTxStats, error)
}

type Service struct {
	clients clientsRepo
	txs     txRepo
	logger  *slog.Logger
}

func New(clients clientsRepo, txs txRepo, logger *slog.Logger) *Service {
	return &Service{clients: clients, txs: txs, logger: logger}
}

// RecalculateAll recalculates RFM scores for all clients in the org.
func (s *Service) RecalculateAll(ctx context.Context, orgID int) error {
	stats, err := s.txs.GetTxStatsPerClient(ctx, orgID)
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return nil
	}

	// Collect monetary amounts for percentile scoring
	amounts := make([]float64, 0, len(stats))
	for _, st := range stats {
		amounts = append(amounts, st.TotalAmount)
	}
	sort.Float64s(amounts)

	now := time.Now()

	for _, st := range stats {
		r := scoreRecency(now, st.LastTxAt)
		f := scoreFrequency(st.TxCount)
		m := scoreMonetary(st.TotalAmount, amounts)
		seg := classifySegment(r + f + m)

		if err := s.clients.UpdateRFM(ctx, st.ClientID, r, f, float64(m), seg); err != nil {
			s.logger.Error("rfm: update client", "client_id", st.ClientID, "error", err)
		}
	}

	return nil
}

// scoreRecency: 5=<7d, 4=<30d, 3=<90d, 2=<180d, 1=else
func scoreRecency(now time.Time, lastTx time.Time) int {
	days := now.Sub(lastTx).Hours() / 24
	switch {
	case days < 7:
		return 5
	case days < 30:
		return 4
	case days < 90:
		return 3
	case days < 180:
		return 2
	default:
		return 1
	}
}

// scoreFrequency: 5=10+, 4=7+, 3=4+, 2=2+, 1=else
func scoreFrequency(count int) int {
	switch {
	case count >= 10:
		return 5
	case count >= 7:
		return 4
	case count >= 4:
		return 3
	case count >= 2:
		return 2
	default:
		return 1
	}
}

// scoreMonetary: 5=top20%, 4=top40%, 3=top60%, 2=top80%, 1=else
// sorted is the sorted slice of all monetary values (ascending).
func scoreMonetary(amount float64, sorted []float64) int {
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

// classifySegment maps total RFM score to a named category.
// champions(≥13), loyal(10–12), at_risk(7–9), cant_lose(5–6), hibernating(3–4), lost(<3)
func classifySegment(total int) string {
	switch {
	case total >= 13:
		return entity.RFMChampions
	case total >= 10:
		return entity.RFMLoyal
	case total >= 7:
		return entity.RFMAtRisk
	case total >= 5:
		return entity.RFMCantLose
	case total >= 3:
		return entity.RFMHibernating
	default:
		return entity.RFMLost
	}
}
