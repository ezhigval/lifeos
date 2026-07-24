-- +goose Up
CREATE TABLE user_memories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    source TEXT NOT NULL DEFAULT 'agent',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT user_memories_kind_check CHECK (kind IN ('preference', 'fact', 'alias', 'pattern')),
    CONSTRAINT user_memories_user_kind_key_uniq UNIQUE (user_id, kind, key)
);

CREATE INDEX user_memories_user_idx ON user_memories (user_id);

CREATE TABLE anon_learning_events (
    id UUID PRIMARY KEY,
    anon_subject TEXT NOT NULL,
    event_type TEXT NOT NULL,
    tool_or_intent TEXT,
    success BOOLEAN,
    ask_rounds INT NOT NULL DEFAULT 0,
    model TEXT,
    meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX anon_learning_events_created_idx ON anon_learning_events (created_at);
CREATE INDEX anon_learning_events_subject_idx ON anon_learning_events (anon_subject);

ALTER TABLE user_settings
    ADD COLUMN memory_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN learning_opt_in BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE telegram_sessions
    DROP CONSTRAINT IF EXISTS telegram_sessions_state_check;

ALTER TABLE telegram_sessions
    ADD CONSTRAINT telegram_sessions_state_check CHECK (
        state IN (
            'idle',
            'await_task_title',
            'await_task_projects',
            'await_project_title',
            'await_project_spheres',
            'await_sphere_name',
            'await_goal_title',
            'await_agent_turn'
        )
    );

-- +goose Down
ALTER TABLE telegram_sessions
    DROP CONSTRAINT IF EXISTS telegram_sessions_state_check;

ALTER TABLE telegram_sessions
    ADD CONSTRAINT telegram_sessions_state_check CHECK (
        state IN (
            'idle',
            'await_task_title',
            'await_task_projects',
            'await_project_title',
            'await_project_spheres',
            'await_sphere_name',
            'await_goal_title'
        )
    );

ALTER TABLE user_settings
    DROP COLUMN IF EXISTS memory_enabled,
    DROP COLUMN IF EXISTS learning_opt_in;

DROP INDEX IF EXISTS anon_learning_events_subject_idx;
DROP INDEX IF EXISTS anon_learning_events_created_idx;
DROP TABLE IF EXISTS anon_learning_events;

DROP INDEX IF EXISTS user_memories_user_idx;
DROP TABLE IF EXISTS user_memories;
