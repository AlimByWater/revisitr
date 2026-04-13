-- +goose Up

-- Create Baratie Demo organization
INSERT INTO organizations (name, created_at, updated_at)
VALUES ('Baratie Demo', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Create demo bot (token filled manually or via managed bots)
INSERT INTO bots (org_id, name, token, username, status, settings, created_at, updated_at)
SELECT o.id, 'Baratie', '', 'baratie_demo_bot', 'inactive',
    '{
        "modules": ["loyalty", "menu", "booking", "feedback", "marketplace"],
        "buttons": [
            {"label": "🍽 Меню", "type": "text", "value": "Используйте кнопку меню для просмотра блюд"},
            {"label": "📍 Как добраться", "type": "text", "value": "Ресторан Baratie находится в открытом море East Blue. Координаты: 12°34''N 56°78''E"}
        ],
        "registration_form": [
            {"name": "first_name", "label": "Имя", "type": "text", "required": true},
            {"name": "phone", "label": "Телефон", "type": "phone", "required": true},
            {"name": "birth_date", "label": "Дата рождения", "type": "date", "required": false}
        ],
        "welcome_message": "Добро пожаловать в Baratie! 🏴‍☠️\n\nМы рады видеть вас на борту лучшего ресторана East Blue.\n\nЗдесь вы можете:\n• Копить бонусы за каждый визит\n• Просматривать наше меню\n• Бронировать столик\n• Оставлять отзывы\n\nДля начала пройдите короткую регистрацию."
    }'::jsonb,
    NOW(), NOW()
FROM organizations o WHERE o.name = 'Baratie Demo'
ON CONFLICT DO NOTHING;

-- Create loyalty program
INSERT INTO loyalty_programs (org_id, name, is_active, config, created_at, updated_at)
SELECT o.id, 'Baratie Rewards', true,
    '{
        "currency_name": "дублонов",
        "earn_rate": 1,
        "earn_per": 100,
        "redeem_rate": 10,
        "welcome_bonus": 50
    }'::jsonb,
    NOW(), NOW()
FROM organizations o WHERE o.name = 'Baratie Demo'
ON CONFLICT DO NOTHING;

-- Create loyalty levels
INSERT INTO loyalty_levels (program_id, name, threshold, bonus_percent, sort_order, created_at, updated_at)
SELECT lp.id, level_name, threshold, bonus_pct, sort_ord, NOW(), NOW()
FROM loyalty_programs lp
JOIN organizations o ON o.id = lp.org_id
CROSS JOIN (VALUES
    ('Cabin Boy', 0, 0, 1),
    ('Navigator', 100, 5, 2),
    ('Captain', 500, 10, 3),
    ('Pirate King', 1500, 15, 4)
) AS levels(level_name, threshold, bonus_pct, sort_ord)
WHERE o.name = 'Baratie Demo'
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM loyalty_levels WHERE program_id IN (
    SELECT lp.id FROM loyalty_programs lp
    JOIN organizations o ON o.id = lp.org_id
    WHERE o.name = 'Baratie Demo'
);

DELETE FROM loyalty_programs WHERE org_id IN (
    SELECT id FROM organizations WHERE name = 'Baratie Demo'
);

DELETE FROM bots WHERE org_id IN (
    SELECT id FROM organizations WHERE name = 'Baratie Demo'
);

DELETE FROM organizations WHERE name = 'Baratie Demo';
