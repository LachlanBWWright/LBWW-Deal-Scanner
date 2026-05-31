import { describe, expect, it } from "vitest";
import { db, SCANNER } from "../globals/PrismaClient.js";
import { checkIfNewCsItem, CsSite, getAllTradeBotItems } from "./csTradeBot.js";

describe("csTradeBot.ts", () => {
  it("returns current csTradeBot queries from the database", async () => {
    const result = await getAllTradeBotItems();
    expect(Array.isArray(result)).toBe(true);
  });

  it("returns false for existing CS item signatures", async () => {
    const itemName = `cs-item-${Date.now()}`;
    const floatValue = 0.1337;
    const itemId = `${itemName}_${floatValue}_${CsSite.CS_TRADE}`;

    await db.ttlItem.create({
      data: {
        itemId,
        scanner: SCANNER.CS_TRADE_BOT,
        lastUpdated: new Date(),
      },
    });

    const result = await checkIfNewCsItem(
      itemName,
      floatValue,
      CsSite.CS_TRADE,
    );
    expect(result).toBe(false);

    await db.ttlItem.deleteMany({
      where: {
        itemId,
        scanner: SCANNER.CS_TRADE_BOT,
      },
    });
  });

  it("creates new CS item signatures and suppresses the second check", async () => {
    const itemName = `new-cs-item-${Date.now()}`;
    const floatValue = 0.2442;
    const itemId = `${itemName}_${floatValue}_${CsSite.TRADEIT_GG}`;

    const firstResult = await checkIfNewCsItem(
      itemName,
      floatValue,
      CsSite.TRADEIT_GG,
    );
    const secondResult = await checkIfNewCsItem(
      itemName,
      floatValue,
      CsSite.TRADEIT_GG,
    );
    const stored = await db.ttlItem.findUnique({
      where: {
        itemId,
        scanner: SCANNER.CS_TRADE_BOT,
      },
    });

    expect(firstResult).toBe(true);
    expect(secondResult).toBe(false);
    expect(stored).not.toBeNull();

    await db.ttlItem.deleteMany({
      where: {
        itemId,
        scanner: SCANNER.CS_TRADE_BOT,
      },
    });
  });
});
