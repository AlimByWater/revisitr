-- +goose Up

ALTER TABLE campaigns
    ADD COLUMN buttons JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN tracking_mode VARCHAR(20) DEFAULT 'utm' CHECK (tracking_mode IN ('utm', 'buttons', 'both', 'none'));

ALTER TABLE campaigns
    DROP CONSTRAINT IF EXISTS campaigns_status_check;
ALTER TABLE campaigns
    ADD CONSTRAINT campaigns_status_check
    CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'completed', 'failed'));

CREATE TABLE campaign_clicks (
    id          SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    client_id   INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    button_idx  INT,
    url         TEXT,
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_campaign_clicks_campaign ON campaign_clicks(campaign_id);
CREATE INDEX idx_campaign_clicks_client   ON campaign_clicks(client_id);

CREATE TABLE campaign_queue_status (
    id          SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    queued_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    worker_id   VARCHAR(50),
    status      VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'done', 'failed'))
);

-- +goose Down
DROP TABLE IF EXISTS campaign_queue_status;
DROP TABLE IF EXISTS campaign_clicks;
ALTER TABLE campaigns DROP COLUMN IF EXISTS buttons;
ALTER TABLE campaigns DROP COLUMN IF EXISTS tracking_mode;
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_status_check;
ALTER TABLE campaigns ADD CONSTRAINT campaigns_status_check
    CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed'));
