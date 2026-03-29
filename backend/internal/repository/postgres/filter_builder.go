package postgres

import (
	"fmt"

	"revisitr/internal/entity"
)

// buildSegmentFilterWhere builds WHERE clauses and args for SegmentFilter.
// Returns: additional WHERE clauses, args, next arg index, whether client_loyalty JOIN is needed.
func buildSegmentFilterWhere(f entity.SegmentFilter, startIdx int) (clauses []string, args []interface{}, nextIdx int, needsLoyaltyJoin bool) {
	idx := startIdx

	if f.BotID != nil {
		clauses = append(clauses, fmt.Sprintf("bc.bot_id = $%d", idx))
		args = append(args, *f.BotID)
		idx++
	}
	if f.Search != nil && *f.Search != "" {
		searchPattern := "%" + *f.Search + "%"
		clauses = append(clauses, fmt.Sprintf(
			"(bc.first_name ILIKE $%d OR bc.last_name ILIKE $%d OR bc.phone ILIKE $%d)",
			idx, idx, idx,
		))
		args = append(args, searchPattern)
		idx++
	}
	if f.Gender != nil {
		clauses = append(clauses, fmt.Sprintf("bc.gender = $%d", idx))
		args = append(args, *f.Gender)
		idx++
	}
	if f.AgeFrom != nil {
		clauses = append(clauses, fmt.Sprintf("DATE_PART('year', AGE(bc.birth_date)) >= $%d", idx))
		args = append(args, *f.AgeFrom)
		idx++
	}
	if f.AgeTo != nil {
		clauses = append(clauses, fmt.Sprintf("DATE_PART('year', AGE(bc.birth_date)) <= $%d", idx))
		args = append(args, *f.AgeTo)
		idx++
	}
	if f.City != nil && *f.City != "" {
		clauses = append(clauses, fmt.Sprintf("bc.city ILIKE $%d", idx))
		args = append(args, "%"+*f.City+"%")
		idx++
	}
	if f.OS != nil && *f.OS != "" {
		clauses = append(clauses, fmt.Sprintf("bc.os = $%d", idx))
		args = append(args, *f.OS)
		idx++
	}
	if f.RegisteredFrom != nil && *f.RegisteredFrom != "" {
		clauses = append(clauses, fmt.Sprintf("bc.registered_at >= $%d::date", idx))
		args = append(args, *f.RegisteredFrom)
		idx++
	}
	if f.RegisteredTo != nil && *f.RegisteredTo != "" {
		clauses = append(clauses, fmt.Sprintf("bc.registered_at < ($%d::date + INTERVAL '1 day')", idx))
		args = append(args, *f.RegisteredTo)
		idx++
	}
	if f.Tags != nil && len(f.Tags) > 0 {
		for _, tag := range f.Tags {
			clauses = append(clauses, fmt.Sprintf("bc.tags @> $%d::jsonb", idx))
			args = append(args, `["`+tag+`"]`)
			idx++
		}
	}
	if f.RFMCategory != nil && *f.RFMCategory != "" {
		clauses = append(clauses, fmt.Sprintf("bc.rfm_segment = $%d", idx))
		args = append(args, *f.RFMCategory)
		idx++
	}
	if f.MinVisits != nil {
		clauses = append(clauses, fmt.Sprintf("bc.total_visits_lifetime >= $%d", idx))
		args = append(args, *f.MinVisits)
		idx++
	}
	if f.MaxVisits != nil {
		clauses = append(clauses, fmt.Sprintf("bc.total_visits_lifetime <= $%d", idx))
		args = append(args, *f.MaxVisits)
		idx++
	}
	if f.MinSpend != nil {
		clauses = append(clauses, fmt.Sprintf("bc.monetary_sum >= $%d", idx))
		args = append(args, *f.MinSpend)
		idx++
	}
	if f.MaxSpend != nil {
		clauses = append(clauses, fmt.Sprintf("bc.monetary_sum <= $%d", idx))
		args = append(args, *f.MaxSpend)
		idx++
	}

	// Loyalty fields — require LEFT JOIN client_loyalty
	if f.LevelID != nil {
		clauses = append(clauses, fmt.Sprintf("cl.level_id = $%d", idx))
		args = append(args, *f.LevelID)
		idx++
		needsLoyaltyJoin = true
	}
	if f.MinBalance != nil {
		clauses = append(clauses, fmt.Sprintf("cl.balance >= $%d", idx))
		args = append(args, *f.MinBalance)
		idx++
		needsLoyaltyJoin = true
	}
	if f.MaxBalance != nil {
		clauses = append(clauses, fmt.Sprintf("cl.balance <= $%d", idx))
		args = append(args, *f.MaxBalance)
		idx++
		needsLoyaltyJoin = true
	}
	if f.MinSpentPoints != nil {
		clauses = append(clauses, fmt.Sprintf("cl.total_spent >= $%d", idx))
		args = append(args, *f.MinSpentPoints)
		idx++
		needsLoyaltyJoin = true
	}
	if f.MaxSpentPoints != nil {
		clauses = append(clauses, fmt.Sprintf("cl.total_spent <= $%d", idx))
		args = append(args, *f.MaxSpentPoints)
		idx++
		needsLoyaltyJoin = true
	}

	nextIdx = idx
	return
}
