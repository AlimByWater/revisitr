package bots

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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
	HasPOSLocations(ctx context.Context, botID int) (bool, error)
	Update(ctx context.Context, bot *entity.Bot) error
	UpdateSettings(ctx context.Context, id int, settings entity.BotSettings) error
	Delete(ctx context.Context, id int) error
}

type botClientsRepo interface {
	CountByBotID(ctx context.Context, botID int) (int, error)
}

type botEventPublisher interface {
	PublishBotReload(ctx context.Context, botID int) error
	PublishBotStop(ctx context.Context, botID int) error
	PublishBotStart(ctx context.Context, botID int) error
	PublishBotSettings(ctx context.Context, botID int, field string) error
}

type Usecase struct {
	bots     botsRepo
	clients  botClientsRepo
	eventBus botEventPublisher
	logger   *slog.Logger
}

func New(bots botsRepo, clients botClientsRepo) *Usecase {
	return &Usecase{
		bots:    bots,
		clients: clients,
	}
}

// SetEventBus sets the event bus for publishing bot events. Optional.
func (uc *Usecase) SetEventBus(eb botEventPublisher) {
	uc.eventBus = eb
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
		Status: "pending",
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

	for i := range bots {
		bots[i].TokenMasked = entity.MaskToken(bots[i].Token)
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

	bot.TokenMasked = entity.MaskToken(bot.Token)

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

	oldStatus := bot.Status

	if req.Name != nil {
		bot.Name = *req.Name
	}
	if req.ProgramID != nil {
		bot.ProgramID = req.ProgramID
	}
	if req.Status != nil {
		hasPOS, err := uc.bots.HasPOSLocations(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("check bot pos locations: %w", err)
		}
		if !hasPOS {
			bot.Status = "pending"
		} else {
			bot.Status = *req.Status
		}
	}

	if err := uc.bots.Update(ctx, bot); err != nil {
		return nil, fmt.Errorf("update bot: %w", err)
	}

	// Publish status change events
	if uc.eventBus != nil && req.Status != nil && oldStatus != bot.Status {
		switch bot.Status {
		case "active":
			_ = uc.eventBus.PublishBotStart(ctx, id)
		case "inactive":
			_ = uc.eventBus.PublishBotStop(ctx, id)
		}
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

	if uc.eventBus != nil {
		_ = uc.eventBus.PublishBotStop(ctx, id)
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
	if req.WelcomeContent != nil {
		if err := req.WelcomeContent.Validate(); err != nil {
			return fmt.Errorf("invalid welcome content: %w", err)
		}
		settings.WelcomeContent = req.WelcomeContent
		if derived := deriveWelcomeMessage(req.WelcomeContent, settings.WelcomeMessage); derived != "" {
			settings.WelcomeMessage = derived
		}
	}
	if req.ModuleConfigs != nil {
		if req.ModuleConfigs.Booking.IntroContent != nil {
			if err := req.ModuleConfigs.Booking.IntroContent.Validate(); err != nil {
				return fmt.Errorf("invalid booking intro content: %w", err)
			}
		}
		settings.ModuleConfigs = *req.ModuleConfigs
	}
	if req.PosSelectorEnabled != nil {
		settings.PosSelectorEnabled = *req.PosSelectorEnabled
	}
	if req.ContactsPOSIDs != nil {
		settings.ContactsPOSIDs = *req.ContactsPOSIDs
	}

	if err := uc.bots.UpdateSettings(ctx, id, settings); err != nil {
		return fmt.Errorf("update bot settings: %w", err)
	}

	// Publish settings change event
	if uc.eventBus != nil {
		field := ""
		switch {
		case req.WelcomeMessage != nil || req.WelcomeContent != nil:
			field = "welcome"
		case req.Buttons != nil:
			field = "buttons"
		case req.Modules != nil:
			field = "modules"
		case req.RegistrationForm != nil:
			field = "registration_form"
		}
		_ = uc.eventBus.PublishBotSettings(ctx, id, field)
	}

	return nil
}

func deriveWelcomeMessage(content *entity.MessageContent, fallback string) string {
	if content == nil || len(content.Parts) == 0 {
		return fallback
	}

	for _, part := range content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			return part.Text
		}
	}

	return fallback
}
