import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const maxPrice = interaction.options.getNumber("maxprice") || 1000;
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
  const search = new URL(query);
  if (search.toString().includes("https://www.gumtree.com.au/")) {
    await db.query.create({
      data: {
        dmOnly,
        gumtree: {
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
