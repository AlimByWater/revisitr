-- +goose Up
-- +goose StatementBegin
UPDATE module_presets
SET description = 'Свернутый индекс категорий с компактным раскрытием по нажатию.'
WHERE module_key = 'menu' AND preset_key = 'list';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE module_presets
SET description = 'Все позиции меню в одном сообщении, сгруппированы по категориям.'
WHERE module_key = 'menu' AND preset_key = 'list';
-- +goose StatementEnd
