/**
 * TypeScript mirror of the Go server's WebSocket message contract
 * (server/internal/protocol/messages.go). Field names and message types
 * must be kept in sync between the two.
 */

// Message types sent from the client to the server.
export const ClientMessageType = {
  Hello: "hello",
  InviteCreate: "invite_create",
  InviteRespond: "invite_respond",
  Move: "move",
  Heartbeat: "heartbeat",
} as const;

// Message types sent from the server to the client.
export const ServerMessageType = {
  Welcome: "welcome",
  LobbyState: "lobby_state",
  PresenceUpdate: "presence_update",
  InviteReceived: "invite_received",
  InviteSent: "invite_sent",
  InviteCanceled: "invite_canceled",
  GameStarted: "game_started",
  MoveApplied: "move_applied",
  GameCompleted: "game_completed",
  HeartbeatAck: "heartbeat_ack",
  Error: "error",
} as const;

export type ErrorCode =
  | "invalid_message"
  | "not_your_turn"
  | "invalid_move"
  | "session_not_found"
  | "participant_offline"
  | "already_in_session"
  | "invitation_not_found"
  | "internal_error";

export interface Envelope<T = unknown> {
  type: string;
  payload?: T;
}

// --- DTOs -------------------------------------------------------------

export interface ParticipantDTO {
  id: string;
  username: string;
}

export type InvitationStatus = "pending" | "accepted" | "declined" | "canceled";

export interface InvitationDTO {
  id: string;
  fromParticipant: ParticipantDTO;
  toParticipant: ParticipantDTO;
  status: InvitationStatus;
  createdAt: string;
}

export type Mark = "X" | "O" | "";

export type SessionStatus = "active" | "completed";

export interface SessionDTO {
  id: string;
  participantX: ParticipantDTO;
  participantO: ParticipantDTO;
  board: Mark[]; // length 9
  turn: "X" | "O";
  status: SessionStatus;
  winnerId?: string;
  isDraw: boolean;
  startedAt: string;
  completedAt?: string;
}

// --- Client -> Server payloads -----------------------------------------

export interface HelloPayload {
  participantId?: string;
  username: string;
}

export interface InviteCreatePayload {
  toParticipantId: string;
}

export interface InviteRespondPayload {
  invitationId: string;
  accept: boolean;
}

export interface MovePayload {
  sessionId: string;
  position: number;
}

// --- Server -> Client payloads -------------------------------------------

export interface WelcomePayload {
  participantId: string;
  username: string;
  reconnected: boolean;
  activeSession?: SessionDTO;
}

export interface LobbyStatePayload {
  online: ParticipantDTO[];
  invitations: InvitationDTO[];
}

export interface PresenceUpdatePayload {
  participant: ParticipantDTO;
  online: boolean;
}

export interface InviteReceivedPayload {
  invitation: InvitationDTO;
}

export interface InviteSentPayload {
  invitation: InvitationDTO;
}

export interface InviteCanceledPayload {
  invitationId: string;
}

export interface GameStartedPayload {
  session: SessionDTO;
}

export interface MoveAppliedPayload {
  session: SessionDTO;
}

export interface GameCompletedPayload {
  session: SessionDTO;
}

export interface ErrorPayload {
  code: ErrorCode;
  message: string;
}
