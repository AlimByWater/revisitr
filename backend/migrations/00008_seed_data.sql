-- +goose Up

-- Organization
INSERT INTO organizations (id, name)
VALUES (1, 'Demo Organization')
ON CONFLICT (id) DO NOTHING;

-- Admin user: admin@revisitr.com / admin123
INSERT INTO users (id, email, phone, name, password_hash, role, org_id)
VALUES (1, 'admin@revisitr.com', '+79001234567', 'Admin', '$2a$10$7lev6hUSLytRZvQ/5UNJlOcilslXIZg1/QH/xqg56IQ9iNx51vw9i', 'owner', 1)
ON CONFLICT (id) DO NOTHING;

UPDATE organizations SET owner_id = 1 WHERE id = 1;

-- Reset sequences
SELECT setval('organizations_id_seq', GREATEST((SELECT MAX(id) FROM organizations), 1));
SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 1));

-- Demo bot (fake token — replace with real one to activate)
INSERT INTO bots (id, org_id, name, token, username, status, settings)
VALUES (1, 1, 'Demo Кафе Бот', 'REPLACE_WITH_REAL_BOT_TOKEN', 'demo_cafe_bot', 'inactive',
  '{"modules": ["loyalty", "registration"], "buttons": [{"label": "Мой баланс", "type": "loyalty", "value": "balance"}, {"label": "Акции", "type": "url", "value": "https://example.com"}], "registration_form": [{"name": "phone", "label": "Телефон", "type": "phone", "required": true}, {"name": "birth_date", "label": "Дата рождения", "type": "date", "required": false}], "welcome_message": "Добро пожаловать! Вы подключены к программе лояльности."}'::jsonb)
ON CONFLICT (id) DO NOTHING;

SELECT setval('bots_id_seq', GREATEST((SELECT MAX(id) FROM bots), 1));

-- Example clients
INSERT INTO bot_clients (id, bot_id, telegram_id, username, first_name, last_name, phone, gender, birth_date, city, tags)
VALUES
  (1, 1, 100001, 'ivan_example', 'Иван', 'Петров', '+79111111111', 'male', '1990-05-15', 'Москва', '["vip", "regular"]'::jsonb),
  (2, 1, 100002, 'maria_example', 'Мария', 'Сидорова', '+79222222222', 'female', '1995-08-22', 'Москва', '["new"]'::jsonb),
  (3, 1, 100003, 'alex_example', 'Алексей', 'Козлов', '+79333333333', 'male', '1988-12-01', 'Санкт-Петербург', '["regular"]'::jsonb),
  (4, 1, 100004, 'elena_example', 'Елена', 'Волкова', NULL, 'female', NULL, NULL, '[]'::jsonb),
  (5, 1, 100005, 'dmitry_example', 'Дмитрий', 'Новиков', '+79555555555', 'male', '1992-03-10', 'Москва', '["vip"]'::jsonb)
ON CONFLICT (id) DO NOTHING;

SELECT setval('bot_clients_id_seq', GREATEST((SELECT MAX(id) FROM bot_clients), 1));

-- Loyalty program: бонусная система
INSERT INTO loyalty_programs (id, org_id, name, type, config, is_active)
VALUES (1, 1, 'Бонусная программа', 'bonus', '{"welcome_bonus": 100, "currency_name": "баллы"}'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

SELECT setval('loyalty_programs_id_seq', GREATEST((SELECT MAX(id) FROM loyalty_programs), 1));

-- Loyalty levels
INSERT INTO loyalty_levels (id, program_id, name, threshold, reward_percent, sort_order)
VALUES
  (1, 1, 'Бронза', 0, 5.00, 1),
  (2, 1, 'Серебро', 5000, 7.50, 2),
  (3, 1, 'Золото', 15000, 10.00, 3),
  (4, 1, 'Платина', 50000, 15.00, 4)
ON CONFLICT (id) DO NOTHING;

SELECT setval('loyalty_levels_id_seq', GREATEST((SELECT MAX(id) FROM loyalty_levels), 1));

-- Client loyalty records
INSERT INTO client_loyalty (id, client_id, program_id, level_id, balance, total_earned, total_spent)
VALUES
  (1, 1, 1, 3, 2450.00, 18500.00, 16050.00),  -- Иван: Золото
  (2, 2, 1, 1, 100.00, 100.00, 0.00),           -- Мария: Бронза (welcome bonus)
  (3, 3, 1, 2, 890.00, 7200.00, 6310.00),        -- Алексей: Серебро
  (4, 5, 1, 4, 5200.00, 62000.00, 56800.00)      -- Дмитрий: Платина
ON CONFLICT (id) DO NOTHING;

SELECT setval('client_loyalty_id_seq', GREATEST((SELECT MAX(id) FROM client_loyalty), 1));

-- Sample transactions
INSERT INTO loyalty_transactions (client_id, program_id, type, amount, balance_after, description, created_by, created_at)
VALUES
  (1, 1, 'earn', 500.00, 2450.00, 'Заказ #1042', 1, NOW() - INTERVAL '2 days'),
  (1, 1, 'spend', 200.00, 1950.00, 'Списание за десерт', 1, NOW() - INTERVAL '1 day'),
  (1, 1, 'earn', 500.00, 2450.00, 'Заказ #1055', 1, NOW()),
  (2, 1, 'earn', 100.00, 100.00, 'Приветственный бонус', 1, NOW() - INTERVAL '3 days'),
  (3, 1, 'earn', 350.00, 890.00, 'Заказ #1038', 1, NOW() - INTERVAL '5 days'),
  (5, 1, 'earn', 1200.00, 5200.00, 'Банкет #87', 1, NOW() - INTERVAL '1 day');

-- POS location
INSERT INTO pos_locations (id, org_id, name, address, phone, schedule, is_active)
VALUES (1, 1, 'Кафе "Центральное"', 'ул. Пушкина, д. 10', '+74951234567',
  '{"mon": {"open": "09:00", "close": "23:00"}, "tue": {"open": "09:00", "close": "23:00"}, "wed": {"open": "09:00", "close": "23:00"}, "thu": {"open": "09:00", "close": "23:00"}, "fri": {"open": "09:00", "close": "00:00"}, "sat": {"open": "10:00", "close": "00:00"}, "sun": {"open": "10:00", "close": "22:00"}}'::jsonb,
  true)
ON CONFLICT (id) DO NOTHING;

SELECT setval('pos_locations_id_seq', GREATEST((SELECT MAX(id) FROM pos_locations), 1));

-- +goose Down

DELETE FROM loyalty_transactions WHERE created_by = 1;
DELETE FROM client_loyalty WHERE program_id = 1;
DELETE FROM loyalty_levels WHERE program_id = 1;
DELETE FROM loyalty_programs WHERE id = 1;
DELETE FROM bot_clients WHERE bot_id = 1;
DELETE FROM bots WHERE id = 1;
DELETE FROM pos_locations WHERE org_id = 1;
DELETE FROM users WHERE id = 1;
DELETE FROM organizations WHERE id = 1;
