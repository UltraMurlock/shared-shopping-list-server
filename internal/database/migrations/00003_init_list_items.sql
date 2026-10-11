-- +goose Up

CREATE TABLE list_items (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id    UUID          NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    name       TEXT          NOT NULL,
    comment    TEXT          NOT NULL DEFAULT '',
    unit_id    SMALLINT      NOT NULL REFERENCES units(id),
    quantity   NUMERIC(12,3) NOT NULL CHECK (quantity > 0),
    is_done    BOOLEAN       NOT NULL DEFAULT false,
    version    BIGINT        NOT NULL DEFAULT 1,
    created_by UUID          NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_by UUID          NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_by UUID          REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX list_items_list_id_idx
    ON list_items(list_id) WHERE deleted_at IS NULL;

-- +goose Down

DROP TABLE list_items;