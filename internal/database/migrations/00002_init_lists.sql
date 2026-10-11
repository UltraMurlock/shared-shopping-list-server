-- +goose Up

CREATE TABLE lists (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL REFERENCES users(id),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    version     BIGINT      NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX lists_owner_id_idx ON lists(owner_id) WHERE deleted_at IS NULL;

-- +goose Down

DROP TABLE lists;