-- +goose Up
CREATE TABLE outbox(
    id bigserial primary key,
    topic text not null,
    key bytea not null,
    value bytea not null,
    headers jsonb,
    sent boolean default false,
    created_at timestamp default CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_unsent ON outbox (sent) WHERE sent = false;

-- +goose Down
DROP TABLE outbox;
