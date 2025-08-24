import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function viewCSMarketQueries(
  interaction: ChatInputCommandInteraction,
) {
  try {
    const results = await db.csMarket.findMany();
    if (!results || results.length === 0) {
      await interaction.editReply(
        `${getResponsePrelude()} there are no saved CS Market queries.`,
      );
      return;
    }

    const list = results
      .map((r) => `- ${r.url} (maxFloat ${r.maxFloat} maxPrice ${r.maxPrice})`)
      .join("\n");
    await interaction.editReply(
      `${getResponsePrelude()} saved CS Market queries:\n${list}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while fetching queries: ${err}`,
    );
  }
}
