package bots

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockBotsRepo struct {
	createFn     func(ctx context.Context, bot *entity.Bot) error
	getByIDFn    func(ctx context.Context, id int) (*entity.Bot, error)
	getByOrgIDFn func(ctx context.Context, orgID int) ([]entity.Bot, error)
	hasPOSFn     func(ctx context.Context, botID int) (bool, error)
	updateFn     func(ctx context.Context, bot *entity.Bot) error
	updateSettFn func(ctx context.Context, id int, s entity.BotSettings) error
	deleteFn     func(ctx context.Context, id int) error
}

func (m *mockBotsRepo) Create(ctx context.Context, bot *entity.Bot) error {
	return m.createFn(ctx, bot)
}
func (m *mockBotsRepo) GetByID(ctx context.Context, id int) (*entity.Bot, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockBotsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockBotsRepo) HasPOSLocations(ctx context.Context, botID int) (bool, error) {
	if m.hasPOSFn != nil {
		return m.hasPOSFn(ctx, botID)
	}
	return false, nil
}
func (m *mockBotsRepo) Update(ctx context.Context, bot *entity.Bot) error {
	return m.updateFn(ctx, bot)
}
func (m *mockBotsRepo) UpdateSettings(ctx context.Context, id int, s entity.BotSettings) error {
	return m.updateSettFn(ctx, id, s)
}
func (m *mockBotsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

type mockBotClientsRepo struct {
	countByBotIDFn func(ctx context.Context, botID int) (int, error)
}

func (m *mockBotClientsRepo) CountByBotID(ctx context.Context, botID int) (int, error) {
	if m.countByBotIDFn != nil {
		return m.countByBotIDFn(ctx, botID)
	}
	return 0, nil
}

// --- helpers ---

func newUC(bots botsRepo, clients botClientsRepo) *Usecase {
	uc := New(bots, clients)
	_ = uc.Init(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return uc
}

func testBot(id, orgID int) *entity.Bot {
	return &entity.Bot{
		ID:    id,
		OrgID: orgID,
		Name:  "Test Bot",
		Token: "test-token",
		Settings: entity.BotSettings{
			Modules:          []string{"loyalty"},
			Buttons:          []entity.BotButton{{Label: "Menu", Type: "url", Value: "https://example.com"}},
			RegistrationForm: []entity.FormField{},
			WelcomeMessage:   "Hello!",
		},
	}
}

func ptr[T any](v T) *T { return &v }

// --- Create ---

func TestCreate_Success(t *testing.T) {
	repo := &mockBotsRepo{
		createFn: func(_ context.Context, b *entity.Bot) error {
			b.ID = 42
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	bot, err := uc.Create(context.Background(), 10, &entity.CreateBotRequest{
		Name:  "My Bot",
		Token: "tok123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bot.ID != 42 {
		t.Errorf("got ID %d, want 42", bot.ID)
	}
	if bot.OrgID != 10 {
		t.Errorf("got OrgID %d, want 10", bot.OrgID)
	}
	if bot.Status != "pending" {
		t.Errorf("got Status %q, want %q", bot.Status, "pending")
	}
	if len(bot.Settings.Modules) != 0 {
		t.Errorf("got %d modules, want empty", len(bot.Settings.Modules))
	}
	if len(bot.Settings.Buttons) != 0 {
		t.Errorf("got %d buttons, want empty", len(bot.Settings.Buttons))
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := &mockBotsRepo{
		createFn: func(_ context.Context, _ *entity.Bot) error {
			return errors.New("db error")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.Create(context.Background(), 10, &entity.CreateBotRequest{Name: "Bot", Token: "tok"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetByOrgID ---

func TestGetByOrgID_Success(t *testing.T) {
	repo := &mockBotsRepo{
		getByOrgIDFn: func(_ context.Context, orgID int) ([]entity.Bot, error) {
			return []entity.Bot{{ID: 1, OrgID: orgID}, {ID: 2, OrgID: orgID}}, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	bots, err := uc.GetByOrgID(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bots) != 2 {
		t.Errorf("got %d bots, want 2", len(bots))
	}
}

func TestGetByOrgID_RepoError(t *testing.T) {
	repo := &mockBotsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Bot, error) {
			return nil, errors.New("db error")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.GetByOrgID(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetByID ---

func TestGetByID_Success(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	got, err := uc.GetByID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("got ID %d, want 1", got.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.GetByID(context.Background(), 999, 10)
	if !errors.Is(err, ErrBotNotFound) {
		t.Errorf("got %v, want ErrBotNotFound", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.GetByID(context.Background(), 1, 999)
	if !errors.Is(err, ErrNotBotOwner) {
		t.Errorf("got %v, want ErrNotBotOwner", err)
	}
}

// --- Update ---

func TestUpdate_NameOnly(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateFn: func(_ context.Context, b *entity.Bot) error {
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	got, err := uc.Update(context.Background(), 1, 10, &entity.UpdateBotRequest{
		Name: ptr("New Name"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "New Name" {
		t.Errorf("got Name %q, want %q", got.Name, "New Name")
	}
}

func TestUpdate_StatusOnly(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		hasPOSFn: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
		updateFn: func(_ context.Context, _ *entity.Bot) error {
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	got, err := uc.Update(context.Background(), 1, 10, &entity.UpdateBotRequest{
		Status: ptr("active"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("got Status %q, want %q", got.Status, "active")
	}
}

func TestUpdate_StatusForcedPendingWithoutPOS(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		hasPOSFn: func(_ context.Context, _ int) (bool, error) {
			return false, nil
		},
		updateFn: func(_ context.Context, _ *entity.Bot) error {
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	got, err := uc.Update(context.Background(), 1, 10, &entity.UpdateBotRequest{
		Status: ptr("active"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "pending" {
		t.Errorf("got Status %q, want %q", got.Status, "pending")
	}
}

func TestUpdate_ProgramID(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateFn: func(_ context.Context, _ *entity.Bot) error {
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	programID := 42
	got, err := uc.Update(context.Background(), 1, 10, &entity.UpdateBotRequest{
		ProgramID: &programID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProgramID == nil || *got.ProgramID != 42 {
		t.Fatalf("got ProgramID %#v, want 42", got.ProgramID)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.Update(context.Background(), 999, 10, &entity.UpdateBotRequest{Name: ptr("x")})
	if !errors.Is(err, ErrBotNotFound) {
		t.Errorf("got %v, want ErrBotNotFound", err)
	}
}

func TestUpdate_WrongOrg(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.Update(context.Background(), 1, 999, &entity.UpdateBotRequest{Name: ptr("x")})
	if !errors.Is(err, ErrNotBotOwner) {
		t.Errorf("got %v, want ErrNotBotOwner", err)
	}
}

func TestUpdate_RepoError(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateFn: func(_ context.Context, _ *entity.Bot) error {
			return errors.New("db error")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.Update(context.Background(), 1, 10, &entity.UpdateBotRequest{Name: ptr("x")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	bot := testBot(1, 10)
	deleted := false
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		deleteFn: func(_ context.Context, id int) error {
			deleted = true
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.Delete(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("repo.Delete was not called")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.Delete(context.Background(), 999, 10)
	if !errors.Is(err, ErrBotNotFound) {
		t.Errorf("got %v, want ErrBotNotFound", err)
	}
}

func TestDelete_WrongOrg(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.Delete(context.Background(), 1, 999)
	if !errors.Is(err, ErrNotBotOwner) {
		t.Errorf("got %v, want ErrNotBotOwner", err)
	}
}

func TestDelete_RepoError(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		deleteFn: func(_ context.Context, _ int) error {
			return errors.New("db error")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.Delete(context.Background(), 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetSettings ---

func TestGetSettings_Success(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	settings, err := uc.GetSettings(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.WelcomeMessage != "Hello!" {
		t.Errorf("got WelcomeMessage %q, want %q", settings.WelcomeMessage, "Hello!")
	}
}

func TestGetSettings_NotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.GetSettings(context.Background(), 999, 10)
	if !errors.Is(err, ErrBotNotFound) {
		t.Errorf("got %v, want ErrBotNotFound", err)
	}
}

func TestGetSettings_WrongOrg(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	_, err := uc.GetSettings(context.Background(), 1, 999)
	if !errors.Is(err, ErrNotBotOwner) {
		t.Errorf("got %v, want ErrNotBotOwner", err)
	}
}

// --- UpdateSettings ---

func TestUpdateSettings_Modules(t *testing.T) {
	bot := testBot(1, 10)
	var savedSettings entity.BotSettings
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateSettFn: func(_ context.Context, _ int, s entity.BotSettings) error {
			savedSettings = s
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	modules := []string{"loyalty", "referral"}
	err := uc.UpdateSettings(context.Background(), 1, 10, &entity.UpdateBotSettingsRequest{
		Modules: &modules,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(savedSettings.Modules) != 2 {
		t.Errorf("got %d modules, want 2", len(savedSettings.Modules))
	}
	// WelcomeMessage should be preserved from original
	if savedSettings.WelcomeMessage != "Hello!" {
		t.Errorf("WelcomeMessage changed to %q, should be preserved", savedSettings.WelcomeMessage)
	}
}

func TestUpdateSettings_WelcomeMessage(t *testing.T) {
	bot := testBot(1, 10)
	var savedSettings entity.BotSettings
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateSettFn: func(_ context.Context, _ int, s entity.BotSettings) error {
			savedSettings = s
			return nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.UpdateSettings(context.Background(), 1, 10, &entity.UpdateBotSettingsRequest{
		WelcomeMessage: ptr("Welcome!"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedSettings.WelcomeMessage != "Welcome!" {
		t.Errorf("got WelcomeMessage %q, want %q", savedSettings.WelcomeMessage, "Welcome!")
	}
	// Modules should be preserved
	if len(savedSettings.Modules) != 1 {
		t.Errorf("Modules changed, should be preserved")
	}
}

func TestUpdateSettings_NotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.UpdateSettings(context.Background(), 999, 10, &entity.UpdateBotSettingsRequest{
		WelcomeMessage: ptr("x"),
	})
	if !errors.Is(err, ErrBotNotFound) {
		t.Errorf("got %v, want ErrBotNotFound", err)
	}
}

func TestUpdateSettings_WrongOrg(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.UpdateSettings(context.Background(), 1, 999, &entity.UpdateBotSettingsRequest{
		WelcomeMessage: ptr("x"),
	})
	if !errors.Is(err, ErrNotBotOwner) {
		t.Errorf("got %v, want ErrNotBotOwner", err)
	}
}

func TestUpdateSettings_RepoError(t *testing.T) {
	bot := testBot(1, 10)
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return bot, nil
		},
		updateSettFn: func(_ context.Context, _ int, _ entity.BotSettings) error {
			return errors.New("db error")
		},
	}
	uc := newUC(repo, &mockBotClientsRepo{})

	err := uc.UpdateSettings(context.Background(), 1, 10, &entity.UpdateBotSettingsRequest{
		WelcomeMessage: ptr("x"),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
