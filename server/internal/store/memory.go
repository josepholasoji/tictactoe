package store

import (
	"sync"
	"time"

	"github.com/oyewunmi/tictactoe/internal/domain"
)

type MemoryStore struct {
	mu sync.RWMutex

	online      map[string]domain.Participant
	invitations map[string]domain.Invitation
	sessions    map[string]*domain.Session
	// sessionByParticipant indexes the one active session a participant is
	// currently part of, so reconnect/session-recovery lookups are O(1).
	sessionByParticipant map[string]string
	heartbeats           map[string]time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		online:               make(map[string]domain.Participant),
		invitations:          make(map[string]domain.Invitation),
		sessions:             make(map[string]*domain.Session),
		sessionByParticipant: make(map[string]string),
		heartbeats:           make(map[string]time.Time),
	}
}

// --- Presence -----------------------------------------------------------

func (m *MemoryStore) SetOnline(p domain.Participant) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.online[p.ID] = p
}

func (m *MemoryStore) SetOffline(participantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.online, participantID)
}

func (m *MemoryStore) GetParticipant(participantID string) (domain.Participant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.online[participantID]
	return p, ok
}

func (m *MemoryStore) ListOnline() []domain.Participant {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Participant, 0, len(m.online))
	for _, p := range m.online {
		out = append(out, p)
	}
	return out
}

// --- Invitations ----------------------------------------------------------

func (m *MemoryStore) SaveInvitation(inv domain.Invitation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invitations[inv.ID] = inv
}

func (m *MemoryStore) GetInvitation(id string) (domain.Invitation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inv, ok := m.invitations[id]
	return inv, ok
}

func (m *MemoryStore) ListInvitationsForParticipant(participantID string) []domain.Invitation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.Invitation
	for _, inv := range m.invitations {
		if inv.Status == domain.InvitationPending && (inv.From.ID == participantID || inv.To.ID == participantID) {
			out = append(out, inv)
		}
	}
	return out
}

func (m *MemoryStore) DeleteInvitation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.invitations, id)
}

// --- Active sessions --------------------------------------------------------

func (m *MemoryStore) SaveSession(s *domain.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = s
	if s.Status == domain.SessionActive {
		m.sessionByParticipant[s.ParticipantX.ID] = s.ID
		m.sessionByParticipant[s.ParticipantO.ID] = s.ID
	} else {
		delete(m.sessionByParticipant, s.ParticipantX.ID)
		delete(m.sessionByParticipant, s.ParticipantO.ID)
	}
}

func (m *MemoryStore) GetSession(id string) (*domain.Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *MemoryStore) GetActiveSessionForParticipant(participantID string) (*domain.Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.sessionByParticipant[participantID]
	if !ok {
		return nil, false
	}
	s, ok := m.sessions[id]
	return s, ok
}

func (m *MemoryStore) DeleteSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		delete(m.sessionByParticipant, s.ParticipantX.ID)
		delete(m.sessionByParticipant, s.ParticipantO.ID)
	}
	delete(m.sessions, id)
}

// --- Heartbeats -------------------------------------------------------------

func (m *MemoryStore) Heartbeat(participantID string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.heartbeats[participantID] = at
}

func (m *MemoryStore) LastHeartbeat(participantID string) (time.Time, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.heartbeats[participantID]
	return t, ok
}
