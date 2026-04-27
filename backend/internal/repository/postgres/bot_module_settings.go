package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type BotModuleSettings struct {
	pg *Module
}

func NewBotModuleSettings(pg *Module) *BotModuleSettings {
	return &BotModuleSettings{pg: pg}
}

func (r *BotModuleSettings) Get(ctx context.Context, botID int, moduleKey string) (*entity.BotModuleSettings, error) {
	var s entity.BotModuleSettings
	err := r.pg.DB().GetContext(ctx, &s,
		"SELECT * FROM bot_module_settings WHERE bot_id = $1 AND module_key = $2", botID, moduleKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("botModuleSettings.Get: %w", err)
	}
	return &s, nil
}

func (r *BotModuleSettings) GetAllForBot(ctx context.Context, botID int) ([]entity.BotModuleSettings, error) {
	var settings []entity.BotModuleSettings
	err := r.pg.DB().SelectContext(ctx, &settings,
		"SELECT * FROM bot_module_settings WHERE bot_id = $1", botID)
	if err != nil {
		return nil, fmt.Errorf("botModuleSettings.GetAllForBot: %w", err)
	}
	return settings, nil
}

func (r *BotModuleSettings) Upsert(ctx context.Context, s *entity.BotModuleSettings) error {
	query := `
		INSERT INTO bot_module_settings (bot_id, module_key, preset_id, preset_key, customized, customizations, config, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (bot_id, module_key) DO UPDATE SET
			preset_id = EXCLUDED.preset_id,
			preset_key = EXCLUDED.preset_key,
			customized = EXCLUDED.customized,
			customizations = EXCLUDED.customizations,
			config = EXCLUDED.config,
			updated_at = NOW()`
	_, err := r.pg.DB().ExecContext(ctx, query,
		s.BotID, s.ModuleKey, s.PresetID, s.PresetKey, s.Customized, s.Customizations, s.Config)
	if err != nil {
		return fmt.Errorf("botModuleSettings.Upsert: %w", err)
	}
	return nil
}

func (r *BotModuleSettings) ResetToPreset(ctx context.Context, botID int, moduleKey, presetKey string, presetID int) error {
	query := `
		UPDATE bot_module_settings
		SET preset_id = $1, preset_key = $2, customized = FALSE, customizations = '{}', updated_at = NOW()
		WHERE bot_id = $3 AND module_key = $4`
	_, err := r.pg.DB().ExecContext(ctx, query, presetID, presetKey, botID, moduleKey)
	if err != nil {
		return fmt.Errorf("botModuleSettings.ResetToPreset: %w", err)
	}
	return nil
}

func (r *BotModuleSettings) Delete(ctx context.Context, botID int, moduleKey string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"DELETE FROM bot_module_settings WHERE bot_id = $1 AND module_key = $2", botID, moduleKey)
	if err != nil {
		return fmt.Errorf("botModuleSettings.Delete: %w", err)
	}
	return nil
}
