import { ChatInputCommandInteraction } from "discord.js";
import { createCs } from "../../scanners/siteScanners/steamMarket.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxFloat = interaction.options.getNumber("maxfloat") || 1;
  const maxPrice = interaction.options.getNumber("maxprice") || 1;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;

  const responseResult = await resultAsync(
    () => createCs(query, maxPrice, maxFloat, dmOnly),
    "Failed to create CS Market query",
  );
  if (responseResult.isErr()) {
    await interaction.editReply(`${getFailurePrelude()} the url was invalid!`);
    return;
  }

  const response = responseResult.value;
  if (response !== "") {
    await interaction.editReply(
      `${getResponsePrelude()} a search has been created${dmOnly ? " (DM only)" : ""} with the URL: ${response}`,
    );
    return;
  }

  await interaction.editReply(`${getFailurePrelude()} the url was invalid!`);
}
