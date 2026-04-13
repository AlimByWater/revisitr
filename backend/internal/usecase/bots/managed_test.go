package bots

import (
	"context"
	"errors"
	"testing"

	"revisitr/internal/entity"
)

type mockAuthTokenRepo struct {
	storeTokenFn func(ctx context.Context, token string, data entity.MasterBotAuthToken) error
}

func (m *mockAuthTokenRepo) StoreToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error {
	if m.storeTokenFn != nil {
		return m.storeTokenFn(ctx, token, data)
	}
	return nil
}

func TestManagedAdapter_StoreAuthToken_Delegates(t *testing.T) {
	called := false
	adapter := NewManagedBotAdapter(newUC(&mockBotsRepo{}, &mockBotClientsRepo{}), &mockAuthTokenRepo{
		storeTokenFn: func(_ context.Context, token string, data entity.MasterBotAuthToken) error {
			called = true
			if token != "abc123" {
				t.Fatalf("unexpected token %q", token)
			}
			if data.OrgID != 7 || data.UserID != 11 {
				t.Fatalf("unexpected data %+v", data)
			}
			return nil
		},
	})

	if err := adapter.StoreAuthToken(context.Background(), "abc123", entity.MasterBotAuthToken{OrgID: 7, UserID: 11}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected StoreToken to be called")
	}
}

func TestManagedAdapter_CreatePendingBot_SetsManagedDefaults(t *testing.T) {
	repo := &mockBotsRepo{
		createFn: func(_ context.Context, bot *entity.Bot) error {
			bot.ID = 55
			if bot.Status != "pending" {
				t.Fatalf("expected pending status, got %q", bot.Status)
			}
			if !bot.IsManagedBot {
				t.Fatal("expected managed bot flag")
			}
			if bot.Settings.Modules == nil || len(bot.Settings.Modules) != 0 {
				t.Fatalf("expected empty modules slice, got %#v", bot.Settings.Modules)
			}
			if bot.Settings.RegistrationForm == nil || len(bot.Settings.RegistrationForm) != 0 {
				t.Fatalf("expected empty registration form slice, got %#v", bot.Settings.RegistrationForm)
			}
			return nil
		},
	}
	adapter := NewManagedBotAdapter(newUC(repo, &mockBotClientsRepo{}), &mockAuthTokenRepo{})

	bot, err := adapter.CreatePendingBot(context.Background(), 9, &entity.CreateManagedBotRequest{
		Name:     "Managed",
		Username: "managedbot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bot.ID != 55 {
		t.Fatalf("expected bot id 55, got %d", bot.ID)
	}
	if bot.OrgID != 9 || bot.Name != "Managed" || bot.Username != "managedbot" {
		t.Fatalf("unexpected bot: %+v", bot)
	}
}

func TestManagedAdapter_CreatePendingBot_PreservesWizardPayload(t *testing.T) {
	welcome := &entity.MessageContent{Parts: []entity.MessagePart{{Type: entity.PartText, Text: "hello"}}}
	fields := []entity.FormField{{Name: "phone", Label: "Телефон", Type: "phone", Required: true}}
	modules := []string{"loyalty", "feedback"}

	repo := &mockBotsRepo{
		createFn: func(_ context.Context, bot *entity.Bot) error {
			if bot.Settings.WelcomeMessage != "Привет" {
				t.Fatalf("expected welcome message, got %q", bot.Settings.WelcomeMessage)
			}
			if bot.Settings.WelcomeContent == nil || len(bot.Settings.WelcomeContent.Parts) != 1 {
				t.Fatalf("expected welcome content, got %#v", bot.Settings.WelcomeContent)
			}
			if len(bot.Settings.RegistrationForm) != 1 || bot.Settings.RegistrationForm[0].Name != "phone" {
				t.Fatalf("unexpected registration form: %#v", bot.Settings.RegistrationForm)
			}
			if len(bot.Settings.Modules) != 2 || bot.Settings.Modules[1] != "feedback" {
				t.Fatalf("unexpected modules: %#v", bot.Settings.Modules)
			}
			return nil
		},
	}
	adapter := NewManagedBotAdapter(newUC(repo, &mockBotClientsRepo{}), &mockAuthTokenRepo{})

	_, err := adapter.CreatePendingBot(context.Background(), 5, &entity.CreateManagedBotRequest{
		Name:             "Managed",
		Username:         "managedbot",
		WelcomeMessage:   "Привет",
		WelcomeContent:   welcome,
		RegistrationForm: fields,
		Modules:          modules,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagedAdapter_GetBotStatus_MapsNotFound(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return nil, errors.New("missing")
		},
	}
	adapter := NewManagedBotAdapter(newUC(repo, &mockBotClientsRepo{}), &mockAuthTokenRepo{})

	_, err := adapter.GetBotStatus(context.Background(), 1, 9)
	if !errors.Is(err, ErrBotNotFound) {
		t.Fatalf("expected ErrBotNotFound, got %v", err)
	}
}

func TestManagedAdapter_GetBotStatus_EnforcesOwnership(t *testing.T) {
	repo := &mockBotsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Bot, error) {
			return &entity.Bot{ID: 3, OrgID: 99, Status: "pending"}, nil
		},
	}
	adapter := NewManagedBotAdapter(newUC(repo, &mockBotClientsRepo{}), &mockAuthTokenRepo{})

	_, err := adapter.GetBotStatus(context.Background(), 3, 1)
	if !errors.Is(err, ErrNotBotOwner) {
		t.Fatalf("expected ErrNotBotOwner, got %v", err)
	}
}
