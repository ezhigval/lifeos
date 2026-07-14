-- +goose Up
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    timezone    VARCHAR(64) NOT NULL DEFAULT 'Europe/Moscow',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_settings (
    user_id            UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    morning_review_at  TIME NOT NULL DEFAULT '08:00',
    evening_review_at  TIME NOT NULL DEFAULT '21:00',
    quiet_hours_start  TIME,
    quiet_hours_end    TIME,
    language           VARCHAR(8) NOT NULL DEFAULT 'ru',
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS users;
