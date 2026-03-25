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
