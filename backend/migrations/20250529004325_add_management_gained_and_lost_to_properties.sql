-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties ADD COLUMN management_gained DATE;
ALTER TABLE properties ADD COLUMN management_lost DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties DROP COLUMN management_gained;
ALTER TABLE properties DROP COLUMN management_lost;
-- +goose StatementEnd
