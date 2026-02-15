import "dotenv/config";
import { PrismaClient } from "@prisma/client";
import { PrismaLibSQL } from "@prisma/adapter-libsql";
import { createClient } from "@libsql/client";

let db: PrismaClient;

try {
  // Use plain Prisma client for tests and local development to avoid adapter runtime issues
  if (
    process.env.NODE_ENV === "test" ||
    process.env.NODE_ENV === "development"
  ) {
    db = new PrismaClient();
  } else if (!process.env.TURSO_DATABASE_URL || !process.env.TURSO_AUTH_TOKEN) {
    // No Turso credentials — fall back to local SQLite via the standard Prisma client
    db = new PrismaClient();
  } else {
    // Use Turso (libsql adapter) for production when credentials are present
    const libsql = createClient({
      url: `${process.env.TURSO_DATABASE_URL}`,
      authToken: `${process.env.TURSO_AUTH_TOKEN}`,
    });
    const adapter = new PrismaLibSQL(libsql);
    db = new PrismaClient({ adapter });
  }
} catch (error) {
  console.warn(
    "Failed to initialize database connection, using fallback Prisma client:",
    error,
  );
  // Fallback to basic Prisma client
  db = new PrismaClient();
}

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
