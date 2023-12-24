import { ChatInputCommandInteraction } from "discord.js";
import { createCs } from "../../../scanners/siteScanners/steamMarket.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxFloat = interaction.options.getNumber("maxfloat") || 1;
  let maxPrice = interaction.options.getNumber("maxprice") || 1;

  try {
    let response = await createCs(query, maxPrice, maxFloat);
    if (response != "")
      await interaction.editReply(
        "Please know a search has been created with the URL: " + response
      );
    else await interaction.editReply("Please know that the url was invalid!");
  } catch (e) {
    await interaction.editReply("Please know that the url was invalid!");
  }
}
