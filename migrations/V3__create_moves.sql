-- Moves: ordered move history per session, enabling replay and audit.
CREATE TABLE moves (
    id             UUID PRIMARY KEY,
    session_id     UUID NOT NULL REFERENCES sessions (id),
    participant_id UUID NOT NULL REFERENCES participants (id),
    position       SMALLINT NOT NULL CHECK (position BETWEEN 0 AND 8),
    move_number    SMALLINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, move_number)
);

CREATE INDEX idx_moves_session ON moves (session_id);
