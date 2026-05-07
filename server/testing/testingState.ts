import type { TestingNotificationResult, TestingScanResult } from "./types.js";

const MAX_ENTRIES = 100;

const notificationHistory: TestingNotificationResult[] = [];
const scanHistory: TestingScanResult[] = [];

export function recordNotificationResult(result: TestingNotificationResult) {
  notificationHistory.unshift(result);
  if (notificationHistory.length > MAX_ENTRIES) {
    notificationHistory.splice(MAX_ENTRIES);
  }
}

export function recordScanResult(result: TestingScanResult) {
  scanHistory.unshift(result);
  if (scanHistory.length > MAX_ENTRIES) {
    scanHistory.splice(MAX_ENTRIES);
  }
}

export function getNotificationHistory(): readonly TestingNotificationResult[] {
  return notificationHistory;
}

export function getScanHistory(): readonly TestingScanResult[] {
  return scanHistory;
}
