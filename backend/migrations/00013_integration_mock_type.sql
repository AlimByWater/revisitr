-- +goose Up
ALTER TABLE integrations
DROP CONSTRAINT IF EXISTS integrations_type_check;

ALTER TABLE integrations
ADD CONSTRAINT integrations_type_check
CHECK (type IN ('iiko', 'rkeeper', '1c', 'mock'));

-- +goose Down
ALTER TABLE integrations
DROP CONSTRAINT IF EXISTS integrations_type_check;

ALTER TABLE integrations
ADD CONSTRAINT integrations_type_check
CHECK (type IN ('iiko', 'rkeeper', '1c'));
