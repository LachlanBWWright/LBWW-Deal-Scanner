import { randomBytes } from "crypto";
import { err, ok } from "neverthrow";
import { db } from "../globals/PrismaClient.js";
import { resultAsync } from "../functions/neverthrowUtils.js";

type ActionType =
  | "delete"
  | "confirm_delete"
  | "cancel_delete"
  | "subscribe_dm"
  | "unsubscribe_dm";

export interface ActionData {
  type: ActionType;
  queryType?: string;
  queryId?: string;
  userId?: string;
  timestamp: number;
  relatedKey?: string;
}

// Database-backed action registry implementation
export const actionRegistry = {
  async get(key: string): Promise<ActionData | undefined> {
    const action = await db.actionRegistry.findUnique({
      where: { id: key },
    });

    if (!action) return undefined;

    const type = parseActionType(action.type).match(
      (value) => value,
      (error) => {
        console.error(error.message);
        return undefined;
      },
    );
    if (!type) return undefined;

    return {
      type,
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
    const deleteResult = await resultAsync(
      () =>
        db.actionRegistry.delete({
          where: { id: key },
        }),
      "Failed to delete action registry entry",
    );

    if (deleteResult.isErr()) {
      const error = deleteResult.error;
      if (isPrismaError(error) && error.code === "P2025") {
        return false;
      }
      console.error("Failed to delete action registry entry:", error.message);
      return false;
    }

    return true;
  },

  async entries(): Promise<[string, ActionData][]> {
    const actions = await db.actionRegistry.findMany();
    return actions.flatMap((action) => {
      const type = parseActionType(action.type).match(
        (value) => value,
        (error) => {
          console.error(error.message);
          return undefined;
        },
      );
      if (!type) return [];
      return [
        [
          action.id,
          {
            type,
            queryType: action.queryType || undefined,
            queryId: action.queryId || undefined,
            userId: action.userId || undefined,
            timestamp: Number(action.timestamp),
            relatedKey: action.relatedKey || undefined,
          },
        ],
      ];
    });
  },

  async cleanupExpired(maxAge: number = 10 * 60 * 1000): Promise<void> {
    const cutoffTime = BigInt(Date.now() - maxAge);
    await db.actionRegistry.deleteMany({
      where: {
        timestamp: {
          lt: cutoffTime,
        },
      },
    });
  },
};

export function generateRandomKey(): string {
  // 10 bytes = 20 hex characters
  return randomBytes(10).toString("hex");
}

function isPrismaError(error: unknown): error is { code: unknown } {
  return typeof error === "object" && error !== null && "code" in error;
}

function parseActionType(type: string) {
  if (
    type === "delete" ||
    type === "confirm_delete" ||
    type === "cancel_delete" ||
    type === "subscribe_dm" ||
    type === "unsubscribe_dm"
  ) {
    return ok(type);
  }
  return err(new Error(`Invalid action type in DB: ${type}`));
}
