import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { actionRegistry, generateRandomKey, ActionData } from "./actionRegistry";
import { db } from "../globals/PrismaClient";

describe("Database-backed ActionRegistry", () => {
  const testActionData: ActionData = {
    type: "delete",
    queryType: "ebay",
    queryId: "test-query-id",
    userId: "test-user-123",
    timestamp: Date.now(),
  };

  beforeEach(async () => {
    // Clean up any existing test data
    await db.actionRegistry.deleteMany({
      where: {
        queryId: testActionData.queryId,
      },
    });
  });

  afterEach(async () => {
    // Clean up after each test
    await db.actionRegistry.deleteMany({
      where: {
        queryId: testActionData.queryId,
      },
    });
  });

  it("should generate random keys", () => {
    const key1 = generateRandomKey();
    const key2 = generateRandomKey();
    
    expect(key1).toMatch(/^[a-f0-9]{20}$/); // 20 hex characters
    expect(key2).toMatch(/^[a-f0-9]{20}$/);
    expect(key1).not.toBe(key2); // Should be different
  });

  it("should store and retrieve action data", async () => {
    const key = generateRandomKey();
    
    await actionRegistry.set(key, testActionData);
    const retrieved = await actionRegistry.get(key);
    
    expect(retrieved).toBeDefined();
    expect(retrieved?.type).toBe(testActionData.type);
    expect(retrieved?.queryType).toBe(testActionData.queryType);
    expect(retrieved?.queryId).toBe(testActionData.queryId);
    expect(retrieved?.userId).toBe(testActionData.userId);
    expect(retrieved?.timestamp).toBe(testActionData.timestamp);
  });

  it("should return undefined for non-existent keys", async () => {
    const nonExistentKey = generateRandomKey();
    const result = await actionRegistry.get(nonExistentKey);
    
    expect(result).toBeUndefined();
  });

  it("should delete action data", async () => {
    const key = generateRandomKey();
    
    await actionRegistry.set(key, testActionData);
    expect(await actionRegistry.get(key)).toBeDefined();
    
    const deleted = await actionRegistry.delete(key);
    expect(deleted).toBe(true);
    
    expect(await actionRegistry.get(key)).toBeUndefined();
  });

  it("should return false when deleting non-existent key", async () => {
    const nonExistentKey = generateRandomKey();
    const deleted = await actionRegistry.delete(nonExistentKey);
    
    expect(deleted).toBe(false);
  });

  it("should handle action data with related keys", async () => {
    const key = generateRandomKey();
    const relatedKey = generateRandomKey();
    
    const actionWithRelatedKey: ActionData = {
      ...testActionData,
      relatedKey,
    };
    
    await actionRegistry.set(key, actionWithRelatedKey);
    const retrieved = await actionRegistry.get(key);
    
    expect(retrieved?.relatedKey).toBe(relatedKey);
  });

  it("should clean up expired actions", async () => {
    const key1 = generateRandomKey();
    const key2 = generateRandomKey();
    
    // Create one old action and one recent action
    const oldTimestamp = Date.now() - (15 * 60 * 1000); // 15 minutes ago
    const recentTimestamp = Date.now();
    
    await actionRegistry.set(key1, { ...testActionData, timestamp: oldTimestamp });
    await actionRegistry.set(key2, { ...testActionData, timestamp: recentTimestamp, queryId: 'recent-query' });
    
    // Clean up actions older than 10 minutes
    await actionRegistry.cleanupExpired(10 * 60 * 1000);
    
    // Old action should be deleted, recent action should remain
    expect(await actionRegistry.get(key1)).toBeUndefined();
    expect(await actionRegistry.get(key2)).toBeDefined();
    
    // Clean up the remaining action
    await actionRegistry.delete(key2);
  });

  it("should retrieve all entries", async () => {
    const key1 = generateRandomKey();
    const key2 = generateRandomKey();
    
    const action1: ActionData = { ...testActionData, queryId: 'query1' };
    const action2: ActionData = { ...testActionData, queryId: 'query2' };
    
    await actionRegistry.set(key1, action1);
    await actionRegistry.set(key2, action2);
    
    const entries = await actionRegistry.entries();
    const entryMap = new Map(entries);
    
    expect(entryMap.has(key1)).toBe(true);
    expect(entryMap.has(key2)).toBe(true);
    expect(entryMap.get(key1)?.queryId).toBe('query1');
    expect(entryMap.get(key2)?.queryId).toBe('query2');
    
    // Clean up
    await actionRegistry.delete(key1);
    await actionRegistry.delete(key2);
  });
});