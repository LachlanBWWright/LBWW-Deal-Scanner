import { REST, Routes, Events } from "discord.js";
import globals, { initGlobals, MONGO_URI } from "./globals/Globals.js";
import runScan from "./scanners/index.js";
import { commandHandler, commandList } from "./commandManager/index.js";
import client from "./globals/DiscordJSClient.js";

//Main function
async function run() {
  await initGlobals(); //Creates the bot's /commands
  const commands = [...commandList].map((command) => command.toJSON());
  const rest = new REST({ version: "9" }).setToken(`${globals.DISCORD_TOKEN}`);
  rest
    .put(
      Routes.applicationGuildCommands(
        `${globals.BOT_CLIENT_ID}`,
        `${globals.DISCORD_GUILD_ID}`,
      ),
      { body: commands },
    )
    .then(() => console.log("Registered the bot's commands successfully"))
    .catch(console.error);

  client.once("ready", runScan);

  //Runs upon a user creating a command
  client.on(Events.InteractionCreate, commandHandler);

  //Starts DiscordJS server
  client.login(globals.DISCORD_TOKEN);
}

run();
