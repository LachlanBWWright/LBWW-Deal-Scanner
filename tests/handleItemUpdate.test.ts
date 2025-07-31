import { describe, it, expect, vi } from "vitest";

// Mock the database (define the mock inline to avoid hoisting issues)
vi.mock("../globals/PrismaClient.js", () => ({
  db: {
    ttlItem: {
      findUnique: vi.fn(),
      update: vi.fn(),
      create: vi.fn(),
    },
  },
  SCANNER: {
    CASH_CONVERTERS: 0,
    EBAY: 1,
    GUMTREE: 2,
    SALVOS: 3,
    STEAM_MARKET_QUERY: 4,
    STEAM_MARKET_CS: 5,
    CS_TRADE_BOT: 6,
  },
}));

import { checkIfNew } from "../functions/handleItemUpdate.js";
import { db, SCANNER } from "../globals/PrismaClient.js";

describe("handleItemUpdate", () => {
  describe("checkIfNew", () => {
    it("should return true for new item and create database entry", async () => {
      // Mock item not found (new item)
      vi.mocked(db.ttlItem.findUnique).mockResolvedValue(null);
      vi.mocked(db.ttlItem.create).mockResolvedValue({} as any);

      const result = await checkIfNew("item-123", SCANNER.EBAY);

      expect(result).toBe(true);
      expect(db.ttlItem.findUnique).toHaveBeenCalledWith({
        where: { itemId: "item-123", scanner: SCANNER.EBAY }
      });
      expect(db.ttlItem.create).toHaveBeenCalledWith({
        data: { 
          itemId: "item-123", 
          scanner: SCANNER.EBAY, 
          lastUpdated: expect.any(Date) 
        }
      });
    });

    it("should return false for existing item and update lastUpdated", async () => {
      // Mock existing item found
      const existingItem = {
        itemId: "item-123",
        scanner: SCANNER.EBAY,
        lastUpdated: new Date(),
      };
      vi.mocked(db.ttlItem.findUnique).mockResolvedValue(existingItem as any);
      vi.mocked(db.ttlItem.update).mockResolvedValue({} as any);

      const result = await checkIfNew("item-123", SCANNER.EBAY);

      expect(result).toBe(false);
      expect(db.ttlItem.findUnique).toHaveBeenCalledWith({
        where: { itemId: "item-123", scanner: SCANNER.EBAY }
      });
      expect(db.ttlItem.update).toHaveBeenCalledWith({
        where: { itemId: "item-123", scanner: SCANNER.EBAY },
        data: {
          lastUpdated: expect.any(Date),
        },
      });
    });

    it("should handle different scanner types", async () => {
      vi.mocked(db.ttlItem.findUnique).mockResolvedValue(null);
      vi.mocked(db.ttlItem.create).mockResolvedValue({} as any);

      await checkIfNew("item-456", SCANNER.CASH_CONVERTERS);

      expect(db.ttlItem.findUnique).toHaveBeenCalledWith({
        where: { itemId: "item-456", scanner: SCANNER.CASH_CONVERTERS }
      });
      expect(db.ttlItem.create).toHaveBeenCalledWith({
        data: { 
          itemId: "item-456", 
          scanner: SCANNER.CASH_CONVERTERS, 
          lastUpdated: expect.any(Date) 
        }
      });
    });

    it("should handle different item IDs", async () => {
      vi.mocked(db.ttlItem.findUnique).mockResolvedValue(null);
      vi.mocked(db.ttlItem.create).mockResolvedValue({} as any);

      await checkIfNew("unique-item-id", SCANNER.GUMTREE);

      expect(db.ttlItem.findUnique).toHaveBeenCalledWith({
        where: { itemId: "unique-item-id", scanner: SCANNER.GUMTREE }
      });
    });
  });
});