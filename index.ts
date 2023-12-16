import {
  Client,
  GuildMember,
  GatewayIntentBits,
  REST,
  Routes,
  Events,
} from "discord.js";

import Dotenv from "dotenv";
import mongoose from "mongoose";
import puppeteer from "puppeteer";

//Schema Imports
import CsDealsItem from "./schema/csDealsItem.js";
import CsTradeItem from "./schema/csTradeItem.js";
import TradeItItem from "./schema/tradeItItem.js";
import LootFarmItem from "./schema/lootFarmItem.js";
import CashConvertersQuery from "./schema/cashConvertersQuery.js";
import SalvosQuery from "./schema/salvosQuery.js";
import EbayQuery from "./schema/ebayQuery.js";
import GumtreeQuery from "./schema/gumtreeQuery.js";
import SteamQuery from "./schema/steamQuery.js";
import CsMarketItem from "./schema/csMarketItem.js";

//Scanner Imports
//TODO: These should be removed
import CsDeals from "./scanners/siteScanners/csDeals.js";
import CsTrade from "./scanners/siteScanners/csTrade.js";
import TradeIt from "./scanners/siteScanners/tradeIt.js";
import LootFarm from "./scanners/siteScanners/lootFarm.js";
import SteamMarket from "./scanners/siteScanners/steamMarket.js";

import runScan from "./scanners/index.js";

import { commandList } from "./commandManager/index.js";

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

//Discord Client Setup
const client = new Client({ intents: [GatewayIntentBits.Guilds] });
client.once("ready", runScan);

