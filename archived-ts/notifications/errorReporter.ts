import type { ErrorNotification } from "../deals/types.js";
import type { NotificationService } from "./types.js";

export function reportError(
  service: NotificationService,
  error: Error,
  source: string,
): void {
  console.error(`An error has occurred in ${source}:`);
  console.error(error);

  const notification: ErrorNotification = {
    kind: "error",
    source,
    message: error.message,
    stack: error.stack,
  };

  void service.publish(notification);
}
