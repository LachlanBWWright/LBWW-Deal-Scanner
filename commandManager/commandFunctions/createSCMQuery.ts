import { ChatInputCommandInteraction } from "discord.js";
import { createQuery } from "../../scanners/siteScanners/steamMarket.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1;

  try {
    let response = await createQuery(query, maxPrice);
    if (response[0])
      await interaction.editReply(
        "Item added successfully! URL generated: " + response[1]
      );
    else await interaction.editReply("The URL is invalid!");
  } catch (e) {
    await interaction.editReply("The URL is invalid!");
  }
}
