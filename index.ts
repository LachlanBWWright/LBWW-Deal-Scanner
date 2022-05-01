import {Client, Intents} from "discord.js";
import {SlashCommandBuilder} from "@discordjs/builders";
import {REST} from "@discordjs/rest";
import {Routes} from "discord-api-types/v9";
import Dotenv from "dotenv";
import mongoose from "mongoose";

//Schema Imports

//Scanner Imports
import Ps5BigW from "./scanners/ps5BigW.js";
import Ps5Target from "./scanners/ps5Target.js";
import XboxBigW from "./scanners/xboxBigW.js";
import CsDeals from "./scanners/csDeals.js";

Dotenv.config();
mongoose.connect(`${process.env.MONGO_URI}`);

//Creates the bot's /commands
const commands = [
    new SlashCommandBuilder()
        .setName("createcssearch")
        .setDescription("Creates a search for a CS:GO item")
        .addStringOption(option =>
            option.setName("website")
                .setDescription("Which website you want to create a search for?")
                .setRequired(true)
                .setChoices(
                    {
                        name: "Cs.Deals", //Just copy and paste this to add more choices
                        value: "csdeals"
                    }
                )
        )
        .addStringOption(option =>
            option.setName("skinname")
                .setDescription("Copy and paste the EXACT name of the skin as it is shown from the website.")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("maximumprice")
                .setDescription("Enter the maximum acceptable price for the skin, according to the website\'s default currency.")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("maxfloat")
                .setDescription("Enter the maximum acceptable float value for the skin.")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("minfloat")
                .setDescription("Enter the minimum acceptable float value for the skin.")
                .setRequired(true)
        )
].map(command => command.toJSON());
const rest = new REST({version: "9"}).setToken(`${process.env.DISCORD_TOKEN}`);
rest.put(Routes.applicationGuildCommands(`${process.env.BOT_CLIENT_ID}`, `${process.env.DISCORD_GUILD_ID}`), {body: commands})
    .then(() => console.log("Registered the bot\'s commands successfully"))
    .catch(console.error);

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
    if(process.env.CS_ITEMS && process.env.CS_CHANNEL_ID && process.env.CS_ROLE_ID) {
        const csDeals = new CsDeals(client, process.env.CS_CHANNEL_ID, process.env.CS_ROLE_ID);
        csDeals.scan();
        setInterval(csDeals.scan, 900000);
    }
});

client.on("interactionCreate", async interaction => {
    try {
        if(!interaction.isCommand()) return; //Cancels if not a command
    }
    catch(error) {
        console.log(error);
    }
});

client.login(process.env.DISCORD_TOKEN);