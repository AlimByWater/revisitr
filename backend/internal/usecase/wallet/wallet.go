package wallet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

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
	GetClientQRCode(ctx context.Context, clientID int) (string, error)
	GetOrgName(ctx context.Context, orgID int) (string, error)
	CreateDeviceRegistration(ctx context.Context, reg *entity.WalletDeviceRegistration) (bool, error)
	DeleteDeviceRegistration(ctx context.Context, deviceID, passTypeID, serial string) error
	GetDeviceSerials(ctx context.Context, deviceID, passTypeID string, sinceEpoch int64) ([]string, int64, error)
	GetRegistrationsBySerial(ctx context.Context, serial string) ([]entity.WalletDeviceRegistration, error)
}

type Usecase struct {
	logger   *slog.Logger
	configs  configRepo
	passes   passRepo
	googleGW *GoogleSaveGenerator
	googleAPI *GoogleWalletAPI
	apns     *APNsClient
}

func New(configs configRepo, passes passRepo) *Usecase {
	return &Usecase{
		configs:  configs,
		passes:   passes,
		googleGW: NewGoogleSaveGenerator(),
		googleAPI: NewGoogleWalletAPI(),
		apns:     NewAPNsClient(),
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	uc.googleAPI.Init(logger)
	uc.apns.Init(logger)
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

	cfg, err := uc.configs.SaveConfig(ctx, orgID, req)
	if err != nil {
		return nil, err
	}

	if req.Platform == "google" && req.IsEnabled && cfg.Credentials["issuer_id"] != "" && cfg.Credentials["service_account_key"] != "" {
		orgName := req.Design.OrganizationName
		if orgName == "" {
			orgName, _ = uc.passes.GetOrgName(ctx, orgID)
		}
		go func() {
			if err := uc.googleAPI.EnsureClass(context.Background(), cfg.Credentials, orgID, orgName, req.Design); err != nil {
				uc.logger.Error("google wallet: ensure class", "error", err, "org_id", orgID)
			}
		}()
	}

	return cfg, nil
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

	if req.Platform == "google" {
		clientQR, _ := uc.passes.GetClientQRCode(ctx, req.ClientID)
		orgName, _ := uc.passes.GetOrgName(ctx, orgID)
		if err := uc.googleAPI.CreateObject(ctx, cfg.Credentials, orgID, pass, clientQR, orgName, cfg.Design); err != nil {
			uc.logger.Error("google wallet: create object", "error", err, "org_id", orgID, "client_id", req.ClientID)
		}
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
			continue
		}
		// Notify Apple Wallet devices to pull the updated pass (fire-and-forget).
		if p.Platform == "apple" {
			go uc.pushAppleUpdate(context.Background(), p.OrgID, p.SerialNumber)
		}
	}
	return nil
}

// pushAppleUpdate sends an APNs update push to every device registered for a
// pass. No-op if APNs is not configured for the org.
func (uc *Usecase) pushAppleUpdate(ctx context.Context, orgID int, serial string) {
	cfg, err := uc.configs.GetConfig(ctx, orgID, "apple")
	if err != nil || cfg == nil || !cfg.IsEnabled {
		return
	}
	if cfg.Credentials["apns_key"] == "" || cfg.Credentials["apns_key_id"] == "" {
		return // push updates not configured
	}
	regs, err := uc.passes.GetRegistrationsBySerial(ctx, serial)
	if err != nil {
		uc.logger.Warn("wallet apns: get registrations", "serial", serial, "error", err)
		return
	}
	for _, reg := range regs {
		if reg.PushToken == nil {
			continue
		}
		if err := uc.apns.SendPush(ctx, cfg.Credentials, *reg.PushToken); err != nil {
			uc.logger.Warn("wallet apns: push failed",
				"serial", serial, "device", reg.DeviceLibraryID, "error", err)
		}
	}
}

// ── Apple Wallet device registration ─────────────────────────────────────────

// RegisterDevice records a device's push token for a pass. Returns true if the
// registration is new (HTTP 201) versus already existing (HTTP 200).
func (uc *Usecase) RegisterDevice(ctx context.Context, deviceLibraryID, passTypeID, serial, authToken, pushToken string) (bool, error) {
	pass, err := uc.passes.GetPassBySerial(ctx, serial)
	if err != nil {
		return false, err
	}
	if pass == nil || pass.AuthToken != authToken {
		return false, ErrPassNotFound
	}
	return uc.passes.CreateDeviceRegistration(ctx, &entity.WalletDeviceRegistration{
		OrgID:           pass.OrgID,
		DeviceLibraryID: deviceLibraryID,
		PassTypeID:      passTypeID,
		SerialNumber:    serial,
		PushToken:       &pushToken,
		AuthToken:       authToken,
	})
}

func (uc *Usecase) UnregisterDevice(ctx context.Context, deviceLibraryID, passTypeID, serial, authToken string) error {
	pass, err := uc.passes.GetPassBySerial(ctx, serial)
	if err != nil {
		return err
	}
	if pass == nil || pass.AuthToken != authToken {
		return ErrPassNotFound
	}
	return uc.passes.DeleteDeviceRegistration(ctx, deviceLibraryID, passTypeID, serial)
}

// GetDeviceSerials returns the serial numbers of passes for a device that changed
// since the given tag, plus the new tag to persist. Empty since = all passes.
func (uc *Usecase) GetDeviceSerials(ctx context.Context, deviceLibraryID, passTypeID, passesUpdatedSince string) ([]string, string, error) {
	var since int64
	if passesUpdatedSince != "" {
		since, _ = strconv.ParseInt(passesUpdatedSince, 10, 64)
	}
	serials, maxEpoch, err := uc.passes.GetDeviceSerials(ctx, deviceLibraryID, passTypeID, since)
	if err != nil {
		return nil, "", err
	}
	tag := passesUpdatedSince
	if maxEpoch > 0 {
		tag = strconv.FormatInt(maxEpoch, 10)
	}
	return serials, tag, nil
}

// GetClientsQRCode returns the QR code string for a client.
func (uc *Usecase) GetClientsQRCode(ctx context.Context, clientID int) (string, error) {
	return uc.passes.GetClientQRCode(ctx, clientID)
}

// GetOrgName returns the organization name by ID.
func (uc *Usecase) GetOrgName(ctx context.Context, orgID int) (string, error) {
	return uc.passes.GetOrgName(ctx, orgID)
}

// GenerateGoogleSaveURL creates a signed "Add to Google Wallet" JWT link.
func (uc *Usecase) GenerateGoogleSaveURL(ctx context.Context, orgID int, pass *entity.WalletPass) (string, error) {
	if pass.Platform != "google" {
		return "", ErrInvalidPlatform
	}
	if pass.Status != "active" {
		return "", ErrPassNotFound
	}

	cfg, err := uc.configs.GetConfig(ctx, orgID, "google")
	if err != nil {
		return "", fmt.Errorf("get google config: %w", err)
	}
	if cfg == nil || !cfg.IsEnabled {
		return "", ErrPlatformDisabled
	}

	return uc.googleAPI.GenerateSaveURL(cfg.Credentials, orgID, pass)
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
