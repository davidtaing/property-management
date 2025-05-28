-- +goose Up
-- +goose StatementBegin
ALTER TABLE landlords ADD COLUMN address_line_1 TEXT NOT NULL;
ALTER TABLE landlords ADD COLUMN address_line_2 TEXT;
ALTER TABLE landlords ADD COLUMN suburb TEXT NOT NULL;
ALTER TABLE landlords ADD COLUMN postcode TEXT NOT NULL;
ALTER TABLE landlords ADD COLUMN state TEXT NOT NULL;
ALTER TABLE landlords ADD COLUMN country TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE landlords DROP COLUMN address_line_1;
ALTER TABLE landlords DROP COLUMN address_line_2;
ALTER TABLE landlords DROP COLUMN suburb;
ALTER TABLE landlords DROP COLUMN postcode;
ALTER TABLE landlords DROP COLUMN state;
ALTER TABLE landlords DROP COLUMN country;
-- +goose StatementEnd
