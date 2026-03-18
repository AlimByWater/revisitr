package entity

import "time"

type User struct {
	ID           int       `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	Phone        string    `db:"phone" json:"phone,omitempty"`
	Name         string    `db:"name" json:"name"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role"`
	OrgID        int       `db:"org_id" json:"org_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
