package modulepresets

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type presetsRepo interface {
	GetByModule(ctx context.Context, moduleKey string) ([]entity.ModulePreset, error)
	GetByKey(ctx context.Context, moduleKey, presetKey string) (*entity.ModulePreset, error)
}

type settingsRepo interface {
	Get(ctx context.Context, botID int, moduleKey string) (*entity.BotModuleSettings, error)
	GetAllForBot(ctx context.Context, botID int) ([]entity.BotModuleSettings, error)
	Upsert(ctx context.Context, s *entity.BotModuleSettings) error
	ResetToPreset(ctx context.Context, botID int, moduleKey, presetKey string, presetID int) error
	Delete(ctx context.Context, botID int, moduleKey string) error
}

type botsRepo interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type botEventPublisher interface {
	PublishBotSettings(ctx context.Context, botID int, field string) error
}

var (
	ErrBotNotFound    = fmt.Errorf("bot not found")
	ErrNotBotOwner    = fmt.Errorf("not authorized to manage this bot")
	ErrPresetNotFound = fmt.Errorf("preset not found")
	ErrInvalidPreset  = fmt.Errorf("invalid preset key")
)

type Usecase struct {
	presets  presetsRepo
	settings settingsRepo
	bots     botsRepo
	eventBus botEventPublisher
	logger   *slog.Logger
}

func New(presets presetsRepo, settings settingsRepo, bots botsRepo) *Usecase {
	return &Usecase{
		presets:  presets,
		settings: settings,
		bots:     bots,
	}
}

func (uc *Usecase) SetEventBus(eb botEventPublisher) {
	uc.eventBus = eb
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) checkBotOwnership(ctx context.Context, botID, orgID int) (*entity.Bot, error) {
	bot, err := uc.bots.GetByID(ctx, botID)
	if err != nil {
		return nil, ErrBotNotFound
	}
	if bot.OrgID != orgID {
		return nil, ErrNotBotOwner
	}
	return bot, nil
}

func (uc *Usecase) publishSettings(ctx context.Context, botID int) {
	if uc.eventBus != nil {
		_ = uc.eventBus.PublishBotSettings(ctx, botID, "module_settings")
	}
}

func (uc *Usecase) ListPresets(ctx context.Context, moduleKey string) ([]entity.ModulePreset, error) {
	return uc.presets.GetByModule(ctx, moduleKey)
}

func (uc *Usecase) GetBotModuleSettings(ctx context.Context, botID, orgID int, moduleKey string) (*entity.BotModuleSettings, error) {
	if _, err := uc.checkBotOwnership(ctx, botID, orgID); err != nil {
		return nil, err
	}
	return uc.settings.Get(ctx, botID, moduleKey)
}

func (uc *Usecase) SelectPreset(ctx context.Context, botID, orgID int, moduleKey, presetKey string) error {
	if _, err := uc.checkBotOwnership(ctx, botID, orgID); err != nil {
		return err
	}

	preset, err := uc.presets.GetByKey(ctx, moduleKey, presetKey)
	if err != nil {
		return fmt.Errorf("select preset: %w", err)
	}
	if preset == nil {
		return ErrPresetNotFound
	}

	// Get existing settings to preserve config
	existing, _ := uc.settings.Get(ctx, botID, moduleKey)
	config := entity.JSONB("{}")
	if existing != nil {
		config = existing.Config
	}

	s := &entity.BotModuleSettings{
		BotID:          botID,
		ModuleKey:      moduleKey,
		PresetID:       &preset.ID,
		PresetKey:      presetKey,
		Customized:     false,
		Customizations: entity.JSONB("{}"),
		Config:         config,
	}
	if err := uc.settings.Upsert(ctx, s); err != nil {
		return fmt.Errorf("select preset: %w", err)
	}

	uc.publishSettings(ctx, botID)
	return nil
}

func (uc *Usecase) UpdateCustomizations(ctx context.Context, botID, orgID int, moduleKey string, customizations json.RawMessage) error {
	if _, err := uc.checkBotOwnership(ctx, botID, orgID); err != nil {
		return err
	}

	existing, err := uc.settings.Get(ctx, botID, moduleKey)
	if err != nil {
		return fmt.Errorf("update customizations: %w", err)
	}
	if existing == nil {
		return ErrPresetNotFound
	}

	existing.Customized = true
	existing.Customizations = entity.JSONB(customizations)
	if err := uc.settings.Upsert(ctx, existing); err != nil {
		return fmt.Errorf("update customizations: %w", err)
	}

	uc.publishSettings(ctx, botID)
	return nil
}

func (uc *Usecase) ResetPreset(ctx context.Context, botID, orgID int, moduleKey string) error {
	if _, err := uc.checkBotOwnership(ctx, botID, orgID); err != nil {
		return err
	}

	existing, err := uc.settings.Get(ctx, botID, moduleKey)
	if err != nil {
		return fmt.Errorf("reset preset: %w", err)
	}
	if existing == nil || existing.PresetID == nil {
		return ErrPresetNotFound
	}

	if err := uc.settings.ResetToPreset(ctx, botID, moduleKey, existing.PresetKey, *existing.PresetID); err != nil {
		return fmt.Errorf("reset preset: %w", err)
	}

	uc.publishSettings(ctx, botID)
	return nil
}
