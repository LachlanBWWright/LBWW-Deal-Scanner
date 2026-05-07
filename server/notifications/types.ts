import type { AppNotification } from "../deals/types.js";

export interface NotificationProvider {
  name: string;
  isEnabled(): boolean;
  send(notification: AppNotification): Promise<void>;
}

export interface NotificationService {
  publish(notification: AppNotification): Promise<void>;
}
