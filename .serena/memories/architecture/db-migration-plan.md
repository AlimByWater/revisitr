# Planned DB Migration — Entity Hierarchy

## New columns needed:
- bots: ADD program_id INT REFERENCES loyalty_programs(id) — NULLABLE
- pos: ADD bot_id INT REFERENCES bots(id) — NULLABLE  
- loyalty_levels: ADD reward_type VARCHAR(10) DEFAULT 'percent' — 'percent' | 'fixed'
- loyalty_levels: ADD reward_amount DECIMAL(10,2) — for fixed bonus
- clients: ALTER phone SET NOT NULL
- clients: ADD qr_code VARCHAR(64) UNIQUE
- clients: ADD phone_normalized VARCHAR(15) + INDEX

## Hierarchy: LoyaltyProgram (1) → Bot (N) → POS (M)
## All via org_id currently — migration adds direct FK relationships
## Existing data: nullable FKs allow gradual migration
