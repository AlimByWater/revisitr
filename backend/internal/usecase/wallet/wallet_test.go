package wallet

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"revisitr/internal/entity"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ── Mock repos ──────────────────────────────────────────────────────────────

type mockConfigRepo struct {
	getConfigs func(ctx context.Context, orgID int) ([]entity.WalletConfig, error)
	getConfig  func(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error)
	saveConfig func(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error)
	deleteConf func(ctx context.Context, orgID int, platform string) error
}

func (m *mockConfigRepo) GetConfigs(ctx context.Context, orgID int) ([]entity.WalletConfig, error) {
	return m.getConfigs(ctx, orgID)
}
func (m *mockConfigRepo) GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
	return m.getConfig(ctx, orgID, platform)
}
func (m *mockConfigRepo) SaveConfig(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error) {
	return m.saveConfig(ctx, orgID, req)
}
func (m *mockConfigRepo) DeleteConfig(ctx context.Context, orgID int, platform string) error {
	return m.deleteConf(ctx, orgID, platform)
}

type mockPassRepo struct {
	createPass           func(ctx context.Context, pass *entity.WalletPass) error
	getPassBySerial      func(ctx context.Context, serial string) (*entity.WalletPass, error)
	getPassByClientPlat  func(ctx context.Context, clientID int, platform string) (*entity.WalletPass, error)
	getPassesByOrg       func(ctx context.Context, orgID int) ([]entity.WalletPass, error)
	getPassesByClient    func(ctx context.Context, clientID int) ([]entity.WalletPass, error)
	updatePassBalance    func(ctx context.Context, id int, balance int, level string) error
	updatePushToken      func(ctx context.Context, serial string, token string) error
	revokePass           func(ctx context.Context, id int) error
	getStats             func(ctx context.Context, orgID int) (*entity.WalletStats, error)
	getPassesWithPush    func(ctx context.Context, orgID int) ([]entity.WalletPass, error)
}

func (m *mockPassRepo) CreatePass(ctx context.Context, pass *entity.WalletPass) error {
	return m.createPass(ctx, pass)
}
func (m *mockPassRepo) GetPassBySerial(ctx context.Context, serial string) (*entity.WalletPass, error) {
	return m.getPassBySerial(ctx, serial)
}
func (m *mockPassRepo) GetPassByClientPlatform(ctx context.Context, clientID int, platform string) (*entity.WalletPass, error) {
	return m.getPassByClientPlat(ctx, clientID, platform)
}
func (m *mockPassRepo) GetPassesByOrg(ctx context.Context, orgID int) ([]entity.WalletPass, error) {
	return m.getPassesByOrg(ctx, orgID)
}
func (m *mockPassRepo) GetPassesByClient(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
	return m.getPassesByClient(ctx, clientID)
}
func (m *mockPassRepo) UpdatePassBalance(ctx context.Context, id int, balance int, level string) error {
	return m.updatePassBalance(ctx, id, balance, level)
}
func (m *mockPassRepo) UpdatePushToken(ctx context.Context, serial string, token string) error {
	return m.updatePushToken(ctx, serial, token)
}
func (m *mockPassRepo) RevokePass(ctx context.Context, id int) error {
	return m.revokePass(ctx, id)
}
func (m *mockPassRepo) GetStats(ctx context.Context, orgID int) (*entity.WalletStats, error) {
	return m.getStats(ctx, orgID)
}
func (m *mockPassRepo) GetPassesWithPushToken(ctx context.Context, orgID int) ([]entity.WalletPass, error) {
	return m.getPassesWithPush(ctx, orgID)
}

func newTestUC(configs *mockConfigRepo, passes *mockPassRepo) *Usecase {
	uc := New(configs, passes)
	_ = uc.Init(context.Background(), testLogger())
	return uc
}

// ── Config tests ────────────────────────────────────────────────────────────

