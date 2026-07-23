package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Wallet struct {
	pg *Module
}

func NewWallet(pg *Module) *Wallet {
	return &Wallet{pg: pg}
}

// ── Config ───────────────────────────────────────────────────────────────────

func (r *Wallet) GetConfigs(ctx context.Context, orgID int) ([]entity.WalletConfig, error) {
	var configs []entity.WalletConfig
	err := r.pg.DB().SelectContext(ctx, &configs,
		"SELECT * FROM wallet_configs WHERE org_id = $1 ORDER BY platform", orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetConfigs: %w", err)
	}
	return configs, nil
}

func (r *Wallet) GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
	var cfg entity.WalletConfig
	err := r.pg.DB().GetContext(ctx, &cfg,
		"SELECT * FROM wallet_configs WHERE org_id = $1 AND platform = $2", orgID, platform)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("wallet.GetConfig: %w", err)
	}
	return &cfg, nil
}

func (r *Wallet) SaveConfig(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error) {
	var cfg entity.WalletConfig
	err := r.pg.DB().GetContext(ctx, &cfg, `
		INSERT INTO wallet_configs (org_id, platform, is_enabled, credentials, design)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (org_id, platform) DO UPDATE SET
			is_enabled = EXCLUDED.is_enabled,
			credentials = EXCLUDED.credentials,
			design = EXCLUDED.design,
			updated_at = now()
		RETURNING *`,
		orgID, req.Platform, req.IsEnabled, req.Credentials, req.Design)
	if err != nil {
		return nil, fmt.Errorf("wallet.SaveConfig: %w", err)
	}
	return &cfg, nil
}

func (r *Wallet) DeleteConfig(ctx context.Context, orgID int, platform string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM wallet_configs WHERE org_id = $1 AND platform = $2", orgID, platform)
	if err != nil {
		return fmt.Errorf("wallet.DeleteConfig: %w", err)
	}
	return nil
}

// ── Passes ───────────────────────────────────────────────────────────────────

func (r *Wallet) CreatePass(ctx context.Context, pass *entity.WalletPass) error {
	err := r.pg.DB().GetContext(ctx, pass, `
		INSERT INTO wallet_passes (org_id, client_id, platform, serial_number, auth_token, last_balance, last_level)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *`,
		pass.OrgID, pass.ClientID, pass.Platform, pass.SerialNumber, pass.AuthToken, pass.LastBalance, pass.LastLevel)
	if err != nil {
		return fmt.Errorf("wallet.CreatePass: %w", err)
	}
	return nil
}

func (r *Wallet) GetPassBySerial(ctx context.Context, serial string) (*entity.WalletPass, error) {
	var pass entity.WalletPass
	err := r.pg.DB().GetContext(ctx, &pass,
		"SELECT * FROM wallet_passes WHERE serial_number = $1", serial)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("wallet.GetPassBySerial: %w", err)
	}
	return &pass, nil
}

func (r *Wallet) GetPassByClientPlatform(ctx context.Context, clientID int, platform string) (*entity.WalletPass, error) {
	var pass entity.WalletPass
	err := r.pg.DB().GetContext(ctx, &pass,
		"SELECT * FROM wallet_passes WHERE client_id = $1 AND platform = $2", clientID, platform)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("wallet.GetPassByClientPlatform: %w", err)
	}
	return &pass, nil
}

func (r *Wallet) GetPassesByOrg(ctx context.Context, orgID int) ([]entity.WalletPass, error) {
	var passes []entity.WalletPass
	err := r.pg.DB().SelectContext(ctx, &passes,
		"SELECT * FROM wallet_passes WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetPassesByOrg: %w", err)
	}
	return passes, nil
}

func (r *Wallet) GetPassesByClient(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
	var passes []entity.WalletPass
	err := r.pg.DB().SelectContext(ctx, &passes,
		"SELECT * FROM wallet_passes WHERE client_id = $1 ORDER BY created_at DESC", clientID)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetPassesByClient: %w", err)
	}
	return passes, nil
}

func (r *Wallet) UpdatePassBalance(ctx context.Context, id int, balance int, level string) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		UPDATE wallet_passes SET last_balance = $1, last_level = $2, last_updated_at = now(), updated_at = now()
		WHERE id = $3`, balance, level, id)
	if err != nil {
		return fmt.Errorf("wallet.UpdatePassBalance: %w", err)
	}
	return nil
}

func (r *Wallet) UpdatePushToken(ctx context.Context, serial string, token string) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		UPDATE wallet_passes SET push_token = $1, updated_at = now()
		WHERE serial_number = $2`, token, serial)
	if err != nil {
		return fmt.Errorf("wallet.UpdatePushToken: %w", err)
	}
	return nil
}

func (r *Wallet) RevokePass(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		UPDATE wallet_passes SET status = 'revoked', updated_at = now()
		WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("wallet.RevokePass: %w", err)
	}
	return nil
}

