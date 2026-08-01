// session owns the lifecycle of a game: creation, move processing,
// win/draw detection, and persistence of the durable summary + move history.
package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/oyewunmi/tictactoe/internal/domain"
	"github.com/oyewunmi/tictactoe/internal/game"
	"github.com/oyewunmi/tictactoe/internal/hub"
	"github.com/oyewunmi/tictactoe/internal/protocol"
	"github.com/oyewunmi/tictactoe/internal/store"
)

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrNotAParticipant  = errors.New("participant is not part of this session")
	ErrSessionCompleted = errors.New("session is already completed")
	ErrNotYourTurn      = errors.New("not your turn")
)

// Repository is the subset of durable persistence the session manager needs.
type Repository interface {
	InsertSession(ctx context.Context, s *domain.Session) error
	InsertMove(ctx context.Context, mv *domain.Move) error
	CompleteSession(ctx context.Context, s *domain.Session) error
}

type Manager struct {
	store store.Store
	hub   *hub.Hub
	repo  Repository
	log   *slog.Logger
}

func NewManager(st store.Store, h *hub.Hub, repo Repository, log *slog.Logger) *Manager {
	return &Manager{store: st, hub: h, repo: repo, log: log}
}

// Create starts a new game between two participants and persists the
// session summary row.
func (m *Manager) Create(ctx context.Context, x, o domain.Participant) (*domain.Session, error) {
	s := &domain.Session{
		ID:           uuid.NewString(),
		ParticipantX: x,
		ParticipantO: o,
		Board:        game.NewBoard(),
		Turn:         game.MarkX,
		Status:       domain.SessionActive,
		StartedAt:    time.Now().UTC(),
	}
	m.store.SaveSession(s)
	if err := m.repo.InsertSession(ctx, s); err != nil {
		m.log.Error("failed to persist new session", "error", err, "session_id", s.ID)
	}
	return s, nil
}

func (m *Manager) Recover(participantID string) (*domain.Session, bool) {
	return m.store.GetActiveSessionForParticipant(participantID)
}

func (m *Manager) ApplyMove(ctx context.Context, participantID, sessionID string, position int) (*domain.Session, bool, error) {
	s, ok := m.store.GetSession(sessionID)
	if !ok {
		return nil, false, ErrSessionNotFound
	}
	if s.Status != domain.SessionActive {
		return nil, false, ErrSessionCompleted
	}
	mark, ok := s.MarkFor(participantID)
	if !ok {
		return nil, false, ErrNotAParticipant
	}
	if s.Turn != mark {
		return nil, false, ErrNotYourTurn
	}

	newBoard, err := game.ApplyMove(s.Board, position, mark)
	if err != nil {
		return nil, false, err
	}
	s.Board = newBoard

	moveNumber := countMoves(newBoard)
	if err := m.repo.InsertMove(ctx, &domain.Move{
		ID:            uuid.NewString(),
		SessionID:     s.ID,
		ParticipantID: participantID,
		Position:      position,
		MoveNumber:    moveNumber,
		CreatedAt:     time.Now().UTC(),
	}); err != nil {
		m.log.Error("failed to persist move", "error", err, "session_id", s.ID)
	}

	completed := false
	if winnerMark, won := game.Winner(newBoard); won {
		winnerID := s.ParticipantX.ID
		if winnerMark == game.MarkO {
			winnerID = s.ParticipantO.ID
		}
		m.finish(ctx, s, &winnerID, false)
		completed = true
	} else if game.IsFull(newBoard) {
		m.finish(ctx, s, nil, true)
		completed = true
	} else {
		s.Turn = game.Opponent(mark)
		m.store.SaveSession(s)
	}

	return s, completed, nil
}

func (m *Manager) finish(ctx context.Context, s *domain.Session, winnerID *string, isDraw bool) {
	now := time.Now().UTC()
	s.Status = domain.SessionCompleted
	s.WinnerID = winnerID
	s.IsDraw = isDraw
	s.CompletedAt = &now
	m.store.SaveSession(s)

	if err := m.repo.CompleteSession(ctx, s); err != nil {
		m.log.Error("failed to persist session completion", "error", err, "session_id", s.ID)
	}
}

func (m *Manager) ToDTO(s *domain.Session) protocol.SessionDTO {
	dto := protocol.SessionDTO{
		ID:           s.ID,
		ParticipantX: protocol.ParticipantDTO{ID: s.ParticipantX.ID, Username: s.ParticipantX.Username},
		ParticipantO: protocol.ParticipantDTO{ID: s.ParticipantO.ID, Username: s.ParticipantO.Username},
		Board:        [9]string(s.Board),
		Turn:         s.Turn,
		Status:       string(s.Status),
		IsDraw:       s.IsDraw,
		StartedAt:    s.StartedAt,
		CompletedAt:  s.CompletedAt,
	}
	if s.WinnerID != nil {
		id := *s.WinnerID
		dto.WinnerID = &id
	}
	return dto
}

func countMoves(b game.Board) int {
	n := 0
	for _, cell := range b {
		if cell != game.Empty {
			n++
		}
	}
	return n
}
