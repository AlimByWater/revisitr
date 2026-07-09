package entity

import "time"

type WalletDeviceRegistration struct {
	ID              int       `db:"id" json:"id"`
	OrgID           int       `db:"org_id" json:"org_id"`
	DeviceLibraryID string    `db:"device_library_id" json:"device_library_id"`
	PassTypeID      string    `db:"pass_type_id" json:"pass_type_id"`
	SerialNumber    string    `db:"serial_number" json:"serial_number"`
	PushToken       *string   `db:"push_token" json:"-"`
	AuthToken       string    `db:"auth_token" json:"-"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
