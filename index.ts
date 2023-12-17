import { Client, GatewayIntentBits, REST, Routes, Events } from "discord.js";
import Dotenv from "dotenv";
import mongoose from "mongoose";
import GlobalsQuery, { globalsInterface } from "./schema/globals.js";
import runScan from "./scanners/index.js";
import { commandHandler, commandList } from "./commandManager/index.js";

//Main function
async function run() {
  Dotenv.config();
  mongoose.connect(`${process.env.MONGO_URI}`);
  //Creates the bot's /commands
  const commands = [...commandList].map((command) => command.toJSON());
  const rest = new REST({ version: "9" }).setToken(
    `${process.env.DISCORD_TOKEN}`
  );
  rest
    .put(
      Routes.applicationGuildCommands(
        `${process.env.BOT_CLIENT_ID}`,
        `${process.env.DISCORD_GUILD_ID}`
      ),
      { body: commands }
    )
    .then(() => console.log("Registered the bot's commands successfully"))
    .catch(console.error);

  //TODO: Consider revisiting type assertion
  const globals =
    (await GlobalsQuery.findOne()) as mongoose.Document<globalsInterface>;

  //Discord Client Setup
  const client = new Client({ intents: [GatewayIntentBits.Guilds] });

  client.once("ready", (client: Client) => {
    runScan(client, globals);
  });

  //Runs upon a user creating a command
  client.on(Events.InteractionCreate, async (interaction) => {
    await commandHandler(interaction, client);
  });

  client.login(process.env.DISCORD_TOKEN);
}

run();
