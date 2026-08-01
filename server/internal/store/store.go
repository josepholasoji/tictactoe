// Package store defines the transient-state interface used for presence,
// invitations, active game sessions, and heartbeats.
package store

import (
	"time"

	"github.com/oyewunmi/tictactoe/internal/domain"
)

type Store interface {
	// Presence
	SetOnline(p domain.Participant)
	SetOffline(participantID string)
	GetParticipant(participantID string) (domain.Participant, bool)
	ListOnline() []domain.Participant

	// Invitations
	SaveInvitation(inv domain.Invitation)
	GetInvitation(id string) (domain.Invitation, bool)
	ListInvitationsForParticipant(participantID string) []domain.Invitation
	DeleteInvitation(id string)

	// Active (in-play) sessions
	SaveSession(s *domain.Session)
	GetSession(id string) (*domain.Session, bool)
	GetActiveSessionForParticipant(participantID string) (*domain.Session, bool)
	DeleteSession(id string)

	// Heartbeats
	Heartbeat(participantID string, at time.Time)
	LastHeartbeat(participantID string) (time.Time, bool)
}
