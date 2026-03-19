package postgres

import (
	"context"
	"fmt"

	"revisitr/internal/entity"
)

type Analytics struct {
	pg *Module
}

func NewAnalytics(pg *Module) *Analytics {
	return &Analytics{pg: pg}
}

func (r *Analytics) GetSalesMetrics(ctx context.Context, f entity.AnalyticsFilter) (*entity.SalesMetrics, error) {
	query := `
		SELECT
			COUNT(*)                          AS transaction_count,
			COUNT(DISTINCT lt.client_id)      AS unique_clients,
			COALESCE(SUM(lt.amount), 0)       AS total_amount,
			COALESCE(AVG(lt.amount), 0)       AS avg_amount,
			CASE WHEN COUNT(DISTINCT lt.client_id) > 0
				THEN COUNT(*)::FLOAT / COUNT(DISTINCT lt.client_id)
				ELSE 0
			END AS buy_frequency
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE lt.type = 'earn'
		  AND b.org_id = $1
		  AND lt.created_at >= $2
		  AND lt.created_at <= $3`

	args := []interface{}{f.OrgID, f.From, f.To}
	idx := 4

	if f.BotID != nil {
		query += fmt.Sprintf(" AND bc.bot_id = $%d", idx)
		args = append(args, *f.BotID)
		idx++
	}

	var m entity.SalesMetrics
	if err := r.pg.DB().GetContext(ctx, &m, query, args...); err != nil {
		return nil, fmt.Errorf("analytics.GetSalesMetrics: %w", err)
	}

	return &m, nil
}

func (r *Analytics) GetSalesCharts(ctx context.Context, f entity.AnalyticsFilter) (map[string][]entity.SalesChartPoint, error) {
	query := `
		SELECT
			DATE_TRUNC('day', lt.created_at) AS day,
			COUNT(*)                          AS value
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE lt.type = 'earn'
		  AND b.org_id = $1
		  AND lt.created_at >= $2
		  AND lt.created_at <= $3`

	args := []interface{}{f.OrgID, f.From, f.To}
	idx := 4

	if f.BotID != nil {
		query += fmt.Sprintf(" AND bc.bot_id = $%d", idx)
		args = append(args, *f.BotID)
		idx++
	}

	_ = idx
	query += " GROUP BY 1 ORDER BY 1"

	var txPoints []entity.SalesChartPoint
	if err := r.pg.DB().SelectContext(ctx, &txPoints, query, args...); err != nil {
		return nil, fmt.Errorf("analytics.GetSalesCharts transactions: %w", err)
	}

	revenueQuery := `
		SELECT
			DATE_TRUNC('day', lt.created_at) AS day,
			COALESCE(SUM(lt.amount), 0)       AS value
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE lt.type = 'earn'
		  AND b.org_id = $1
		  AND lt.created_at >= $2
		  AND lt.created_at <= $3`

	if f.BotID != nil {
		revenueQuery += fmt.Sprintf(" AND bc.bot_id = $%d", 4)
	}
	revenueQuery += " GROUP BY 1 ORDER BY 1"

	var revPoints []entity.SalesChartPoint
	if err := r.pg.DB().SelectContext(ctx, &revPoints, revenueQuery, args...); err != nil {
		return nil, fmt.Errorf("analytics.GetSalesCharts revenue: %w", err)
	}

	avgQuery := `
		SELECT
			DATE_TRUNC('day', lt.created_at) AS day,
			COALESCE(AVG(lt.amount), 0)       AS value
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE lt.type = 'earn'
		  AND b.org_id = $1
		  AND lt.created_at >= $2
		  AND lt.created_at <= $3`

	if f.BotID != nil {
		avgQuery += fmt.Sprintf(" AND bc.bot_id = $%d", 4)
	}
	avgQuery += " GROUP BY 1 ORDER BY 1"

	var avgPoints []entity.SalesChartPoint
	if err := r.pg.DB().SelectContext(ctx, &avgPoints, avgQuery, args...); err != nil {
		return nil, fmt.Errorf("analytics.GetSalesCharts avg_amount: %w", err)
	}

	return map[string][]entity.SalesChartPoint{
		"transactions": txPoints,
		"revenue":      revPoints,
		"avg_amount":   avgPoints,
	}, nil
}

