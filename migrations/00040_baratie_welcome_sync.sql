-- +goose Up
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        settings,
        '{welcome_content}',
        jsonb_build_object(
            'parts',
            jsonb_build_array(
                jsonb_build_object(
                    'type', 'text',
                    'text', settings->>'welcome_message',
                    'parse_mode', 'Markdown'
                )
            )
        ),
        true
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot'
  AND COALESCE(settings->>'welcome_message', '') <> '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Irreversible data sync migration: keep current synced state.
SELECT 1;

-- +goose StatementEnd
