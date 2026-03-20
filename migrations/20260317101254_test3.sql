-- +goose Up
ALTER TABLE processed_commands DROP CONSTRAINT processed_commands_payment_id_fkey;

-- +goose Down

