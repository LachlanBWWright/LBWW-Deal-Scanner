import { ChatInputCommandInteraction, Client } from "discord.js";
import SteamMarket from "../../../scanners/siteScanners/steamMarket.js";

export default async function (
  client: Client,
  interaction: ChatInputCommandInteraction
) {
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
}
