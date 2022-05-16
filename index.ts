import {Client, Intents} from "discord.js";
import {SlashCommandBuilder} from "@discordjs/builders";
import {REST} from "@discordjs/rest";
import {Routes} from "discord-api-types/v9";
import Dotenv from "dotenv";
import mongoose from "mongoose";

//Schema Imports
import CsDealsItem from "./schema/csDealsItem.js";
import CsTradeItem from "./schema/csTradeItem.js";
import TradeItItem from "./schema/tradeItItem.js";
import LootFarmItem from "./schema/lootFarmItem.js";

//Scanner Imports
import Ps5BigW from "./scanners/ps5BigW.js";
import Ps5Target from "./scanners/ps5Target.js";
import XboxBigW from "./scanners/xboxBigW.js";
import CsDeals from "./scanners/csDeals.js";
import CsTrade from "./scanners/csTrade.js";
import TradeIt from "./scanners/tradeIt.js";
import LootFarm from "./scanners/lootFarm.js";

Dotenv.config();
mongoose.connect(`${process.env.MONGO_URI}`);

//Creates the bot's /commands
const commands = [
    new SlashCommandBuilder()
        .setName("createmultisearch")
        .setDescription("Creates a search for a CS:GO item on multiple trade bots")
        .addStringOption(option =>
            option.setName("skinname")
                .setDescription("Copy and paste from the SCM, E.G. \"StatTrak™ XM1014 | Seasons (Minimal Wear)\".")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("minfloat")
                .setDescription("Enter the minimum acceptable float value for the skin.")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("maxfloat")
                .setDescription("Enter the maximum acceptable float value for the skin.")
                .setRequired(true)
        )
        .addNumberOption(option => 
            option.setName("maxpricecsdeals")
                .setDescription("Enter the max price for the skin on cs.deals, no search will be created if it's $0.")
                .setRequired(false)
        )
        .addNumberOption(option => 
            option.setName("maxpricecstrade")
                .setDescription("Enter the max price for the skin on cs.trade, no search will be created if it's $0.")
                .setRequired(false)
        )
        .addNumberOption(option => 
            option.setName("maxpricetradeit")
                .setDescription("Enter the max price for the skin on tradeit.gg, no search will be created if it's $0.")
                .setRequired(false)
        )
        .addNumberOption(option => 
            option.setName("maxpricelootfarm")
                .setDescription("Enter the max price for the skin on loot.farm, no search will be created if it's $0.")
                .setRequired(false)
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
        setInterval(csDeals.scan, 900000); //TODO: Revert interval

        const csTrade = new CsTrade(client, process.env.CS_CHANNEL_ID, process.env.CS_ROLE_ID);
        csTrade.scan();
        setInterval(csTrade.scan, 900000);

        const tradeIt = new TradeIt(client, process.env.CS_CHANNEL_ID, process.env.CS_ROLE_ID);
        tradeIt.scan();
        setInterval(tradeIt.scan, 900000);

        const lootFarm = new LootFarm(client, process.env.CS_CHANNEL_ID, process.env.CS_ROLE_ID);
        lootFarm.scan();
    }
});

