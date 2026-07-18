-- +goose Up
-- +goose StatementBegin

ALTER TABLE organizations
    ADD COLUMN timezone VARCHAR(64) NOT NULL DEFAULT 'Europe/Moscow';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE organizations DROP COLUMN timezone;

-- +goose StatementEnd
