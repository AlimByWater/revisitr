package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type ModulePresets struct {
	pg *Module
}

func NewModulePresets(pg *Module) *ModulePresets {
	return &ModulePresets{pg: pg}
}

func (r *ModulePresets) GetByModule(ctx context.Context, moduleKey string) ([]entity.ModulePreset, error) {
	var presets []entity.ModulePreset
	err := r.pg.DB().SelectContext(ctx, &presets,
		"SELECT * FROM module_presets WHERE module_key = $1 ORDER BY sort_order", moduleKey)
	if err != nil {
		return nil, fmt.Errorf("modulePresets.GetByModule: %w", err)
	}
	return presets, nil
}

func (r *ModulePresets) GetByKey(ctx context.Context, moduleKey, presetKey string) (*entity.ModulePreset, error) {
	var preset entity.ModulePreset
	err := r.pg.DB().GetContext(ctx, &preset,
		"SELECT * FROM module_presets WHERE module_key = $1 AND preset_key = $2", moduleKey, presetKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("modulePresets.GetByKey: %w", err)
	}
	return &preset, nil
}
