import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function viewSCMQueries(
  interaction: ChatInputCommandInteraction,
) {
  const resultsResult = await fromThrowableAsync(
    () => db.steamMarket.findMany(),
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
}
