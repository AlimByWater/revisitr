-- +goose Up
CREATE TABLE IF NOT EXISTS pos_locations (
    id SERIAL PRIMARY KEY,
    org_id INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    schedule JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_pos_locations_org_id ON pos_locations(org_id);

-- +goose Down
DROP TABLE IF EXISTS pos_locations;
