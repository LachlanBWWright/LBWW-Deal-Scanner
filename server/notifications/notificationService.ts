import type { AppNotification } from "../deals/types.js";
import { resultAsync } from "../functions/neverthrowUtils.js";
import type { NotificationProvider, NotificationService } from "./types.js";

export class DefaultNotificationService implements NotificationService {
  private providers: NotificationProvider[];

  constructor(providers: NotificationProvider[]) {
    this.providers = providers;
  }

  getProviderNames(): string[] {
    return this.providers.map((provider) => provider.name);
  }

  async publish(notification: AppNotification): Promise<void> {
    const enabled = this.providers.filter((p) => p.isEnabled());

    await Promise.all(
      enabled.map(async (provider) => {
        const result = await resultAsync(() =>
          provider.send(notification),
        );

        if (result.isErr()) {
          console.error(
            `Notification provider "${provider.name}" failed:`,
            result.error,
          );
        }
      }),
    );
  }
}
