import Dotenv from "dotenv";
import { db } from "./PrismaClient.js";
Dotenv.config();

if (!process.env.BOT_CLIENT_ID) {
  throw new Error("BOT_CLIENT_ID is not defined in .env");
}
if (!process.env.DISCORD_GUILD_ID) {
  throw new Error("DISCORD_GUILD_ID is not defined in .env");
}
if (!process.env.DISCORD_TOKEN) {
  throw new Error("DISCORD_TOKEN is not defined in .env");
}
if (!process.env.DISCORD_TOKEN) {
  throw new Error("DISCORD_TOKEN is not defined in .env");
}
if (!process.env.DISCORD_TOKEN) {
  throw new Error("DISCORD_TOKEN is not defined in .env");
}
if (!process.env.TURSO_DATABASE_URL) {
  throw new Error("TURSO_DATABASE_URL is not defined in .env");
}
if (!process.env.TURSO_AUTH_TOKEN) {
  throw new Error("TURSO_AUTH_TOKEN is not defined in .env");
}
const defaultGlobals = {
  BOT_CLIENT_ID: process.env.BOT_CLIENT_ID,
  CASH_CONVERTERS: false,
  CASH_CONVERTERS_CHANNEL_ID: null,
  CASH_CONVERTERS_ROLE_ID: null,
  COMMAND_PERMISSION_ROLE_ID: null,
  CS_CHANNEL_ID: null,
  CS_ITEMS: false,
  CS_MARKET_CHANNEL_ID: null,
  CS_MARKET_ROLE_ID: null,
  CS_ROLE_ID: null,
  DISCORD_GUILD_ID: process.env.DISCORD_GUILD_ID,
  DISCORD_TOKEN: process.env.DISCORD_TOKEN,
  EBAY: false,
  EBAY_CHANNEL_ID: null,
  EBAY_ROLE_ID: null,
  ERROR_CHANNEL_ID: null,
  GUMTREE: false,
  GUMTREE_CHANNEL_ID: null,
  GUMTREE_ROLE_ID: null,
  SALVOS: false,
  SALVOS_CHANNEL_ID: null,
  SALVOS_ROLE_ID: null,
  STEAM_QUERY: false,
  STEAM_QUERY_CHANNEL_ID: null,
  STEAM_QUERY_ROLE_ID: null,
};

export const MONGO_URI = process.env.MONGO_URI;

export async function initGlobals() {
  const globals = await db.globals.findFirst();

  if (!globals) return;

  // Copy every non-null and non-undefined value from globals to defaultGlobals
  for (const key in defaultGlobals) {
    if (
      globals[key as keyof typeof defaultGlobals] !== null &&
      globals[key as keyof typeof defaultGlobals] !== undefined
    ) {
      (defaultGlobals as any)[key as keyof typeof defaultGlobals] =
        globals[key as keyof typeof defaultGlobals];
    }
  }
}

export default defaultGlobals;
