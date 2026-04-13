package bots

import (
	"context"
	"fmt"

	"revisitr/internal/entity"
)

type authTokenRepo interface {
	StoreToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error
}

// ManagedBotAdapter implements the managedBotDeps interface for the HTTP controller.
type ManagedBotAdapter struct {
	uc       *Usecase
	authRepo authTokenRepo
}

func NewManagedBotAdapter(uc *Usecase, authRepo authTokenRepo) *ManagedBotAdapter {
	return &ManagedBotAdapter{uc: uc, authRepo: authRepo}
}

func (a *ManagedBotAdapter) StoreAuthToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error {
	return a.authRepo.StoreToken(ctx, token, data)
}

func (a *ManagedBotAdapter) CreatePendingBot(ctx context.Context, orgID int, req *entity.CreateManagedBotRequest) (*entity.Bot, error) {
	bot := &entity.Bot{
		OrgID:        orgID,
		Name:         req.Name,
		Username:     req.Username,
		Status:       "pending",
		IsManagedBot: true,
		Settings: entity.BotSettings{
			Modules:          req.Modules,
			RegistrationForm: req.RegistrationForm,
			WelcomeMessage:   req.WelcomeMessage,
			WelcomeContent:   req.WelcomeContent,
			Buttons:          []entity.BotButton{},
		},
	}

	if bot.Settings.Modules == nil {
		bot.Settings.Modules = []string{}
	}
	if bot.Settings.RegistrationForm == nil {
		bot.Settings.RegistrationForm = []entity.FormField{}
	}

	if err := a.uc.bots.Create(ctx, bot); err != nil {
		return nil, fmt.Errorf("create pending bot: %w", err)
	}

	return bot, nil
}

func (a *ManagedBotAdapter) GetBotStatus(ctx context.Context, botID, orgID int) (string, error) {
	bot, err := a.uc.bots.GetByID(ctx, botID)
	if err != nil {
		return "", ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return "", ErrNotBotOwner
	}

	return bot.Status, nil
}
