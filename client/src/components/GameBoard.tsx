import type { SessionDTO } from "../shared/protocol";

interface Props {
  session: SessionDTO;
  selfId: string;
  onMove: (position: number) => void;
  onLeave: () => void;
}

export function GameBoard({ session, selfId, onMove, onLeave }: Props) {
  const selfMark = session.participantX.id === selfId ? "X" : "O";
  const opponent = session.participantX.id === selfId ? session.participantO : session.participantX;
  const isActive = session.status === "active";
  const isMyTurn = isActive && session.turn === selfMark;

  let statusText: string;
  if (session.status === "completed") {
    if (session.isDraw) {
      statusText = "It's a draw";
    } else if (session.winnerId === selfId) {
      statusText = "You won!";
    } else {
      statusText = "You lost";
    }
  } else {
    statusText = isMyTurn ? "Your turn" : `Waiting for ${opponent.username}…`;
  }

  return (
    <div className="mx-auto flex max-w-md flex-col items-center gap-6">
      <div className="text-center">
        <p className="text-sm text-slate-400">
          You ({selfMark}) vs {opponent.username}
        </p>
        <p className="mt-1 text-lg font-semibold">{statusText}</p>
      </div>

      <div className="grid grid-cols-3 gap-2 rounded-lg bg-slate-900 p-2">
        {session.board.map((cell, position) => (
          <button
            key={position}
            type="button"
            disabled={!isMyTurn || cell !== ""}
            onClick={() => onMove(position)}
            className="flex h-24 w-24 items-center justify-center rounded-md bg-slate-800 text-4xl font-bold text-slate-100 transition enabled:hover:bg-slate-700 disabled:cursor-not-allowed"
          >
            <span className={cell === "X" ? "text-sky-400" : cell === "O" ? "text-rose-400" : ""}>{cell}</span>
          </button>
        ))}
      </div>

      {session.status === "completed" && (
        <button
          type="button"
          onClick={onLeave}
          className="rounded-md bg-slate-800 px-4 py-2 text-sm font-medium hover:bg-slate-700"
        >
          Back to Lobby
        </button>
      )}
    </div>
  );
}
