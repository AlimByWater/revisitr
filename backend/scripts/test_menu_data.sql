-- Test menu data for local development
-- Run: psql -U revisitr -d revisitr -f backend/scripts/test_menu_data.sql

-- Menu
INSERT INTO menus (id, org_id, name, source)
VALUES (1, 1, 'Основное меню', 'manual')
ON CONFLICT (id) DO NOTHING;
SELECT setval('menus_id_seq', GREATEST((SELECT MAX(id) FROM menus), 1));

-- Categories
INSERT INTO menu_categories (id, menu_id, name, sort_order) VALUES
(1, 1, 'Закуски', 0),
(2, 1, 'Салаты', 1),
(3, 1, 'Горячее', 2),
(4, 1, 'Напитки', 3),
(5, 1, 'Десерты', 4)
ON CONFLICT (id) DO NOTHING;
SELECT setval('menu_categories_id_seq', GREATEST((SELECT MAX(id) FROM menu_categories), 1));

-- Items: Закуски
INSERT INTO menu_items (id, category_id, name, description, price, weight, image_url, is_available, sort_order) VALUES
(1, 1, 'Blinis Demidoff', 'Гречневые блины с крем-фрешем и икрой осетра, подаются в холодной подаче.', 1390, '140 г', 'https://picsum.photos/seed/blinis/400/300', true, 0),
(2, 1, 'Bruschetta Classica', 'Хрустящий багет с томатами, базиликом и оливковым маслом.', 690, '180 г', 'https://picsum.photos/seed/bruschetta/400/300', true, 1),
(3, 1, 'Tartare di Manzo', 'Тартар из мраморной говядины с каперсами и перепелиным яйцом.', 1590, '200 г', 'https://picsum.photos/seed/tartare/400/300', true, 2)
ON CONFLICT (id) DO NOTHING;

-- Items: Салаты
INSERT INTO menu_items (id, category_id, name, description, price, weight, image_url, is_available, sort_order) VALUES
(4, 2, 'Цезарь с курицей', 'Классический салат с куриным филе, пармезаном и соусом цезарь.', 890, '250 г', 'https://picsum.photos/seed/caesar/400/300', true, 0),
(5, 2, 'Греческий салат', 'Свежие овощи с сыром фета и оливками.', 750, '220 г', 'https://picsum.photos/seed/greek/400/300', true, 1)
ON CONFLICT (id) DO NOTHING;

-- Items: Горячее
INSERT INTO menu_items (id, category_id, name, description, price, weight, image_url, is_available, sort_order) VALUES
(6, 3, 'Ribeye Steak', 'Мраморный рибай с трюфельным соусом и овощами гриль.', 3290, '350 г', 'https://picsum.photos/seed/ribeye/400/300', true, 0),
(7, 3, 'Лосось на гриле', 'Филе лосося с лимонным маслом и спаржей.', 2590, '280 г', 'https://picsum.photos/seed/salmon/400/300', true, 1),
(8, 3, 'Паста Карбонара', 'Домашняя паста с беконом, пармезаном и перепелиным желтком.', 1290, '320 г', 'https://picsum.photos/seed/carbonara/400/300', true, 2)
ON CONFLICT (id) DO NOTHING;

-- Items: Напитки
INSERT INTO menu_items (id, category_id, name, description, price, weight, image_url, is_available, sort_order) VALUES
(9, 4, 'Апероль Шприц', 'Классический итальянский аперитив с просекко.', 890, '200 мл', 'https://picsum.photos/seed/aperol/400/300', true, 0),
(10, 4, 'Эспрессо', 'Классический итальянский эспрессо 30 мл.', 250, '30 мл', 'https://picsum.photos/seed/espresso/400/300', true, 1),
(11, 4, 'Матча Лаванда', 'Латте на альтернативном молоке с сиропом лаванды.', 590, '300 мл', 'https://picsum.photos/seed/matcha/400/300', true, 2),
(12, 4, 'Лимонад Маракуйя', 'Домашний лимонад со свежей маракуйей.', 490, '400 мл', 'https://picsum.photos/seed/lemonade/400/300', true, 3)
ON CONFLICT (id) DO NOTHING;

-- Items: Десерты
INSERT INTO menu_items (id, category_id, name, description, price, weight, image_url, is_available, sort_order) VALUES
(13, 5, 'Тирамису', 'Классический итальянский десерт со сливочным маскарпоне.', 690, '180 г', 'https://picsum.photos/seed/tiramisu/400/300', true, 0),
(14, 5, 'Шоколадный фондан', 'Горячий шоколадный десерт с жидкой сердцевиной.', 790, '150 г', 'https://picsum.photos/seed/fondant/400/300', true, 1),
(15, 5, 'Сорбет Манго', 'Итальянский сорбет из манго.', 490, '120 г', 'https://picsum.photos/seed/sorbet/400/300', true, 2)
ON CONFLICT (id) DO NOTHING;
SELECT setval('menu_items_id_seq', GREATEST((SELECT MAX(id) FROM menu_items), 1));

-- Bind menu to POS location
INSERT INTO menu_pos_bindings (menu_id, pos_id, is_active) VALUES
(1, 1, true)
ON CONFLICT DO NOTHING;

-- Bind bot to POS location (for bot_id = 1 from seed data)
INSERT INTO bot_pos_locations (bot_id, pos_id) VALUES
(1, 1)
ON CONFLICT DO NOTHING;
