-- +goose Up
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_daily_sales AS
SELECT
    DATE_TRUNC('day', lt.created_at) AS day,
    bc.bot_id,
    b.org_id,
    COUNT(DISTINCT lt.client_id)     AS unique_clients,
    COUNT(*)                         AS transaction_count,
    SUM(lt.amount)                   AS total_amount,
    AVG(lt.amount)                   AS avg_amount
FROM loyalty_transactions lt
JOIN bot_clients bc ON lt.client_id = bc.id
JOIN bots b ON bc.bot_id = b.id
WHERE lt.type = 'earn'
GROUP BY 1, 2, 3
WITH DATA;
CREATE UNIQUE INDEX ON mv_daily_sales(day, bot_id);

-- mv_loyalty_stats stores only registration aggregates per day.
-- active_clients is computed at runtime (analytics repo), not in MV,
-- because NOW() in MV is fixed at REFRESH time and becomes stale between refreshes.
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_loyalty_stats AS
SELECT
    b.org_id,
    bc.bot_id,
    DATE_TRUNC('day', bc.registered_at) AS day,
    COUNT(*)                             AS new_clients
FROM bot_clients bc
JOIN bots b ON bc.bot_id = b.id
GROUP BY 1, 2, 3
WITH DATA;
CREATE UNIQUE INDEX ON mv_loyalty_stats(org_id, bot_id, day);

-- +goose Down
DROP MATERIALIZED VIEW IF EXISTS mv_loyalty_stats;
DROP MATERIALIZED VIEW IF EXISTS mv_daily_sales;
