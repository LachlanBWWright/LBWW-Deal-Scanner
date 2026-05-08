import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";

export default async function viewMultiSearchQueries(
  interaction: ChatInputCommandInteraction,
) {
  const resultsResult = await resultAsync(
    () => db.csTradeBot.findMany(),
    "Failed to fetch saved queries",
  );
  if (resultsResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while fetching queries: ${resultsResult.error.message}`,
    );
    return;
  }

  const results = resultsResult.value;
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
}
