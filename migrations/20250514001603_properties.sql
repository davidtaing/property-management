-- +goose Up
-- +goose StatementBegin
CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address_line_1 TEXT NOT NULL,
    address_line_2 TEXT,
    suburb TEXT NOT NULL,
    postcode TEXT NOT NULL,
    state TEXT NOT NULL,
    management_fee DOUBLE PRECISION NOT NULL,
    is_archived TIMESTAMP WITH TIME ZONE,
    landlord_id UUID NOT NULL REFERENCES landlords(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE properties;
-- +goose StatementEnd
