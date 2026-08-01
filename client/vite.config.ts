import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

// Relative base so the built renderer can be loaded from a file:// URL by
// Electron's BrowserWindow in production, not just from an HTTP dev server.
export default defineConfig({
  base: "./",
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
  },
  build: {
    outDir: "dist",
  },
  test: {
    environment: "node",
  },
});
