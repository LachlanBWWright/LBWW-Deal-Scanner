import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function deleteCSMarket(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  const existingResult = await fromThrowableAsync(
    () => db.csMarket.findUnique({ where: { url: id } }),
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
      `${getFailurePrelude()} no saved CS Market query found with that id.`,
    );
    return;
  }

  const deleteResult = await fromThrowableAsync(
    () => db.csMarket.delete({ where: { url: id } }),
    "Failed to delete query",
  );
  if (deleteResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting: ${deleteResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} deleted CS Market query: ${id}`,
  );
}
