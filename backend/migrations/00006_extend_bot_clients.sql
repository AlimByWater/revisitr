-- +goose Up
ALTER TABLE bot_clients
    ADD COLUMN gender VARCHAR(10),
    ADD COLUMN birth_date DATE,
    ADD COLUMN city VARCHAR(100),
    ADD COLUMN tags JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN os VARCHAR(20);

CREATE INDEX idx_bot_clients_tags ON bot_clients USING GIN (tags);
CREATE INDEX idx_bot_clients_birth_date ON bot_clients (birth_date);

-- +goose Down
DROP INDEX IF EXISTS idx_bot_clients_birth_date;
DROP INDEX IF EXISTS idx_bot_clients_tags;

ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS os,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS gender;
