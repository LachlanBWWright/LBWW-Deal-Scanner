import { ChatInputCommandInteraction } from "discord.js";
import GumtreeQuery from "../../mongoSchema/gumtreeQuery.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1000;
  let search = new URL(query);
  if (search.toString().includes("https://www.gumtree.com.au/")) {
    let gumtreeQuery = new GumtreeQuery({
      name: search.toString(),
      maxPrice: maxPrice,
    });
    gumtreeQuery.save();
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created: ${search.toString()}`,
    );
  } else
    interaction.editReply(`${getFailurePrelude()} your search was invalid!`);
}
