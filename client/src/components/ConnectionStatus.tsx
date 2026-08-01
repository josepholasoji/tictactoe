import type { ConnectionStatus as Status } from "../ws/client";

const LABEL: Record<Status, string> = {
  connecting: "Connecting…",
  connected: "Connected",
  reconnecting: "Reconnecting…",
  disconnected: "Disconnected",
};

const DOT_CLASS: Record<Status, string> = {
  connecting: "bg-amber-400 animate-pulse",
  connected: "bg-emerald-400",
  reconnecting: "bg-amber-400 animate-pulse",
  disconnected: "bg-rose-500",
};

export function ConnectionStatus({ status }: { status: Status }) {
  return (
    <div className="flex items-center gap-2 rounded-full bg-slate-900 px-3 py-1 text-xs text-slate-300">
      <span className={`h-2 w-2 rounded-full ${DOT_CLASS[status]}`} />
      {LABEL[status]}
    </div>
  );
}
