import { startApiServer } from "./api.js";
import { startDiscordBot } from "./discordBot.js";

async function run() {
  const api = await startApiServer();
  console.log(`Fastify API listening on http://${api.host}:${api.port}`);

  void startDiscordBot().catch((error: unknown) => {
    console.error("Discord bot startup failed:", error);
  });
}

run().catch((error: unknown) => {
  console.error(error);
});
