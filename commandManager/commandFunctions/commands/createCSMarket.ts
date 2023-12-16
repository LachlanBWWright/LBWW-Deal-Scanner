import { Client, ChatInputCommandInteraction } from "discord.js";
import SteamMarket from "../../../scanners/siteScanners/steamMarket.js";

export default async function (
  client: Client,
  interaction: ChatInputCommandInteraction
) {
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
    else await interaction.editReply("Please know that the url was invalid!");
  } catch (e) {
    await interaction.editReply("Please know that the url was invalid!");
  }
}