func (r *Analytics) GetLoyaltyAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error) {
	var result entity.LoyaltyAnalytics

	// New clients in period
	newQuery := `
		SELECT COUNT(*)
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1
		  AND bc.registered_at >= $2
		  AND bc.registered_at <= $3`
	args := []interface{}{f.OrgID, f.From, f.To}
	if f.BotID != nil {
		newQuery += " AND bc.bot_id = $4"
		args = append(args, *f.BotID)
	}
	if err := r.pg.DB().GetContext(ctx, &result.NewClients, newQuery, args...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics new_clients: %w", err)
	}

	// Active clients (activity in last 30 days from end of period)
	activeQuery := `
		SELECT COUNT(DISTINCT lt.client_id)
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1
		  AND lt.created_at > $2::timestamptz - INTERVAL '30 days'
		  AND lt.created_at <= $2::timestamptz`
	activeArgs := []interface{}{f.OrgID, f.To}
	if f.BotID != nil {
		activeQuery += " AND bc.bot_id = $3"
		activeArgs = append(activeArgs, *f.BotID)
	}
	if err := r.pg.DB().GetContext(ctx, &result.ActiveClients, activeQuery, activeArgs...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics active_clients: %w", err)
	}

	// Bonus earned/spent
	bonusQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN lt.type = 'earn' THEN lt.amount ELSE 0 END), 0) AS bonus_earned,
			COALESCE(SUM(CASE WHEN lt.type = 'spend' THEN lt.amount ELSE 0 END), 0) AS bonus_spent
		FROM loyalty_transactions lt
		JOIN bot_clients bc ON lt.client_id = bc.id
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1
		  AND lt.created_at >= $2
		  AND lt.created_at <= $3`
	bonusArgs := []interface{}{f.OrgID, f.From, f.To}
	if f.BotID != nil {
		bonusQuery += " AND bc.bot_id = $4"
		bonusArgs = append(bonusArgs, *f.BotID)
	}

	var bonusRow struct {
		BonusEarned float64 `db:"bonus_earned"`
		BonusSpent  float64 `db:"bonus_spent"`
	}
	if err := r.pg.DB().GetContext(ctx, &bonusRow, bonusQuery, bonusArgs...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics bonus: %w", err)
	}
	result.BonusEarned = bonusRow.BonusEarned
	result.BonusSpent = bonusRow.BonusSpent

	// Demographics: by gender
	genderQuery := `
		SELECT COALESCE(gender, 'unknown') AS label, COUNT(*) AS value
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1`
	gArgs := []interface{}{f.OrgID}
	if f.BotID != nil {
		genderQuery += " AND bc.bot_id = $2"
		gArgs = append(gArgs, *f.BotID)
	}
	genderQuery += " GROUP BY 1"

	var genderSlices []struct {
		Label string `db:"label"`
		Value int64  `db:"value"`
	}
	if err := r.pg.DB().SelectContext(ctx, &genderSlices, genderQuery, gArgs...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics gender: %w", err)
	}

	var total int64
	for _, s := range genderSlices {
		total += s.Value
	}
	for _, s := range genderSlices {
		pct := float64(0)
		if total > 0 {
			pct = float64(s.Value) / float64(total) * 100
		}
		result.Demographics.ByGender = append(result.Demographics.ByGender, entity.PieSlice{
			Label:   s.Label,
			Value:   s.Value,
			Percent: pct,
		})
	}

	// Demographics: by OS
	osQuery := `
		SELECT COALESCE(os, 'unknown') AS label, COUNT(*) AS value
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1`
	osArgs := []interface{}{f.OrgID}
	if f.BotID != nil {
		osQuery += " AND bc.bot_id = $2"
		osArgs = append(osArgs, *f.BotID)
	}
	osQuery += " GROUP BY 1"

	var osSlices []struct {
		Label string `db:"label"`
		Value int64  `db:"value"`
	}
	if err := r.pg.DB().SelectContext(ctx, &osSlices, osQuery, osArgs...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics os: %w", err)
	}
	var osTotal int64
	for _, s := range osSlices {
		osTotal += s.Value
	}
	for _, s := range osSlices {
		pct := float64(0)
		if osTotal > 0 {
			pct = float64(s.Value) / float64(osTotal) * 100
		}
		result.Demographics.ByOS = append(result.Demographics.ByOS, entity.PieSlice{
			Label:   s.Label,
			Value:   s.Value,
			Percent: pct,
		})
	}

	// Bot funnel: registered → has_loyalty → transacted
	funnelQuery := `
		SELECT
			COUNT(*)                                                               AS registered,
			COUNT(cl.client_id)                                                    AS has_loyalty,
			COUNT(DISTINCT lt.client_id)                                           AS transacted
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		LEFT JOIN client_loyalty cl ON cl.client_id = bc.id
		LEFT JOIN loyalty_transactions lt ON lt.client_id = bc.id
		WHERE b.org_id = $1`
	fArgs := []interface{}{f.OrgID}
	if f.BotID != nil {
		funnelQuery += " AND bc.bot_id = $2"
		fArgs = append(fArgs, *f.BotID)
	}

	var funnelRow struct {
		Registered int64 `db:"registered"`
		HasLoyalty int64 `db:"has_loyalty"`
		Transacted int64 `db:"transacted"`
	}
	if err := r.pg.DB().GetContext(ctx, &funnelRow, funnelQuery, fArgs...); err != nil {
		return nil, fmt.Errorf("analytics.GetLoyaltyAnalytics funnel: %w", err)
	}

	steps := []struct {
		step  string
		count int64
	}{
		{"registered", funnelRow.Registered},
		{"has_loyalty", funnelRow.HasLoyalty},
		{"transacted", funnelRow.Transacted},
	}
	for _, s := range steps {
		pct := float64(0)
		if funnelRow.Registered > 0 {
			pct = float64(s.count) / float64(funnelRow.Registered) * 100
		}
		result.BotFunnel = append(result.BotFunnel, entity.FunnelStep{
			Step:    s.step,
			Count:   s.count,
			Percent: pct,
		})
	}

	// Loyalty percent
	if total > 0 {
		result.Demographics.LoyaltyPercent = float64(funnelRow.HasLoyalty) / float64(funnelRow.Registered) * 100
	}

	return &result, nil
}

func (r *Analytics) GetCampaignAnalytics(ctx context.Context, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error) {
	query := `
		SELECT
			c.id   AS campaign_id,
			c.name AS campaign_name,
			COUNT(cm.id)                                                          AS sent,
			COALESCE(
				COUNT(cm.id) FILTER (WHERE cm.status = 'sent')::FLOAT / NULLIF(COUNT(cm.id), 0) * 100,
				0
			)                                                                     AS open_rate,
			0::BIGINT                                                             AS conversions
		FROM campaigns c
		LEFT JOIN campaign_messages cm ON cm.campaign_id = c.id
		WHERE c.org_id = $1
		  AND c.created_at >= $2
		  AND c.created_at <= $3
		GROUP BY c.id, c.name
		ORDER BY sent DESC`

	var stats []entity.CampaignStat
	if err := r.pg.DB().SelectContext(ctx, &stats, query, f.OrgID, f.From, f.To); err != nil {
		return nil, fmt.Errorf("analytics.GetCampaignAnalytics: %w", err)
	}

	var result entity.CampaignAnalytics
	result.ByCampaign = stats

	for _, s := range stats {
		result.TotalSent += s.Sent
	}
	if len(stats) > 0 && result.TotalSent > 0 {
		var totalOpenRate float64
		for _, s := range stats {
			totalOpenRate += s.OpenRate
		}
		result.OpenRate = totalOpenRate / float64(len(stats))
	}

	return &result, nil
}

func (r *Analytics) RefreshMaterializedViews(ctx context.Context) error {
	for _, view := range []string{"mv_daily_sales", "mv_loyalty_stats"} {
		if _, err := r.pg.DB().ExecContext(ctx,
			"REFRESH MATERIALIZED VIEW CONCURRENTLY "+view); err != nil {
			return fmt.Errorf("analytics.RefreshMaterializedViews %s: %w", view, err)
		}
	}
	return nil
}
