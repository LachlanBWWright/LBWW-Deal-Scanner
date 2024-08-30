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
        `${getFailurePrelude()} no scanner was found. Aborting.`,
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
        `${getFailurePrelude()} the item you are searching for has already been deleted. Or didn't exist in the first place.`,
      );
      return;
    }

    interaction.editReply(
      `${getResponsePrelude()} the following item has been deleted \n Name: ${
        isScmQuery ? item.displayUrl : item.name
      }` +
        `\n${item.maxPrice ? `Max Price: ${item.maxPrice}\n` : ""}` +
        `\n${item.minPrice ? `Mix Price: ${item.minPrice}\n` : ""}` +
        `\n${item.maxFloat ? `Max Float: ${item.maxFloat}\n` : ""}` +
        `\n${item.minFloat ? `Max Float: ${item.minFloat}\n` : ""}` +
        `\n${item.maxDistance ? `Max Distance: ${item.maxDistance}` : ""}`,
    );
  } catch (e) {
    console.error(e);
  }

  if (!searchName) {
    await interaction.editReply(
      `${getFailurePrelude()} no search name was found.`,
    );
    return;
  }
}
