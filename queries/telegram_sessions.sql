-- name: GetTelegramSession :one
SELECT user_id, chat_id, dashboard_message_id, state, state_payload, updated_at
FROM telegram_sessions
WHERE user_id = $1;

-- name: UpsertTelegramSession :exec
INSERT INTO telegram_sessions (user_id, chat_id, dashboard_message_id, state, state_payload, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (user_id) DO UPDATE
SET chat_id = EXCLUDED.chat_id,
    dashboard_message_id = COALESCE(EXCLUDED.dashboard_message_id, telegram_sessions.dashboard_message_id),
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
