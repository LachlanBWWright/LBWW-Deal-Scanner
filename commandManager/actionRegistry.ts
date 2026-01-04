import { randomBytes } from "crypto";
import { db } from "../globals/PrismaClient.js";

export type ActionData = {
  type: "delete" | "confirm_delete" | "cancel_delete";
  queryType?: string;
  queryId?: string;
  userId?: string;
  timestamp: number;
  relatedKey?: string;
};

// Database-backed action registry implementation
export const actionRegistry = {
  async get(key: string): Promise<ActionData | undefined> {
    const action = await db.actionRegistry.findUnique({
      where: { id: key }
    });
    
    if (!action) return undefined;
    
    return {
      type: action.type as ActionData["type"],
      queryType: action.queryType || undefined,
      queryId: action.queryId || undefined,
      userId: action.userId || undefined,
      timestamp: Number(action.timestamp),
      relatedKey: action.relatedKey || undefined,
    };
  },

  async set(key: string, data: ActionData): Promise<void> {
    await db.actionRegistry.upsert({
      where: { id: key },
      update: {
        type: data.type,
        queryType: data.queryType,
        queryId: data.queryId,
        userId: data.userId,
        timestamp: BigInt(data.timestamp),
        relatedKey: data.relatedKey,
      },
      create: {
        id: key,
        type: data.type,
        queryType: data.queryType,
        queryId: data.queryId,
        userId: data.userId,
        timestamp: BigInt(data.timestamp),
        relatedKey: data.relatedKey,
      },
    });
  },

  async delete(key: string): Promise<boolean> {
    try {
      await db.actionRegistry.delete({
        where: { id: key }
      });
      return true;
    } catch (error) {
      if ((error as { code?: string })?.code === 'P2025') {
        // Record not found
        return false;
      }
      throw error;
    }
  },

  async entries(): Promise<[string, ActionData][]> {
    const actions = await db.actionRegistry.findMany();
    return actions.map(action => [
      action.id,
      {
        type: action.type as ActionData["type"],
        queryType: action.queryType || undefined,
        queryId: action.queryId || undefined,
        userId: action.userId || undefined,
        timestamp: Number(action.timestamp),
        relatedKey: action.relatedKey || undefined,
      }
    ]);
  },

  async cleanupExpired(maxAge: number = 10 * 60 * 1000): Promise<void> {
    const cutoffTime = BigInt(Date.now() - maxAge);
    await db.actionRegistry.deleteMany({
      where: {
        timestamp: {
          lt: cutoffTime
        }
      }
    });
  }
};

export function generateRandomKey(): string {
  // 10 bytes = 20 hex characters
  return randomBytes(10).toString("hex");
}
