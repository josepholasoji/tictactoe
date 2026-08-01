// ws implements the WebSocket gateway: connection upgrade,
// participant registration/reconnection, and dispatch of incoming messages
// to the matchmaking and session services.
package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oyewunmi/tictactoe/internal/domain"
	"github.com/oyewunmi/tictactoe/internal/game"
	"github.com/oyewunmi/tictactoe/internal/hub"
	"github.com/oyewunmi/tictactoe/internal/matchmaking"
	"github.com/oyewunmi/tictactoe/internal/protocol"
	"github.com/oyewunmi/tictactoe/internal/session"
	"github.com/oyewunmi/tictactoe/internal/store"
	"nhooyr.io/websocket"
)

const (
	helloTimeout     = 10 * time.Second
	heartbeatTimeout = 30 * time.Second
)

// Repository is the narrow persistence interface the gateway itself needs.
type Repository interface {
	UpsertParticipant(ctx context.Context, p domain.Participant) error
}

type Gateway struct {
	hub         *hub.Hub
	store       store.Store
	matchmaking *matchmaking.Service
	sessions    *session.Manager
	repo        Repository
	log         *slog.Logger
}

func NewGateway(h *hub.Hub, st store.Store, mm *matchmaking.Service, sess *session.Manager, repo Repository, log *slog.Logger) *Gateway {
	return &Gateway{hub: h, store: st, matchmaking: mm, sessions: sess, repo: repo, log: log}
}

func (g *Gateway) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// The Electron client's renderer runs on a file:// or app:// origin,
		// so the browser's default same-origin check has to be relaxed here.
		InsecureSkipVerify: true,
	})
	if err != nil {
		g.log.Warn("websocket accept failed", "error", err)
		return
	}

	participant, reconnected, ok := g.performHandshake(r.Context(), conn)
	if !ok {
		return
	}

	client := g.hub.Register(participant.ID, participant.Username, conn)
	pumpCtx, cancelPump := context.WithCancel(context.Background())
	go client.WritePump(pumpCtx)
	defer cancelPump()

	g.log.Info("participant connected", "participant_id", participant.ID, "username", participant.Username, "reconnected", reconnected)

	g.sendWelcome(client, participant, reconnected)
	g.send(client, protocol.TypeLobbyState, g.lobbyState(participant.ID))
	g.broadcastPresence(participant, true)

	g.readLoop(r.Context(), client, participant)

	g.handleDisconnect(client, participant)
}

// performHandshake reads the mandatory first "hello" frame and registers
// (or resumes) the participant's identity.
func (g *Gateway) performHandshake(ctx context.Context, conn *websocket.Conn) (domain.Participant, bool, bool) {
	helloCtx, cancel := context.WithTimeout(ctx, helloTimeout)
	defer cancel()

	_, data, err := conn.Read(helloCtx)
	if err != nil {
		conn.Close(websocket.StatusPolicyViolation, "expected hello message")
		return domain.Participant{}, false, false
	}

	var env protocol.Envelope
	if err := json.Unmarshal(data, &env); err != nil || env.Type != protocol.TypeHello {
		conn.Close(websocket.StatusPolicyViolation, "first message must be hello")
		return domain.Participant{}, false, false
	}

	var hello protocol.HelloPayload
	_ = json.Unmarshal(env.Payload, &hello)

	username := strings.TrimSpace(hello.Username)
	if username == "" {
		username = "Player"
	}

	participantID := strings.TrimSpace(hello.ParticipantID)
	reconnected := participantID != ""
	if participantID == "" {
		participantID = uuid.NewString()
	}

	participant := domain.Participant{ID: participantID, Username: username}
	g.store.SetOnline(participant)
	g.store.Heartbeat(participantID, time.Now())
	if err := g.repo.UpsertParticipant(ctx, participant); err != nil {
		g.log.Error("failed to persist participant", "error", err, "participant_id", participantID)
	}

	return participant, reconnected, true
}

func (g *Gateway) sendWelcome(c *hub.Client, p domain.Participant, reconnected bool) {
	var activeDTO *protocol.SessionDTO
	if s, ok := g.sessions.Recover(p.ID); ok {
		dto := g.sessions.ToDTO(s)
		activeDTO = &dto
	}
	g.send(c, protocol.TypeWelcome, protocol.WelcomePayload{
		ParticipantID: p.ID,
		Username:      p.Username,
		Reconnected:   reconnected,
		ActiveSession: activeDTO,
	})
}

