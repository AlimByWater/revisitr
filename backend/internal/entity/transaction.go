package entity

import "time"

type LoyaltyTransaction struct {
	ID           int       `db:"id" json:"id"`
	ClientID     int       `db:"client_id" json:"client_id"`
	ProgramID    int       `db:"program_id" json:"program_id"`
	Type         string    `db:"type" json:"type"`
	Amount       float64   `db:"amount" json:"amount"`
	BalanceAfter float64   `db:"balance_after" json:"balance_after"`
	Description  string    `db:"description" json:"description,omitempty"`
	CreatedBy    *int      `db:"created_by" json:"created_by,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
