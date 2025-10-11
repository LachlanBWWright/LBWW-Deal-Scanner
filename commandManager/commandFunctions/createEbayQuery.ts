import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  let query = interaction.options.getString("query") || "placeholder";
  let maxPrice = interaction.options.getNumber("maxprice") || 1000;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
  let search = new URL(query);
  if (search.toString().includes("https://www.ebay.com.au/")) {
    await db.query.create({
      data: {
        dmOnly,
        ebay: {
          create: {
            url: search.toString(),
            maxPrice,
          },
        },
      },
    });
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created${dmOnly ? " (DM only)" : ""}: ${search.toString()}`,
    );
  } else
    interaction.editReply(`${getFailurePrelude()} your search was invalid!`);
}
