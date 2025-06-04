-- +goose Up
-- +goose StatementBegin
ALTER TABLE landlords ADD COLUMN organisation_id TEXT;
CREATE INDEX idx_landlords_organisation_id ON landlords(organisation_id);

ALTER TABLE tenants ADD COLUMN organisation_id TEXT;
CREATE INDEX idx_tenants_organisation_id ON tenants(organisation_id);

ALTER TABLE properties ADD COLUMN organisation_id TEXT;
CREATE INDEX idx_properties_organisation_id ON properties(organisation_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_landlords_organisation_id;
ALTER TABLE landlords DROP COLUMN organisation_id;

DROP INDEX idx_tenants_organisation_id;
ALTER TABLE tenants DROP COLUMN organisation_id;

DROP INDEX idx_properties_organisation_id;
ALTER TABLE properties DROP COLUMN organisation_id;
-- +goose StatementEnd
