import { ChatInputCommandInteraction } from "discord.js";
import { createCs } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxFloat = interaction.options.getNumber("maxfloat") || 1;
  let maxPrice = interaction.options.getNumber("maxprice") || 1;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;

  try {
    let response = await createCs(query, maxPrice, maxFloat, dmOnly);
    if (response != "")
      await interaction.editReply(
        `${getResponsePrelude()} a search has been created${dmOnly ? " (DM only)" : ""} with the URL: ${response}`,
      );
    else
      await interaction.editReply(
        `${getFailurePrelude()} the url was invalid!`,
      );
  } catch (e) {
    await interaction.editReply(`${getFailurePrelude()} the url was invalid!`);
  }
}
