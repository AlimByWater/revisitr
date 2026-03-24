-- +goose Up
-- +goose StatementBegin

CREATE TABLE balance_reserves (
    id SERIAL PRIMARY KEY,
    client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    program_id INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'cancelled', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_balance_reserves_client_program ON balance_reserves(client_id, program_id);
CREATE INDEX idx_balance_reserves_status ON balance_reserves(status) WHERE status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS balance_reserves;

-- +goose StatementEnd
