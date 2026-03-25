-- +goose Up

-- Products available for purchase with loyalty points
CREATE TABLE marketplace_products (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    image_url   TEXT NOT NULL DEFAULT '',
    price_points INT NOT NULL CHECK (price_points > 0),
    stock       INT, -- NULL = unlimited
    is_active   BOOLEAN NOT NULL DEFAULT true,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_marketplace_products_org ON marketplace_products(org_id);

-- Orders placed by clients
CREATE TABLE marketplace_orders (
    id           SERIAL PRIMARY KEY,
    org_id       INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id    INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled')),
    total_points INT NOT NULL CHECK (total_points >= 0),
    items        JSONB NOT NULL DEFAULT '[]',
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_marketplace_orders_org ON marketplace_orders(org_id);
CREATE INDEX idx_marketplace_orders_client ON marketplace_orders(client_id);

-- +goose Down
DROP TABLE IF EXISTS marketplace_orders;
DROP TABLE IF EXISTS marketplace_products;
