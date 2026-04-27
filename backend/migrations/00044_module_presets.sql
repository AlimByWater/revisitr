-- +goose Up

CREATE TABLE module_presets (
    id          SERIAL PRIMARY KEY,
    module_key  TEXT NOT NULL,
    preset_key  TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    definition  JSONB NOT NULL DEFAULT '{}',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(module_key, preset_key)
);

CREATE TABLE bot_module_settings (
    bot_id         INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    module_key     TEXT NOT NULL,
    preset_id      INT REFERENCES module_presets(id) ON DELETE SET NULL,
    preset_key     TEXT NOT NULL DEFAULT '',
    customized     BOOLEAN NOT NULL DEFAULT FALSE,
    customizations JSONB NOT NULL DEFAULT '{}',
    config         JSONB NOT NULL DEFAULT '{}',
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bot_id, module_key)
);

CREATE INDEX idx_bot_module_settings_bot ON bot_module_settings(bot_id);

-- Seed menu presets
INSERT INTO module_presets (module_key, preset_key, name, description, definition, sort_order) VALUES
('menu', 'tabs', 'Таб-категории', 'Категории как inline-кнопки с переключением через editMessageText.', '{"render_mode": "tabs"}', 0),
('menu', 'list', 'Список', 'Все позиции меню в одном сообщении, сгруппированы по категориям.', '{"render_mode": "list"}', 1),
('menu', 'carousel', 'Карусель', 'Фото + описание с навигацией ←/→ по позициям.', '{"render_mode": "carousel"}', 2);

-- Migrate existing bots with menu module to tabs preset (current behavior)
INSERT INTO bot_module_settings (bot_id, module_key, preset_id, preset_key, config)
SELECT b.id, 'menu',
       (SELECT id FROM module_presets WHERE module_key = 'menu' AND preset_key = 'tabs'),
       'tabs',
       COALESCE(b.settings->'module_configs'->'menu', '{}')
FROM bots b
WHERE b.settings->'modules' ? 'menu'
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS bot_module_settings;
DROP TABLE IF EXISTS module_presets;
