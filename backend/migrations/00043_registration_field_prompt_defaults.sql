-- +goose Up
-- +goose StatementBegin
UPDATE bots
SET settings = jsonb_set(
    settings,
    '{registration_form}',
    (
        SELECT COALESCE(jsonb_agg(
            CASE field->>'name'
                WHEN 'first_name' THEN jsonb_set(field, '{label}', to_jsonb('Как вас зовут?'::text))
                WHEN 'phone' THEN jsonb_set(field, '{label}', to_jsonb('Ваш номер телефона?'::text))
                WHEN 'birth_date' THEN jsonb_set(
                    jsonb_set(field, '{name}', to_jsonb('birthday'::text)),
                    '{label}',
                    to_jsonb('Когда у вас день рождения?'::text)
                )
                WHEN 'birthday' THEN jsonb_set(field, '{label}', to_jsonb('Когда у вас день рождения?'::text))
                ELSE field
            END
            ORDER BY ordinality
        ), '[]'::jsonb)
        FROM jsonb_array_elements(COALESCE(settings->'registration_form', '[]'::jsonb)) WITH ORDINALITY AS fields(field, ordinality)
    ),
    true
)
WHERE settings ? 'registration_form'
  AND EXISTS (
      SELECT 1
      FROM jsonb_array_elements(COALESCE(settings->'registration_form', '[]'::jsonb)) AS field
      WHERE field->>'name' IN ('first_name', 'phone', 'birth_date', 'birthday')
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE bots
SET settings = jsonb_set(
    settings,
    '{registration_form}',
    (
        SELECT COALESCE(jsonb_agg(
            CASE field->>'name'
                WHEN 'first_name' THEN jsonb_set(field, '{label}', to_jsonb('Имя'::text))
                WHEN 'phone' THEN jsonb_set(field, '{label}', to_jsonb('Телефон'::text))
                WHEN 'birthday' THEN jsonb_set(field, '{label}', to_jsonb('Дата рождения'::text))
                ELSE field
            END
            ORDER BY ordinality
        ), '[]'::jsonb)
        FROM jsonb_array_elements(COALESCE(settings->'registration_form', '[]'::jsonb)) WITH ORDINALITY AS fields(field, ordinality)
    ),
    true
)
WHERE settings ? 'registration_form'
  AND EXISTS (
      SELECT 1
      FROM jsonb_array_elements(COALESCE(settings->'registration_form', '[]'::jsonb)) AS field
      WHERE field->>'name' IN ('first_name', 'phone', 'birthday')
  );
-- +goose StatementEnd
