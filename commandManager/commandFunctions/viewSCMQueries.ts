import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function viewSCMQueries(
  interaction: ChatInputCommandInteraction,
) {
  try {
    const results = await db.steamMarket.findMany();
    if (!results || results.length === 0) {
      await interaction.editReply(
        `${getResponsePrelude()} there are no saved SCM queries.`,
      );
      return;
    }

    const list = results
      .map(
        (r) =>
          `- ${r.name} (display ${r.displayUrl} maxPrice ${r.maxPrice} lastPrice ${r.lastPrice})`,
      )
      .join("\n");
    await interaction.editReply(
      `${getResponsePrelude()} saved SCM queries:\n${list}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while fetching queries: ${err}`,
    );
  }
}
