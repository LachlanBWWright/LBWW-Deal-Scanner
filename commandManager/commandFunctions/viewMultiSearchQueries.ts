import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function viewMultiSearchQueries(
  interaction: ChatInputCommandInteraction,
) {
  try {
    const results = await db.csTradeBot.findMany();
    if (!results || results.length === 0) {
      await interaction.editReply(
        `${getResponsePrelude()} there are no saved multi-search queries.`,
      );
      return;
    }

    const list = results
      .map(
        (r) =>
          `- ${r.name} (min ${r.minFloat} max ${r.maxFloat} maxPrice ${r.maxPrice})`,
      )
      .join("\n");
    await interaction.editReply(
      `${getResponsePrelude()} saved multi-search queries:\n${list}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while fetching queries: ${err}`,
    );
  }
}
