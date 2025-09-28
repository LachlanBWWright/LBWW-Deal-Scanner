import "dotenv/config";
import { PrismaClient } from "@prisma/client";
import { PrismaLibSQL } from "@prisma/adapter-libsql";
import { createClient } from "@libsql/client";

let libsql;
let adapter;

// Check if we're in a test environment or if environment variables are missing
if (process.env.NODE_ENV === 'test' || !process.env.TURSO_DATABASE_URL || !process.env.TURSO_AUTH_TOKEN) {
  // Use SQLite for testing or when env vars are missing
  libsql = createClient({
    url: "file:./prisma/test.db"
  });
} else {
  // Use Turso for production
  libsql = createClient({
    url: `${process.env.TURSO_DATABASE_URL}`,
    authToken: `${process.env.TURSO_AUTH_TOKEN}`,
  });
}

adapter = new PrismaLibSQL(libsql);
export const db = new PrismaClient({ adapter });

export enum SCANNER {
  CASH_CONVERTERS = 0,
  EBAY = 1,
  GUMTREE = 2,
  SALVOS = 3,
  STEAM_MARKET_QUERY = 4,
  STEAM_MARKET_CS = 5,
  CS_TRADE_BOT = 6,
}