func TestGetConfigs(t *testing.T) {
	configs := &mockConfigRepo{
		getConfigs: func(_ context.Context, orgID int) ([]entity.WalletConfig, error) {
			return []entity.WalletConfig{{ID: 1, OrgID: orgID, Platform: "apple"}}, nil
		},
	}
	uc := newTestUC(configs, &mockPassRepo{})
	result, err := uc.GetConfigs(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Platform != "apple" {
		t.Fatalf("unexpected configs: %+v", result)
	}
}

func TestGetConfig_NotFound(t *testing.T) {
	configs := &mockConfigRepo{
		getConfig: func(_ context.Context, _ int, _ string) (*entity.WalletConfig, error) {
			return nil, nil
		},
	}
	uc := newTestUC(configs, &mockPassRepo{})
	_, err := uc.GetConfig(context.Background(), 1, "apple")
	if err != ErrConfigNotFound {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestSaveConfig_InvalidPlatform(t *testing.T) {
	uc := newTestUC(&mockConfigRepo{}, &mockPassRepo{})
	_, err := uc.SaveConfig(context.Background(), 1, entity.SaveWalletConfigRequest{Platform: "windows"})
	if err != ErrInvalidPlatform {
		t.Fatalf("expected ErrInvalidPlatform, got %v", err)
	}
}

func TestSaveConfig_Success(t *testing.T) {
	configs := &mockConfigRepo{
		saveConfig: func(_ context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{ID: 1, OrgID: orgID, Platform: req.Platform, IsEnabled: req.IsEnabled}, nil
		},
	}
	uc := newTestUC(configs, &mockPassRepo{})
	cfg, err := uc.SaveConfig(context.Background(), 1, entity.SaveWalletConfigRequest{
		Platform:  "apple",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.IsEnabled || cfg.Platform != "apple" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

// ── Pass tests ──────────────────────────────────────────────────────────────

func TestIssuePass_PlatformDisabled(t *testing.T) {
	configs := &mockConfigRepo{
		getConfig: func(_ context.Context, _ int, _ string) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{IsEnabled: false}, nil
		},
	}
	uc := newTestUC(configs, &mockPassRepo{})
	_, err := uc.IssuePass(context.Background(), 1, entity.IssueWalletPassRequest{
		ClientID: 10, Platform: "apple",
	})
	if err != ErrPlatformDisabled {
		t.Fatalf("expected ErrPlatformDisabled, got %v", err)
	}
}

func TestIssuePass_NoConfig(t *testing.T) {
	configs := &mockConfigRepo{
		getConfig: func(_ context.Context, _ int, _ string) (*entity.WalletConfig, error) {
			return nil, nil
		},
	}
	uc := newTestUC(configs, &mockPassRepo{})
	_, err := uc.IssuePass(context.Background(), 1, entity.IssueWalletPassRequest{
		ClientID: 10, Platform: "google",
	})
	if err != ErrPlatformDisabled {
		t.Fatalf("expected ErrPlatformDisabled, got %v", err)
	}
}

func TestIssuePass_AlreadyExists(t *testing.T) {
	configs := &mockConfigRepo{
		getConfig: func(_ context.Context, _ int, _ string) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{IsEnabled: true}, nil
		},
	}
	passes := &mockPassRepo{
		getPassByClientPlat: func(_ context.Context, _ int, _ string) (*entity.WalletPass, error) {
			return &entity.WalletPass{ID: 1}, nil
		},
	}
	uc := newTestUC(configs, passes)
	_, err := uc.IssuePass(context.Background(), 1, entity.IssueWalletPassRequest{
		ClientID: 10, Platform: "apple",
	})
	if err != ErrPassAlreadyExists {
		t.Fatalf("expected ErrPassAlreadyExists, got %v", err)
	}
}

func TestIssuePass_InvalidPlatform(t *testing.T) {
	uc := newTestUC(&mockConfigRepo{}, &mockPassRepo{})
	_, err := uc.IssuePass(context.Background(), 1, entity.IssueWalletPassRequest{
		ClientID: 10, Platform: "blackberry",
	})
	if err != ErrInvalidPlatform {
		t.Fatalf("expected ErrInvalidPlatform, got %v", err)
	}
}

func TestIssuePass_Success(t *testing.T) {
	configs := &mockConfigRepo{
		getConfig: func(_ context.Context, _ int, _ string) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{IsEnabled: true}, nil
		},
	}
	var createdPass *entity.WalletPass
	passes := &mockPassRepo{
		getPassByClientPlat: func(_ context.Context, _ int, _ string) (*entity.WalletPass, error) {
			return nil, nil
		},
		createPass: func(_ context.Context, pass *entity.WalletPass) error {
			pass.ID = 42
			createdPass = pass
			return nil
		},
	}
	uc := newTestUC(configs, passes)
	result, err := uc.IssuePass(context.Background(), 1, entity.IssueWalletPassRequest{
		ClientID: 10, Platform: "google",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 42 || result.Platform != "google" || result.ClientID != 10 {
		t.Fatalf("unexpected pass: %+v", result)
	}
	if createdPass.SerialNumber == "" || createdPass.AuthToken == "" {
		t.Fatal("serial and auth token must be generated")
	}
}

func TestGetPass_NotFound(t *testing.T) {
	passes := &mockPassRepo{
		getPassBySerial: func(_ context.Context, _ string) (*entity.WalletPass, error) {
			return nil, nil
		},
	}
	uc := newTestUC(&mockConfigRepo{}, passes)
	_, err := uc.GetPass(context.Background(), "nonexistent")
	if err != ErrPassNotFound {
		t.Fatalf("expected ErrPassNotFound, got %v", err)
	}
}

func TestRegisterPushToken_WrongAuth(t *testing.T) {
	passes := &mockPassRepo{
		getPassBySerial: func(_ context.Context, _ string) (*entity.WalletPass, error) {
			return &entity.WalletPass{ID: 1, AuthToken: "correct-token"}, nil
		},
	}
	uc := newTestUC(&mockConfigRepo{}, passes)
	err := uc.RegisterPushToken(context.Background(), "serial123", "wrong-token", "push-token")
	if err != ErrPassNotFound {
		t.Fatalf("expected ErrPassNotFound for wrong auth, got %v", err)
	}
}

func TestRegisterPushToken_Success(t *testing.T) {
	var updatedSerial, updatedToken string
	passes := &mockPassRepo{
		getPassBySerial: func(_ context.Context, serial string) (*entity.WalletPass, error) {
			return &entity.WalletPass{ID: 1, SerialNumber: serial, AuthToken: "correct-token"}, nil
		},
		updatePushToken: func(_ context.Context, serial string, token string) error {
			updatedSerial = serial
			updatedToken = token
			return nil
		},
	}
	uc := newTestUC(&mockConfigRepo{}, passes)
	err := uc.RegisterPushToken(context.Background(), "serial123", "correct-token", "new-push-token")
	if err != nil {
		t.Fatal(err)
	}
	if updatedSerial != "serial123" || updatedToken != "new-push-token" {
		t.Fatalf("unexpected update: serial=%s token=%s", updatedSerial, updatedToken)
	}
}

func TestRefreshPassBalance(t *testing.T) {
	var updates []int
	passes := &mockPassRepo{
		getPassesByClient: func(_ context.Context, _ int) ([]entity.WalletPass, error) {
			return []entity.WalletPass{
				{ID: 1, Status: "active"},
				{ID: 2, Status: "revoked"},
				{ID: 3, Status: "active"},
			}, nil
		},
		updatePassBalance: func(_ context.Context, id int, _ int, _ string) error {
			updates = append(updates, id)
			return nil
		},
	}
	uc := newTestUC(&mockConfigRepo{}, passes)
	err := uc.RefreshPassBalance(context.Background(), 10, 500, "Gold")
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 2 || updates[0] != 1 || updates[1] != 3 {
		t.Fatalf("expected updates for passes 1 and 3, got %v", updates)
	}
}

func TestGetStats(t *testing.T) {
	passes := &mockPassRepo{
		getStats: func(_ context.Context, _ int) (*entity.WalletStats, error) {
			return &entity.WalletStats{TotalPasses: 10, ApplePasses: 6, GooglePasses: 4, ActivePasses: 8}, nil
		},
	}
	uc := newTestUC(&mockConfigRepo{}, passes)
	stats, err := uc.GetStats(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPasses != 10 || stats.ActivePasses != 8 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
