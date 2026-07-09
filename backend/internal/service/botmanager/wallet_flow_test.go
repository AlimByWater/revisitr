package botmanager

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"revisitr/internal/entity"
)

func testLoggerBM() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ── Mocks ────────────────────────────────────────────────────────────────────

type mockWalletUsecase struct {
	getConfig          func(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error)
	issuePass          func(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error)
	getClientPasses    func(ctx context.Context, clientID int) ([]entity.WalletPass, error)
	refreshPassBalance func(ctx context.Context, clientID int, balance int, level string) error
	getClientsQRCode   func(ctx context.Context, clientID int) (string, error)
	getOrgName         func(ctx context.Context, orgID int) (string, error)
	generateGoogleURL  func(ctx context.Context, orgID int, pass *entity.WalletPass) (string, error)
	issuePassCalled    bool
}

func (m *mockWalletUsecase) GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
	return m.getConfig(ctx, orgID, platform)
}
func (m *mockWalletUsecase) IssuePass(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
	m.issuePassCalled = true
	return m.issuePass(ctx, orgID, req)
}
func (m *mockWalletUsecase) GetClientPasses(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
	return m.getClientPasses(ctx, clientID)
}
func (m *mockWalletUsecase) RefreshPassBalance(ctx context.Context, clientID int, balance int, level string) error {
	return m.refreshPassBalance(ctx, clientID, balance, level)
}
func (m *mockWalletUsecase) GetClientsQRCode(ctx context.Context, clientID int) (string, error) {
	return m.getClientsQRCode(ctx, clientID)
}
func (m *mockWalletUsecase) GetOrgName(ctx context.Context, orgID int) (string, error) {
	return m.getOrgName(ctx, orgID)
}
func (m *mockWalletUsecase) GenerateGoogleSaveURL(ctx context.Context, orgID int, pass *entity.WalletPass) (string, error) {
	return m.generateGoogleURL(ctx, orgID, pass)
}

type mockLoyaltyRepo struct {
	getProgramsByOrgID func(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	getClientLoyalty   func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	getLevelsByProgram func(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error)
}

func (m *mockLoyaltyRepo) GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
	return m.getProgramsByOrgID(ctx, orgID)
}
func (m *mockLoyaltyRepo) GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	return m.getClientLoyalty(ctx, clientID, programID)
}
func (m *mockLoyaltyRepo) GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error) {
	return m.getLevelsByProgram(ctx, programID)
}
func (m *mockLoyaltyRepo) UpsertClientLoyalty(ctx context.Context, cl *entity.ClientLoyalty) error {
	return nil
}
func (m *mockLoyaltyRepo) CreateTransaction(ctx context.Context, tx *entity.LoyaltyTransaction) error {
	return nil
}

func testHandler(wallet walletUsecase, loyalty loyaltyRepository) *handler {
	return &handler{
		mgr:    &Manager{wallet: wallet, loyaltyRepo: loyalty},
		info:   &entity.Bot{ID: 1, OrgID: 42},
		logger: testLoggerBM(),
	}
}

// ── getOrIssuePass ───────────────────────────────────────────────────────────

func TestGetOrIssuePass_ReusesExisting(t *testing.T) {
	existing := entity.WalletPass{ID: 7, ClientID: 5, Platform: "apple", SerialNumber: "abc", Status: "active"}
	wallet := &mockWalletUsecase{
		getClientPasses: func(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
			return []entity.WalletPass{
				{ID: 1, ClientID: 5, Platform: "google"},
				existing,
			}, nil
		},
		issuePass: func(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
			t.Fatal("IssuePass should not be called when a pass already exists")
			return nil, nil
		},
	}
	h := testHandler(wallet, nil)

	pass, err := h.getOrIssuePass(context.Background(), 5, "apple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass.SerialNumber != "abc" {
		t.Fatalf("expected existing pass to be reused, got serial %q", pass.SerialNumber)
	}
	if wallet.issuePassCalled {
		t.Fatal("IssuePass should not have been called")
	}
}

func TestGetOrIssuePass_IssuesNewWhenAbsent(t *testing.T) {
	wallet := &mockWalletUsecase{
		getClientPasses: func(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
			return nil, nil
		},
		issuePass: func(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
			if orgID != 42 {
				t.Fatalf("expected org id 42, got %d", orgID)
			}
			if req.Platform != "apple" || req.ClientID != 5 {
				t.Fatalf("unexpected issue request: %+v", req)
			}
			return &entity.WalletPass{ID: 9, ClientID: 5, Platform: "apple", SerialNumber: "new", Status: "active"}, nil
		},
	}
	h := testHandler(wallet, nil)

	pass, err := h.getOrIssuePass(context.Background(), 5, "apple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wallet.issuePassCalled {
		t.Fatal("expected IssuePass to be called")
	}
	if pass.SerialNumber != "new" {
		t.Fatalf("expected newly issued pass, got serial %q", pass.SerialNumber)
	}
}

