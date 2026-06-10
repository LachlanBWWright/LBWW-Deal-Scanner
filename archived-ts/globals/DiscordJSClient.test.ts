import { Client, GatewayIntentBits } from "discord.js";
import { describe, expect, it } from "vitest";
import client from "./DiscordJSClient.js";

describe("DiscordJSClient.ts", () => {
  it("creates a Discord client configured with Guild intent", () => {
    expect(client instanceof Client).toBe(true);
    expect(client.options.intents.has(GatewayIntentBits.Guilds)).toBe(true);
  });
});
