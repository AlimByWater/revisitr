package wallet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrConfigNotFound     = errors.New("wallet config not found")
	ErrPlatformDisabled   = errors.New("wallet platform is not enabled")
	ErrPassAlreadyExists  = errors.New("client already has a pass for this platform")
	ErrPassNotFound       = errors.New("wallet pass not found")
	ErrInvalidPlatform    = errors.New("invalid platform: must be apple or google")
)

type configRepo interface {
	GetConfigs(ctx context.Context, orgID int) ([]entity.WalletConfig, error)
	GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error)
	SaveConfig(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error)
	DeleteConfig(ctx context.Context, orgID int, platform string) error
}

type passRepo interface {
	CreatePass(ctx context.Context, pass *entity.WalletPass) error
	GetPassBySerial(ctx context.Context, serial string) (*entity.WalletPass, error)
	GetPassByClientPlatform(ctx context.Context, clientID int, platform string) (*entity.WalletPass, error)
	GetPassesByOrg(ctx context.Context, orgID int) ([]entity.WalletPass, error)
	GetPassesByClient(ctx context.Context, clientID int) ([]entity.WalletPass, error)
	UpdatePassBalance(ctx context.Context, id int, balance int, level string) error
	UpdatePushToken(ctx context.Context, serial string, token string) error
	RevokePass(ctx context.Context, id int) error
	GetStats(ctx context.Context, orgID int) (*entity.WalletStats, error)
	GetPassesWithPushToken(ctx context.Context, orgID int) ([]entity.WalletPass, error)
}

type Usecase struct {
	logger  *slog.Logger
	configs configRepo
	passes  passRepo
}

func New(configs configRepo, passes passRepo) *Usecase {
	return &Usecase{configs: configs, passes: passes}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// ── Config ───────────────────────────────────────────────────────────────────

func (uc *Usecase) GetConfigs(ctx context.Context, orgID int) ([]entity.WalletConfig, error) {
	return uc.configs.GetConfigs(ctx, orgID)
}

func (uc *Usecase) GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
	cfg, err := uc.configs.GetConfig(ctx, orgID, platform)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrConfigNotFound
	}
	return cfg, nil
}

func (uc *Usecase) SaveConfig(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error) {
	if req.Platform != "apple" && req.Platform != "google" {
		return nil, ErrInvalidPlatform
	}
	return uc.configs.SaveConfig(ctx, orgID, req)
}

func (uc *Usecase) DeleteConfig(ctx context.Context, orgID int, platform string) error {
	return uc.configs.DeleteConfig(ctx, orgID, platform)
}

// ── Passes ───────────────────────────────────────────────────────────────────

func (uc *Usecase) IssuePass(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
	if req.Platform != "apple" && req.Platform != "google" {
		return nil, ErrInvalidPlatform
	}

	// Check platform is enabled
	cfg, err := uc.configs.GetConfig(ctx, orgID, req.Platform)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.IsEnabled {
		return nil, ErrPlatformDisabled
	}

	// Check no existing pass
	existing, err := uc.passes.GetPassByClientPlatform(ctx, req.ClientID, req.Platform)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPassAlreadyExists
	}

	serial, err := generateSerial()
	if err != nil {
		return nil, fmt.Errorf("generate serial: %w", err)
	}
	authToken, err := generateAuthToken()
	if err != nil {
		return nil, fmt.Errorf("generate auth token: %w", err)
	}

	pass := &entity.WalletPass{
		OrgID:        orgID,
		ClientID:     req.ClientID,
		Platform:     req.Platform,
		SerialNumber: serial,
		AuthToken:    authToken,
		LastBalance:  0,
		LastLevel:    "",
		Status:       "active",
	}

	if err := uc.passes.CreatePass(ctx, pass); err != nil {
		return nil, err
	}

	uc.logger.Info("wallet pass issued",
		"org_id", orgID, "client_id", req.ClientID,
		"platform", req.Platform, "serial", serial)

	return pass, nil
}

func (uc *Usecase) GetPass(ctx context.Context, serial string) (*entity.WalletPass, error) {
	pass, err := uc.passes.GetPassBySerial(ctx, serial)
	if err != nil {
		return nil, err
	}
	if pass == nil {
		return nil, ErrPassNotFound
	}
	return pass, nil
}

func (uc *Usecase) GetPasses(ctx context.Context, orgID int) ([]entity.WalletPass, error) {
	return uc.passes.GetPassesByOrg(ctx, orgID)
}

func (uc *Usecase) GetClientPasses(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
	return uc.passes.GetPassesByClient(ctx, clientID)
}

func (uc *Usecase) RegisterPushToken(ctx context.Context, serial string, authToken string, pushToken string) error {
	pass, err := uc.passes.GetPassBySerial(ctx, serial)
	if err != nil {
		return err
	}
	if pass == nil {
		return ErrPassNotFound
	}
	if pass.AuthToken != authToken {
		return ErrPassNotFound // Don't leak existence
	}
	return uc.passes.UpdatePushToken(ctx, serial, pushToken)
}

func (uc *Usecase) RevokePass(ctx context.Context, orgID int, passID int) error {
	return uc.passes.RevokePass(ctx, passID)
}

func (uc *Usecase) GetStats(ctx context.Context, orgID int) (*entity.WalletStats, error) {
	return uc.passes.GetStats(ctx, orgID)
}

// RefreshPassBalance updates a pass's cached balance/level — called after loyalty changes.
func (uc *Usecase) RefreshPassBalance(ctx context.Context, clientID int, balance int, level string) error {
	passes, err := uc.passes.GetPassesByClient(ctx, clientID)
	if err != nil {
		return err
	}
	for _, p := range passes {
		if p.Status != "active" {
			continue
		}
		if err := uc.passes.UpdatePassBalance(ctx, p.ID, balance, level); err != nil {
			uc.logger.Warn("wallet: failed to update pass balance",
				"pass_id", p.ID, "client_id", clientID, "error", err)
		}
	}
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func generateSerial() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateAuthToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
