import {Client, Intents} from "discord.js";
import Dotenv from "dotenv";
import {MongoClient} from "mongodb";
//const dbClient = new MongoClient(process.env.MONGO_URI); TODO: Re-enable
Dotenv.config();

//Scanner Imports
import Ps5BigW from "./scanners/ps5BigW";
import Ps5Target from "./scanners/ps5Target";
import XboxBigW from "./scanners/xboxBigW";
import CsDeals from "./scanners/csDeals";

const csDeals = new CsDeals();
csDeals.scan();

//Discord Client Setup
const client = new Client({ intents: [Intents.FLAGS.GUILDS]});
client.once('ready', () => {
    console.log("Discord client is ready.")

    //Actives scanners as specified in .env
    if(process.env.PS5BIGW && process.env.PS5_CHANNEL_ID && process.env.PS5_ROLE_ID) {
        const ps5BigW = new Ps5BigW(client, process.env.PS5_CHANNEL_ID, process.env.PS5_ROLE_ID);
        setInterval(ps5BigW.scan, 10000);
    }
    if(process.env.XBOXBIGW && process.env.XBOX_CHANNEL_ID && process.env.XBOX_ROLE_ID) {
        const xboxBigW = new XboxBigW(client, process.env.XBOX_CHANNEL_ID, process.env.XBOX_ROLE_ID);
        setInterval(xboxBigW.scan, 10000);
    }
    if(process.env.PS5TARGET && process.env.PS5_CHANNEL_ID && process.env.PS5_ROLE_ID) {
        const ps5Target = new Ps5Target(client, process.env.PS5_CHANNEL_ID, process.env.PS5_ROLE_ID);
        setInterval(ps5Target.scan, 10000);
    }
});

client.on("interactionCreate", interaction => {

});

client.login(process.env.DISCORD_TOKEN);