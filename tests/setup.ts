import { vi } from "vitest";

// Mock environment variables for tests
process.env.TURSO_DATABASE_URL = "file::memory:";
process.env.TURSO_AUTH_TOKEN = "test-token";
process.env.DISCORD_TOKEN = "test-discord-token";
process.env.BOT_CLIENT_ID = "test-bot-id";
process.env.DISCORD_GUILD_ID = "test-guild-id";

// Mock Prisma client to avoid database connections in tests
vi.mock("../globals/PrismaClient.js", () => ({
  db: {
    // Mock database methods as needed
    item: {
      findMany: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
    // Add other tables as needed
  },
  SCANNER: {
    CASH_CONVERTERS: 0,
    EBAY: 1,
    GUMTREE: 2,
    SALVOS: 3,
    STEAM_MARKET_QUERY: 4,
    STEAM_MARKET_CS: 5,
    CS_TRADE_BOT: 6,
  }
}));

// Mock Discord client to avoid Discord connections in tests
vi.mock("../globals/DiscordJSClient.js", () => ({
  default: {
    login: vi.fn(),
    once: vi.fn(),
    on: vi.fn(),
    channels: {
      cache: {
        get: vi.fn(),
      },
    },
  }
}));

// Mock globals to avoid initialization issues
vi.mock("../globals/Globals.js", () => ({
  default: {
    DISCORD_TOKEN: "test-token",
    BOT_CLIENT_ID: "test-bot-id", 
    DISCORD_GUILD_ID: "test-guild-id",
    ERROR_CHANNEL_ID: "test-error-channel",
  },
  initGlobals: vi.fn(),
}));