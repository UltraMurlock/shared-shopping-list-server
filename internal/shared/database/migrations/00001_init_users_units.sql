-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    login         TEXT        NOT NULL UNIQUE,
    display_name  TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE units (
    id   SMALLINT PRIMARY KEY,
    name TEXT     NOT NULL UNIQUE
);

-- +goose Down

DROP TABLE units;
DROP TABLE users;
-- DROP EXTENSION pgcrypto; -- не откатываем, может использоваться другими