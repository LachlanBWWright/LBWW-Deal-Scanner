import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function viewSalvosQueries(
  interaction: ChatInputCommandInteraction,
) {
  const resultsResult = await fromThrowableAsync(
    () => db.salvos.findMany(),
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
      `${getResponsePrelude()} there are no saved Salvos queries.`,
    );
    return;
  }

  const list = results
    .map(
      (r) =>
        `- ${r.name} (min ${r.minPrice ?? "n/a"} max ${r.maxPrice ?? "n/a"})`,
    )
    .join("\n");
  await interaction.editReply(
    `${getResponsePrelude()} saved Salvos queries:\n${list}`,
  );
}
