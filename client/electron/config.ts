import { promises as fs } from "node:fs";
import path from "node:path";
import type { App } from "electron";

export interface ClientConfig {
  serverUrl: string;
  maxReconnectAttempts: number;
  baseReconnectDelayMs: number;
  maxReconnectDelayMs: number;
}

export const DEFAULT_CONFIG: ClientConfig = {
  serverUrl: "ws://localhost:8080/ws",
  maxReconnectAttempts: 20,
  baseReconnectDelayMs: 500,
  maxReconnectDelayMs: 10000,
};

/**
 * Loads client/config.json from the user's data directory if present,
 * otherwise falls back to DEFAULT_CONFIG. This lets a packaged build be
 * pointed at a different server without rebuilding, per the README's
 * documented client configuration.
 */
export async function loadConfig(app: App): Promise<ClientConfig> {
  const configPath = path.join(app.getPath("userData"), "config.json");
  try {
    const raw = await fs.readFile(configPath, "utf-8");
    const parsed = JSON.parse(raw);
    return { ...DEFAULT_CONFIG, ...parsed };
  } catch {
    if (process.env.SERVER_WS_URL) {
      return { ...DEFAULT_CONFIG, serverUrl: process.env.SERVER_WS_URL };
    }
    return DEFAULT_CONFIG;
  }
}