func TestGetOrIssuePass_PropagatesIssueError(t *testing.T) {
	wallet := &mockWalletUsecase{
		getClientPasses: func(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
			return nil, nil
		},
		issuePass: func(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
			return nil, errors.New("platform disabled")
		},
	}
	h := testHandler(wallet, nil)

	if _, err := h.getOrIssuePass(context.Background(), 5, "apple"); err == nil {
		t.Fatal("expected error to propagate")
	}
}

// ── currentLoyaltySnapshot ───────────────────────────────────────────────────

func TestCurrentLoyaltySnapshot_ReturnsBalanceAndLevel(t *testing.T) {
	levelID := 3
	loyalty := &mockLoyaltyRepo{
		getProgramsByOrgID: func(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
			return []entity.LoyaltyProgram{{ID: 1, IsActive: true}}, nil
		},
		getClientLoyalty: func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{ClientID: clientID, ProgramID: programID, Balance: 150, LevelID: &levelID}, nil
		},
		getLevelsByProgram: func(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error) {
			return []entity.LoyaltyLevel{{ID: 3, Name: "Gold"}}, nil
		},
	}
	h := testHandler(nil, loyalty)

	balance, level := h.currentLoyaltySnapshot(context.Background(), 5)
	if balance != 150 {
		t.Fatalf("expected balance 150, got %d", balance)
	}
	if level != "Gold" {
		t.Fatalf("expected level Gold, got %q", level)
	}
}

func TestCurrentLoyaltySnapshot_NoActiveProgram(t *testing.T) {
	loyalty := &mockLoyaltyRepo{
		getProgramsByOrgID: func(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
			return []entity.LoyaltyProgram{{ID: 1, IsActive: false}}, nil
		},
	}
	h := testHandler(nil, loyalty)

	balance, level := h.currentLoyaltySnapshot(context.Background(), 5)
	if balance != 0 || level != "" {
		t.Fatalf("expected zero snapshot, got balance=%d level=%q", balance, level)
	}
}

// ── appleWalletButtonRow ─────────────────────────────────────────────────────

func TestAppleWalletButtonRow_HiddenWhenDisabled(t *testing.T) {
	wallet := &mockWalletUsecase{
		getConfig: func(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{IsEnabled: false}, nil
		},
	}
	h := testHandler(wallet, nil)

	if rows := h.appleWalletButtonRow(context.Background()); rows != nil {
		t.Fatalf("expected no button row, got %+v", rows)
	}
}

func TestAppleWalletButtonRow_ShownWhenEnabled(t *testing.T) {
	wallet := &mockWalletUsecase{
		getConfig: func(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
			return &entity.WalletConfig{IsEnabled: true}, nil
		},
	}
	h := testHandler(wallet, nil)

	rows := h.appleWalletButtonRow(context.Background())
	if len(rows) != 1 || len(rows[0]) != 1 {
		t.Fatalf("expected exactly one button, got %+v", rows)
	}
	if rows[0][0].Data != callbackWalletAddApple {
		t.Fatalf("unexpected callback data: %q", rows[0][0].Data)
	}
}

// ── googleWalletButtonRow ────────────────────────────────────────────────────

func TestGoogleWalletButtonRow_ShownWhenEnabled(t *testing.T) {
	wallet := &mockWalletUsecase{
		getConfig: func(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error) {
			if platform != "google" {
				t.Fatalf("expected platform google, got %q", platform)
			}
			return &entity.WalletConfig{IsEnabled: true}, nil
		},
	}
	h := testHandler(wallet, nil)

	rows := h.googleWalletButtonRow(context.Background())
	if len(rows) != 1 || len(rows[0]) != 1 {
		t.Fatalf("expected exactly one button, got %+v", rows)
	}
	if rows[0][0].Data != callbackWalletAddGoogle {
		t.Fatalf("unexpected callback data: %q", rows[0][0].Data)
	}
}

func TestGetOrIssuePass_MatchesRequestedPlatformOnly(t *testing.T) {
	wallet := &mockWalletUsecase{
		getClientPasses: func(ctx context.Context, clientID int) ([]entity.WalletPass, error) {
			return []entity.WalletPass{
				{ID: 1, ClientID: 5, Platform: "apple", SerialNumber: "apple-serial"},
			}, nil
		},
		issuePass: func(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error) {
			if req.Platform != "google" {
				t.Fatalf("expected to issue a google pass, got %q", req.Platform)
			}
			return &entity.WalletPass{ID: 2, ClientID: 5, Platform: "google", SerialNumber: "google-serial", Status: "active"}, nil
		},
	}
	h := testHandler(wallet, nil)

	pass, err := h.getOrIssuePass(context.Background(), 5, "google")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass.SerialNumber != "google-serial" {
		t.Fatalf("expected a new google pass to be issued, got %q", pass.SerialNumber)
	}
	if !wallet.issuePassCalled {
		t.Fatal("expected IssuePass to be called since only an apple pass existed")
	}
}
