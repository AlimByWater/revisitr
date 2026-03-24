package entity

import "time"

type BalanceReserve struct {
	ID        int       `db:"id" json:"id"`
	ClientID  int       `db:"client_id" json:"client_id"`
	ProgramID int       `db:"program_id" json:"program_id"`
	Amount    float64   `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
