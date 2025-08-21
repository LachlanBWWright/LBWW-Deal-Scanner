import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function viewCashQueries(
  interaction: ChatInputCommandInteraction,
) {
  try {
    const results = await db.cashConverters.findMany();
    if (!results || results.length === 0) {
      await interaction.editReply(
        `${getResponsePrelude()} there are no saved Cash Converters queries.`,
      );
      return;
    }

    const list = results.map((r) => `- ${r.url}`).join("\n");
    await interaction.editReply(
      `${getResponsePrelude()} saved Cash Converters queries:\n${list}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while fetching queries.`,
    );
  }
}
