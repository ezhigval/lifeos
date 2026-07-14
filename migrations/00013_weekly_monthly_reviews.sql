-- +goose Up
ALTER TABLE user_settings
    ADD COLUMN weekly_review_at  TIME NOT NULL DEFAULT '10:00',
    ADD COLUMN monthly_review_at TIME NOT NULL DEFAULT '10:00';

ALTER TABLE scheduled_jobs DROP CONSTRAINT scheduled_jobs_job_type_check;
ALTER TABLE scheduled_jobs ADD CONSTRAINT scheduled_jobs_job_type_check
    CHECK (job_type IN ('reminder', 'morning_review', 'evening_review', 'weekly_review', 'monthly_review'));

-- +goose Down
ALTER TABLE scheduled_jobs DROP CONSTRAINT scheduled_jobs_job_type_check;
ALTER TABLE scheduled_jobs ADD CONSTRAINT scheduled_jobs_job_type_check
    CHECK (job_type IN ('reminder', 'morning_review', 'evening_review'));

ALTER TABLE user_settings
    DROP COLUMN IF EXISTS weekly_review_at,
    DROP COLUMN IF EXISTS monthly_review_at;
