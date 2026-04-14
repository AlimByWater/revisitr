-- +goose Up
-- +goose StatementBegin

DO $$
DECLARE
    v_org_id INT;
    v_program_id INT;
    v_main_bot_id INT;
    v_menu_id INT;
    v_cat_id INT;
    v_pos_id INT;
BEGIN
    SELECT b.org_id
      INTO v_org_id
      FROM bots b
     WHERE b.username = 'baratie_demo_bot'
     ORDER BY CASE WHEN b.status = 'active' THEN 0 ELSE 1 END, b.id DESC
     LIMIT 1;

    IF v_org_id IS NULL THEN
        SELECT o.id
          INTO v_org_id
          FROM organizations o
         WHERE o.name = 'Baratie Demo'
         ORDER BY o.id DESC
         LIMIT 1;
    END IF;

    IF v_org_id IS NULL THEN
        INSERT INTO organizations (name, created_at)
        VALUES ('Baratie Demo', NOW())
        RETURNING id INTO v_org_id;
    ELSE
        UPDATE organizations
           SET name = 'Baratie Demo'
         WHERE id = v_org_id;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM bots WHERE org_id = v_org_id AND username = 'baratie_demo_bot'
    ) THEN
        INSERT INTO bots (
            org_id, name, token, username, status, settings,
            created_at, updated_at, is_managed
        ) VALUES (
            v_org_id,
            'Baratie',
            '',
            'baratie_demo_bot',
            'inactive',
            '{}'::jsonb,
            NOW(),
            NOW(),
            true
        );
    END IF;

    UPDATE bots
       SET name = 'Baratie',
           username = 'baratie_demo_bot',
           is_managed = COALESCE(is_managed, false) OR username = 'baratie_demo_bot',
           updated_at = NOW(),
           settings = jsonb_build_object(
               'modules', jsonb_build_array('loyalty', 'menu', 'booking', 'feedback', 'marketplace'),
               'buttons', jsonb_build_array(
                   jsonb_build_object(
                       'label', '🍽 Меню',
                       'type', 'text',
                       'value', E'🍽 Меню Baratie\\n\\nЗакуски\\n• Sniper Prawns — 490 ₽\\n• Grand Line Bruschetta — 420 ₽\\n• East Blue Oyster Trio — 560 ₽\\n\\nОсновные блюда\\n• Sea King Steak — 1390 ₽\\n• All Blue Sashimi — 1250 ₽\\n• Diable Jambe Pasta — 990 ₽\\n\\nДесерты\\n• Mera Mera Crème Brûlée — 450 ₽\\n• Going Merry Cheesecake — 470 ₽\\n\\nНапитки\\n• Cola Float Franky — 320 ₽\\n• Baratie Citrus Spritz — 390 ₽\\n\\nПолное меню уже загружено в систему Revisitr для демо.'
                   ),
                   jsonb_build_object(
                       'label', '🪑 Забронировать',
                       'type', 'text',
                       'value', E'🪑 Бронирование в Baratie\\n\\nДоступные зоны:\\n• Палуба — до 4 гостей\\n• Каюта — до 2 гостей\\n• VIP-каюта — до 6 гостей\\n\\nПопулярные слоты сегодня:\\n• 18:00\\n• 19:30\\n• 21:00\\n\\nДля демо скажите менеджеру: «Хочу столик в Baratie на 19:30».'
                   ),
                   jsonb_build_object(
                       'label', '🛍 Мерч',
                       'type', 'text',
                       'value', E'🛍 Мерч Baratie за дублоны\\n\\n• Фартук шефа Sanji — 900 дублонов\\n• Кружка All Blue — 450 дублонов\\n• Футболка Baratie Crew — 1200 дублонов\\n• Набор фирменных стикеров — 250 дублонов\\n• Подарочный сертификат капитана — 1500 дублонов\\n\\nКаталог товаров уже загружен в marketplace.'
                   ),
                   jsonb_build_object(
                       'label', '⭐ Оставить отзыв',
                       'type', 'text',
                       'value', E'⭐ Спасибо, что были у нас на борту!\\n\\nНам важно узнать:\\n• как вас встретили\\n• какое блюдо понравилось больше всего\\n• готовы ли вы рекомендовать Baratie друзьям\\n\\nДля демо ответьте одним сообщением в формате:\\n5 / Sea King Steak / Вернусь ещё'
                   )
               ),
               'registration_form', jsonb_build_array(
                   jsonb_build_object('name', 'first_name', 'label', 'Имя', 'type', 'text', 'required', true),
                   jsonb_build_object('name', 'phone', 'label', 'Телефон', 'type', 'phone', 'required', true),
                   jsonb_build_object('name', 'birth_date', 'label', 'Дата рождения', 'type', 'date', 'required', false)
               ),
               'welcome_message', E'Добро пожаловать в Baratie! 🏴‍☠️\\n\\nМы рады видеть вас на борту легендарного ресторана на воде.\\n\\nЗдесь вы можете:\\n• копить дублоны за каждый визит\\n• посмотреть авторское меню Sanji\\n• забронировать столик\\n• выбрать фирменный мерч\\n\\nДля начала пройдите короткую регистрацию.'
           )
     WHERE org_id = v_org_id
       AND username = 'baratie_demo_bot';

    SELECT b.id
      INTO v_main_bot_id
      FROM bots b
     WHERE b.org_id = v_org_id
       AND b.username = 'baratie_demo_bot'
     ORDER BY CASE WHEN b.status = 'active' THEN 0 ELSE 1 END, b.id DESC
     LIMIT 1;

    SELECT lp.id
      INTO v_program_id
      FROM loyalty_programs lp
     WHERE lp.org_id = v_org_id
     ORDER BY CASE WHEN lp.name = 'Baratie Rewards' THEN 0 ELSE 1 END, lp.id DESC
     LIMIT 1;

    IF v_program_id IS NULL THEN
        INSERT INTO loyalty_programs (org_id, name, type, is_active, config, created_at, updated_at)
        VALUES (
            v_org_id,
            'Baratie Rewards',
            'bonus',
            true,
            '{
                "currency_name": "дублонов",
                "welcome_bonus": 150
            }'::jsonb,
            NOW(),
            NOW()
        )
        RETURNING id INTO v_program_id;
    ELSE
        UPDATE loyalty_programs
           SET name = 'Baratie Rewards',
               type = 'bonus',
               is_active = true,
               config = '{
                   "currency_name": "дублонов",
                   "welcome_bonus": 150
               }'::jsonb,
               updated_at = NOW()
         WHERE id = v_program_id;
    END IF;

    DELETE FROM loyalty_levels WHERE program_id = v_program_id;

    INSERT INTO loyalty_levels (program_id, name, threshold, reward_percent, reward_type, reward_amount, sort_order)
    VALUES
        (v_program_id, 'Cabin Boy', 0, 0, 'percent', 0, 1),
        (v_program_id, 'Navigator', 100, 5, 'percent', 5, 2),
        (v_program_id, 'Captain', 500, 10, 'percent', 10, 3),
        (v_program_id, 'Pirate King', 1500, 15, 'percent', 15, 4);

    UPDATE bots
       SET program_id = v_program_id,
           updated_at = NOW()
     WHERE org_id = v_org_id
       AND username = 'baratie_demo_bot';

    DELETE FROM bot_pos_locations
     WHERE bot_id IN (SELECT id FROM bots WHERE org_id = v_org_id);

    DELETE FROM pos_locations WHERE org_id = v_org_id;

    INSERT INTO pos_locations (org_id, name, address, phone, schedule, is_active, created_at, updated_at)
    VALUES
        (
            v_org_id,
            'Baratie — Главная палуба',
            'Pier 39, East Blue Marina',
            '+7 (495) 555-17-17',
            '{
                "monday": {"open": "12:00", "close": "23:00"},
                "tuesday": {"open": "12:00", "close": "23:00"},
                "wednesday": {"open": "12:00", "close": "23:00"},
                "thursday": {"open": "12:00", "close": "23:00"},
                "friday": {"open": "12:00", "close": "00:00"},
                "saturday": {"open": "10:00", "close": "00:00"},
                "sunday": {"open": "10:00", "close": "22:00"}
            }'::jsonb,
            true,
            NOW(),
            NOW()
        ),
        (
            v_org_id,
            'Baratie — VIP-каюта',
            'East Blue Marina, Deck C',
            '+7 (495) 555-18-18',
            '{
                "monday": {"open": "18:00", "close": "23:00"},
                "tuesday": {"open": "18:00", "close": "23:00"},
                "wednesday": {"open": "18:00", "close": "23:00"},
                "thursday": {"open": "18:00", "close": "23:00"},
                "friday": {"open": "18:00", "close": "00:00"},
                "saturday": {"open": "15:00", "close": "00:00"},
                "sunday": {"open": "15:00", "close": "22:00"}
            }'::jsonb,
            true,
            NOW(),
            NOW()
        );

    IF v_main_bot_id IS NOT NULL THEN
        FOR v_pos_id IN
            SELECT id FROM pos_locations WHERE org_id = v_org_id
        LOOP
            INSERT INTO bot_pos_locations (bot_id, pos_id)
            VALUES (v_main_bot_id, v_pos_id)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;

    DELETE FROM menu_items
     WHERE category_id IN (
         SELECT mc.id
           FROM menu_categories mc
           JOIN menus m ON m.id = mc.menu_id
          WHERE m.org_id = v_org_id
     );
    DELETE FROM menu_categories
     WHERE menu_id IN (SELECT id FROM menus WHERE org_id = v_org_id);
    DELETE FROM menus WHERE org_id = v_org_id;

    INSERT INTO menus (org_id, name, source, created_at, updated_at)
    VALUES (v_org_id, 'Baratie Signature Menu', 'manual', NOW(), NOW())
    RETURNING id INTO v_menu_id;

    INSERT INTO menu_categories (menu_id, name, sort_order)
    VALUES (v_menu_id, 'Закуски', 1)
    RETURNING id INTO v_cat_id;
    INSERT INTO menu_items (category_id, name, description, price, tags, sort_order)
    VALUES
        (v_cat_id, 'Sniper Prawns', 'Тигровые креветки в остром citrus glaze.', 490, '["seafood", "signature"]'::jsonb, 1),
        (v_cat_id, 'Grand Line Bruschetta', 'Хрустящий багет с томатами, анчоусами и базиликом.', 420, '["starter"]'::jsonb, 2),
        (v_cat_id, 'East Blue Oyster Trio', 'Три свежие устрицы с юдзу и морской солью.', 560, '["seafood"]'::jsonb, 3);

    INSERT INTO menu_categories (menu_id, name, sort_order)
    VALUES (v_menu_id, 'Основные блюда', 2)
    RETURNING id INTO v_cat_id;
    INSERT INTO menu_items (category_id, name, description, price, tags, sort_order)
    VALUES
        (v_cat_id, 'Sea King Steak', 'Фирменный стейк Sanji с перечным соусом и картофелем.', 1390, '["main", "best_seller"]'::jsonb, 1),
        (v_cat_id, 'All Blue Sashimi', 'Ассорти из тунца, лосося и хамачи с цитрусовым понзу.', 1250, '["main", "seafood"]'::jsonb, 2),
        (v_cat_id, 'Diable Jambe Pasta', 'Паста с острым томатным соусом и морепродуктами.', 990, '["main", "spicy"]'::jsonb, 3),
        (v_cat_id, 'Merry Roast Chicken', 'Курица с травами и сливочным соусом.', 870, '["main"]'::jsonb, 4);

    INSERT INTO menu_categories (menu_id, name, sort_order)
    VALUES (v_menu_id, 'Десерты', 3)
    RETURNING id INTO v_cat_id;
    INSERT INTO menu_items (category_id, name, description, price, tags, sort_order)
    VALUES
        (v_cat_id, 'Mera Mera Crème Brûlée', 'Крем-брюле с карамельной корочкой и апельсином.', 450, '["dessert"]'::jsonb, 1),
        (v_cat_id, 'Going Merry Cheesecake', 'Ванильный чизкейк с солёной карамелью.', 470, '["dessert", "best_seller"]'::jsonb, 2),
        (v_cat_id, 'Chopper Cotton Candy', 'Нежный мусс с ягодами и облаком сахарной ваты.', 390, '["dessert"]'::jsonb, 3);

    INSERT INTO menu_categories (menu_id, name, sort_order)
    VALUES (v_menu_id, 'Напитки', 4)
    RETURNING id INTO v_cat_id;
    INSERT INTO menu_items (category_id, name, description, price, tags, sort_order)
    VALUES
        (v_cat_id, 'Cola Float Franky', 'Фирменная cola float с ванильным мороженым.', 320, '["drink"]'::jsonb, 1),
        (v_cat_id, 'Baratie Citrus Spritz', 'Освежающий спритц с цитрусами и розмарином.', 390, '["drink", "signature"]'::jsonb, 2),
        (v_cat_id, 'All Blue Tea', 'Синий чай анчан с лимоном и мёдом.', 290, '["drink"]'::jsonb, 3);

    DELETE FROM marketplace_products WHERE org_id = v_org_id;

    INSERT INTO marketplace_products (org_id, name, description, image_url, price_points, stock, is_active, sort_order, created_at, updated_at)
    VALUES
        (v_org_id, 'Фартук шефа Sanji', 'Фирменный фартук Baratie для кухни и гриля.', '', 900, 12, true, 1, NOW(), NOW()),
        (v_org_id, 'Кружка All Blue', 'Керамическая кружка с картой All Blue.', '', 450, 30, true, 2, NOW(), NOW()),
        (v_org_id, 'Футболка Baratie Crew', 'Чёрная футболка экипажа Baratie.', '', 1200, 20, true, 3, NOW(), NOW()),
        (v_org_id, 'Набор стикеров East Blue', 'Пак стикеров с героями и блюдами Baratie.', '', 250, 100, true, 4, NOW(), NOW()),
        (v_org_id, 'Подарочный сертификат капитана', 'Сертификат на ужин для двоих в VIP-каюте.', '', 1500, 10, true, 5, NOW(), NOW());
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
DECLARE
    v_org_id INT;
