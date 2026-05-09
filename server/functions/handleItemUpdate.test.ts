import { describe, expect, it } from "vitest";
import { db, SCANNER } from "../globals/PrismaClient.js";
import { checkIfNew } from "./handleItemUpdate.js";

describe("handleItemUpdate.ts", () => {
  it("creates new TTL records and marks duplicate checks as not-new", async () => {
    const itemId = `ttl-item-${Date.now()}`;

    const firstResult = await checkIfNew(itemId, SCANNER.EBAY);
    const firstRecord = await db.ttlItem.findUnique({
      where: { itemId, scanner: SCANNER.EBAY },
    });

    expect(firstResult).toBe(true);
    expect(firstRecord).not.toBeNull();

    const beforeUpdate = firstRecord?.lastUpdated;
    const secondResult = await checkIfNew(itemId, SCANNER.EBAY);
    const secondRecord = await db.ttlItem.findUnique({
      where: { itemId, scanner: SCANNER.EBAY },
    });

    expect(secondResult).toBe(false);
    expect(secondRecord).not.toBeNull();
    expect((secondRecord?.lastUpdated.getTime() ?? 0) >= (beforeUpdate?.getTime() ?? 0)).toBe(true);

    await db.ttlItem.deleteMany({ where: { itemId, scanner: SCANNER.EBAY } });
  });
});
