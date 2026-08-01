import type { ParticipantDTO } from "../shared/protocol";

interface Props {
  online: ParticipantDTO[];
  pendingParticipantIds: Set<string>;
  onSelect: (participantId: string) => void;
  onClose: () => void;
}

export function PlayGameDialog({ online, pendingParticipantIds, onSelect, onClose }: Props) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-6" onClick={onClose}>
      <div
        className="w-full max-w-sm rounded-lg bg-slate-900 p-4 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400">Play Game</h2>
          <button type="button" onClick={onClose} className="text-slate-500 hover:text-slate-300">
            ✕
          </button>
        </div>

        <p className="mb-3 text-xs text-slate-500">Select a participant to invite.</p>

        <ul className="max-h-72 space-y-1.5 overflow-y-auto">
          {online.length === 0 && <li className="text-sm text-slate-500">No one else is online yet.</li>}
          {online.map((p) => {
            const pending = pendingParticipantIds.has(p.id);
            return (
              <li key={p.id} className="flex items-center justify-between rounded-md bg-slate-800 px-3 py-2 text-sm">
                <span>{p.username}</span>
                <button
                  type="button"
                  disabled={pending}
                  onClick={() => onSelect(p.id)}
                  className="rounded bg-sky-600 px-2 py-1 text-xs hover:enabled:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {pending ? "Invite sent" : "Play"}
                </button>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