client.on("interactionCreate", async interaction => {
    try {
        if(!interaction.isCommand()) return; //Cancels if not a command

        if(interaction.commandName === "createcssearch") {
            await interaction.deferReply();
            let website = interaction.options.getString("website") || "csdeals";
            let skinName = interaction.options.getString("skinname") || "placeholder";
            let maxPrice = interaction.options.getNumber("maxprice") || -1;
            let maxFloat = interaction.options.getNumber("maxfloat") || -1;
            let minFloat = interaction.options.getNumber("minfloat") || 0;
            //console.log(`${website}` + `${skinName}` + maxPrice + maxFloat + minFloat);

            if(maxPrice <= 0 || maxPrice > 500) {
                interaction.editReply("I cannot stress enough: The price must be above $0, and no more than $500.");
            }
            else if(minFloat > maxFloat) {
                interaction.editReply("I cannot stress enough: The minimum float cannot be higher than the maximum float value.");
            }
            else if(maxFloat <= 0 || maxFloat >= 1) {
                interaction.editReply("I cannot stress enough: The maximum float must be between 0 and 1.");
            }
            else if(minFloat < 0 || minFloat >= 1) {
                interaction.editReply("I cannot stress enough: The minimum float must be between 0 and 1.");
            }
            else if(website == "csdeals") {
                const csDeals = new CsDeals(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                if(await csDeals.skinExists(skinName)) {
                    const csDealsItem = new CsDealsItem({
                        name: skinName,
                        maxPrice: maxPrice,
                        minFloat: minFloat,
                        maxFloat: maxFloat
                    })
                    csDealsItem.save(err => console.log(err));
                    console.log(csDealsItem);

                    interaction.editReply("Please know that the skin has been added to the watchlist. Hopefully.");
                }
                else {
                    interaction.editReply("The skin was not added to the database, as it could not be found on the site. Try finding it on the site, right click it, and copy and paste. I cannot stress enough: It should look something like: ★ Moto Gloves | Cool Mint (Minimal Wear)");    
                }
            }
            else if(website == "cstrade") {
                const csTrade = new CsTrade(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                if(await csTrade.skinExists(skinName)) {
                    const csTradeItem = new CsTradeItem({
                        name: skinName,
                        maxPrice: maxPrice,
                        minFloat: minFloat,
                        maxFloat: maxFloat
                    })
                    csTradeItem.save(err => console.log(err));
                    console.log(csTradeItem);
                    interaction.editReply("Please know that the skin has been added to the watchlist. Hopefully.");
                }
                else {
                    interaction.editReply("The skin was not added to the database, as it could not be found on the site. Try finding it on the site, right click it, and copy and paste. I cannot stress enough: It should look something like: ★ Moto Gloves | Cool Mint (Minimal Wear)");
                }
                
            }
            else {
                interaction.editReply("Please know that something went wrong.");
            }
        }

        if(interaction.commandName === "createmultisearch") {
            await interaction.deferReply();
            let replyText = "Please know the results of your attempt to create new searches - ";
            let website = interaction.options.getString("website") || "csdeals";
            let skinName = interaction.options.getString("skinname") || "placeholder";
            let maxPriceCsDeals = interaction.options.getNumber("maxpricecsdeals") || -1;
            let maxPriceCsTrade = interaction.options.getNumber("maxpricecstrade") || -1;
            let maxPriceTradeIt = interaction.options.getNumber("maxpricetradeit") || -1;
            let maxPriceLootFarm = interaction.options.getNumber("maxpricelootfarm") || -1;
            let maxFloat = interaction.options.getNumber("maxfloat") || -1;
            let minFloat = interaction.options.getNumber("minfloat") || 0;
            //console.log(`${website}` + `${skinName}` + maxPrice + maxFloat + minFloat);
            if(minFloat > maxFloat) {
                interaction.editReply("I cannot stress enough: The minimum float cannot be higher than the maximum float value.");
            }
            else if(maxFloat <= 0 || maxFloat >= 1) {
                interaction.editReply("I cannot stress enough: The maximum float must be between 0 and 1.");
            }
            else if(minFloat < 0 || minFloat >= 1) {
                interaction.editReply("I cannot stress enough: The minimum float must be between 0 and 1.");
            }
            else { //Attempt creating new searches
                interaction.editReply(replyText);
                if(maxPriceCsDeals > 0) {
                    const csDeals = new CsDeals(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                    if(await csDeals.skinExists(skinName)) {
                        const csDealsItem = new CsDealsItem({
                            name: skinName,
                            maxPrice: maxPriceCsDeals,
                            minFloat: minFloat,
                            maxFloat: maxFloat
                        })
                        csDealsItem.save(err => console.log(err));
                        console.log(csDealsItem);
    
                        replyText = replyText.concat("Cs.Deals: Successful, ");
                    }
                    else {
                        replyText = replyText.concat("Cs.Deals: Skin not found on site, ");
                    }
                    interaction.editReply(replyText);
                }
                if(maxPriceCsTrade > 0) {
                    const csTrade = new CsTrade(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                    if(await csTrade.skinExists(skinName)) {
                        const csTradeItem = new CsTradeItem({
                            name: skinName,
                            maxPrice: maxPriceCsTrade,
                            minFloat: minFloat,
                            maxFloat: maxFloat
                        })
                        csTradeItem.save(err => console.log(err));
                        console.log(csTradeItem);
    
                        replyText = replyText.concat("Cs.Trade: Successful, ");
                    }
                    else {
                        replyText = replyText.concat("Cs.Trade: Skin not found on site, ");
                    }
                    interaction.editReply(replyText);
                }
                if(maxPriceTradeIt > 0) {
                    const tradeIt = new TradeIt(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                    if(await tradeIt.skinExists(skinName)) {
                        const tradeItItem = new TradeItItem({
                            name: skinName,
                            maxPrice: maxPriceTradeIt,
                            minFloat: minFloat,
                            maxFloat: maxFloat
                        })
                        tradeItItem.save(err => console.log(err));
                        console.log(tradeItItem);
    
                        replyText = replyText.concat("Tradeit.gg: Successful, ");
                    }
                    else {
                        replyText = replyText.concat("Tradeit.gg: Skin not found on site (This site's a bit fickle, consider retrying), ");
                    }
                    interaction.editReply(replyText);
                }
                if(maxPriceLootFarm > 0) {
                    const lootFarm = new LootFarm(client, `${process.env.CS_CHANNEL_ID}`, `${process.env.CS_ROLE_ID}`);
                    if(await lootFarm.skinExists(skinName)) {
                        const lootFarmItem = new LootFarmItem({
                            name: skinName,
                            maxPrice: maxPriceLootFarm,
                            minFloat: minFloat,
                            maxFloat: maxFloat
                        })
                        lootFarmItem.save(err => console.log(err));
                        console.log(lootFarmItem);
    
                        replyText = replyText.concat("Loot.farm: Successful, ");
                    }
                    else {
                        replyText = replyText.concat("Loot.farm: Skin not found on site, ");
                    }
                    interaction.editReply(replyText);
                }
            }
        }
    }
    catch(error) {
        console.log(error);
    }
});

client.login(process.env.DISCORD_TOKEN);