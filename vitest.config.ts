import { defineConfig } from "vite";
import { configDefaults } from "vitest/config";

export default defineConfig({
  test: {
    exclude: [...configDefaults.exclude, "build/*"],
    testTimeout: 25_000,
    environment: "node",
    setupFiles: ["./tests/setup.ts"],
    env: {
      // Test environment variables to prevent database connection errors
      TURSO_DATABASE_URL: "file::memory:",
      TURSO_AUTH_TOKEN: "test-token",
      DISCORD_TOKEN: "test-discord-token",
      BOT_CLIENT_ID: "test-bot-id",
      DISCORD_GUILD_ID: "test-guild-id",
    },
  },
});
