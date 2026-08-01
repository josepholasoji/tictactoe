import { app, BrowserWindow, ipcMain } from "electron";
import net from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { loadConfig } from "./config.js";
import { loadIdentity, saveIdentity, type Identity } from "./identity.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const isDev = process.env.NODE_ENV === "development";

const IDENTITY_SLOT_BASE_PORT = 47850;
const MAX_IDENTITY_SLOTS = 32;

/**
 * Multiple instances of this app can run at once (the README's dev workflow
 * explicitly opens several to simulate different participants locally).
 * Left alone, they'd all resolve Electron's default userData path to the
 * same directory and therefore share one identity.json - every instance
 * would reconnect to the server as the *same* participant, and the
 * server's one-live-connection-per-participant rule would make them
 * repeatedly kick each other off (a reconnect war).
 *
 * To keep normal single-instance use (identity persists across restarts)
 * unaffected, only instances *beyond the first* get redirected to an
 * isolated, uniquely-suffixed userData directory. "First" is decided by
 * claiming a loopback TCP port as a liveness mutex: unlike a lock file, the
 * OS releases the port automatically if a process crashes or is killed, so
 * this can never get stuck pointing every future launch at a fresh
 * directory forever.
 */
function claimUserDataSlot(): Promise<void> {
  return new Promise((resolve) => {
    let slot = 0;

    const tryPort = () => {
      const server = net.createServer();
      server.unref();
      server.once("error", (err: NodeJS.ErrnoException) => {
        if (err.code !== "EADDRINUSE" || slot >= MAX_IDENTITY_SLOTS - 1) {
          resolve(); // don't block startup over this - worst case, shares identity
          return;
        }
        slot += 1;
        tryPort();
      });
      server.listen(IDENTITY_SLOT_BASE_PORT + slot, "127.0.0.1", () => {
        if (slot > 0) {
          app.setPath("userData", `${app.getPath("userData")}-instance${slot}`);
        }
        resolve();
      });
    };

    tryPort();
  });
}

async function createWindow(): Promise<void> {
  const win = new BrowserWindow({
    width: 1100,
    height: 800,
    minWidth: 720,
    minHeight: 560,
    title: "Tic-Tac-Toe",
    webPreferences: {
      // preload runs in Electron's sandboxed preload environment, which
      // only understands CommonJS - it's compiled from preload.cts (not
      // .ts) specifically so tsc emits .cjs regardless of this package's
      // "type": "module".
      preload: path.join(__dirname, "preload.cjs"),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  if (isDev) {
    await win.loadURL("http://localhost:5173");
    win.webContents.openDevTools({ mode: "detach" });
  } else {
    await win.loadFile(path.join(__dirname, "../dist/index.html"));
  }
}

function registerIpcHandlers(): void {
  ipcMain.handle("config:get", async () => loadConfig(app));

  ipcMain.handle("identity:get", async () => loadIdentity(app));

  ipcMain.handle("identity:set", async (_event, identity: Identity) => {
    await saveIdentity(app, identity);
  });
}

// Must resolve before app.whenReady() - app.setPath("userData", ...) only
// has an effect if called before the app finishes initializing.
claimUserDataSlot().then(() =>
  app.whenReady().then(() => {
    registerIpcHandlers();
    void createWindow();

    app.on("activate", () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        void createWindow();
      }
    });
  }),
);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
