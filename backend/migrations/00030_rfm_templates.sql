-- +goose Up

-- RFM template configuration fields for rfm_configs
ALTER TABLE rfm_configs
    ADD COLUMN active_template_type VARCHAR(10) NOT NULL DEFAULT 'standard',
    ADD COLUMN active_template_key  VARCHAR(30) NOT NULL DEFAULT 'tsr',
    ADD COLUMN custom_template_name VARCHAR(100),
    ADD COLUMN custom_r_thresholds  JSONB,
    ADD COLUMN custom_f_thresholds  JSONB;

ALTER TABLE rfm_configs
    ADD CONSTRAINT chk_rfm_template_type CHECK (active_template_type IN ('standard', 'custom')),
    ADD CONSTRAINT chk_rfm_template_key CHECK (
        active_template_key IN ('coffeegng', 'qsr', 'tsr', 'bar')
        OR active_template_type = 'custom'
    );

-- RFM v2 score fields for bot_clients (replace legacy rfm_recency/rfm_frequency/rfm_monetary)
ALTER TABLE bot_clients
    ADD COLUMN r_score              SMALLINT,
    ADD COLUMN f_score              SMALLINT,
    ADD COLUMN m_score              SMALLINT,
    ADD COLUMN recency_days         INT,
    ADD COLUMN frequency_count      INT,
    ADD COLUMN monetary_sum         NUMERIC(12,2),
    ADD COLUMN total_visits_lifetime INT NOT NULL DEFAULT 0,
    ADD COLUMN last_visit_date      DATE;

ALTER TABLE bot_clients
    ADD CONSTRAINT chk_r_score CHECK (r_score IS NULL OR (r_score >= 1 AND r_score <= 5)),
    ADD CONSTRAINT chk_f_score CHECK (f_score IS NULL OR (f_score >= 1 AND f_score <= 5)),
    ADD CONSTRAINT chk_m_score CHECK (m_score IS NULL OR (m_score >= 1 AND m_score <= 5));

CREATE INDEX idx_bot_clients_rfm_scores ON bot_clients(rfm_segment) WHERE rfm_segment IS NOT NULL;

-- Drop legacy v1 fields (superseded by r_score/f_score/m_score + recency_days/frequency_count/monetary_sum)
ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS rfm_recency,
    DROP COLUMN IF EXISTS rfm_frequency,
    DROP COLUMN IF EXISTS rfm_monetary;

-- +goose Down

-- Restore legacy v1 fields
ALTER TABLE bot_clients
    ADD COLUMN IF NOT EXISTS rfm_recency    INT,
    ADD COLUMN IF NOT EXISTS rfm_frequency  INT,
    ADD COLUMN IF NOT EXISTS rfm_monetary   NUMERIC(12,2);

DROP INDEX IF EXISTS idx_bot_clients_rfm_scores;

ALTER TABLE bot_clients
    DROP CONSTRAINT IF EXISTS chk_r_score,
    DROP CONSTRAINT IF EXISTS chk_f_score,
    DROP CONSTRAINT IF EXISTS chk_m_score;

ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS r_score,
    DROP COLUMN IF EXISTS f_score,
    DROP COLUMN IF EXISTS m_score,
    DROP COLUMN IF EXISTS recency_days,
    DROP COLUMN IF EXISTS frequency_count,
    DROP COLUMN IF EXISTS monetary_sum,
    DROP COLUMN IF EXISTS total_visits_lifetime,
    DROP COLUMN IF EXISTS last_visit_date;

ALTER TABLE rfm_configs
    DROP CONSTRAINT IF EXISTS chk_rfm_template_type,
    DROP CONSTRAINT IF EXISTS chk_rfm_template_key;

ALTER TABLE rfm_configs
    DROP COLUMN IF EXISTS active_template_type,
    DROP COLUMN IF EXISTS active_template_key,
    DROP COLUMN IF EXISTS custom_template_name,
    DROP COLUMN IF EXISTS custom_r_thresholds,
    DROP COLUMN IF EXISTS custom_f_thresholds;
