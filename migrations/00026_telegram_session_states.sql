-- +goose Up
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

-- +goose Down
ALTER TABLE telegram_sessions
    DROP CONSTRAINT IF EXISTS telegram_sessions_state_check;

ALTER TABLE telegram_sessions
    ADD CONSTRAINT telegram_sessions_state_check CHECK (
        state IN ('idle', 'await_task_title', 'await_goal_title')
    );
