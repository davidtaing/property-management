-- +goose Up
-- +goose StatementBegin
ALTER TABLE landlords ALTER COLUMN organisation_id SET NOT NULL;
ALTER TABLE tenants ALTER COLUMN organisation_id SET NOT NULL;
ALTER TABLE properties ALTER COLUMN organisation_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE landlords ALTER COLUMN organisation_id DROP NOT NULL;
ALTER TABLE tenants ALTER COLUMN organisation_id DROP NOT NULL;
ALTER TABLE properties ALTER COLUMN organisation_id DROP NOT NULL;
-- +goose StatementEnd
