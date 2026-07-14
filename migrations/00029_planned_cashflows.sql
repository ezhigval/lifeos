-- +goose Up
CREATE TABLE planned_cashflows (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    interval TEXT NOT NULL DEFAULT 'monthly',
    next_date DATE NOT NULL,
    debt_id UUID REFERENCES debts (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT planned_cashflows_kind_check CHECK (kind IN ('income', 'expense')),
    CONSTRAINT planned_cashflows_amount_check CHECK (amount_cents > 0),
    CONSTRAINT planned_cashflows_interval_check CHECK (interval IN ('once', 'weekly', 'monthly')),
    CONSTRAINT planned_cashflows_title_check CHECK (char_length(trim(title)) > 0)
);

CREATE INDEX planned_cashflows_user_next_idx ON planned_cashflows (user_id, next_date);

-- +goose Down
DROP INDEX IF EXISTS planned_cashflows_user_next_idx;
DROP TABLE IF EXISTS planned_cashflows;
