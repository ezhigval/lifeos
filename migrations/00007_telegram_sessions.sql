-- +goose Up
CREATE TABLE telegram_sessions (
    user_id               UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    chat_id               BIGINT NOT NULL,
    dashboard_message_id  BIGINT,
    state                 VARCHAR(64) NOT NULL DEFAULT 'idle',
    state_payload         JSONB NOT NULL DEFAULT '{}',
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT telegram_sessions_state_check CHECK (
        state IN ('idle', 'await_task_title', 'await_goal_title')
    )
);

-- +goose Down
DROP TABLE IF EXISTS telegram_sessions;
