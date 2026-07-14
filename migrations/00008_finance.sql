-- +goose Up
CREATE TABLE finance_categories (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(128) NOT NULL,
    kind       VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT finance_categories_kind_check CHECK (kind IN ('income', 'expense')),
    UNIQUE (user_id, name, kind)
);

CREATE TABLE finance_transactions (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id  UUID NOT NULL REFERENCES finance_categories(id),
    kind         VARCHAR(16) NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency     VARCHAR(8) NOT NULL DEFAULT 'RUB',
    description  TEXT,
    occurred_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT finance_transactions_kind_check CHECK (kind IN ('income', 'expense')),
    CONSTRAINT finance_transactions_amount_check CHECK (amount_cents > 0)
);

CREATE INDEX finance_transactions_user_occurred_idx
    ON finance_transactions (user_id, occurred_at DESC);

-- +goose Down
DROP TABLE IF EXISTS finance_transactions;
DROP TABLE IF EXISTS finance_categories;
