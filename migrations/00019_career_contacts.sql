-- +goose Up
CREATE TABLE career_contacts (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    company    TEXT NOT NULL DEFAULT '',
    role       TEXT NOT NULL DEFAULT '',
    notes      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT career_contact_name_check CHECK (length(trim(name)) > 0)
);

CREATE INDEX career_contacts_user_created_idx ON career_contacts (user_id, created_at DESC);
CREATE INDEX career_contacts_user_name_idx ON career_contacts (user_id, lower(name));

-- +goose Down
DROP TABLE IF EXISTS career_contacts;
