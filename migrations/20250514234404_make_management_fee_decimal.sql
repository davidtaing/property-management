-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties
ALTER COLUMN management_fee TYPE DECIMAL(18, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties
ALTER COLUMN management_fee TYPE DOUBLE PRECISION;
-- +goose StatementEnd
