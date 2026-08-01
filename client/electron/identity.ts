import { promises as fs } from "node:fs";
import path from "node:path";
import type { App } from "electron";

export interface Identity {
  participantId?: string;
  username: string;
}

function identityPath(app: App): string {
  return path.join(app.getPath("userData"), "identity.json");
}

/**
 * Reads the persisted participant identity (participantId + username) so a
 * relaunched client reconnects as the same participant instead of the
 * server treating it as brand new. The renderer's "New Connection" dialog
 * is responsible for actually supplying a username - this never invents
 * one, so a first run (or a saved-but-empty username) returns "" and lets
 * the user type their own.
 */
export async function loadIdentity(app: App): Promise<Identity> {
  try {
    const raw = await fs.readFile(identityPath(app), "utf-8");
    const parsed = JSON.parse(raw) as Partial<Identity>;
    if (parsed.username) {
      return { participantId: parsed.participantId, username: parsed.username };
    }
    return { participantId: parsed.participantId, username: "" };
  } catch {
    return { username: "" };
  }
}

export async function saveIdentity(app: App, identity: Identity): Promise<void> {
  await fs.writeFile(identityPath(app), JSON.stringify(identity, null, 2), "utf-8");
}
