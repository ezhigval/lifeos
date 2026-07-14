-- name: GetTelegramSession :one
SELECT user_id, chat_id, dashboard_message_id, state, state_payload, updated_at
FROM telegram_sessions
WHERE user_id = $1;

-- name: UpsertTelegramSession :exec
-- Always write dashboard_message_id (including NULL). Callers load-modify-save;
-- COALESCE made ClearDashboard(/start keyboard reset) a no-op.
INSERT INTO telegram_sessions (user_id, chat_id, dashboard_message_id, state, state_payload, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (user_id) DO UPDATE
SET chat_id = EXCLUDED.chat_id,
    dashboard_message_id = EXCLUDED.dashboard_message_id,
    state = EXCLUDED.state,
    state_payload = EXCLUDED.state_payload,
    updated_at = now();

-- name: SetTelegramDashboard :exec
UPDATE telegram_sessions
SET chat_id = $2,
    dashboard_message_id = $3,
    updated_at = now()
WHERE user_id = $1;

-- name: SetTelegramState :exec
UPDATE telegram_sessions
SET state = $2,
    state_payload = $3,
    updated_at = now()
WHERE user_id = $1;

-- name: ResetTelegramSession :exec
-- Clears only conversation UI state. Domain user data is untouched.
INSERT INTO telegram_sessions (user_id, chat_id, dashboard_message_id, state, state_payload, updated_at)
VALUES ($1, $2, NULL, 'idle', '{}', now())
ON CONFLICT (user_id) DO UPDATE
SET chat_id = EXCLUDED.chat_id,
    dashboard_message_id = NULL,
    state = 'idle',
    state_payload = '{}',
    updated_at = now();
