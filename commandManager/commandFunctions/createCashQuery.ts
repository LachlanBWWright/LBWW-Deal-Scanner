import { ChatInputCommandInteraction } from "discord.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const query = interaction.options.getString("query") || "placeholder";
  const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
  const search = new URL(query);
  if (search.toString().includes("https://www.cashconverters.com.au/")) {
    await db.query.create({
      data: {
        dmOnly,
        cashConverters: {
          create: { url: search.toString() },
        },
      },
    });
    await interaction.editReply(
      `${getResponsePrelude()} the search has been created${dmOnly ? " (DM only)" : ""}: ${search.toString()}`,
    );
  } else
    interaction.editReply(`${getFailurePrelude()} your search was invalid!`);
}
