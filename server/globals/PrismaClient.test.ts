import { describe, expect, it } from "vitest";
import { createDbClient, db, SCANNER } from "./PrismaClient.js";

describe("PrismaClient.ts", () => {
  it("exports prisma db client", () => {
    expect(typeof db).toBe("object");
    expect(typeof db.$connect).toBe("function");
    expect(typeof createDbClient).toBe("function");
  });

  it("exports scanner model enum", () => {
    expect(typeof SCANNER).toBe("object");
    expect(SCANNER.EBAY).toBe(1);
    expect(SCANNER.CS_TRADE_BOT).toBe(6);
  });
});
