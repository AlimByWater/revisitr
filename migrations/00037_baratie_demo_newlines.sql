-- +goose Up
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(
                        settings,
                        '{welcome_message}',
                        to_jsonb(replace(settings->>'welcome_message', chr(92) || 'n', chr(10)))
                    ),
                    '{buttons,0,value}',
                    to_jsonb(replace(settings->'buttons'->0->>'value', chr(92) || 'n', chr(10)))
                ),
                '{buttons,1,value}',
                to_jsonb(replace(settings->'buttons'->1->>'value', chr(92) || 'n', chr(10)))
            ),
            '{buttons,2,value}',
            to_jsonb(replace(settings->'buttons'->2->>'value', chr(92) || 'n', chr(10)))
        ),
        '{buttons,3,value}',
        to_jsonb(replace(settings->'buttons'->3->>'value', chr(92) || 'n', chr(10)))
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(
                        settings,
                        '{welcome_message}',
                        to_jsonb(replace(settings->>'welcome_message', chr(10), chr(92) || 'n'))
                    ),
                    '{buttons,0,value}',
                    to_jsonb(replace(settings->'buttons'->0->>'value', chr(10), chr(92) || 'n'))
                ),
                '{buttons,1,value}',
                to_jsonb(replace(settings->'buttons'->1->>'value', chr(10), chr(92) || 'n'))
            ),
            '{buttons,2,value}',
            to_jsonb(replace(settings->'buttons'->2->>'value', chr(10), chr(92) || 'n'))
        ),
        '{buttons,3,value}',
        to_jsonb(replace(settings->'buttons'->3->>'value', chr(10), chr(92) || 'n'))
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot';

-- +goose StatementEnd
