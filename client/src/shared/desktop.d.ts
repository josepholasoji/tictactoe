import type { ClientConfig } from "../../electron/config";
import type { Identity } from "../../electron/identity";

export type { ClientConfig, Identity };

export interface DesktopBridge {
  getConfig(): Promise<ClientConfig>;
  getIdentity(): Promise<Identity>;
  setIdentity(identity: Identity): Promise<void>;
}

declare global {
  interface Window {
    desktop: DesktopBridge;
  }
}
