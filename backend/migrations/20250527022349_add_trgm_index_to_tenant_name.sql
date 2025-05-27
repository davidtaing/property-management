-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_tenant_name_trgm ON tenants USING GIN (name gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_tenant_name_trgm;
-- +goose StatementEnd
