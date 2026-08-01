import { useState, type FormEvent } from "react";

interface Props {
  initialUsername: string;
  onConnect: (username: string) => void;
}

export function ConnectDialog({ initialUsername, onConnect }: Props) {
  const [name, setName] = useState(initialUsername);
  const trimmed = name.trim();

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!trimmed) return;
    onConnect(trimmed);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-6">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-lg bg-slate-900 p-4 shadow-xl"
      >
        <fieldset className="rounded-md border border-slate-700 p-4">
          <legend className="px-2 text-sm font-semibold uppercase tracking-wide text-slate-400">
            New Connection
          </legend>

          <label htmlFor="participant-name" className="mb-1 block text-sm text-slate-300">
            Participant name
          </label>
          <input
            id="participant-name"
            type="text"
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
            className="mb-4 w-full rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-sky-500 focus:outline-none"
          />

          <button
            type="submit"
            disabled={!trimmed}
            className="w-full rounded-md bg-sky-600 px-3 py-2 text-sm font-medium hover:enabled:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Connect
          </button>
        </fieldset>
      </form>
    </div>
  );
}
