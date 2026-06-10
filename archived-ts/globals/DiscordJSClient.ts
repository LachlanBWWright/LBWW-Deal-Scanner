//Discord Client Setup
import { GatewayIntentBits, Client } from "discord.js";

const client = new Client({ intents: [GatewayIntentBits.Guilds] });

export default client;