func (g *Gateway) readLoop(ctx context.Context, c *hub.Client, p domain.Participant) {
	for {
		data, err := c.Read(ctx)
		if err != nil {
			return
		}

		var env protocol.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			g.sendError(c, protocol.ErrCodeInvalidMessage, "malformed message envelope")
			continue
		}

		g.dispatch(ctx, c, p, env)
	}
}

func (g *Gateway) dispatch(ctx context.Context, c *hub.Client, p domain.Participant, env protocol.Envelope) {
	switch env.Type {
	case protocol.TypeHeartbeat:
		g.store.Heartbeat(p.ID, time.Now())
		g.send(c, protocol.TypeHeartbeatAck, struct{}{})

	case protocol.TypeInviteCreate:
		g.handleInviteCreate(c, p, env.Payload)

	case protocol.TypeInviteRespond:
		g.handleInviteRespond(ctx, c, p, env.Payload)

	case protocol.TypeMove:
		g.handleMove(ctx, c, p, env.Payload)

	default:
		g.sendError(c, protocol.ErrCodeInvalidMessage, "unknown message type: "+env.Type)
	}
}

func (g *Gateway) handleInviteCreate(c *hub.Client, p domain.Participant, raw json.RawMessage) {
	var payload protocol.InviteCreatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		g.sendError(c, protocol.ErrCodeInvalidMessage, "malformed invite_create payload")
		return
	}
	inv, err := g.matchmaking.CreateInvitation(p.ID, payload.ToParticipantID)
	if err != nil {
		g.sendError(c, matchmakingErrorCode(err), err.Error())
		return
	}
	dto := invitationDTO(inv)
	g.send(c, protocol.TypeInviteSent, protocol.InviteSentPayload{Invitation: dto})
	g.hub.Send(inv.To.ID, mustEncode(protocol.TypeInviteReceived, protocol.InviteReceivedPayload{Invitation: dto}))
}

func (g *Gateway) handleInviteRespond(ctx context.Context, c *hub.Client, p domain.Participant, raw json.RawMessage) {
	var payload protocol.InviteRespondPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		g.sendError(c, protocol.ErrCodeInvalidMessage, "malformed invite_respond payload")
		return
	}
	inv, sess, err := g.matchmaking.RespondInvitation(ctx, p.ID, payload.InvitationID, payload.Accept)
	if err != nil {
		g.sendError(c, matchmakingErrorCode(err), err.Error())
		return
	}
	if sess != nil {
		dto := g.sessions.ToDTO(sess)
		g.broadcastToParticipants(sess, protocol.TypeGameStarted, protocol.GameStartedPayload{Session: dto})
		g.announceAvailability(sess, false)
		return
	}
	g.hub.Send(inv.From.ID, mustEncode(protocol.TypeInviteCanceled, protocol.InviteCanceledPayload{InvitationID: inv.ID}))
}

func (g *Gateway) handleMove(ctx context.Context, c *hub.Client, p domain.Participant, raw json.RawMessage) {
	var payload protocol.MovePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		g.sendError(c, protocol.ErrCodeInvalidMessage, "malformed move payload")
		return
	}
	sess, completed, err := g.sessions.ApplyMove(ctx, p.ID, payload.SessionID, payload.Position)
	if err != nil {
		g.sendError(c, sessionErrorCode(err), err.Error())
		return
	}
	dto := g.sessions.ToDTO(sess)
	if completed {
		g.broadcastToParticipants(sess, protocol.TypeGameCompleted, protocol.GameCompletedPayload{Session: dto})
		g.announceAvailability(sess, true)
	} else {
		g.broadcastToParticipants(sess, protocol.TypeMoveApplied, protocol.MoveAppliedPayload{Session: dto})
	}
}

func (g *Gateway) handleDisconnect(c *hub.Client, p domain.Participant) {
	if !g.hub.Unregister(p.ID, c) {
		return
	}
	g.store.SetOffline(p.ID)
	g.broadcastPresence(p, false)
	g.log.Info("participant disconnected", "participant_id", p.ID, "username", p.Username)
}