//Runs upon a user creating a command
client.on(Events.InteractionCreate, async (interaction) => {
  if (!interaction.isChatInputCommand()) return; //Cancels if not a command
  try {
    await interaction.deferReply(); //Creates the loading '...'

    let roleFound = false;
    let member = interaction.member;
    member = <GuildMember>member;
    member.roles.cache.map((role) => {
      if (role.id == process.env.COMMAND_PERMISSION_ROLE_ID) {
        roleFound = true;
      }
    });
    if (!roleFound) {
      await interaction.editReply(
        "Please know you don't have the role needed to make commands. That's not acceptable."
      );
      return;
    }

    if (interaction.commandName === "createmultisearch") {
      let replyText =
        "Please know the results of your attempt to create new searches - ";
      let skinName = interaction.options.getString("skinname") ?? "placeholder";
      let maxPriceCsDeals =
        interaction.options.getNumber("maxpricecsdeals") ?? -1;
      let maxPriceCsTrade =
        interaction.options.getNumber("maxpricecstrade") ?? -1;
      let maxPriceTradeIt =
        interaction.options.getNumber("maxpricetradeit") ?? -1;
      let maxPriceLootFarm =
        interaction.options.getNumber("maxpricelootfarm") ?? -1;
      let maxFloat = interaction.options.getNumber("maxfloat") ?? -1;
      let minFloat = interaction.options.getNumber("minfloat") ?? 0;

      if (minFloat > maxFloat)
        await interaction.editReply(
          "I cannot stress enough: The minimum float cannot be higher than the maximum float value."
        );
      else if (maxFloat <= 0 || maxFloat >= 1)
        await interaction.editReply(
          "I cannot stress enough: The maximum float must be between 0 and 1."
        );
      else if (minFloat < 0 || minFloat >= 1)
        await interaction.editReply(
          "I cannot stress enough: The minimum float must be between 0 and 1."
        );
      else {
        //Attempt creating new searches
        await interaction.editReply(replyText);
        if (maxPriceCsDeals > 0) {
          const csDeals = new CsDeals(
            client,
            `${process.env.CS_CHANNEL_ID}`,
            `${process.env.CS_ROLE_ID}`
          );
          if (await csDeals.skinExists(skinName)) {
            const csDealsItem = new CsDealsItem({
              name: skinName,
              maxPrice: maxPriceCsDeals,
              minFloat: minFloat,
              maxFloat: maxFloat,
            });
            csDealsItem.save((err) => console.error(err));

            replyText = replyText.concat("Cs.Deals: Successful, ");
          } else {
            replyText = replyText.concat("Cs.Deals: Skin not found on site, ");
          }
          await interaction.editReply(replyText);
        }
        if (maxPriceCsTrade > 0) {
          const csTrade = new CsTrade(
            client,
            `${process.env.CS_CHANNEL_ID}`,
            `${process.env.CS_ROLE_ID}`
          );
          if (await csTrade.skinExists(skinName)) {
            const csTradeItem = new CsTradeItem({
              name: skinName,
              maxPrice: maxPriceCsTrade,
              minFloat: minFloat,
              maxFloat: maxFloat,
            });
            csTradeItem.save((err) => console.error(err));

            replyText = replyText.concat("Cs.Trade: Successful, ");
          } else {
            replyText = replyText.concat("Cs.Trade: Skin not found on site, ");
          }
          await interaction.editReply(replyText);
        }
        if (maxPriceTradeIt > 0) {
          const tradeIt = new TradeIt(
            client,
            `${process.env.CS_CHANNEL_ID}`,
            `${process.env.CS_ROLE_ID}`
          );
          if (await tradeIt.skinExists(skinName)) {
            const tradeItItem = new TradeItItem({
              name: skinName,
              maxPrice: maxPriceTradeIt,
              minFloat: minFloat,
              maxFloat: maxFloat,
            });
            tradeItItem.save((err) => console.error(err));

            replyText = replyText.concat("Tradeit.gg: Successful, ");
          } else {
            replyText = replyText.concat(
              "Tradeit.gg: Skin not found on site (This site's a bit fickle, consider retrying), "
            );
          }
          await interaction.editReply(replyText);
        }
        if (maxPriceLootFarm > 0) {
          const lootFarm = new LootFarm(
            client,
            `${process.env.CS_CHANNEL_ID}`,
            `${process.env.CS_ROLE_ID}`
          );
          if (await lootFarm.skinExists(skinName)) {
            const lootFarmItem = new LootFarmItem({
              name: skinName,
              maxPrice: maxPriceLootFarm,
              minFloat: minFloat,
              maxFloat: maxFloat,
            });
            lootFarmItem.save((err) => console.error(err));

            replyText = replyText.concat("Loot.farm: Successful, ");
          } else {
            replyText = replyText.concat("Loot.farm: Skin not found on site, ");
          }
          await interaction.editReply(replyText);
        }
      }
    }
    // createscmquery  createcsmarket createcashquery createsalvosquery
    else if (interaction.commandName === "createscmquery") {
      let query = interaction.options.getString("query") || "placeholder";
      let maxPrice = interaction.options.getNumber("maxprice") || 1;

      try {
        const steamMarket = new SteamMarket(
          client,
          `${process.env.STEAM_QUERY_CHANNEL_ID}`,
          `${process.env.STEAM_QUERY_ROLE_ID}`,
          `${process.env.CS_MARKET_CHANNEL_ID}`,
          `${process.env.CS_MARKET_ROLE_ID}`
        );
        let response = await steamMarket.createQuery(query, maxPrice);
        if (response[0])
          await interaction.editReply(
            "Item added successfully! URL generated: " + response[1]
          );
        else await interaction.editReply("The URL is invalid!");
      } catch (e) {
        await interaction.editReply("The URL is invalid!");
      }
    } else if (interaction.commandName === "createcsmarket") {
      let query = interaction.options.getString("query") || "placeholder";
      let maxFloat = interaction.options.getNumber("maxfloat") || 1;
      let maxPrice = interaction.options.getNumber("maxprice") || 1;

      try {
        const steamMarket = new SteamMarket(
          client,
          `${process.env.STEAM_QUERY_CHANNEL_ID}`,
          `${process.env.STEAM_QUERY_ROLE_ID}`,
          `${process.env.CS_MARKET_CHANNEL_ID}`,
          `${process.env.CS_MARKET_ROLE_ID}`
        );
        let response = await steamMarket.createCs(query, maxPrice, maxFloat);
        if (response != "")
          await interaction.editReply(
            "Please know a search has been created with the URL: " + response
          );
        else
          await interaction.editReply("Please know that the url was invalid!");
      } catch (e) {
        await interaction.editReply("Please know that the url was invalid!");
      }
    } else if (interaction.commandName === "createcashquery") {
      let query = interaction.options.getString("query") || "placeholder";
      let search = new URL(query);
      if (search.toString().includes("https://www.cashconverters.com.au/")) {
        let cashQuery = new CashConvertersQuery({
          name: search.toString(),
        });
        cashQuery.save();
        await interaction.editReply(
          "Please know that the search has been created: " + search.toString()
        );
      } else interaction.editReply("Please know that your search was invalid!");
    } else if (interaction.commandName === "createsalvosquery") {
      let query = interaction.options.getString("query") || "placeholder";
      let search = new URL(query);
      if (
        search
          .toString()
          .includes("https://www.salvosstores.com.au/shop?search=")
      ) {
        let salvosQuery = new SalvosQuery({
          name: search.toString(),
        });
        salvosQuery.save();
        await interaction.editReply(
          "Please know that the search has been created: " + search.toString()
        );
      } else interaction.editReply("Please know that your search was invalid!");
    } else if (interaction.commandName === "createebayquery") {
      let query = interaction.options.getString("query") || "placeholder";
      let maxPrice = interaction.options.getNumber("maxprice") || 1000;
      let search = new URL(query);
      if (search.toString().includes("https://www.ebay.com.au/")) {
        let ebayQuery = new EbayQuery({
          name: search.toString(),
          maxPrice: maxPrice,
        });
        ebayQuery.save();
        await interaction.editReply(
          "Please know that the search has been created: " + search.toString()
        );
      } else interaction.editReply("Please know that your search was invalid!");
    } else if (interaction.commandName === "creategumtreequery") {
      let query = interaction.options.getString("query") || "placeholder";
      let maxPrice = interaction.options.getNumber("maxprice") || 1000;
      let search = new URL(query);
      if (search.toString().includes("https://www.gumtree.com.au/")) {
        let gumtreeQuery = new GumtreeQuery({
          name: search.toString(),
          maxPrice: maxPrice,
        });
        gumtreeQuery.save();
        await interaction.editReply(
          "Please know that the search has been created: " + search.toString()
        );
      } else interaction.editReply("Please know that your search was invalid!");
    } else if (interaction.commandName === "viewqueries") {
      let scanner =
        (await interaction.options.getString("whichscanner")) || "csdeals";
      let searchName = await interaction.options.getString("searchname");
      let model: mongoose.Model<any>;
      try {
        if (scanner === "csdeals") model = CsDealsItem;
        else if (scanner === "cstrade") model = CsTradeItem;
        else if (scanner === "tradeit") model = TradeItItem;
        else if (scanner === "lootfarm") model = LootFarmItem;
        else if (scanner === "scmquery") model = SteamQuery;
        else if (scanner === "csmarket") model = CsMarketItem;
        else if (scanner === "cashquery") model = CashConvertersQuery;
        else if (scanner === "ebayquery") model = EbayQuery;
        else if (scanner === "salvosquery") model = SalvosQuery;
        else if (scanner === "gumtreequery") model = GumtreeQuery;
        else {
          await interaction.editReply(
            "No scanner found. That's not acceptable. Aborting"
          );
          return;
        }

        if (!searchName) {
          //Find multiple items
          await interaction.editReply(
            "Returning all items for the chosen scanner: "
          );

          for await (const item of model.find()) {
            interaction.followUp(
              `Name: ${scanner === "scmquery" ? item.displayUrl : item.name}` +
                `\n${item.maxPrice ? `Max Price: ${item.maxPrice} ` : ""}` +
                `\n${item.minPrice ? `Mix Price: ${item.minPrice} ` : ""}` +
                `\n${item.maxFloat ? `Max Float: ${item.maxFloat} ` : ""}` +
                `\n${item.minFloat ? `Max Float: ${item.minFloat} ` : ""}` +
                `\n${
                  item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""
                }`
            );
          }
        } else {
          //Find an item by name
          let item: any;
          let isScmQuery = scanner === "scmquery";

          if (isScmQuery)
            item = await model.findOne({ displayUrl: searchName });
          else item = await model.findOne({ name: searchName });

          if (!item) {
            interaction.editReply(
              "Please know that there is no search for the item you entered. That's not acceptable."
            );
            return;
          }

          interaction.editReply(
            `Name: ${isScmQuery ? item.displayUrl : item.name}` +
              `\n${item.maxPrice ? `Max Price: ${item.maxPrice}` : ""}` +
              `\n${item.minPrice ? `Mix Price: ${item.minPrice}` : ""}` +
              `\n${item.maxFloat ? `Max Float: ${item.maxFloat}` : ""}` +
              `\n${item.minFloat ? `Max Float: ${item.minFloat}` : ""}` +
              `\n${item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""}`
          );
        }
      } catch (e) {
        console.error(e);
      }
    } else if (interaction.commandName === "deletequery") {
      let scanner = await interaction.options.getString("whichscanner");
      let searchName = await interaction.options.getString("searchname");
      let model: mongoose.Model<any>;

      try {
        if (scanner === "csdeals") model = CsDealsItem;
        else if (scanner === "cstrade") model = CsTradeItem;
        else if (scanner === "tradeit") model = TradeItItem;
        else if (scanner === "lootfarm") model = LootFarmItem;
        else if (scanner === "scmquery") model = SteamQuery;
        else if (scanner === "csmarket") model = CsMarketItem;
        else if (scanner === "cashquery") model = CashConvertersQuery;
        else if (scanner === "ebayquery") model = EbayQuery;
        else if (scanner === "salvosquery") model = SalvosQuery;
        else if (scanner === "gumtreequery") model = GumtreeQuery;
        else {
          await interaction.editReply(
            "No scanner found. That's not acceptable. Aborting."
          );
          return;
        }

        let item: any;
        let isScmQuery = scanner === "scmquery";

        if (isScmQuery)
          item = await model.findOneAndDelete({ displayUrl: searchName });
        else item = await model.findOneAndDelete({ name: searchName });

        if (!item) {
          interaction.editReply(
            "Please know that the item you are searching for has already been deleted. Or didn't exist in the first place. That's not acceptable."
          );
          return;
        }

        interaction.editReply(
          `The following item has been deleted \n Name: ${
            isScmQuery ? item.displayUrl : item.name
          }` +
            `\n${item.maxPrice ? `Max Price: ${item.maxPrice}\n` : ""}` +
            `\n${item.minPrice ? `Mix Price: ${item.minPrice}\n` : ""}` +
            `\n${item.maxFloat ? `Max Float: ${item.maxFloat}\n` : ""}` +
            `\n${item.minFloat ? `Max Float: ${item.minFloat}\n` : ""}` +
            `\n${item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""}`
        );
      } catch (e) {
        console.error(e);
      }

      if (!searchName) {
        await interaction.editReply(
          "No search name found. That's not acceptable."
        );
        return;
      }
    }
  } catch (err) {
    console.error(err);
    await interaction.editReply(
      `Please know that an error has occurred: ${err}`
    );
  }
});

client.login(process.env.DISCORD_TOKEN);
