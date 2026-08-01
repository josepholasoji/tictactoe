// Package protocol defines the JSON message contracts
// for the system communication over WebSocket.
package protocol

import (
	"encoding/json"
	"time"
)

// Message types sent from the client to the server.
const (
	TypeHello         = "hello"
	TypeInviteCreate  = "invite_create"
	TypeInviteRespond = "invite_respond"
	TypeMove          = "move"
	TypeHeartbeat     = "heartbeat"
)

// Message types sent from the server to the client.
const (
	TypeWelcome        = "welcome"
	TypeLobbyState     = "lobby_state"
	TypePresenceUpdate = "presence_update"
	TypeInviteReceived = "invite_received"
	TypeInviteSent     = "invite_sent"
	TypeInviteCanceled = "invite_canceled"
	TypeGameStarted    = "game_started"
	TypeMoveApplied    = "move_applied"
	TypeGameCompleted  = "game_completed"
	TypeHeartbeatAck   = "heartbeat_ack"
	TypeError          = "error"
)

// Error codes returned in ErrorPayload.Code.
const (
	ErrCodeInvalidMessage     = "invalid_message"
	ErrCodeNotYourTurn        = "not_your_turn"
	ErrCodeInvalidMove        = "invalid_move"
	ErrCodeSessionNotFound    = "session_not_found"
	ErrCodeParticipantOffline = "participant_offline"
	ErrCodeAlreadyInSession   = "already_in_session"
	ErrCodeInvitationNotFound = "invitation_not_found"
	ErrCodeInternal           = "internal_error"
)

// Envelope is the outer shape of every WebSocket frame in both directions.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func Encode(msgType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Payload: raw})
}

// In a real system, this implementation would have a separate dto folder with
// each of these types in its own file, and a mapping layer to convert between
// the wire protocol (internal/protocol) and the storage backend (internal/store).
// --- DTOs -------------------------------------------------------------

type ParticipantDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type InvitationDTO struct {
	ID              string         `json:"id"`
	FromParticipant ParticipantDTO `json:"fromParticipant"`
	ToParticipant   ParticipantDTO `json:"toParticipant"`
	Status          string         `json:"status"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type SessionDTO struct {
	ID           string         `json:"id"`
	ParticipantX ParticipantDTO `json:"participantX"`
	ParticipantO ParticipantDTO `json:"participantO"`
	Board        [9]string      `json:"board"`
	Turn         string         `json:"turn"`
	Status       string         `json:"status"`
	WinnerID     *string        `json:"winnerId,omitempty"`
	IsDraw       bool           `json:"isDraw"`
	StartedAt    time.Time      `json:"startedAt"`
	CompletedAt  *time.Time     `json:"completedAt,omitempty"`
}

// --- Client -> Server payloads -----------------------------------------

type HelloPayload struct {
	ParticipantID string `json:"participantId,omitempty"`
	Username      string `json:"username"`
}

type InviteCreatePayload struct {
	ToParticipantID string `json:"toParticipantId"`
}

type InviteRespondPayload struct {
	InvitationID string `json:"invitationId"`
	Accept       bool   `json:"accept"`
}

type MovePayload struct {
	SessionID string `json:"sessionId"`
	Position  int    `json:"position"`
}

// --- Server -> Client payloads -------------------------------------------

type WelcomePayload struct {
	ParticipantID string      `json:"participantId"`
	Username      string      `json:"username"`
	Reconnected   bool        `json:"reconnected"`
	ActiveSession *SessionDTO `json:"activeSession,omitempty"`
}

type LobbyStatePayload struct {
	Online      []ParticipantDTO `json:"online"`
	Invitations []InvitationDTO  `json:"invitations"`
}

type PresenceUpdatePayload struct {
	Participant ParticipantDTO `json:"participant"`
	Online      bool           `json:"online"`
}

type InviteReceivedPayload struct {
	Invitation InvitationDTO `json:"invitation"`
}

type InviteSentPayload struct {
	Invitation InvitationDTO `json:"invitation"`
}

type InviteCanceledPayload struct {
	InvitationID string `json:"invitationId"`
}

type GameStartedPayload struct {
	Session SessionDTO `json:"session"`
}

type MoveAppliedPayload struct {
	Session SessionDTO `json:"session"`
}

type GameCompletedPayload struct {
	Session SessionDTO `json:"session"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
