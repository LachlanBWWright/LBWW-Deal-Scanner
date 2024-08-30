import { ChatInputCommandInteraction } from "discord.js";
import mongoose from "mongoose";
import CsDealsItem from "../../mongoSchema/csDealsItem.js";
import CsTradeItem from "../../mongoSchema/csTradeItem.js";
import TradeItItem from "../../mongoSchema/tradeItItem.js";
import LootFarmItem from "../../mongoSchema/lootFarmItem.js";
import SteamQuery from "../../mongoSchema/steamQuery.js";
import CsMarketItem from "../../mongoSchema/csMarketItem.js";
import CashConvertersQuery from "../../mongoSchema/cashConvertersQuery.js";
import EbayQuery from "../../mongoSchema/ebayQuery.js";
import SalvosQuery from "../../mongoSchema/salvosQuery.js";
import GumtreeQuery from "../../mongoSchema/gumtreeQuery.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
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
        `${getFailurePrelude()} no scanner was found. Aborting.`,
      );
      return;
    }

    if (!searchName) {
      //Find multiple items
      await interaction.editReply(
        `${getResponsePrelude()} returning all items for the chosen scanner: `,
      );

      for await (const item of model.find()) {
        interaction.followUp(
          `Name: ${scanner === "scmquery" ? item.displayUrl : item.name}` +
            `\n${item.maxPrice ? `Max Price: ${item.maxPrice} ` : ""}` +
            `\n${item.minPrice ? `Mix Price: ${item.minPrice} ` : ""}` +
            `\n${item.maxFloat ? `Max Float: ${item.maxFloat} ` : ""}` +
            `\n${item.minFloat ? `Max Float: ${item.minFloat} ` : ""}` +
            `\n${item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""}`,
        );
      }
    } else {
      //Find an item by name
      let item: any;
      let isScmQuery = scanner === "scmquery";

      if (isScmQuery) item = await model.findOne({ displayUrl: searchName });
      else item = await model.findOne({ name: searchName });

      if (!item) {
        interaction.editReply(
          `${getFailurePrelude()} there is no search for the item you entered.`,
        );
        return;
      }

      interaction.editReply(
        `Name: ${isScmQuery ? item.displayUrl : item.name}` +
          `\n${item.maxPrice ? `Max Price: ${item.maxPrice}` : ""}` +
          `\n${item.minPrice ? `Mix Price: ${item.minPrice}` : ""}` +
          `\n${item.maxFloat ? `Max Float: ${item.maxFloat}` : ""}` +
          `\n${item.minFloat ? `Max Float: ${item.minFloat}` : ""}` +
          `\n${item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""}`,
      );
    }
  } catch (e) {
    console.error(e);
  }
}
