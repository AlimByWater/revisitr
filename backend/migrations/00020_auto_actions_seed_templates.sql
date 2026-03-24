-- +goose Up

-- Templates are global (not owned by any org/bot), so allow NULLs.
ALTER TABLE auto_scenarios ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE auto_scenarios ALTER COLUMN bot_id DROP NOT NULL;
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_org_id_fkey;
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_bot_id_fkey;
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_org_id_fkey FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_bot_id_fkey FOREIGN KEY (bot_id) REFERENCES bots(id);

INSERT INTO auto_scenarios (org_id, bot_id, name, trigger_type, trigger_config, message, is_template, template_key, actions, timing, is_active)
VALUES
(NULL, NULL, 'День рождения клиента', 'birthday',
 '{}', '{name}, с днём рождения! 🎂',
 true, 'tpl_birthday',
 '[{"type":"bonus","amount":500},{"type":"campaign","template":"birthday"}]',
 '{"days_before":0,"days_after":0}',
 false),

(NULL, NULL, '8 Марта', 'holiday',
 '{}', '{name}, поздравляем с 8 Марта! 💐',
 true, 'tpl_8march',
 '[{"type":"bonus","amount":300},{"type":"campaign","template":"holiday"}]',
 '{"month":3,"day":8}',
 false),

(NULL, NULL, 'Новый год', 'holiday',
 '{}', '{name}, с Новым годом! 🎄',
 true, 'tpl_newyear',
 '[{"type":"promo_code","discount":10},{"type":"campaign","template":"newyear"}]',
 '{"month":12,"day":31}',
 false),

(NULL, NULL, 'Неактивность 30 дней', 'inactive_days',
 '{"days":30}', '{name}, мы скучаем! Вот вам подарок 🎁',
 true, 'tpl_inactive30',
 '[{"type":"bonus","amount":200},{"type":"campaign","template":"comeback"}]',
 '{}',
 false);

-- +goose Down
DELETE FROM auto_scenarios WHERE is_template = true AND template_key IN ('tpl_birthday', 'tpl_8march', 'tpl_newyear', 'tpl_inactive30');
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_org_id_fkey;
ALTER TABLE auto_scenarios DROP CONSTRAINT IF EXISTS auto_scenarios_bot_id_fkey;
ALTER TABLE auto_scenarios ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE auto_scenarios ALTER COLUMN bot_id SET NOT NULL;
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_org_id_fkey FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE auto_scenarios
    ADD CONSTRAINT auto_scenarios_bot_id_fkey FOREIGN KEY (bot_id) REFERENCES bots(id);
