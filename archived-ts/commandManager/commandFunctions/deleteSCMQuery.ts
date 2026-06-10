import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";

export default async function deleteSCMQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  const existingResult = await resultAsync(
    () => db.steamMarket.findUnique({ where: { name: id } }),
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
      `${getFailurePrelude()} no saved SCM query found with that id.`,
    );
    return;
  }

  const deleteResult = await resultAsync(
    () => db.steamMarket.delete({ where: { name: id } }),
    "Failed to delete query",
  );
  if (deleteResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting: ${deleteResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} deleted SCM query: ${id}`,
  );
}
