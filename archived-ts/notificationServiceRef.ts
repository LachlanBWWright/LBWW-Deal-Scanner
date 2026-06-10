import type { NotificationService } from "./notifications/types.js";

let service: NotificationService | null = null;

export function setSharedNotificationService(s: NotificationService) {
  service = s;
}

export function getNotificationService(): NotificationService | null {
  return service;
}
