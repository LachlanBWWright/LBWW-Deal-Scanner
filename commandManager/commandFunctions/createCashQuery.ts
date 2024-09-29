import { ChatInputCommandInteraction } from "discord.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let search = new URL(query);
  if (search.toString().includes("https://www.cashconverters.com.au/")) {
    await db.cashConverters.create({ data: { url: search.toString() } });
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created:  + ${search.toString()}`,
    );
  } else
    interaction.editReply(`${getFailurePrelude()} your search was invalid!`);
}
