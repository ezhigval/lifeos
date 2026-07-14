-- name: GetUserSettingsByUserID :one
SELECT user_id, morning_review_at, evening_review_at, weekly_review_at, monthly_review_at,
       quiet_hours_start, quiet_hours_end, language, updated_at
FROM user_settings
WHERE user_id = $1;

-- name: UpdateMorningReviewAt :exec
UPDATE user_settings
SET morning_review_at = $2, updated_at = now()
WHERE user_id = $1;

-- name: UpdateEveningReviewAt :exec
UPDATE user_settings
SET evening_review_at = $2, updated_at = now()
WHERE user_id = $1;

-- name: UpdateQuietHours :exec
UPDATE user_settings
SET quiet_hours_start = $2, quiet_hours_end = $3, updated_at = now()
WHERE user_id = $1;
