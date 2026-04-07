/// <reference types="vitest/config" />

import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@verin/ui": path.resolve(__dirname, "../../packages/ui/src/index.ts"),
      "@verin/api-client": path.resolve(__dirname, "../../packages/api-client/src/index.ts")
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: "./src/test/setup.ts",
    env: {
      NODE_ENV: "test",
    },
  }
});
