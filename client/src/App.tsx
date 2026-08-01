import { useEffect, useState } from "react";
import { useStore } from "./store/useStore";
import { ConnectionStatus } from "./components/ConnectionStatus";
import { ConnectDialog } from "./components/ConnectDialog";
import { Lobby } from "./components/Lobby";
import { GameBoard } from "./components/GameBoard";
import type { ClientConfig, Identity } from "./shared/desktop";

export default function App() {
  const status = useStore((s) => s.status);
  const self = useStore((s) => s.self);
  const online = useStore((s) => s.online);
  const invitations = useStore((s) => s.invitations);
  const session = useStore((s) => s.session);
  const lastError = useStore((s) => s.lastError);

  const initialize = useStore((s) => s.initialize);
  const invite = useStore((s) => s.invite);
  const respondInvite = useStore((s) => s.respondInvite);
  const move = useStore((s) => s.move);
  const dismissError = useStore((s) => s.dismissError);
  const leaveGame = useStore((s) => s.leaveGame);

  const [startup, setStartup] = useState<{ config: ClientConfig; identity: Identity } | null>(null);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    let cancelled = false;
    Promise.all([window.desktop.getConfig(), window.desktop.getIdentity()]).then(([config, identity]) => {
      if (cancelled) return;
      setStartup({ config, identity });
    });
    return () => {
      cancelled = true;
    };
  }, []);

  function handleConnect(username: string) {
    if (!startup) return;
    setSubmitted(true);
    initialize(startup.config, { participantId: startup.identity.participantId, username });
  }

  return (
    <div className="flex min-h-screen flex-col gap-6 p-6">
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold">Tic-Tac-Toe</h1>
          {self && <p className="text-xs text-slate-500">Playing as {self.username}</p>}
        </div>
        <ConnectionStatus status={status} />
      </header>

      {lastError && (
        <div className="flex items-center justify-between rounded-md bg-rose-950 px-4 py-2 text-sm text-rose-200">
          <span>{lastError.message}</span>
          <button type="button" onClick={dismissError} className="text-rose-300 hover:text-rose-100">
            Dismiss
          </button>
        </div>
      )}

      <main className="flex-1">
        {!startup || !submitted ? (
          startup && <ConnectDialog initialUsername={startup.identity.username} onConnect={handleConnect} />
        ) : !self ? (
          <p className="text-center text-slate-500">Connecting to server…</p>
        ) : session ? (
          <GameBoard session={session} selfId={self.participantId} onMove={move} onLeave={leaveGame} />
        ) : (
          <Lobby
            selfId={self.participantId}
            online={online}
            invitations={invitations}
            onInvite={invite}
            onRespondInvite={respondInvite}
          />
        )}
      </main>
    </div>
  );
}
