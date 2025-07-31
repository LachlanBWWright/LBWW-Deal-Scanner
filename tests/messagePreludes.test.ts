import { describe, it, expect, vi } from "vitest";
import { 
  getNotificationPrelude, 
  getResponsePrelude, 
  getFailurePrelude 
} from "../functions/messagePreludes.js";

describe("messagePreludes", () => {
  describe("getNotificationPrelude", () => {
    it("should return a string", () => {
      const result = getNotificationPrelude();
      expect(typeof result).toBe("string");
      expect(result.length).toBeGreaterThan(0);
    });

    it("should return one of the expected standard preludes most of the time", () => {
      const standardPreludes = [
        "Listen up, Jack,",
        "My fellow Americans,", 
        "Folks,",
        "Here's the deal,",
      ];
      
      // Test multiple times to increase chance of getting standard prelude
      const results = Array.from({ length: 50 }, () => getNotificationPrelude());
      const hasStandardPrelude = results.some(result => 
        standardPreludes.includes(result)
      );
      
      expect(hasStandardPrelude).toBe(true);
    });

    it("should occasionally return rare prelude", () => {
      // Mock Math.random to return value that triggers rare prelude
      const mockRandom = vi.spyOn(Math, "random").mockReturnValue(0.005);
      
      const result = getNotificationPrelude();
      expect(result).toBe("This is a big fu- ...uh... flippable deal,");
      
      mockRandom.mockRestore();
    });
  });

  describe("getResponsePrelude", () => {
    it("should return a string", () => {
      const result = getResponsePrelude();
      expect(typeof result).toBe("string");
      expect(result.length).toBeGreaterThan(0);
    });

    it("should return one of the expected standard preludes most of the time", () => {
      const standardPreludes = [
        "Listen up, Jack,",
        "My fellow Americans,",
        "Folks,", 
        "Here's the deal,",
      ];
      
      // Test multiple times to increase chance of getting standard prelude
      const results = Array.from({ length: 50 }, () => getResponsePrelude());
      const hasStandardPrelude = results.some(result => 
        standardPreludes.includes(result)
      );
      
      expect(hasStandardPrelude).toBe(true);
    });

    it("should occasionally return rare prelude", () => {
      // Mock Math.random to return value that triggers rare prelude
      const mockRandom = vi.spyOn(Math, "random").mockReturnValue(0.005);
      
      const result = getResponsePrelude();
      expect(result).toBe("Look, fat,");
      
      mockRandom.mockRestore();
    });
  });

  describe("getFailurePrelude", () => {
    it("should return a string", () => {
      const result = getFailurePrelude();
      expect(typeof result).toBe("string");
      expect(result.length).toBeGreaterThan(0);
    });

    it("should return one of the expected negative preludes most of the time", () => {
      const negativePreludes = [
        "Malarkey,",
        "This is a bunch of malarkey,", 
        "Uh-oh, Jack,",
      ];
      
      // Test multiple times to increase chance of getting standard prelude
      const results = Array.from({ length: 50 }, () => getFailurePrelude());
      const hasNegativePrelude = results.some(result => 
        negativePreludes.includes(result)
      );
      
      expect(hasNegativePrelude).toBe(true);
    });

    it("should occasionally return rare negative prelude", () => {
      // Mock Math.random to return value that triggers rare prelude
      const mockRandom = vi.spyOn(Math, "random").mockReturnValue(0.005);
      
      const result = getFailurePrelude();
      expect(result).toBe("Stupid son of a bitch,");
      
      mockRandom.mockRestore();
    });
  });
});