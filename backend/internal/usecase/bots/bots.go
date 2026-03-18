package bots

import (
	"context"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrBotNotFound = fmt.Errorf("bot not found")
	ErrNotBotOwner = fmt.Errorf("not authorized to manage this bot")
)

type botsRepo interface {
	Create(ctx context.Context, bot *entity.Bot) error
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error)
	Update(ctx context.Context, bot *entity.Bot) error
	UpdateSettings(ctx context.Context, id int, settings entity.BotSettings) error
	Delete(ctx context.Context, id int) error
}

type botClientsRepo interface {
	CountByBotID(ctx context.Context, botID int) (int, error)
}

type Usecase struct {
	bots    botsRepo
	clients botClientsRepo
	logger  *slog.Logger
}

func New(bots botsRepo, clients botClientsRepo) *Usecase {
	return &Usecase{
		bots:    bots,
		clients: clients,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreateBotRequest) (*entity.Bot, error) {
	bot := &entity.Bot{
		OrgID:  orgID,
		Name:   req.Name,
		Token:  req.Token,
		Status: "inactive",
		Settings: entity.BotSettings{
			Modules:          []string{},
			Buttons:          []entity.BotButton{},
			RegistrationForm: []entity.FormField{},
		},
	}

	if err := uc.bots.Create(ctx, bot); err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return bot, nil
}

func (uc *Usecase) GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error) {
	bots, err := uc.bots.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get bots by org: %w", err)
	}

	return bots, nil
}

func (uc *Usecase) GetByID(ctx context.Context, id, orgID int) (*entity.Bot, error) {
	bot, err := uc.bots.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return nil, ErrNotBotOwner
	}

	return bot, nil
}

func (uc *Usecase) Update(ctx context.Context, id, orgID int, req *entity.UpdateBotRequest) (*entity.Bot, error) {
	bot, err := uc.bots.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return nil, ErrNotBotOwner
	}

	if req.Name != nil {
		bot.Name = *req.Name
	}
	if req.Status != nil {
		bot.Status = *req.Status
	}

	if err := uc.bots.Update(ctx, bot); err != nil {
		return nil, fmt.Errorf("update bot: %w", err)
	}

	return bot, nil
}

func (uc *Usecase) Delete(ctx context.Context, id, orgID int) error {
	bot, err := uc.bots.GetByID(ctx, id)
	if err != nil {
		return ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return ErrNotBotOwner
	}

	if err := uc.bots.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete bot: %w", err)
	}

	return nil
}

func (uc *Usecase) GetSettings(ctx context.Context, id, orgID int) (*entity.BotSettings, error) {
	bot, err := uc.bots.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return nil, ErrNotBotOwner
	}

	return &bot.Settings, nil
}

func (uc *Usecase) UpdateSettings(ctx context.Context, id, orgID int, req *entity.UpdateBotSettingsRequest) error {
	bot, err := uc.bots.GetByID(ctx, id)
	if err != nil {
		return ErrBotNotFound
	}

	if bot.OrgID != orgID {
		return ErrNotBotOwner
	}

	settings := bot.Settings

	if req.Modules != nil {
		settings.Modules = *req.Modules
	}
	if req.Buttons != nil {
		settings.Buttons = *req.Buttons
	}
	if req.RegistrationForm != nil {
		settings.RegistrationForm = *req.RegistrationForm
	}
	if req.WelcomeMessage != nil {
		settings.WelcomeMessage = *req.WelcomeMessage
	}

	if err := uc.bots.UpdateSettings(ctx, id, settings); err != nil {
		return fmt.Errorf("update bot settings: %w", err)
	}

	return nil
}
