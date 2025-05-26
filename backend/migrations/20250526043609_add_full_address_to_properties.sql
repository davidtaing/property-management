-- +goose Up
-- +goose StatementBegin
ALTER TABLE properties
ADD COLUMN full_address TEXT GENERATED ALWAYS AS (
  COALESCE(address_line_1, '') || ' ' ||
  COALESCE(address_line_2, '') || ' ' ||
  suburb || ' ' ||
  postcode || ' ' ||
  state
) STORED;

CREATE INDEX properties_full_address_trgm_idx ON properties USING GIN (full_address gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE properties
DROP INDEX IF EXISTS properties_full_address_trgm_idx;

ALTER TABLE properties
DROP COLUMN full_address;
-- +goose StatementEnd
