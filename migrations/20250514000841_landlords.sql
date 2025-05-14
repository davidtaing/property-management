-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE landlords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    mobile TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    is_archived TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE landlords;
-- +goose StatementEnd
