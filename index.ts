import { REST, Routes, Events } from "discord.js";
import globals, { initGlobals } from "./globals/Globals.js";
import runScan from "./scanners/index.js";
import { commandHandler, commandList } from "./commandManager/index.js";
import { buttonInteractionHandler } from "./commandManager/buttonHandler.js";
import client from "./globals/DiscordJSClient.js";

//Main function
async function run() {
  await initGlobals(); //Creates the bot's /commands
  const commands = [...commandList].map((command) => command.toJSON());
  const rest = new REST({ version: "9" }).setToken(`${globals.DISCORD_TOKEN}`);
  await rest.put(
    Routes.applicationGuildCommands(
      `${globals.BOT_CLIENT_ID}`,
      `${globals.DISCORD_GUILD_ID}`,
    ),
    { body: commands },
  );
  console.log("Registered the bot's commands successfully");

  client.once("ready", () => {
    void runScan();
  });

  //Runs upon a user creating a command
  client.on(Events.InteractionCreate, (interaction) => {
    if (interaction.isChatInputCommand()) {
      void commandHandler(interaction);
    } else if (interaction.isButton()) {
      void buttonInteractionHandler(interaction);
    }
  });

  //Starts DiscordJS server
  await client.login(globals.DISCORD_TOKEN);
}

run().catch((error: unknown) => {
  console.error(error);
});
