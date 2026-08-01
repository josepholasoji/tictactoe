// Package domain holds the core game entities, independent of both the
// wire protocol (internal/protocol) and the storage backend (internal/store).
package domain

import (
	"time"

	"github.com/oyewunmi/tictactoe/internal/game"
)

type Participant struct {
	ID       string
	Username string
}

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
	InvitationCanceled InvitationStatus = "canceled"
)

type Invitation struct {
	ID        string
	From      Participant
	To        Participant
	Status    InvitationStatus
	CreatedAt time.Time
}

type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
)

// Session is the live, in-play representation of a game. It is kept in the
// transient store (in-memory map, or Redis when horizontally scaled) for
// the duration of the game; a durable summary row is persisted to
// PostgreSQL on creation and again on completion.
type Session struct {
	ID           string
	ParticipantX Participant
	ParticipantO Participant
	Board        game.Board
	Turn         string // game.MarkX or game.MarkO
	Status       SessionStatus
	WinnerID     *string
	IsDraw       bool
	StartedAt    time.Time
	CompletedAt  *time.Time
}

// MarkFor returns the mark ("X"/"O") played by the given participant, or
// ok=false if they are not part of this session.
func (s *Session) MarkFor(participantID string) (string, bool) {
	switch participantID {
	case s.ParticipantX.ID:
		return game.MarkX, true
	case s.ParticipantO.ID:
		return game.MarkO, true
	default:
		return "", false
	}
}

func (s *Session) OpponentOf(participantID string) (Participant, bool) {
	switch participantID {
	case s.ParticipantX.ID:
		return s.ParticipantO, true
	case s.ParticipantO.ID:
		return s.ParticipantX, true
	default:
		return Participant{}, false
	}
}

type Move struct {
	ID            string
	SessionID     string
	ParticipantID string
	Position      int
	MoveNumber    int
	CreatedAt     time.Time
}
