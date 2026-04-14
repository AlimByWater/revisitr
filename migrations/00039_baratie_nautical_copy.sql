-- +goose Up
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(
                        jsonb_set(
                            jsonb_set(
                                settings,
                                '{welcome_message}',
                                to_jsonb((
                                    E'⋆｡°✩ ⚓︎ Добро пожаловать на борт Baratie ⚓︎ ✩°｡⋆\n\n╭──────༺🪸༻──────╮\nЗдесь вас ждут:\n• дублоны за визиты\n• авторское меню Sanji\n• бронь палубы и кают\n• корабельная лавка\n╰──────༺🦈༻──────╯\n\nПоделитесь номером телефона, и мы впишем вас в список гостей.'
                                )::text),
                                true
                            ),
                            '{buttons,1,label}',
                            to_jsonb(('🪑 Бронь каюты')::text),
                            true
                        ),
                        '{buttons,1,value}',
                        to_jsonb((
                            E'╭──────༺⚓༻──────╮\nБронь каюты в Baratie\n╰──────༺⚓༻──────╯\n\n🪑 Палуба — до 4 гостей\n🪑 Каюта — до 2 гостей\n🪑 VIP-каюта — до 6 гостей\n\nПопулярные слоты:\n• 18:00\n• 19:30\n• 21:00\n\nНапишите менеджеру: «Хочу столик в Baratie на 19:30».'
                        )::text),
                        true
                    ),
                    '{buttons,2,label}',
                    to_jsonb(('🛍 Корабельная лавка')::text),
                    true
                ),
                '{buttons,2,value}',
                to_jsonb((
                    E'╭──────༺🐚༻──────╮\nКорабельная лавка Baratie\n╰──────༺🐚༻──────╯\n\n🪙 За дублоны можно взять:\n• Фартук шефа Sanji — 900\n• Кружка All Blue — 450\n• Футболка Baratie Crew — 1200\n• Набор стикеров East Blue — 250\n• Подарочный сертификат капитана — 1500\n\nСпросите у экипажа, если хотите оформить заказ.'
                )::text),
                true
            ),
            '{buttons,3,label}',
            to_jsonb(('📖 Бортовой отзыв')::text),
            true
        ),
        '{buttons,3,value}',
        to_jsonb((
            E'╭──────༺📖༻──────╮\nЗапись в бортовой журнал\n╰──────༺📖༻──────╯\n\nПоделитесь впечатлением о визите:\n• как встретили на борту\n• какое блюдо стало любимым\n• вернётесь ли вы в Baratie снова\n\nДля демо можно ответить так:\n5 / Sea King Steak / Вернусь ещё'
        )::text),
        true
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(
                        jsonb_set(
                            jsonb_set(
                                settings,
                                '{welcome_message}',
                                to_jsonb((
                                    E'Добро пожаловать в Baratie! 🏴‍☠️\n\nМы рады видеть вас на борту легендарного ресторана на воде.\n\nЗдесь вы можете:\n• копить дублоны за каждый визит\n• посмотреть авторское меню Sanji\n• забронировать столик\n• выбрать фирменный мерч\n\nДля начала пройдите короткую регистрацию.'
                                )::text),
                                true
                            ),
                            '{buttons,1,label}',
                            to_jsonb(('🪑 Забронировать')::text),
                            true
                        ),
                        '{buttons,1,value}',
                        to_jsonb((
                            E'🪑 Бронирование в Baratie\n\nДоступные зоны:\n• Палуба — до 4 гостей\n• Каюта — до 2 гостей\n• VIP-каюта — до 6 гостей\n\nПопулярные слоты сегодня:\n• 18:00\n• 19:30\n• 21:00\n\nДля демо скажите менеджеру: «Хочу столик в Baratie на 19:30».'
                        )::text),
                        true
                    ),
                    '{buttons,2,label}',
                    to_jsonb(('🛍 Мерч')::text),
                    true
                ),
                '{buttons,2,value}',
                to_jsonb((
                    E'🛍 Мерч Baratie за дублоны\n\n• Фартук шефа Sanji — 900 дублонов\n• Кружка All Blue — 450 дублонов\n• Футболка Baratie Crew — 1200 дублонов\n• Набор фирменных стикеров — 250 дублонов\n• Подарочный сертификат капитана — 1500 дублонов\n\nКаталог товаров уже загружен в marketplace.'
                )::text),
                true
            ),
            '{buttons,3,label}',
            to_jsonb(('⭐ Оставить отзыв')::text),
            true
        ),
        '{buttons,3,value}',
        to_jsonb((
            E'⭐ Спасибо, что были у нас на борту!\n\nНам важно узнать:\n• как вас встретили\n• какое блюдо понравилось больше всего\n• готовы ли вы рекомендовать Baratie друзьям\n\nДля демо ответьте одним сообщением в формате:\n5 / Sea King Steak / Вернусь ещё'
        )::text),
        true
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot';

-- +goose StatementEnd
