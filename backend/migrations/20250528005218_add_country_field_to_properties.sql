-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties ADD COLUMN country TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties DROP COLUMN country;
-- +goose StatementEnd
