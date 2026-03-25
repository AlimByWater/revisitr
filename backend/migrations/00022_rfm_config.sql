-- +goose Up

-- Конфигурация RFM-расчёта для организации
CREATE TABLE rfm_configs (
    id                SERIAL PRIMARY KEY,
    org_id            INT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    period_days       INT NOT NULL DEFAULT 365,
    recalc_interval   VARCHAR(20) NOT NULL DEFAULT '24h',
    last_calc_at      TIMESTAMPTZ,
    clients_processed INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- История изменений RFM-сегментов (для трендового графика)
CREATE TABLE rfm_history (
    id            SERIAL PRIMARY KEY,
    org_id        INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    segment       VARCHAR(50) NOT NULL,
    client_count  INT NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_rfm_history_org_date ON rfm_history(org_id, calculated_at);

-- +goose Down
DROP TABLE IF EXISTS rfm_history;
DROP TABLE IF EXISTS rfm_configs;
