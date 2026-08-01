import { contextBridge, ipcRenderer } from "electron";
import type { ClientConfig } from "./config.js";
import type { Identity } from "./identity.js";

export interface DesktopBridge {
  getConfig(): Promise<ClientConfig>;
  getIdentity(): Promise<Identity>;
  setIdentity(identity: Identity): Promise<void>;
}

const bridge: DesktopBridge = {
  getConfig: () => ipcRenderer.invoke("config:get"),
  getIdentity: () => ipcRenderer.invoke("identity:get"),
  setIdentity: (identity: Identity) => ipcRenderer.invoke("identity:set", identity),
};

contextBridge.exposeInMainWorld("desktop", bridge);
