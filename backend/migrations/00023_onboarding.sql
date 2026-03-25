-- +goose Up

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS onboarding_state JSONB NOT NULL DEFAULT '{}';

-- +goose Down

ALTER TABLE organizations
    DROP COLUMN IF EXISTS onboarding_completed,
    DROP COLUMN IF EXISTS onboarding_state;
