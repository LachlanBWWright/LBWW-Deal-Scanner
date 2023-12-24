import { ChatInputCommandInteraction } from "discord.js";
import mongoose from "mongoose";
import CsDealsItem from "../../../schema/csDealsItem.js";
import CsTradeItem from "../../../schema/csTradeItem.js";
import TradeItItem from "../../../schema/tradeItItem.js";
import LootFarmItem from "../../../schema/lootFarmItem.js";
import SteamQuery from "../../../schema/steamQuery.js";
import CsMarketItem from "../../../schema/csMarketItem.js";
import CashConvertersQuery from "../../../schema/cashConvertersQuery.js";
import EbayQuery from "../../../schema/ebayQuery.js";
import SalvosQuery from "../../../schema/salvosQuery.js";
import GumtreeQuery from "../../../schema/gumtreeQuery.js";

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
    await interaction.editReply("No search name found. That's not acceptable.");
    return;
  }
}
