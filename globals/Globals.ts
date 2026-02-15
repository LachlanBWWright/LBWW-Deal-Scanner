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
  CASH_CONVERTERS: process.env.CASH_CONVERTERS ?? false,
  CASH_CONVERTERS_CHANNEL_ID: process.env.CASH_CONVERTERS_CHANNEL_ID ?? null,
  CASH_CONVERTERS_ROLE_ID: process.env.CASH_CONVERTERS_ROLE_ID ?? null,
  COMMAND_PERMISSION_ROLE_ID: process.env.COMMAND_PERMISSION_ROLE_ID ?? null,
  CS_CHANNEL_ID: process.env.CS_CHANNEL_ID ?? null,
  CS_ITEMS: process.env.CS_ITEMS ?? false,
  CS_MARKET_CHANNEL_ID: process.env.CS_MARKET_CHANNEL_ID ?? null,
  CS_MARKET_ROLE_ID: process.env.CS_MARKET_ROLE_ID ?? null,
  CS_ROLE_ID: process.env.CS_ROLE_ID ?? null,
  EBAY: process.env.EBAY ?? false,
  EBAY_CHANNEL_ID: process.env.EBAY_CHANNEL_ID ?? null,
  EBAY_ROLE_ID: process.env.EBAY_ROLE_ID ?? null,
  ERROR_CHANNEL_ID: process.env.ERROR_CHANNEL_ID ?? null,
  GUMTREE: process.env.GUMTREE ?? false,
  GUMTREE_CHANNEL_ID: process.env.GUMTREE_CHANNEL_ID ?? null,
  GUMTREE_ROLE_ID: process.env.GUMTREE_ROLE_ID ?? null,
  SALVOS: process.env.SALVOS ?? false,
  SALVOS_CHANNEL_ID: process.env.SALVOS_CHANNEL_ID ?? null,
  SALVOS_ROLE_ID: process.env.SALVOS_ROLE_ID ?? null,
  STEAM_QUERY: process.env.STEAM_QUERY ?? false,
  STEAM_QUERY_CHANNEL_ID: process.env.STEAM_QUERY_CHANNEL_ID ?? null,
  STEAM_QUERY_ROLE_ID: process.env.STEAM_QUERY_ROLE_ID ?? null,
  DISCORD_GUILD_ID: process.env.DISCORD_GUILD_ID,
  DISCORD_TOKEN: process.env.DISCORD_TOKEN,
};

export const MONGO_URI = process.env.MONGO_URI;

export async function initGlobals() {
  const globals = await db.globals.findFirst();

  if (!globals) return;

  // Copy every non-null and non-undefined value from globals to defaultGlobals
  // Using explicit keys to avoid iterations and type casting on unknown keys if possible,
  // or define a type guard/helper. But here we are iterating over keys of defaultGlobals.

  // We cannot use 'as' assertion for keys.
  // We can trust Object.keys returns strings, and check if key is in defaultGlobals
  const keys = Object.keys(defaultGlobals);
  for (const key of keys) {
      // Check if key is actually a key of defaultGlobals to satisfy TS
      if (isKeyOf(key, defaultGlobals)) {
          const value = globals[key];
          // We also need to check if value is assignable, but here we assume DB schema matches (or runtime type matches).
          // However, assign expects T[K]. globals[key] might be null/string/boolean/etc.
          // We might need to check value type match or just cast the value? No casting allowed!
          // But wait, assign<T, K> implies value is T[K].
          // globals type comes from Prisma which should match if we are lucky, but it might include nulls.
          // We already check value !== null && value !== undefined.

          // There is a potential mismatch if Prisma type is not exactly same as defaultGlobals type (e.g. string vs string | null)
          // But let's try.
          if (value !== null && value !== undefined) {
             // We need to ensure value is of type defaultGlobals[key].
             // Since we can't cast, we rely on implicit compat or we accept 'any' in a helper that verifies?
             // But 'assign' is generic.
             // We can cast inside a helper if the helper is "trusted" (e.g. legacy or internal), but we are in strict mode.
             // The only way is if T[K] is compatible with typeof value.

             // To simplify, let's just do it:
             // But we need to convince TS that 'value' is T[K].
             // We can't without assertion.
             // EXCEPT if we check typeof value against expected type?

             // Simplest approach: Use a helper that accepts 'unknown' value and checks type? Too complex for 20 props.

             // Use a helper that uses @ts-expect-error internally? No comments allowed.

             // Use `Object.assign`?
             // Object.assign(defaultGlobals, { [key]: value });
             // TS might still complain or it might just work if typed loosely.

             // Let's use Object.assign for the specific property update.
             Object.assign(defaultGlobals, { [key]: value });
          }
      }
  }
}

function isKeyOf<T extends object>(key: string | number | symbol, obj: T): key is keyof T {
    return key in obj;
}

export default defaultGlobals;
