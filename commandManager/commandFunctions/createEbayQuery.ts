import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1000;
  let search = new URL(query);
  if (search.toString().includes("https://www.ebay.com.au/")) {
    await db.ebay.create({
      data: {
        url: search.toString(),
        maxPrice,
      },
    });
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created: ${search.toString()}`,
    );
  } else
    interaction.editReply(`${getFailurePrelude()} your search was invalid!`);
}
