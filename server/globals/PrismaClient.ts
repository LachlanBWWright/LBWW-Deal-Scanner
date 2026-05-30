import "dotenv/config";
import { PrismaClient } from "@prisma/client";
import { PrismaLibSql } from "@prisma/adapter-libsql";

function readOptionalEnv(name: string): string | undefined {
  const value = process.env[name];
  if (value === undefined || value.trim() === "") return undefined;
  return value;
}

export function createDbClient() {
  const url =
    readOptionalEnv("TURSO_DATABASE_URL") ??
    readOptionalEnv("DATABASE_URL") ??
    "file:./prisma/dev.db";
  const authToken = readOptionalEnv("TURSO_AUTH_TOKEN");
  const adapter = new PrismaLibSql({
    url,
    ...(authToken ? { authToken } : {}),
  });
  return new PrismaClient({ adapter });
}

const db = createDbClient();

export { db };

export enum SCANNER {
  CASH_CONVERTERS = 0,
  EBAY = 1,
  GUMTREE = 2,
  SALVOS = 3,
  STEAM_MARKET_QUERY = 4,
  STEAM_MARKET_CS = 5,
  CS_TRADE_BOT = 6,
}
