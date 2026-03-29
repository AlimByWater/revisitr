-- +goose Up
ALTER TABLE promotions
    ADD COLUMN filter JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN triggers JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN actions JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN combinable_with_loyalty BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE promotions
    DROP COLUMN IF EXISTS filter,
    DROP COLUMN IF EXISTS triggers,
    DROP COLUMN IF EXISTS actions,
    DROP COLUMN IF EXISTS combinable_with_loyalty;
