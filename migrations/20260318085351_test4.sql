-- +goose Up
ALTER TABLE payment_events DROP CONSTRAINT payment_events_payment_id_fkey;

-- +goose Down
SELECT 'down SQL query';