func (g *Gateway) SweepStaleConnections() {
	now := time.Now()
	for _, p := range g.store.ListOnline() {
		last, ok := g.store.LastHeartbeat(p.ID)
		if !ok || now.Sub(last) <= heartbeatTimeout {
			continue
		}
		if c, ok := g.hub.Get(p.ID); ok {
			g.log.Info("closing stale connection", "participant_id", p.ID)
			c.Close()
		}
	}
}

//helpers

func (g *Gateway) send(c *hub.Client, msgType string, payload any) {
	c.Enqueue(mustEncode(msgType, payload))
}

func (g *Gateway) sendError(c *hub.Client, code, message string) {
	g.send(c, protocol.TypeError, protocol.ErrorPayload{Code: code, Message: message})
}

func (g *Gateway) broadcastToParticipants(s *domain.Session, msgType string, payload any) {
	msg := mustEncode(msgType, payload)
	g.hub.Send(s.ParticipantX.ID, msg)
	g.hub.Send(s.ParticipantO.ID, msg)
}

func (g *Gateway) broadcastPresence(p domain.Participant, online bool) {
	msg := mustEncode(protocol.TypePresenceUpdate, protocol.PresenceUpdatePayload{
		Participant: protocol.ParticipantDTO{ID: p.ID, Username: p.Username},
		Online:      online,
	})
	g.hub.Broadcast(msg)
}

func (g *Gateway) announceAvailability(s *domain.Session, available bool) {
	for _, participant := range [2]domain.Participant{s.ParticipantX, s.ParticipantO} {
		if available && !g.hub.IsConnected(participant.ID) {
			continue
		}
		g.broadcastPresence(participant, available)
	}
}

func (g *Gateway) lobbyState(participantID string) protocol.LobbyStatePayload {
	online := g.store.ListOnline()
	onlineDTOs := make([]protocol.ParticipantDTO, 0, len(online))
	for _, p := range online {
		if p.ID == participantID {
			continue
		}
		if _, inGame := g.sessions.Recover(p.ID); inGame {
			continue
		}
		onlineDTOs = append(onlineDTOs, protocol.ParticipantDTO{ID: p.ID, Username: p.Username})
	}
	invitations := g.store.ListInvitationsForParticipant(participantID)
	invDTOs := make([]protocol.InvitationDTO, 0, len(invitations))
	for _, inv := range invitations {
		invDTOs = append(invDTOs, invitationDTO(inv))
	}
	return protocol.LobbyStatePayload{Online: onlineDTOs, Invitations: invDTOs}
}

func invitationDTO(inv domain.Invitation) protocol.InvitationDTO {
	return protocol.InvitationDTO{
		ID:              inv.ID,
		FromParticipant: protocol.ParticipantDTO{ID: inv.From.ID, Username: inv.From.Username},
		ToParticipant:   protocol.ParticipantDTO{ID: inv.To.ID, Username: inv.To.Username},
		Status:          string(inv.Status),
		CreatedAt:       inv.CreatedAt,
	}
}

func mustEncode(msgType string, payload any) []byte {
	msg, err := protocol.Encode(msgType, payload)
	if err != nil {
		panic(err)
	}
	return msg
}

func matchmakingErrorCode(err error) string {
	switch {
	case errors.Is(err, matchmaking.ErrAlreadyInSession):
		return protocol.ErrCodeAlreadyInSession
	case errors.Is(err, matchmaking.ErrUnknownParticipant):
		return protocol.ErrCodeParticipantOffline
	case errors.Is(err, matchmaking.ErrInvitationNotFound), errors.Is(err, matchmaking.ErrNotInvitationTarget):
		return protocol.ErrCodeInvitationNotFound
	default:
		return protocol.ErrCodeInvalidMessage
	}
}

func sessionErrorCode(err error) string {
	switch {
	case errors.Is(err, session.ErrSessionNotFound):
		return protocol.ErrCodeSessionNotFound
	case errors.Is(err, session.ErrNotYourTurn):
		return protocol.ErrCodeNotYourTurn
	case errors.Is(err, session.ErrNotAParticipant):
		return protocol.ErrCodeInvalidMessage
	case errors.Is(err, session.ErrSessionCompleted):
		return protocol.ErrCodeInvalidMove
	case errors.Is(err, game.ErrCellTaken), errors.Is(err, game.ErrPositionOutOfRange):
		return protocol.ErrCodeInvalidMove
	default:
		return protocol.ErrCodeInternal
	}
}
