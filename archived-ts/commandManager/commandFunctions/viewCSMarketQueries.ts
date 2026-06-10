import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";

export default async function viewCSMarketQueries(
  interaction: ChatInputCommandInteraction,
) {
  const resultsResult = await resultAsync(
    () => db.csMarket.findMany(),
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
}