func (r *Wallet) GetStats(ctx context.Context, orgID int) (*entity.WalletStats, error) {
	var stats entity.WalletStats
	err := r.pg.DB().GetContext(ctx, &stats, `
		SELECT
			COUNT(*) AS total_passes,
			COUNT(*) FILTER (WHERE platform = 'apple') AS apple_passes,
			COUNT(*) FILTER (WHERE platform = 'google') AS google_passes,
			COUNT(*) FILTER (WHERE status = 'active') AS active_passes
		FROM wallet_passes WHERE org_id = $1`, orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetStats: %w", err)
	}
	return &stats, nil
}

func (r *Wallet) GetPassesWithPushToken(ctx context.Context, orgID int) ([]entity.WalletPass, error) {
	var passes []entity.WalletPass
	err := r.pg.DB().SelectContext(ctx, &passes, `
		SELECT * FROM wallet_passes
		WHERE org_id = $1 AND status = 'active' AND push_token IS NOT NULL`,
		orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetPassesWithPushToken: %w", err)
	}
	return passes, nil
}

func (r *Wallet) GetClientQRCode(ctx context.Context, clientID int) (string, error) {
	var qrCode *string
	err := r.pg.DB().GetContext(ctx, &qrCode,
		"SELECT qr_code FROM bot_clients WHERE id = $1", clientID)
	if err != nil {
		return "", fmt.Errorf("wallet.GetClientQRCode: %w", err)
	}
	if qrCode == nil {
		return "", nil
	}
	return *qrCode, nil
}

func (r *Wallet) GetOrgName(ctx context.Context, orgID int) (string, error) {
	var name string
	err := r.pg.DB().GetContext(ctx, &name,
		"SELECT name FROM organizations WHERE id = $1", orgID)
	if err != nil {
		return "", fmt.Errorf("wallet.GetOrgName: %w", err)
	}
	return name, nil
}

// ── Device Registrations (Apple Wallet web service) ─────────────────────────

// CreateDeviceRegistration upserts a device↔pass registration.
// Returns true if a new registration was created, false if it already existed.
func (r *Wallet) CreateDeviceRegistration(ctx context.Context, reg *entity.WalletDeviceRegistration) (bool, error) {
	var created bool
	err := r.pg.DB().GetContext(ctx, &created, `
		INSERT INTO wallet_device_registrations (org_id, device_library_id, pass_type_id, serial_number, push_token, auth_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (device_library_id, pass_type_id, serial_number) DO UPDATE SET
			push_token = EXCLUDED.push_token, updated_at = now()
		RETURNING (xmax = 0) AS created`,
		reg.OrgID, reg.DeviceLibraryID, reg.PassTypeID, reg.SerialNumber, reg.PushToken, reg.AuthToken)
	if err != nil {
		return false, fmt.Errorf("wallet.CreateDeviceRegistration: %w", err)
	}
	return created, nil
}

func (r *Wallet) DeleteDeviceRegistration(ctx context.Context, deviceID, passTypeID, serial string) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		DELETE FROM wallet_device_registrations
		WHERE device_library_id = $1 AND pass_type_id = $2 AND serial_number = $3`,
		deviceID, passTypeID, serial)
	if err != nil {
		return fmt.Errorf("wallet.DeleteDeviceRegistration: %w", err)
	}
	return nil
}

// GetDeviceSerials returns serial numbers of passes registered to a device for a
// given pass type, updated strictly after sinceEpoch (0 = all). It also returns
// the max updated_at epoch across matched passes — the tag handed back to the device.
func (r *Wallet) GetDeviceSerials(ctx context.Context, deviceID, passTypeID string, sinceEpoch int64) ([]string, int64, error) {
	var rows []struct {
		Serial string `db:"serial_number"`
		Epoch  int64  `db:"epoch"`
	}
	err := r.pg.DB().SelectContext(ctx, &rows, `
		SELECT wp.serial_number, EXTRACT(EPOCH FROM wp.updated_at)::bigint AS epoch
		FROM wallet_device_registrations dr
		JOIN wallet_passes wp ON wp.serial_number = dr.serial_number
		WHERE dr.device_library_id = $1 AND dr.pass_type_id = $2
		  AND ($3 = 0 OR wp.updated_at > to_timestamp($3))`,
		deviceID, passTypeID, sinceEpoch)
	if err != nil {
		return nil, 0, fmt.Errorf("wallet.GetDeviceSerials: %w", err)
	}
	serials := make([]string, 0, len(rows))
	var maxEpoch int64
	for _, row := range rows {
		serials = append(serials, row.Serial)
		if row.Epoch > maxEpoch {
			maxEpoch = row.Epoch
		}
	}
	return serials, maxEpoch, nil
}

// GetRegistrationsBySerial returns all device registrations (with push tokens) for a pass.
func (r *Wallet) GetRegistrationsBySerial(ctx context.Context, serial string) ([]entity.WalletDeviceRegistration, error) {
	var regs []entity.WalletDeviceRegistration
	err := r.pg.DB().SelectContext(ctx, &regs, `
		SELECT * FROM wallet_device_registrations
		WHERE serial_number = $1 AND push_token IS NOT NULL`, serial)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetRegistrationsBySerial: %w", err)
	}
	return regs, nil
}
