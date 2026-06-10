import type { AppNotification } from "../../deals/types.js";
import type { NotificationProvider } from "../../notifications/types.js";

/**
 * A no-op notification provider that logs notifications to console.
 * Useful for local development or as a proof-of-extension template.
 * Enable by setting DESKTOP_NOTIFICATIONS=true in environment.
 */
export class DesktopNotificationProvider implements NotificationProvider {
  name = "desktop";

  isEnabled(): boolean {
    return !!process.env.DESKTOP_NOTIFICATIONS;
  }

  async send(notification: AppNotification): Promise<void> {
    if (notification.kind === "error") {
      console.log(
        `[Desktop Notification] ERROR from ${notification.source}: ${notification.message}`,
      );
    } else {
      console.log(
        `[Desktop Notification] DEAL from ${notification.source}: ${notification.title}`,
      );
    }
  }
}
