-- +goose Up
-- +goose StatementBegin
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    mobile TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    original_start_date DATE NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    termination_date DATE,
    termination_reason TEXT,
    vacate_date DATE,
    is_archived TIMESTAMP WITH TIME ZONE,
    property_id UUID NOT NULL REFERENCES properties(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tenants;
-- +goose StatementEnd
