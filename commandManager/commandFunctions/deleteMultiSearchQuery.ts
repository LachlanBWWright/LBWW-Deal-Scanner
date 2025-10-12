import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function deleteMultiSearchQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  try {
    const existing = await db.csTradeBot.findUnique({ where: { name: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved multi-search query found with that id.`,
      );
      return;
    }

    await db.csTradeBot.delete({ where: { name: id } });
    await interaction.editReply(
      `${getResponsePrelude()} deleted multi-search query: ${id}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while deleting: ${err}`,
    );
  }
}
