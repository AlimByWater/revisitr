-- +goose Up

ALTER TABLE auto_scenarios
    ADD COLUMN actions JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN timing JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN conditions JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN is_template BOOLEAN DEFAULT false,
    ADD COLUMN template_key VARCHAR(50);

ALTER TABLE auto_scenarios
    DROP CONSTRAINT IF EXISTS auto_scenarios_trigger_type_check;
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_trigger_type_check
    CHECK (trigger_type IN (
        'inactive_days', 'visit_count', 'bonus_threshold', 'level_up',
        'birthday', 'holiday', 'registration', 'level_change'
    ));

CREATE TABLE auto_action_log (
    id           SERIAL PRIMARY KEY,
    scenario_id  INT NOT NULL REFERENCES auto_scenarios(id) ON DELETE CASCADE,
    client_id    INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    action_type  VARCHAR(30) NOT NULL,
    action_data  JSONB DEFAULT '{}'::jsonb,
    result       VARCHAR(20) NOT NULL CHECK (result IN ('success', 'failed', 'skipped')),
    error_msg    TEXT,
    executed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_auto_action_log_scenario ON auto_action_log(scenario_id);
CREATE INDEX idx_auto_action_log_client   ON auto_action_log(client_id);
CREATE INDEX idx_auto_action_log_date     ON auto_action_log(executed_at);

CREATE TABLE auto_action_dedup (
    id           SERIAL PRIMARY KEY,
    scenario_id  INT NOT NULL REFERENCES auto_scenarios(id) ON DELETE CASCADE,
    client_id    INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    trigger_key  VARCHAR(100) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scenario_id, client_id, trigger_key)
);

-- +goose Down
DROP TABLE IF EXISTS auto_action_dedup;
DROP TABLE IF EXISTS auto_action_log;
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_trigger_type_check;
ALTER TABLE auto_scenarios ADD CONSTRAINT auto_scenarios_trigger_type_check
    CHECK (trigger_type IN ('inactive_days', 'visit_count', 'bonus_threshold', 'level_up', 'birthday'));
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS template_key;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS is_template;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS conditions;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS timing;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS actions;
