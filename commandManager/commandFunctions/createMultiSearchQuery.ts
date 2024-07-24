import { Client, ChatInputCommandInteraction } from "discord.js";
import { csDealsSkinExists } from "../../scanners/siteScanners/csDeals.js";
import CsDealsItem from "../../schema/csDealsItem.js";
import { tradeItSkinExists } from "../../scanners/siteScanners/tradeIt.js";
import TradeItItem from "../../schema/tradeItItem.js";
import { csTradeSkinExists } from "../../scanners/siteScanners/csTrade.js";
import CsTradeItem from "../../schema/csTradeItem.js";
import { lootFarmSkinExists } from "../../scanners/siteScanners/lootFarm.js";
import LootFarmItem from "../../schema/lootFarmItem.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let replyText = `${getResponsePrelude()} the results of your attempt to create new searches - `;
  let skinName = interaction.options.getString("skinname") ?? "placeholder";
  let maxPriceCsDeals = interaction.options.getNumber("maxpricecsdeals") ?? -1;
  let maxPriceCsTrade = interaction.options.getNumber("maxpricecstrade") ?? -1;
  let maxPriceTradeIt = interaction.options.getNumber("maxpricetradeit") ?? -1;
  let maxPriceLootFarm =
    interaction.options.getNumber("maxpricelootfarm") ?? -1;
  let maxFloat = interaction.options.getNumber("maxfloat") ?? -1;
  let minFloat = interaction.options.getNumber("minfloat") ?? 0;

  if (minFloat > maxFloat)
    await interaction.editReply(
      `${getFailurePrelude()} the minimum float cannot be higher than the maximum float value.`,
    );
  else if (maxFloat <= 0 || maxFloat >= 1)
    await interaction.editReply(
      `${getFailurePrelude()} the maximum float must be between 0 and 1.`,
    );
  else if (minFloat < 0 || minFloat >= 1)
    await interaction.editReply(
      `${getFailurePrelude()} the minimum float must be between 0 and 1.`,
    );
  else {
    //Attempt creating new searches
    await interaction.editReply(replyText);
    if (maxPriceCsDeals > 0) {
      if (await csDealsSkinExists(skinName)) {
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
      if (await csTradeSkinExists(skinName)) {
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
      if (await tradeItSkinExists(skinName)) {
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
          "Tradeit.gg: Skin not found on site (This site's a bit fickle, consider retrying), ",
        );
      }
      await interaction.editReply(replyText);
    }
    if (maxPriceLootFarm > 0) {
      if (await lootFarmSkinExists(skinName)) {
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
