import type { NotificationService } from "../notifications/types.js";

export default function handleError(
  error: Error,
  name: string,
  notificationService?: NotificationService,
) {
  console.error("An error has occurred in " + name + ":");
  console.error(error);

  if (notificationService) {
    void notificationService.publish({
      kind: "error",
      source: name,
      message: error.message,
      stack: error.stack,
    });
  }
}
