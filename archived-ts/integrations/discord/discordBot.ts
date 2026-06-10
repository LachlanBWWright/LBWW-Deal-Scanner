import { Events, REST, Routes } from "discord.js";

import client from "../../globals/DiscordJSClient.js";
import globals from "../../globals/Globals.js";
import { buttonInteractionHandler } from "../../commandManager/buttonHandler.js";
import { commandHandler, commandList } from "../../commandManager/index.js";
import { setBotStatus, setCommandNames } from "../../controlState.js";

export async function startDiscordBot() {
  if (
    !globals.DISCORD_TOKEN ||
    !globals.BOT_CLIENT_ID ||
    !globals.DISCORD_GUILD_ID
  ) {
    setBotStatus("disabled", "Missing Discord configuration");
    console.warn("Missing Discord configuration, skipping Discord startup.");
    return false;
  }

  const commands = [...commandList].map((command) => command.toJSON());
  setCommandNames(commands.map((command) => command.name));

  const rest = new REST({ version: "9" }).setToken(`${globals.DISCORD_TOKEN}`);
  await rest.put(
    Routes.applicationGuildCommands(
      `${globals.BOT_CLIENT_ID}`,
      `${globals.DISCORD_GUILD_ID}`,
    ),
    { body: commands },
  );
  console.log("Registered the bot's commands successfully");

  client.once("clientReady", () => {
    setBotStatus("ready");
  });

  client.on(Events.InteractionCreate, (interaction) => {
    if (interaction.isChatInputCommand()) {
      void commandHandler(interaction);
    } else if (interaction.isButton()) {
      void buttonInteractionHandler(interaction);
    }
  });

  client.on("error", (error) => {
    setBotStatus("error", error.message);
  });

  await client.login(globals.DISCORD_TOKEN);
  return true;
}
