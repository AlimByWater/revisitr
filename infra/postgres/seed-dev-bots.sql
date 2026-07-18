-- Seed dev client bot: @baratie_devbot (8726329303)
-- Run AFTER migrations are applied on dev DB.
-- Idempotent — safe to re-run.

INSERT INTO bots (org_id, name, token, username, status, settings, created_at, updated_at)
SELECT
    1,
    'Baratie (dev)',
    '8726329303:AAH8vBQwUDkgCaeiMm3qPOTgfdrDxarYq8k',
    '@baratie_devbot',
    'active',
    '{}'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM bots WHERE token = '8726329303:AAH8vBQwUDkgCaeiMm3qPOTgfdrDxarYq8k'
);
