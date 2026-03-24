-- +goose Up

ALTER TABLE promo_codes
    ADD COLUMN channel VARCHAR(50),
    ADD COLUMN per_user_limit INT DEFAULT 1,
    ADD COLUMN description TEXT;

ALTER TABLE promotions
    ADD COLUMN recurrence VARCHAR(20) DEFAULT 'one_time' CHECK (recurrence IN ('one_time', 'daily', 'weekly', 'monthly'));

CREATE VIEW promo_channel_analytics AS
SELECT
    pc.org_id,
    pc.channel,
    COUNT(DISTINCT pc.id) AS code_count,
    COALESCE(SUM(pc.usage_count), 0) AS total_usages,
    COUNT(DISTINCT pu.client_id) AS unique_clients
FROM promo_codes pc
LEFT JOIN promotion_usages pu ON pu.promo_code_id = pc.id
GROUP BY pc.org_id, pc.channel;

-- +goose Down
DROP VIEW IF EXISTS promo_channel_analytics;
ALTER TABLE promotions DROP COLUMN IF EXISTS recurrence;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS description;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS per_user_limit;
ALTER TABLE promo_codes DROP COLUMN IF EXISTS channel;
