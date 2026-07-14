-- +goose Up
ALTER TABLE tasks
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'task',
    ADD COLUMN address TEXT,
    ADD COLUMN note_id UUID REFERENCES notes (id) ON DELETE SET NULL;

ALTER TABLE tasks
    ADD CONSTRAINT tasks_kind_check CHECK (kind IN ('task', 'reminder', 'meeting'));

CREATE INDEX tasks_due_date_idx ON tasks (user_id, due_date)
    WHERE deleted_at IS NULL AND due_date IS NOT NULL;

ALTER TABLE debts
    ADD COLUMN installment_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN installment_interval TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN next_payment_date DATE;

ALTER TABLE debts
    ADD CONSTRAINT debts_installment_cents_check CHECK (installment_cents >= 0),
    ADD CONSTRAINT debts_installment_interval_check CHECK (installment_interval IN ('none', 'weekly', 'monthly'));

-- +goose Down
ALTER TABLE debts DROP CONSTRAINT IF EXISTS debts_installment_interval_check;
ALTER TABLE debts DROP CONSTRAINT IF EXISTS debts_installment_cents_check;
ALTER TABLE debts DROP COLUMN IF EXISTS next_payment_date;
ALTER TABLE debts DROP COLUMN IF EXISTS installment_interval;
ALTER TABLE debts DROP COLUMN IF EXISTS installment_cents;

DROP INDEX IF EXISTS tasks_due_date_idx;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_kind_check;
ALTER TABLE tasks DROP COLUMN IF EXISTS note_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS address;
ALTER TABLE tasks DROP COLUMN IF EXISTS kind;
