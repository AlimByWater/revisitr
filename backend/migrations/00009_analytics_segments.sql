-- +goose Up
CREATE TABLE segments (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('rfm', 'custom')),
    filter      JSONB NOT NULL DEFAULT '{}',
    auto_assign BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_segments_org_id ON segments(org_id);

CREATE TABLE segment_clients (
    segment_id  INT NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    client_id   INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (segment_id, client_id)
);

ALTER TABLE bot_clients
    ADD COLUMN IF NOT EXISTS rfm_recency    INT,
    ADD COLUMN IF NOT EXISTS rfm_frequency  INT,
    ADD COLUMN IF NOT EXISTS rfm_monetary   NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS rfm_segment    VARCHAR(50),
    ADD COLUMN IF NOT EXISTS rfm_updated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE bot_clients
    DROP COLUMN IF EXISTS rfm_recency,
    DROP COLUMN IF EXISTS rfm_frequency,
    DROP COLUMN IF EXISTS rfm_monetary,
    DROP COLUMN IF EXISTS rfm_segment,
    DROP COLUMN IF EXISTS rfm_updated_at;
DROP TABLE IF EXISTS segment_clients;
DROP TABLE IF EXISTS segments;
