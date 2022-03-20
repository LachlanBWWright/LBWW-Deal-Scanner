//Imports
import {Client, Intents} from "discord.js";
import Dotenv from "dotenv";
Dotenv.config();

//Scanner Imports
import Ps5BigW from "./scanners/ps5BigW";

//Discord Client Setup
const client = new Client({ intents: [Intents.FLAGS.GUILDS]});
client.once('ready', () => {
    console.log("Discord client is ready.")

    //Actives scanners as specified in .env
    if(process.env.PS5BIGW && process.env.PS5BIGW_CHANNEL_ID) {
        const ps5BigW = new Ps5BigW(client, process.env.PS5BIGW_CHANNEL_ID);
        setInterval(ps5BigW.scan, 10000);
    }
});

client.on("interactionCreate", interaction => {

});

client.login(process.env.DISCORD_TOKEN);