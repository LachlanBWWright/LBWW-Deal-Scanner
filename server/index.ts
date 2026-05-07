import { startApiServer } from "./api.js";
import { startDiscordBot } from "./integrations/discord/discordBot.js";
import { startBackgroundScanLoop, setNotificationService } from "./scannerRuntime.js";
import { DefaultNotificationService } from "./notifications/notificationService.js";
import { DiscordNotificationProvider } from "./integrations/discord/discordNotifier.js";
import { DesktopNotificationProvider } from "./integrations/desktop/desktopNotifier.js";
import { initGlobals } from "./globals/Globals.js";

async function run() {
  await initGlobals();

  // Construct notification providers
  const providers = [
    new DiscordNotificationProvider(),
    new DesktopNotificationProvider(),
  ];

  const notificationService = new DefaultNotificationService(providers);
  setNotificationService(notificationService);

  // Start API server
  const api = await startApiServer();
  console.log(`Fastify API listening on http://${api.host}:${api.port}`);

  // Start background scan loop (independent of Discord)
  void startBackgroundScanLoop();

  // Start Discord bot (independent of scan loop)
  void startDiscordBot().catch((error: unknown) => {
    console.error("Discord bot startup failed:", error);
  });
}

run().catch((error: unknown) => {
  console.error(error);
});

