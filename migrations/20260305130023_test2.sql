-- +goose Up
ALTER TABLE payment_events ALTER COLUMN payload TYPE bytea USING payload::text::bytea;

-- +goose Down
SELECT 'down SQL query';
