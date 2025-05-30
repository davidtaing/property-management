-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties
  RENAME COLUMN address_line_1 TO street_number;

ALTER TABLE properties
  RENAME COLUMN address_line_2 TO street_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties
  RENAME COLUMN street_number TO address_line_1;

ALTER TABLE properties
  RENAME COLUMN street_name TO address_line_2;
-- +goose StatementEnd
