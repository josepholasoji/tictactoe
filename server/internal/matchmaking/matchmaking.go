// Package matchmaking pairs participants into a session via a direct
// invitation from one to the other.
package matchmaking

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/oyewunmi/tictactoe/internal/domain"
	"github.com/oyewunmi/tictactoe/internal/hub"
	"github.com/oyewunmi/tictactoe/internal/session"
	"github.com/oyewunmi/tictactoe/internal/store"
)

var (
	ErrUnknownParticipant  = errors.New("participant not found or offline")
	ErrCannotInviteSelf    = errors.New("cannot invite yourself")
	ErrAlreadyInSession    = errors.New("participant already has an active session")
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrNotInvitationTarget = errors.New("only the invited participant can respond")
)

type Service struct {
	store    store.Store
	hub      *hub.Hub
	sessions *session.Manager
	log      *slog.Logger
}

func NewService(st store.Store, h *hub.Hub, sessions *session.Manager, log *slog.Logger) *Service {
	return &Service{store: st, hub: h, sessions: sessions, log: log}
}

// CreateInvitation sends a direct challenge from one participant to another.
func (s *Service) CreateInvitation(fromID, toID string) (domain.Invitation, error) {
	if fromID == toID {
		return domain.Invitation{}, ErrCannotInviteSelf
	}
	from, ok := s.store.GetParticipant(fromID)
	if !ok {
		return domain.Invitation{}, ErrUnknownParticipant
	}
	to, ok := s.store.GetParticipant(toID)
	if !ok {
		return domain.Invitation{}, ErrUnknownParticipant
	}
	if _, active := s.sessions.Recover(fromID); active {
		return domain.Invitation{}, ErrAlreadyInSession
	}
	if _, active := s.sessions.Recover(toID); active {
		return domain.Invitation{}, ErrAlreadyInSession
	}

	inv := domain.Invitation{
		ID:        uuid.NewString(),
		From:      from,
		To:        to,
		Status:    domain.InvitationPending,
		CreatedAt: time.Now().UTC(),
	}
	s.store.SaveInvitation(inv)
	return inv, nil
}

// RespondInvitation accepts or declines a pending invitation.
func (s *Service) RespondInvitation(ctx context.Context, participantID, invitationID string, accept bool) (domain.Invitation, *domain.Session, error) {
	inv, ok := s.store.GetInvitation(invitationID)
	if !ok || inv.Status != domain.InvitationPending {
		return domain.Invitation{}, nil, ErrInvitationNotFound
	}
	if inv.To.ID != participantID {
		return domain.Invitation{}, nil, ErrNotInvitationTarget
	}

	s.store.DeleteInvitation(invitationID)

	if !accept {
		inv.Status = domain.InvitationDeclined
		return inv, nil, nil
	}

	inv.Status = domain.InvitationAccepted
	sess, err := s.sessions.Create(ctx, inv.From, inv.To)
	if err != nil {
		return inv, nil, err
	}
	return inv, sess, nil
}
