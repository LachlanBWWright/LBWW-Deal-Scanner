import Dotenv from "dotenv";
import { err, ok, Result } from "neverthrow";
import { db } from "./PrismaClient.js";
import { resultAsync } from "../functions/neverthrowUtils.js";
Dotenv.config({ quiet: true });

const requiredEnvVars = [
  "BOT_CLIENT_ID",
  "DISCORD_GUILD_ID",
  "DISCORD_TOKEN",
  "TURSO_DATABASE_URL",
  "TURSO_AUTH_TOKEN",
] as const;

function getRequiredEnv(
  name: (typeof requiredEnvVars)[number],
): Result<string, Error> {
  const value = process.env[name];
  if (!value) return err(new Error(`${name} is not defined in .env`));
  return ok(value);
}

for (const envName of requiredEnvVars) {
  getRequiredEnv(envName).match(
    () => undefined,
    (error) => {
      console.warn(error.message);
      return undefined;
    },
  );
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

export async function initGlobals() {
  const globalsResult = await resultAsync(
    () => db.globals.findFirst(),
    "Unable to load globals from database",
  );
  if (globalsResult.isErr()) {
    console.warn(globalsResult.error.message);
    return;
  }
  const globals = globalsResult.value;

  if (!globals) return;

  const keys = Object.keys(defaultGlobals);
  for (const key of keys) {
    if (isKeyOf(key, defaultGlobals)) {
      const value = globals[key];
      if (value !== null && value !== undefined) {
        Object.assign(defaultGlobals, { [key]: value });
      }
    }
  }
}

function isKeyOf<T extends object>(
  key: string | number | symbol,
  obj: T,
): key is keyof T {
  return key in obj;
}

export default defaultGlobals;
