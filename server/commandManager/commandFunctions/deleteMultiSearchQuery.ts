import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function deleteMultiSearchQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  const existingResult = await fromThrowableAsync(
    () => db.csTradeBot.findUnique({ where: { name: id } }),
    "Failed to load existing query",
  );
  if (existingResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting: ${existingResult.error.message}`,
    );
    return;
  }

  if (!existingResult.value) {
    await interaction.editReply(
      `${getFailurePrelude()} no saved multi-search query found with that id.`,
    );
    return;
  }

  const deleteResult = await fromThrowableAsync(
    () => db.csTradeBot.delete({ where: { name: id } }),
    "Failed to delete query",
  );
  if (deleteResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting: ${deleteResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} deleted multi-search query: ${id}`,
  );
}
