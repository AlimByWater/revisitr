-- +goose Up
-- +goose StatementBegin

-- Loyalty Program (1) -> Bot (N)
ALTER TABLE bots
    ADD COLUMN program_id INT REFERENCES loyalty_programs(id) ON DELETE SET NULL;
CREATE INDEX idx_bots_program_id ON bots(program_id);

-- Bot (1) -> POS (M)
ALTER TABLE pos_locations
    ADD COLUMN bot_id INT REFERENCES bots(id) ON DELETE SET NULL;
CREATE INDEX idx_pos_locations_bot_id ON pos_locations(bot_id);

-- Reward type expansion
ALTER TABLE loyalty_levels
    ADD COLUMN reward_type VARCHAR(10) NOT NULL DEFAULT 'percent',
    ADD COLUMN reward_amount DECIMAL(10,2) NOT NULL DEFAULT 0;
UPDATE loyalty_levels SET reward_amount = reward_percent WHERE reward_type = 'percent';
ALTER TABLE loyalty_levels
    ADD CONSTRAINT chk_reward_type CHECK (reward_type IN ('percent', 'fixed'));

-- Client phone normalization + QR
ALTER TABLE bot_clients
    ADD COLUMN phone_normalized VARCHAR(15),
    ADD COLUMN qr_code VARCHAR(64) UNIQUE;
CREATE INDEX idx_bot_clients_phone_normalized ON bot_clients(phone_normalized);
CREATE INDEX idx_bot_clients_qr_code ON bot_clients(qr_code);

-- Make phone NOT NULL (set placeholder for existing nulls)
UPDATE bot_clients SET phone = 'unknown' WHERE phone IS NULL OR phone = '';
ALTER TABLE bot_clients ALTER COLUMN phone SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE bot_clients ALTER COLUMN phone DROP NOT NULL;
DROP INDEX IF EXISTS idx_bot_clients_qr_code;
DROP INDEX IF EXISTS idx_bot_clients_phone_normalized;
ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS qr_code,
    DROP COLUMN IF EXISTS phone_normalized;

ALTER TABLE loyalty_levels
    DROP CONSTRAINT IF EXISTS chk_reward_type,
    DROP COLUMN IF EXISTS reward_amount,
    DROP COLUMN IF EXISTS reward_type;

DROP INDEX IF EXISTS idx_pos_locations_bot_id;
ALTER TABLE pos_locations DROP COLUMN IF EXISTS bot_id;

DROP INDEX IF EXISTS idx_bots_program_id;
ALTER TABLE bots DROP COLUMN IF EXISTS program_id;

-- +goose StatementEnd
