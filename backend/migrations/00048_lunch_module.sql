-- +goose Up
-- +goose StatementBegin

CREATE TABLE lunch_programs (
    id          SERIAL PRIMARY KEY,
    bot_id      INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL DEFAULT 'Бизнес-ланч',
    description TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (bot_id)
);

CREATE TABLE lunch_courses (
    id               SERIAL PRIMARY KEY,
    program_id       INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    code             VARCHAR(8) NOT NULL,
    title            VARCHAR(255) NOT NULL,
    menu_category_id INT NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    sort_order       INT NOT NULL DEFAULT 0,
    UNIQUE (program_id, code)
);

CREATE INDEX idx_lunch_courses_program ON lunch_courses(program_id);

CREATE TABLE lunch_course_items (
    course_id    INT NOT NULL REFERENCES lunch_courses(id) ON DELETE CASCADE,
    menu_item_id INT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    surcharge    NUMERIC(10,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (course_id, menu_item_id)
);

CREATE TABLE lunch_formats (
    id         SERIAL PRIMARY KEY,
    program_id INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    price_mode VARCHAR(32) NOT NULL DEFAULT 'fixed'
               CHECK (price_mode IN ('fixed', 'sum_of_items', 'base_plus_surcharge')),
    base_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_lunch_formats_program ON lunch_formats(program_id);

CREATE TABLE lunch_format_courses (
    format_id INT NOT NULL REFERENCES lunch_formats(id) ON DELETE CASCADE,
    course_id INT NOT NULL REFERENCES lunch_courses(id) ON DELETE CASCADE,
    position  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (format_id, course_id)
);

CREATE TABLE lunch_availability (
    id         SERIAL PRIMARY KEY,
    program_id INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    weekday    SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7), -- ISO: 1 = понедельник
    time_from  TIME NOT NULL,
    time_to    TIME NOT NULL,
    CHECK (time_from < time_to)
);

CREATE INDEX idx_lunch_availability_program ON lunch_availability(program_id);

CREATE TABLE lunch_orders (
    id            SERIAL PRIMARY KEY,
    bot_id        INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    bot_client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    format_id     INT REFERENCES lunch_formats(id) ON DELETE SET NULL,
    format_name   VARCHAR(255) NOT NULL DEFAULT '',
    table_num     VARCHAR(16) NOT NULL,
    total_price   NUMERIC(10,2) NOT NULL,
    status        VARCHAR(16) NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new', 'sent', 'cancelled')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lunch_orders_bot_status ON lunch_orders(bot_id, status);

CREATE TABLE lunch_order_items (
    id             SERIAL PRIMARY KEY,
    lunch_order_id INT NOT NULL REFERENCES lunch_orders(id) ON DELETE CASCADE,
    course_id      INT REFERENCES lunch_courses(id) ON DELETE SET NULL,
    course_title   VARCHAR(255) NOT NULL DEFAULT '',
    menu_item_id   INT REFERENCES menu_items(id) ON DELETE SET NULL,
    item_name      VARCHAR(255) NOT NULL DEFAULT '',
    price          NUMERIC(10,2) NOT NULL DEFAULT 0,
    surcharge      NUMERIC(10,2) NOT NULL DEFAULT 0
);

CREATE INDEX idx_lunch_order_items_order ON lunch_order_items(lunch_order_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lunch_order_items;
DROP TABLE IF EXISTS lunch_orders;
DROP TABLE IF EXISTS lunch_availability;
DROP TABLE IF EXISTS lunch_format_courses;
DROP TABLE IF EXISTS lunch_formats;
DROP TABLE IF EXISTS lunch_course_items;
DROP TABLE IF EXISTS lunch_courses;
DROP TABLE IF EXISTS lunch_programs;

-- +goose StatementEnd
