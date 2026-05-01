import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function viewGumtreeQueries(
  interaction: ChatInputCommandInteraction,
) {
  const resultsResult = await fromThrowableAsync(
    () => db.gumtree.findMany(),
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
      `${getResponsePrelude()} there are no saved Gumtree queries.`,
    );
    return;
  }

  const list = results.map((r) => `- ${r.url} (max ${r.maxPrice})`).join("\n");
  await interaction.editReply(
    `${getResponsePrelude()} saved Gumtree queries:\n${list}`,
  );
}
