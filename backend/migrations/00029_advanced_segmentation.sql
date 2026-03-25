-- +goose Up

-- Behavioral segment rules — more expressive than SegmentFilter JSONB
CREATE TABLE segment_rules (
    id          SERIAL PRIMARY KEY,
    segment_id  INT NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    field       VARCHAR(50) NOT NULL, -- e.g. 'days_since_visit', 'total_orders', 'avg_check', 'loyalty_level'
    operator    VARCHAR(10) NOT NULL CHECK (operator IN ('eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'in', 'not_in', 'between')),
    value       JSONB NOT NULL,       -- scalar or array depending on operator
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_segment_rules_segment ON segment_rules(segment_id);

-- Client predictions — computed periodically by scheduler
CREATE TABLE client_predictions (
    id              SERIAL PRIMARY KEY,
    org_id          INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id       INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    churn_risk      REAL NOT NULL DEFAULT 0 CHECK (churn_risk >= 0 AND churn_risk <= 1),
    upsell_score    REAL NOT NULL DEFAULT 0 CHECK (upsell_score >= 0 AND upsell_score <= 1),
    predicted_value REAL NOT NULL DEFAULT 0,
    factors         JSONB NOT NULL DEFAULT '{}',
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(client_id)
);

CREATE INDEX idx_client_predictions_org ON client_predictions(org_id);
CREATE INDEX idx_client_predictions_churn ON client_predictions(org_id, churn_risk DESC);

-- +goose Down
DROP TABLE IF EXISTS client_predictions;
DROP TABLE IF EXISTS segment_rules;