BEGIN
    SELECT b.org_id
      INTO v_org_id
      FROM bots b
     WHERE b.username = 'baratie_demo_bot'
     ORDER BY b.id DESC
     LIMIT 1;

    IF v_org_id IS NULL THEN
        SELECT o.id
          INTO v_org_id
          FROM organizations o
         WHERE o.name = 'Baratie Demo'
         ORDER BY o.id DESC
         LIMIT 1;
    END IF;

    IF v_org_id IS NULL THEN
        RETURN;
    END IF;

    DELETE FROM bot_pos_locations
     WHERE bot_id IN (SELECT id FROM bots WHERE org_id = v_org_id);
    DELETE FROM pos_locations WHERE org_id = v_org_id;
    DELETE FROM menu_items
     WHERE category_id IN (
         SELECT mc.id
           FROM menu_categories mc
           JOIN menus m ON m.id = mc.menu_id
          WHERE m.org_id = v_org_id
     );
    DELETE FROM menu_categories
     WHERE menu_id IN (SELECT id FROM menus WHERE org_id = v_org_id);
    DELETE FROM menus WHERE org_id = v_org_id;
    DELETE FROM marketplace_products WHERE org_id = v_org_id;

    UPDATE bots
       SET settings = '{}'::jsonb,
           program_id = NULL,
           updated_at = NOW()
     WHERE org_id = v_org_id
       AND username = 'baratie_demo_bot';

    DELETE FROM loyalty_levels
     WHERE program_id IN (
         SELECT id FROM loyalty_programs WHERE org_id = v_org_id AND name = 'Baratie Rewards'
     );
    DELETE FROM loyalty_programs
     WHERE org_id = v_org_id
       AND name = 'Baratie Rewards';
END $$;
-- +goose StatementEnd
