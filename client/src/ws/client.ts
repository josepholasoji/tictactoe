import type { ClientConfig } from "../shared/desktop";
import {
  ClientMessageType,
  type Envelope,
  type HelloPayload,
  type InviteCreatePayload,
  type InviteRespondPayload,
  type MovePayload,
} from "../shared/protocol";

export type ConnectionStatus = "connecting" | "connected" | "reconnecting" | "disconnected";

type Listener = (payload: unknown) => void;

const HEARTBEAT_INTERVAL_MS = 10_000;

/** Exponential backoff delay (pre-jitter), capped at maxMs. Exported for unit testing. */
export function capExponential(attempt: number, baseMs: number, maxMs: number): number {
  return Math.min(baseMs * 2 ** attempt, maxMs);
}

/**
 * Thin WebSocket client: owns the socket lifecycle, reconnection with
 * exponential backoff + jitter, and the periodic heartbeat. Message
 * handling is left to whoever calls `.on(type, handler)` (the Zustand
 * store) so this class stays free of game/lobby domain logic.
 */
export class WSClient {
  private socket: WebSocket | null = null;
  private attempts = 0;
  private manuallyClosed = false;
  private heartbeatHandle: ReturnType<typeof setInterval> | null = null;
  private reconnectHandle: ReturnType<typeof setTimeout> | null = null;
  private listeners = new Map<string, Set<Listener>>();
  private statusListeners = new Set<(status: ConnectionStatus) => void>();
  private status: ConnectionStatus = "disconnected";

  constructor(
    private readonly config: ClientConfig,
    private getHello: () => HelloPayload,
  ) {}

  connect(): void {
    this.manuallyClosed = false;
    this.open();
  }

  close(): void {
    this.manuallyClosed = true;
    this.stopHeartbeat();
    if (this.reconnectHandle) clearTimeout(this.reconnectHandle);
    this.socket?.close();
  }

  onStatusChange(listener: (status: ConnectionStatus) => void): () => void {
    this.statusListeners.add(listener);
    return () => this.statusListeners.delete(listener);
  }

  on<T = unknown>(type: string, listener: (payload: T) => void): () => void {
    const set = this.listeners.get(type) ?? new Set();
    set.add(listener as Listener);
    this.listeners.set(type, set);
    return () => set.delete(listener as Listener);
  }

  createInvitation(toParticipantId: string): void {
    this.send(ClientMessageType.InviteCreate, { toParticipantId } satisfies InviteCreatePayload);
  }

  respondInvitation(invitationId: string, accept: boolean): void {
    this.send(ClientMessageType.InviteRespond, { invitationId, accept } satisfies InviteRespondPayload);
  }

  sendMove(sessionId: string, position: number): void {
    this.send(ClientMessageType.Move, { sessionId, position } satisfies MovePayload);
  }

  private open(): void {
    this.setStatus(this.attempts > 0 ? "reconnecting" : "connecting");
    const socket = new WebSocket(this.config.serverUrl);
    this.socket = socket;

    socket.addEventListener("open", () => {
      this.attempts = 0;
      this.setStatus("connected");
      this.send(ClientMessageType.Hello, this.getHello());
      this.startHeartbeat();
    });

    socket.addEventListener("message", (event) => {
      this.handleMessage(String(event.data));
    });

    socket.addEventListener("close", () => {
      this.stopHeartbeat();
      if (this.manuallyClosed) {
        this.setStatus("disconnected");
        return;
      }
      this.scheduleReconnect();
    });

    socket.addEventListener("error", () => {
      socket.close();
    });
  }

  private handleMessage(raw: string): void {
    let envelope: Envelope;
    try {
      envelope = JSON.parse(raw) as Envelope;
    } catch {
      return;
    }
    const set = this.listeners.get(envelope.type);
    if (!set) return;
    for (const listener of set) listener(envelope.payload);
  }

  private send(type: string, payload: unknown): void {
    if (this.socket?.readyState !== WebSocket.OPEN) return;
    const envelope: Envelope = { type, payload };
    this.socket.send(JSON.stringify(envelope));
  }

  private scheduleReconnect(): void {
    if (this.attempts >= this.config.maxReconnectAttempts) {
      this.setStatus("disconnected");
      return;
    }
    this.setStatus("reconnecting");
    const capped = capExponential(this.attempts, this.config.baseReconnectDelayMs, this.config.maxReconnectDelayMs);
    const jitter = capped * (0.5 + Math.random() * 0.5);
    this.attempts += 1;
    this.reconnectHandle = setTimeout(() => this.open(), jitter);
  }

  private startHeartbeat(): void {
    this.heartbeatHandle = setInterval(() => {
      this.send(ClientMessageType.Heartbeat, {});
    }, HEARTBEAT_INTERVAL_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatHandle) {
      clearInterval(this.heartbeatHandle);
      this.heartbeatHandle = null;
    }
  }

  private setStatus(status: ConnectionStatus): void {
    this.status = status;
    for (const listener of this.statusListeners) listener(status);
  }

  getStatus(): ConnectionStatus {
    return this.status;
  }
}
