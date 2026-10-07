-- +goose Up
-- Публичный @ник, чтобы веб-вход нашёл чат. Коды — одноразовые, в базе только HMAC.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS telegram_username TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS users_telegram_username_uidx
    ON users (telegram_username)
    WHERE telegram_username IS NOT NULL;

CREATE TABLE IF NOT EXISTS login_codes (
    id           UUID PRIMARY KEY,
    telegram_id  BIGINT NOT NULL,
    username     TEXT NOT NULL,
    code_hash    TEXT NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    attempts     INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS login_codes_username_created_idx
    ON login_codes (username, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS login_codes;
DROP INDEX IF EXISTS users_telegram_username_uidx;
ALTER TABLE users DROP COLUMN IF EXISTS telegram_username;
