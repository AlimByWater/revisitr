package postgres

import (
	"context"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type Dashboard struct {
	pg *Module
}

func NewDashboard(pg *Module) *Dashboard {
	return &Dashboard{pg: pg}
}

func parsePeriod(filter entity.DashboardFilter) (from, to, prevFrom, prevTo time.Time) {
	to = time.Now()
	if filter.To != nil {
		to = *filter.To
	}

	days := 30
	switch filter.Period {
	case "7d":
		days = 7
	case "90d":
		days = 90
	}

	if filter.From != nil {
		from = *filter.From
	} else {
		from = to.AddDate(0, 0, -days)
	}

	duration := to.Sub(from)
	prevTo = from
	prevFrom = prevTo.Add(-duration)
	return
}

func (r *Dashboard) GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error) {
	from, to, prevFrom, prevTo := parsePeriod(filter)

	var widgets entity.DashboardWidgets

	// Revenue: SUM of earn transactions for the org
	revenueQuery := `
		SELECT COALESCE(SUM(lt.amount), 0)
		FROM loyalty_transactions lt
		JOIN client_loyalty cl ON lt.client_id = cl.client_id AND lt.program_id = cl.program_id
		JOIN loyalty_programs lp ON cl.program_id = lp.id
		WHERE lp.org_id = $1
		  AND lt.type = 'earn'
		  AND lt.created_at >= $2 AND lt.created_at < $3`

	if err := r.pg.DB().GetContext(ctx, &widgets.Revenue.Value, revenueQuery, orgID, from, to); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets revenue: %w", err)
	}
	if err := r.pg.DB().GetContext(ctx, &widgets.Revenue.Previous, revenueQuery, orgID, prevFrom, prevTo); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets revenue prev: %w", err)
	}
	widgets.Revenue.Trend = calcTrend(widgets.Revenue.Value, widgets.Revenue.Previous)

	// AvgCheck: AVG of earn transactions
	avgQuery := `
		SELECT COALESCE(AVG(lt.amount), 0)
		FROM loyalty_transactions lt
		JOIN client_loyalty cl ON lt.client_id = cl.client_id AND lt.program_id = cl.program_id
		JOIN loyalty_programs lp ON cl.program_id = lp.id
		WHERE lp.org_id = $1
		  AND lt.type = 'earn'
		  AND lt.created_at >= $2 AND lt.created_at < $3`

	if err := r.pg.DB().GetContext(ctx, &widgets.AvgCheck.Value, avgQuery, orgID, from, to); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets avg_check: %w", err)
	}
	if err := r.pg.DB().GetContext(ctx, &widgets.AvgCheck.Previous, avgQuery, orgID, prevFrom, prevTo); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets avg_check prev: %w", err)
	}
	widgets.AvgCheck.Trend = calcTrend(widgets.AvgCheck.Value, widgets.AvgCheck.Previous)

	// NewClients: COUNT of bot_clients registered in period
	newClientsQuery := `
		SELECT COUNT(*)
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE b.org_id = $1
		  AND bc.registered_at >= $2 AND bc.registered_at < $3`

	if err := r.pg.DB().GetContext(ctx, &widgets.NewClients.Value, newClientsQuery, orgID, from, to); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets new_clients: %w", err)
	}
	if err := r.pg.DB().GetContext(ctx, &widgets.NewClients.Previous, newClientsQuery, orgID, prevFrom, prevTo); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets new_clients prev: %w", err)
	}
	widgets.NewClients.Trend = calcTrend(widgets.NewClients.Value, widgets.NewClients.Previous)

	// ActiveClients: COUNT DISTINCT client_id from loyalty_transactions in period
	activeClientsQuery := `
		SELECT COUNT(DISTINCT lt.client_id)
		FROM loyalty_transactions lt
		JOIN client_loyalty cl ON lt.client_id = cl.client_id AND lt.program_id = cl.program_id
		JOIN loyalty_programs lp ON cl.program_id = lp.id
		WHERE lp.org_id = $1
		  AND lt.created_at >= $2 AND lt.created_at < $3`

	if err := r.pg.DB().GetContext(ctx, &widgets.ActiveClients.Value, activeClientsQuery, orgID, from, to); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets active_clients: %w", err)
	}
	if err := r.pg.DB().GetContext(ctx, &widgets.ActiveClients.Previous, activeClientsQuery, orgID, prevFrom, prevTo); err != nil {
		return nil, fmt.Errorf("dashboard.GetWidgets active_clients prev: %w", err)
	}
	widgets.ActiveClients.Trend = calcTrend(widgets.ActiveClients.Value, widgets.ActiveClients.Previous)

	return &widgets, nil
}

func (r *Dashboard) GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error) {
	from, to, _, _ := parsePeriod(filter)

	var charts entity.DashboardCharts

	// Revenue chart: daily SUM of earn transactions with generate_series for filling gaps
	revenueChartQuery := `
		SELECT d::date AS date, COALESCE(SUM(lt.amount), 0) AS value
		FROM generate_series($1::date, $2::date, '1 day') d
		LEFT JOIN loyalty_transactions lt ON DATE(lt.created_at) = d::date
			AND lt.type = 'earn'
			AND lt.client_id IN (
				SELECT bc.id FROM bot_clients bc
				JOIN bots b ON bc.bot_id = b.id
				WHERE b.org_id = $3
			)
		GROUP BY d::date
		ORDER BY d::date`

	rows, err := r.pg.DB().QueryContext(ctx, revenueChartQuery, from, to, orgID)
	if err != nil {
		return nil, fmt.Errorf("dashboard.GetCharts revenue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var point entity.ChartPoint
		var date time.Time
		if err := rows.Scan(&date, &point.Value); err != nil {
			return nil, fmt.Errorf("dashboard.GetCharts revenue scan: %w", err)
		}
		point.Date = date.Format("2006-01-02")
		charts.Revenue = append(charts.Revenue, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard.GetCharts revenue rows: %w", err)
	}

	// New clients chart: daily COUNT of bot_clients with generate_series
	newClientsChartQuery := `
		SELECT d::date AS date, COALESCE(COUNT(bc.id), 0) AS value
		FROM generate_series($1::date, $2::date, '1 day') d
		LEFT JOIN bot_clients bc ON DATE(bc.registered_at) = d::date
			AND bc.bot_id IN (
				SELECT b.id FROM bots b WHERE b.org_id = $3
			)
		GROUP BY d::date
		ORDER BY d::date`

	rows2, err := r.pg.DB().QueryContext(ctx, newClientsChartQuery, from, to, orgID)
	if err != nil {
		return nil, fmt.Errorf("dashboard.GetCharts new_clients: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var point entity.ChartPoint
		var date time.Time
		if err := rows2.Scan(&date, &point.Value); err != nil {
			return nil, fmt.Errorf("dashboard.GetCharts new_clients scan: %w", err)
		}
		point.Date = date.Format("2006-01-02")
		charts.NewClients = append(charts.NewClients, point)
	}
	if err := rows2.Err(); err != nil {
		return nil, fmt.Errorf("dashboard.GetCharts new_clients rows: %w", err)
	}

	return &charts, nil
}

func calcTrend(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return ((current - previous) / previous) * 100
}
