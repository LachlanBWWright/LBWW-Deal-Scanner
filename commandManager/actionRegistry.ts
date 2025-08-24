import { randomBytes } from "crypto";

export type ActionData = {
  type: "delete" | "confirm_delete" | "cancel_delete";
  queryType?: string;
  queryId?: string;
  userId?: string;
  timestamp: number;
  relatedKey?: string;
};

export const actionRegistry = new Map<string, ActionData>();

export function generateRandomKey(): string {
  // 10 bytes = 20 hex characters
  return randomBytes(10).toString("hex");
}
