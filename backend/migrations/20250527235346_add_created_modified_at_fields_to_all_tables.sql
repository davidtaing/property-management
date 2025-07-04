-- +goose Up
-- +goose StatementBegin
ALTER TABLE landlords ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE landlords ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE properties ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE properties ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE tenants ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE tenants ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE landlords DROP COLUMN created_at;
ALTER TABLE landlords DROP COLUMN updated_at;

ALTER TABLE properties DROP COLUMN created_at;
ALTER TABLE properties DROP COLUMN updated_at;

ALTER TABLE tenants DROP COLUMN created_at;
ALTER TABLE tenants DROP COLUMN updated_at;
-- +goose StatementEnd
