-- +goose Up
CREATE TABLE IF NOT EXISTS bots (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_bots_org_id ON bots(org_id);

CREATE TABLE IF NOT EXISTS bot_clients (
    id SERIAL PRIMARY KEY,
    bot_id INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    telegram_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    phone VARCHAR(50),
    data JSONB DEFAULT '{}',
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(bot_id, telegram_id)
);

CREATE INDEX idx_bot_clients_bot_id ON bot_clients(bot_id);

-- +goose Down
DROP TABLE IF EXISTS bot_clients;
DROP TABLE IF EXISTS bots;
