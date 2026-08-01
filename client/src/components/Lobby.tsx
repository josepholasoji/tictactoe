import { useMemo, useState } from "react";
import type { InvitationDTO, ParticipantDTO } from "../shared/protocol";
import { PlayGameDialog } from "./PlayGameDialog";

interface Props {
  selfId: string;
  online: ParticipantDTO[];
  invitations: InvitationDTO[];
  onInvite: (participantId: string) => void;
  onRespondInvite: (invitationId: string, accept: boolean) => void;
}

export function Lobby({ selfId, online, invitations, onInvite, onRespondInvite }: Props) {
  const [showPlayDialog, setShowPlayDialog] = useState(false);

  const incoming = invitations.filter((inv) => inv.toParticipant.id === selfId);
  const outgoing = invitations.filter((inv) => inv.fromParticipant.id === selfId);
  const pendingParticipantIds = useMemo(
    () => new Set(outgoing.map((inv) => inv.toParticipant.id)),
    [outgoing],
  );

  function handleSelect(participantId: string) {
    onInvite(participantId);
    setShowPlayDialog(false);
  }

  return (
    <div className="mx-auto max-w-xl">
      <section className="rounded-lg bg-slate-900 p-4">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400">Play</h2>
          <button
            type="button"
            onClick={() => setShowPlayDialog(true)}
            className="rounded-md bg-sky-600 px-3 py-1.5 text-sm font-medium hover:bg-sky-500"
          >
            Play Game
          </button>
        </div>

        {incoming.length > 0 && (
          <div className="mb-4 space-y-2">
            <h3 className="text-xs font-semibold uppercase text-slate-500">Invitations</h3>
            {incoming.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between rounded-md bg-slate-800 px-3 py-2 text-sm">
                <span>{inv.fromParticipant.username} challenged you</span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => onRespondInvite(inv.id, true)}
                    className="rounded bg-emerald-600 px-2 py-1 text-xs hover:bg-emerald-500"
                  >
                    Accept
                  </button>
                  <button
                    type="button"
                    onClick={() => onRespondInvite(inv.id, false)}
                    className="rounded bg-slate-700 px-2 py-1 text-xs hover:bg-slate-600"
                  >
                    Decline
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        <h3 className="mb-2 text-xs font-semibold uppercase text-slate-500">Online ({online.length})</h3>
        <ul className="space-y-1.5">
          {online.length === 0 && <li className="text-sm text-slate-500">No one else is online yet.</li>}
          {online.map((p) => (
            <li key={p.id} className="flex items-center justify-between rounded-md bg-slate-800 px-3 py-2 text-sm">
              <span>{p.username}</span>
              {pendingParticipantIds.has(p.id) && <span className="text-xs text-slate-500">Invite sent</span>}
            </li>
          ))}
        </ul>
      </section>

      {showPlayDialog && (
        <PlayGameDialog
          online={online}
          pendingParticipantIds={pendingParticipantIds}
          onSelect={handleSelect}
          onClose={() => setShowPlayDialog(false)}
        />
      )}
    </div>
  );
}
