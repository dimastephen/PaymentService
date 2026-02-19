-- +goose Up
-- +goose StatementBegin
CREATE TABLE payments(
    id uuid NOT NULL PRIMARY KEY,
    status varchar(15) NOT NULL,
    amount bigint NOT NULL,
    currency varchar(3) NOT NULL,
    merchant text NOT NULL,
    idempotency_key text UNIQUE NOT NULL ,
    psp_transaction_id text UNIQUE,
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP,
    error_message text
);

CREATE TABLE payment_events(
    id bigserial PRIMARY KEY,
    payment_id uuid references payments(id) NOT NULL ,
    occurred_at timestamp default CURRENT_TIMESTAMP,
    event_type text NOT NULL,
    payload jsonb
);

CREATE TABLE processed_commands(
    idempotency_key text UNIQUE PRIMARY KEY NOT NULL,
    payment_id uuid references payments(id) NOT NULL,
    processed_at timestamp default CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_events_payment_id ON payment_events(payment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processed_commands;
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
