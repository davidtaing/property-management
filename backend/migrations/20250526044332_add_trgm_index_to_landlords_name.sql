-- +goose Up
-- +goose StatementBegin
CREATE INDEX landlords_name_trgm_idx ON landlords USING GIN ("name" gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS landlords_name_trgm_idx;
-- +goose StatementEnd
