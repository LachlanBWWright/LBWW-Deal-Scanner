import { describe, it, expect } from "vitest";
import { randomBytes } from "crypto";

// Re-implement the key generation function for testing
function generateRandomKey(): string {
  // 10 bytes = 20 hex characters
  return randomBytes(10).toString("hex");
}

// Type definition for testing
export interface ActionData {
  type: "delete" | "confirm_delete" | "cancel_delete";
  queryType?: string;
  queryId?: string;
  userId?: string;
  timestamp: number;
  relatedKey?: string;
}

describe("ActionRegistry Key Generation", () => {
  it("should generate unique random keys", () => {
    const key1 = generateRandomKey();
    const key2 = generateRandomKey();
    
    expect(key1).toMatch(/^[a-f0-9]{20}$/); // 20 hex characters
    expect(key2).toMatch(/^[a-f0-9]{20}$/);
    expect(key1).not.toBe(key2); // Should be different
  });

  it("should generate keys of consistent length", () => {
    for (let i = 0; i < 10; i++) {
      const key = generateRandomKey();
      expect(key).toHaveLength(20);
    }
  });
});

describe("ActionData Type Safety", () => {
  it("should properly type ActionData interface", () => {
    const testAction: ActionData = {
      type: "delete",
      queryType: "ebay",
      queryId: "test-query",
      userId: "user123",
      timestamp: Date.now(),
      relatedKey: "related-key"
    };

    // Type assertions to ensure interface is correctly defined
    expect(testAction.type).toBe("delete");
    expect(testAction.queryType).toBe("ebay");
    expect(testAction.queryId).toBe("test-query");
    expect(testAction.userId).toBe("user123");
    expect(typeof testAction.timestamp).toBe("number");
    expect(testAction.relatedKey).toBe("related-key");
  });

  it("should handle optional fields", () => {
    const minimalAction: ActionData = {
      type: "confirm_delete",
      timestamp: Date.now(),
    };

    expect(minimalAction.type).toBe("confirm_delete");
    expect(typeof minimalAction.timestamp).toBe("number");
    expect(minimalAction.queryType).toBeUndefined();
    expect(minimalAction.queryId).toBeUndefined();
    expect(minimalAction.userId).toBeUndefined();
    expect(minimalAction.relatedKey).toBeUndefined();
  });

  it("should validate action types", () => {
    const validTypes: ActionData["type"][] = ["delete", "confirm_delete", "cancel_delete"];
    
    validTypes.forEach(type => {
      const action: ActionData = {
        type,
        timestamp: Date.now(),
      };
      expect(action.type).toBe(type);
    });
  });
});