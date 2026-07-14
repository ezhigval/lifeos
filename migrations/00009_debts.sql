-- +goose Up
CREATE TABLE debts (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creditor     VARCHAR(256) NOT NULL,
    amount_cents BIGINT NOT NULL,
    paid_cents   BIGINT NOT NULL DEFAULT 0,
    due_date     DATE,
    status       VARCHAR(16) NOT NULL DEFAULT 'open',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT debts_amount_check CHECK (amount_cents > 0),
    CONSTRAINT debts_paid_check CHECK (paid_cents >= 0 AND paid_cents <= amount_cents),
    CONSTRAINT debts_status_check CHECK (status IN ('open', 'paid', 'cancelled'))
);

CREATE INDEX debts_user_status_idx ON debts (user_id, status);

-- +goose Down
DROP TABLE IF EXISTS debts;
