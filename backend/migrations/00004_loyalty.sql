-- +goose Up
-- +goose StatementBegin

CREATE TABLE loyalty_programs (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('bonus', 'discount')),
    config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_loyalty_programs_org_id ON loyalty_programs(org_id);

CREATE TABLE loyalty_levels (
    id SERIAL PRIMARY KEY,
    program_id INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    threshold INT NOT NULL DEFAULT 0,
    reward_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_loyalty_levels_program_id ON loyalty_levels(program_id);

CREATE TABLE client_loyalty (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    program_id INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    level_id INT REFERENCES loyalty_levels(id) ON DELETE SET NULL,
    balance DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_earned DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_spent DECIMAL(12,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(client_id, program_id)
);

CREATE TABLE loyalty_transactions (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    program_id INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('earn', 'spend', 'adjust')),
    amount DECIMAL(12,2) NOT NULL,
    balance_after DECIMAL(12,2) NOT NULL,
    description TEXT,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_loyalty_transactions_client_program ON loyalty_transactions(client_id, program_id);
CREATE INDEX idx_loyalty_transactions_created_at ON loyalty_transactions(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS loyalty_transactions;
DROP TABLE IF EXISTS client_loyalty;
DROP TABLE IF EXISTS loyalty_levels;
DROP TABLE IF EXISTS loyalty_programs;

-- +goose StatementEnd
