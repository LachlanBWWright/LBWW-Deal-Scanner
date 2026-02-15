import { ChatInputCommandInteraction } from "discord.js";
import { createCs } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxFloat = interaction.options.getNumber("maxfloat") || 1;
  const maxPrice = interaction.options.getNumber("maxprice") || 1;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;

  try {
    const response = await createCs(query, maxPrice, maxFloat, dmOnly);
    if (response != "")
      await interaction.editReply(
        `${getResponsePrelude()} a search has been created${dmOnly ? " (DM only)" : ""} with the URL: ${response}`,
      );
    else
      await interaction.editReply(
        `${getFailurePrelude()} the url was invalid!`,
      );
  } catch {
    await interaction.editReply(`${getFailurePrelude()} the url was invalid!`);
  }
}
