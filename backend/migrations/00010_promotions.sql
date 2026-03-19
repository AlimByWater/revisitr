-- +goose Up
CREATE TABLE promotions (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(30) NOT NULL CHECK (type IN ('discount','bonus','tag_update','campaign')),
    conditions  JSONB NOT NULL DEFAULT '{}',
    result      JSONB NOT NULL DEFAULT '{}',
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,
    usage_limit INT,
    combinable  BOOLEAN NOT NULL DEFAULT true,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_promotions_org_id ON promotions(org_id);

CREATE TABLE promo_codes (
    id               SERIAL PRIMARY KEY,
    org_id           INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    promotion_id     INT REFERENCES promotions(id) ON DELETE SET NULL,
    code             VARCHAR(50) NOT NULL,
    discount_percent NUMERIC(5,2),
    bonus_amount     INT,
    starts_at        TIMESTAMPTZ,
    ends_at          TIMESTAMPTZ,
    conditions       JSONB NOT NULL DEFAULT '{}',
    usage_count      INT NOT NULL DEFAULT 0,
    usage_limit      INT,
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, code)
);

CREATE TABLE promotion_usages (
    id            SERIAL PRIMARY KEY,
    promotion_id  INT NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    client_id     INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    promo_code_id INT REFERENCES promo_codes(id) ON DELETE SET NULL,
    used_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_promotion_usages_promotion_id ON promotion_usages(promotion_id);
CREATE INDEX idx_promotion_usages_client_id    ON promotion_usages(client_id);

-- +goose Down
DROP TABLE IF EXISTS promotion_usages;
DROP TABLE IF EXISTS promo_codes;
DROP TABLE IF EXISTS promotions;
