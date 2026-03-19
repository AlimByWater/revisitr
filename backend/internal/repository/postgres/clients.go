package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"revisitr/internal/entity"
)

type Clients struct {
	pg *Module
}

func NewClients(pg *Module) *Clients {
	return &Clients{pg: pg}
}

// clientProfileRow is a scan struct that includes total_count for windowed pagination.
type clientProfileRow struct {
	entity.BotClient
	BotName        string  `db:"bot_name"`
	LoyaltyBalance float64 `db:"loyalty_balance"`
	LoyaltyLevel   *string `db:"loyalty_level"`
	TotalPurchases float64 `db:"total_purchases"`
	PurchaseCount  int     `db:"purchase_count"`
	TotalCount     int     `db:"total_count"`
}

var allowedSortColumns = map[string]string{
	"name":          "bc.first_name",
	"balance":       "loyalty_balance",
	"registered_at": "bc.registered_at",
}

func (r *Clients) GetByOrgID(ctx context.Context, orgID int, filter entity.ClientFilter) ([]entity.ClientProfile, int, error) {
	args := []interface{}{orgID}
	argIdx := 2

	where := []string{"b.org_id = $1"}

	if filter.BotID != nil {
		where = append(where, fmt.Sprintf("bc.bot_id = $%d", argIdx))
		args = append(args, *filter.BotID)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		where = append(where, fmt.Sprintf(
			"(bc.first_name ILIKE $%d OR bc.last_name ILIKE $%d OR bc.phone ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, searchPattern)
		argIdx++
	}

	if filter.Segment != nil && *filter.Segment != "" {
		where = append(where, fmt.Sprintf("bc.tags @> $%d::jsonb", argIdx))
		segmentJSON := fmt.Sprintf(`[%q]`, *filter.Segment)
		args = append(args, segmentJSON)
		argIdx++
	}

	orderBy := "bc.registered_at DESC"
	if col, ok := allowedSortColumns[filter.SortBy]; ok {
		dir := "ASC"
		if strings.EqualFold(filter.SortOrder, "desc") {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", col, dir)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT bc.*, b.name as bot_name,
			COALESCE(cl.balance, 0) as loyalty_balance,
			ll.name as loyalty_level,
			COALESCE(cl.total_earned, 0) as total_purchases,
			0 as purchase_count,
			COUNT(*) OVER() as total_count
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		LEFT JOIN client_loyalty cl ON bc.id = cl.client_id
		LEFT JOIN loyalty_levels ll ON cl.level_id = ll.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "),
		orderBy,
		argIdx, argIdx+1,
	)

	args = append(args, limit, offset)

	var rows []clientProfileRow
	if err := r.pg.DB().SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("clients.GetByOrgID: %w", err)
	}

	total := 0
	profiles := make([]entity.ClientProfile, len(rows))
	for i, row := range rows {
		if i == 0 {
			total = row.TotalCount
		}
		profiles[i] = entity.ClientProfile{
			BotClient:      row.BotClient,
			BotName:        row.BotName,
			LoyaltyBalance: row.LoyaltyBalance,
			LoyaltyLevel:   row.LoyaltyLevel,
			TotalPurchases: row.TotalPurchases,
			PurchaseCount:  row.PurchaseCount,
		}
	}

	return profiles, total, nil
}

func (r *Clients) GetByID(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error) {
	var row clientProfileRow
	query := `
		SELECT bc.*, b.name as bot_name,
			COALESCE(cl.balance, 0) as loyalty_balance,
			ll.name as loyalty_level,
			COALESCE(cl.total_earned, 0) as total_purchases,
			0 as purchase_count,
			0 as total_count
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		LEFT JOIN client_loyalty cl ON bc.id = cl.client_id
		LEFT JOIN loyalty_levels ll ON cl.level_id = ll.id
		WHERE bc.id = $1 AND b.org_id = $2`

	if err := r.pg.DB().GetContext(ctx, &row, query, clientID, orgID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("clients.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("clients.GetByID: %w", err)
	}

	profile := &entity.ClientProfile{
		BotClient:      row.BotClient,
		BotName:        row.BotName,
		LoyaltyBalance: row.LoyaltyBalance,
		LoyaltyLevel:   row.LoyaltyLevel,
		TotalPurchases: row.TotalPurchases,
		PurchaseCount:  row.PurchaseCount,
	}

	return profile, nil
}

func (r *Clients) Update(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error {
	if req.Tags == nil {
		return nil
	}

	tagsVal, err := req.Tags.Value()
	if err != nil {
		return fmt.Errorf("clients.Update tags value: %w", err)
	}

	query := `
		UPDATE bot_clients SET tags = $1
		WHERE id = $2 AND bot_id IN (SELECT id FROM bots WHERE org_id = $3)`

	result, err := r.pg.DB().ExecContext(ctx, query, tagsVal, clientID, orgID)
	if err != nil {
		return fmt.Errorf("clients.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("clients.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("clients.Update: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Clients) GetStats(ctx context.Context, orgID int) (*entity.ClientStats, error) {
	var stats entity.ClientStats
	query := `
		SELECT
			COUNT(*) as total_clients,
			COALESCE(SUM(cl.balance), 0) as total_balance,
			COUNT(*) FILTER (WHERE bc.registered_at >= date_trunc('month', CURRENT_DATE)) as new_this_month,
			COUNT(*) FILTER (WHERE bc.registered_at >= CURRENT_DATE - INTERVAL '7 days') as active_this_week
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		LEFT JOIN client_loyalty cl ON bc.id = cl.client_id
		WHERE b.org_id = $1`

	if err := r.pg.DB().GetContext(ctx, &stats, query, orgID); err != nil {
		return nil, fmt.Errorf("clients.GetStats: %w", err)
	}

	return &stats, nil
}

func (r *Clients) CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error) {
	args := []interface{}{orgID}
	argIdx := 2

	where := []string{"b.org_id = $1"}

	if filter.BotID != nil {
		where = append(where, fmt.Sprintf("bc.bot_id = $%d", argIdx))
		args = append(args, *filter.BotID)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		where = append(where, fmt.Sprintf(
			"(bc.first_name ILIKE $%d OR bc.last_name ILIKE $%d OR bc.phone ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, searchPattern)
		argIdx++
	}

	if filter.Segment != nil && *filter.Segment != "" {
		where = append(where, fmt.Sprintf("bc.tags @> $%d::jsonb", argIdx))
		segmentJSON := fmt.Sprintf(`[%q]`, *filter.Segment)
		args = append(args, segmentJSON)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE %s`,
		strings.Join(where, " AND "),
	)

	var count int
	if err := r.pg.DB().GetContext(ctx, &count, query, args...); err != nil {
		return 0, fmt.Errorf("clients.CountByFilter: %w", err)
	}

	return count, nil
}

func (r *Clients) GetTransactionsByClientID(ctx context.Context, clientID int, limit, offset int) ([]entity.LoyaltyTransaction, error) {
	var txs []entity.LoyaltyTransaction
	query := `
		SELECT * FROM loyalty_transactions
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.pg.DB().SelectContext(ctx, &txs, query, clientID, limit, offset); err != nil {
		return nil, fmt.Errorf("clients.GetTransactionsByClientID: %w", err)
	}

	return txs, nil
}

// GetIDsByFilter returns client IDs matching a segment filter for a given org.
func (r *Clients) GetIDsByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error) {
	args := []interface{}{orgID}
	idx := 2
	where := []string{"b.org_id = $1"}

	if f.Gender != nil {
		where = append(where, fmt.Sprintf("bc.gender = $%d", idx))
		args = append(args, *f.Gender)
		idx++
	}
	if f.AgeFrom != nil {
		where = append(where, fmt.Sprintf("DATE_PART('year', AGE(bc.birth_date)) >= $%d", idx))
		args = append(args, *f.AgeFrom)
		idx++
	}
	if f.AgeTo != nil {
		where = append(where, fmt.Sprintf("DATE_PART('year', AGE(bc.birth_date)) <= $%d", idx))
		args = append(args, *f.AgeTo)
		idx++
	}
	if f.Tags != nil && len(f.Tags) > 0 {
		for _, tag := range f.Tags {
			where = append(where, fmt.Sprintf("bc.tags @> $%d::jsonb", idx))
			args = append(args, `["`+tag+`"]`)
			idx++
		}
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT bc.id
		FROM bot_clients bc
		JOIN bots b ON bc.bot_id = b.id
		WHERE %s`,
		strings.Join(where, " AND "),
	)

	var ids []int
	if err := r.pg.DB().SelectContext(ctx, &ids, query, args...); err != nil {
		return nil, fmt.Errorf("clients.GetIDsByFilter: %w", err)
	}
	return ids, nil
}
