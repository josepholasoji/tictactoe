-- Sessions: one row per game, created when it starts and finalized on completion.
CREATE TABLE sessions (
    id                 UUID PRIMARY KEY,
    participant_x_id   UUID NOT NULL REFERENCES participants (id),
    participant_o_id    UUID NOT NULL REFERENCES participants (id),
    status              TEXT NOT NULL CHECK (status IN ('active', 'completed')),
    winner_id           UUID REFERENCES participants (id),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_sessions_participant_x ON sessions (participant_x_id);
CREATE INDEX idx_sessions_participant_o ON sessions (participant_o_id);

-- Invitations: a pending or resolved direct challenge between two participants.
CREATE TABLE invitations (
    id                   UUID PRIMARY KEY,
    from_participant_id  UUID NOT NULL REFERENCES participants (id),
    to_participant_id    UUID NOT NULL REFERENCES participants (id),
    status               TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'declined', 'canceled')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invitations_to_participant ON invitations (to_participant_id);
