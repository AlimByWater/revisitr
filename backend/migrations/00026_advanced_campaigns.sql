-- +goose Up

-- Campaign A/B testing: variants for a campaign
CREATE TABLE campaign_variants (
    id            SERIAL PRIMARY KEY,
    campaign_id   INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name          VARCHAR(100) NOT NULL,       -- e.g. "Variant A", "Variant B"
    audience_pct  INT NOT NULL DEFAULT 50,      -- percentage of audience (0-100)
    message       TEXT NOT NULL,
    media_url     VARCHAR(500),
    buttons       JSONB NOT NULL DEFAULT '[]',
    stats         JSONB NOT NULL DEFAULT '{"total":0,"sent":0,"failed":0}',
    is_winner     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaign_variants_campaign ON campaign_variants(campaign_id);

-- Add variant_id to campaign_messages for tracking which variant was sent
ALTER TABLE campaign_messages ADD COLUMN variant_id INT REFERENCES campaign_variants(id);

-- Campaign templates: reusable campaign blueprints
CREATE TABLE campaign_templates (
    id              SERIAL PRIMARY KEY,
    org_id          INT REFERENCES organizations(id) ON DELETE CASCADE,  -- NULL = system template
    name            VARCHAR(200) NOT NULL,
    category        VARCHAR(50) NOT NULL DEFAULT 'general',  -- general, welcome, promo, holiday, reactivation
    description     TEXT,
    message         TEXT NOT NULL,
    media_url       VARCHAR(500),
    buttons         JSONB NOT NULL DEFAULT '[]',
    audience_filter JSONB NOT NULL DEFAULT '{}',
    tracking_mode   VARCHAR(20) NOT NULL DEFAULT 'none',
    is_system       BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaign_templates_org ON campaign_templates(org_id);
CREATE INDEX idx_campaign_templates_category ON campaign_templates(category);

-- Seed system templates
INSERT INTO campaign_templates (org_id, name, category, description, message, buttons, is_system) VALUES
(NULL, 'Приветственная рассылка', 'welcome',
 'Рассылка для новых клиентов с приветствием и бонусом',
 'Привет, {name}! 🎉 Добро пожаловать! Мы рады видеть вас среди наших гостей. В подарок — бонусные баллы на первый заказ!',
 '[]', true),

(NULL, 'Промо-акция', 'promo',
 'Шаблон промо-рассылки со скидкой',
 '🔥 Специальное предложение, {name}! Только сегодня — скидка на все меню. Не упустите шанс!',
 '[{"text":"Подробнее","url":""}]', true),

(NULL, 'С праздником!', 'holiday',
 'Поздравительная рассылка к празднику',
 '🎄 {name}, поздравляем с праздником! Желаем тепла и уюта. Ждём вас в гости — мы приготовили кое-что особенное!',
 '[]', true),

(NULL, 'Реактивация', 'reactivation',
 'Возврат неактивных клиентов',
 'Мы скучаем, {name}! 💫 Давно не виделись. Специально для вас — бонусные баллы за визит. Ждём!',
 '[]', true);

-- +goose Down

DELETE FROM campaign_templates WHERE is_system = true;
DROP TABLE IF EXISTS campaign_templates;
ALTER TABLE campaign_messages DROP COLUMN IF EXISTS variant_id;
DROP TABLE IF EXISTS campaign_variants;
