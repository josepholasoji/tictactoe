// Package db is the PostgreSQL persistence layer: the durable system of
// record for participants, completed sessions, and move history. Schema is
// owned by the Flyway migrations in /migrations.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oyewunmi/tictactoe/internal/domain"
	"github.com/oyewunmi/tictactoe/internal/metrics"
)

type Repository struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) UpsertParticipant(ctx context.Context, p domain.Participant) error {
	defer observe("upsert_participant", time.Now())
	_, err := r.pool.Exec(ctx, `
		INSERT INTO participants (id, username, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username
	`, p.ID, p.Username)
	return err
}

func (r *Repository) InsertSession(ctx context.Context, s *domain.Session) error {
	defer observe("insert_session", time.Now())
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, participant_x_id, participant_o_id, status, started_at)
		VALUES ($1, $2, $3, $4, $5)
	`, s.ID, s.ParticipantX.ID, s.ParticipantO.ID, string(s.Status), s.StartedAt)
	return err
}

func (r *Repository) InsertMove(ctx context.Context, mv *domain.Move) error {
	defer observe("insert_move", time.Now())
	_, err := r.pool.Exec(ctx, `
		INSERT INTO moves (id, session_id, participant_id, position, move_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, mv.ID, mv.SessionID, mv.ParticipantID, mv.Position, mv.MoveNumber, mv.CreatedAt)
	return err
}

func (r *Repository) CompleteSession(ctx context.Context, s *domain.Session) error {
	defer observe("complete_session", time.Now())
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET status = $2, winner_id = $3, completed_at = $4
		WHERE id = $1
	`, s.ID, string(s.Status), s.WinnerID, s.CompletedAt)
	return err
}

func observe(operation string, start time.Time) {
	metrics.DatabaseLatencySeconds.WithLabelValues(operation).Observe(time.Since(start).Seconds())
}
