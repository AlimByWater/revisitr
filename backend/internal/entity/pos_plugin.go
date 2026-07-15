package entity

import "time"

type PluginKey struct {
	ID            int        `db:"id" json:"id"`
	OrgID         int        `db:"org_id" json:"org_id"`
	IntegrationID int        `db:"integration_id" json:"integration_id"`
	KeyHash       string     `db:"key_hash" json:"-"`
	Label         string     `db:"label" json:"label"`
	LastUsedAt    *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	RevokedAt     *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

type PluginOperation struct {
	ID              int       `db:"id" json:"id"`
	IntegrationID   int       `db:"integration_id" json:"integration_id"`
	ExternalOrderID string    `db:"external_order_id" json:"external_order_id"`
	OpType          string    `db:"op_type" json:"op_type"`
	ClientID        int       `db:"client_id" json:"client_id"`
	ProgramID       int       `db:"program_id" json:"program_id"`
	Amount          float64   `db:"amount" json:"amount"`
	BalanceAfter    float64   `db:"balance_after" json:"balance_after"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
