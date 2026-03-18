-- +goose Up
CREATE TABLE campaigns (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id),
    bot_id INT NOT NULL REFERENCES bots(id),
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('manual', 'auto')),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed')),
    audience_filter JSONB DEFAULT '{}'::jsonb,
    message TEXT NOT NULL,
    media_url VARCHAR(500),
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    stats JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE campaign_messages (
    id SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    client_id INT NOT NULL REFERENCES bot_clients(id),
    telegram_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auto_scenarios (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id),
    bot_id INT NOT NULL REFERENCES bots(id),
    name VARCHAR(200) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('inactive_days', 'visit_count', 'bonus_threshold', 'level_up', 'birthday')),
    trigger_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    message TEXT NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_org_id ON campaigns(org_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaign_messages_campaign_id ON campaign_messages(campaign_id);
CREATE INDEX idx_campaign_messages_status ON campaign_messages(status);
CREATE INDEX idx_auto_scenarios_org_id ON auto_scenarios(org_id);

-- +goose Down
DROP TABLE IF EXISTS auto_scenarios;
DROP TABLE IF EXISTS campaign_messages;
DROP TABLE IF EXISTS campaigns;
