-- +goose Up
-- +goose StatementBegin

UPDATE bots
SET settings = jsonb_set(
        settings,
        '{buttons,0,value}',
        to_jsonb((
            E'╔═*.·:·. ﹏﹏𓂁🦈𓂁﹏﹏.·:·.*═╗\n║\n║ ⟼ Sea King Steak\n║ ⟼ All Blue Sashimi\n║ ⟼ Diable Jambe Pasta\n║ ⟼ Sniper Prawns\n║ ⟼ Grand Line Bruschetta\n║ ⟼ East Blue Oyster Trio\n║ ⟼ Mera Mera Crème Brûlée\n║ ⟼ Baratie Citrus Spritz\n║\n╚═*.·:·. ﹏﹏𓂁🪸𓂁﹏﹏.·:·.*═╝'
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
        settings,
        '{buttons,0,value}',
        to_jsonb((
            E'🍽 Меню Baratie\n\nЗакуски\n• Sniper Prawns — 490 ₽\n• Grand Line Bruschetta — 420 ₽\n• East Blue Oyster Trio — 560 ₽\n\nОсновные блюда\n• Sea King Steak — 1390 ₽\n• All Blue Sashimi — 1250 ₽\n• Diable Jambe Pasta — 990 ₽\n\nДесерты\n• Mera Mera Crème Brûlée — 450 ₽\n• Going Merry Cheesecake — 470 ₽\n\nНапитки\n• Cola Float Franky — 320 ₽\n• Baratie Citrus Spritz — 390 ₽\n\nПолное меню уже загружено в систему Revisitr для демо.'
        )::text),
        true
    ),
    updated_at = NOW()
WHERE username = 'baratie_demo_bot';

-- +goose StatementEnd
