import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function viewEbayQueries(
  interaction: ChatInputCommandInteraction,
) {
  try {
    const results = await db.ebay.findMany();
    if (!results || results.length === 0) {
      await interaction.editReply(
        `${getResponsePrelude()} there are no saved eBay queries.`,
      );
      return;
    }

    const list = results
      .map((r) => `- ${r.url} (max ${r.maxPrice})`)
      .join("\n");
    await interaction.editReply(
      `${getResponsePrelude()} saved eBay queries:\n${list}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while fetching queries.`,
    );
  }
}
