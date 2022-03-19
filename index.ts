import {Client, Intents} from "discord.js";
import Dotenv from "dotenv";
Dotenv.config();

const client = new Client({ intents: [Intents.FLAGS.GUILDS]});
client.once('ready', () => console.log('Discord client is ready.'));

client.login(process.env.DISCORD_TOKEN);