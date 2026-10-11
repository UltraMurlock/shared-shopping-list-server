-- +goose Up

CREATE TABLE list_events (
    id             BIGSERIAL   PRIMARY KEY,
    list_id        UUID        NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    actor_id       UUID        NOT NULL REFERENCES users(id),
    entity_type    TEXT        NOT NULL CHECK (entity_type IN ('list', 'item')),
    entity_id      UUID        NOT NULL,
    operation      TEXT        NOT NULL CHECK (operation IN ('created', 'updated', 'deleted')),
    entity_version BIGINT      NOT NULL,
    payload        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX list_events_list_id_id_idx ON list_events(list_id, id);

-- +goose Down

DROP TABLE list_events;