-- +goose Up
-- +goose StatementBegin
ALTER TABLE tenants ADD COLUMN paid_from DATE;
ALTER TABLE tenants ADD COLUMN paid_to DATE;
ALTER TABLE tenants ADD COLUMN rental_amount DECIMAL(18, 2);
CREATE TYPE rental_frequency AS ENUM ('weekly', 'fortnightly', 'monthly');
ALTER TABLE tenants ADD COLUMN frequency rental_frequency;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tenants DROP COLUMN paid_from;
ALTER TABLE tenants DROP COLUMN paid_to;
ALTER TABLE tenants DROP COLUMN rental_amount;
ALTER TABLE tenants DROP COLUMN frequency;
DROP TYPE rental_frequency;
-- +goose StatementEnd
