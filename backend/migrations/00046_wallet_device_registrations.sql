-- +goose Up
-- +goose StatementBegin

CREATE TABLE wallet_device_registrations (
    id                  SERIAL PRIMARY KEY,
    org_id              INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    device_library_id   VARCHAR(128) NOT NULL,
    pass_type_id        VARCHAR(64) NOT NULL,
    serial_number       VARCHAR(64) NOT NULL REFERENCES wallet_passes(serial_number) ON DELETE CASCADE,
    push_token          TEXT,
    auth_token          VARCHAR(128) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(device_library_id, pass_type_id, serial_number)
);

CREATE INDEX idx_wallet_device_registrations_serial ON wallet_device_registrations(serial_number);
CREATE INDEX idx_wallet_device_registrations_device ON wallet_device_registrations(device_library_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS wallet_device_registrations;

-- +goose StatementEnd
