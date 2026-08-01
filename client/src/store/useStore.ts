import { create } from "zustand";
import { WSClient, type ConnectionStatus } from "../ws/client";
import type { ClientConfig, Identity } from "../shared/desktop";
import {
  ServerMessageType,
  type ErrorPayload,
  type GameCompletedPayload,
  type GameStartedPayload,
  type InvitationDTO,
  type InviteCanceledPayload,
  type InviteReceivedPayload,
  type InviteSentPayload,
  type LobbyStatePayload,
  type MoveAppliedPayload,
  type ParticipantDTO,
  type PresenceUpdatePayload,
  type SessionDTO,
  type WelcomePayload,
} from "../shared/protocol";

interface AppState {
  status: ConnectionStatus;
  self: { participantId: string; username: string } | null;
  online: ParticipantDTO[];
  invitations: InvitationDTO[];
  session: SessionDTO | null;
  lastError: ErrorPayload | null;

  initialize: (config: ClientConfig, identity: Identity) => void;
  invite: (toParticipantId: string) => void;
  respondInvite: (invitationId: string, accept: boolean) => void;
  move: (position: number) => void;
  dismissError: () => void;
  leaveGame: () => void;
}

let client: WSClient | null = null;

export const useStore = create<AppState>((set, get) => ({
  status: "disconnected",
  self: null,
  online: [],
  invitations: [],
  session: null,
  lastError: null,

  initialize: (config, identity) => {
    if (client) return; // already initialized (e.g. React StrictMode double-invoke)

    client = new WSClient(config, () => ({
      participantId: identity.participantId,
      username: identity.username,
    }));

    client.onStatusChange((status) => set({ status }));

    client.on<WelcomePayload>(ServerMessageType.Welcome, (payload) => {
      set({
        self: { participantId: payload.participantId, username: payload.username },
        session: payload.activeSession ?? null,
      });
      void window.desktop.setIdentity({
        participantId: payload.participantId,
        username: payload.username,
      });
    });

    client.on<LobbyStatePayload>(ServerMessageType.LobbyState, (payload) => {
      set({ online: payload.online, invitations: payload.invitations });
    });

    client.on<PresenceUpdatePayload>(ServerMessageType.PresenceUpdate, (payload) => {
      set((state) => {
        // Presence broadcasts go to every connected client, including the
        // one the update is about - never let a participant add themselves
        // to their own online/invitable list.
        if (payload.participant.id === state.self?.participantId) return state;
        const withoutParticipant = state.online.filter((p) => p.id !== payload.participant.id);
        return {
          online: payload.online ? [...withoutParticipant, payload.participant] : withoutParticipant,
        };
      });
    });

    client.on<GameStartedPayload>(ServerMessageType.GameStarted, (payload) => {
      set({ session: payload.session });
    });

    client.on<MoveAppliedPayload>(ServerMessageType.MoveApplied, (payload) => {
      set({ session: payload.session });
    });

    client.on<GameCompletedPayload>(ServerMessageType.GameCompleted, (payload) => {
      set({ session: payload.session });
    });

    client.on<InviteReceivedPayload>(ServerMessageType.InviteReceived, (payload) => {
      set((state) => ({ invitations: [...state.invitations, payload.invitation] }));
    });

    client.on<InviteSentPayload>(ServerMessageType.InviteSent, (payload) => {
      set((state) => ({ invitations: [...state.invitations, payload.invitation] }));
    });

    client.on<InviteCanceledPayload>(ServerMessageType.InviteCanceled, (payload) => {
      set((state) => ({
        invitations: state.invitations.filter((inv) => inv.id !== payload.invitationId),
      }));
    });

    client.on<ErrorPayload>(ServerMessageType.Error, (payload) => {
      set({ lastError: payload });
    });

    client.connect();
  },

  invite: (toParticipantId) => client?.createInvitation(toParticipantId),
  respondInvite: (invitationId, accept) => {
    client?.respondInvitation(invitationId, accept);
    set((state) => ({ invitations: state.invitations.filter((inv) => inv.id !== invitationId) }));
  },
  move: (position) => {
    const session = get().session;
    if (!session) return;
    client?.sendMove(session.id, position);
  },
  dismissError: () => set({ lastError: null }),
  leaveGame: () => set({ session: null }),
}));
