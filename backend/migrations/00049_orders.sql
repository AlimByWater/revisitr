-- +goose Up
-- +goose StatementBegin

-- Заказы гостей через бота (source — какой флоу создал заказ).
-- Не путать с external_orders (импорт из POS).
CREATE TABLE orders (
    id            SERIAL PRIMARY KEY,
    bot_id        INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    bot_client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    source        VARCHAR(16) NOT NULL CHECK (source IN ('lunch', 'menu')),
    format_id     INT REFERENCES lunch_formats(id) ON DELETE SET NULL, -- lunch only
    format_name   VARCHAR(255) NOT NULL DEFAULT '',                    -- lunch only
    table_num     VARCHAR(16) NOT NULL,
    total_price   NUMERIC(10,2) NOT NULL,
    status        VARCHAR(16) NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new', 'sent', 'cancelled')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_bot_status ON orders(bot_id, status);

CREATE TABLE order_items (
    id           SERIAL PRIMARY KEY,
    order_id     INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    course_id    INT REFERENCES lunch_courses(id) ON DELETE SET NULL,  -- lunch only
    course_title VARCHAR(255) NOT NULL DEFAULT '',                     -- lunch only
    menu_item_id INT REFERENCES menu_items(id) ON DELETE SET NULL,
    item_name    VARCHAR(255) NOT NULL DEFAULT '',
    price        NUMERIC(10,2) NOT NULL DEFAULT 0,
    surcharge    NUMERIC(10,2) NOT NULL DEFAULT 0                      -- lunch only
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

-- +goose StatementEnd
