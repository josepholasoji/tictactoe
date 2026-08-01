-- Participants: one row per registered player identity.
CREATE TABLE participants (
    id         UUID PRIMARY KEY,
    username   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
