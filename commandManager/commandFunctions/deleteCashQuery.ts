import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function deleteCashQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  try {
    const existing = await db.cashConverters.findUnique({ where: { url: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved query found with that id.`,
      );
      return;
    }

    await db.cashConverters.delete({ where: { url: id } });
    await interaction.editReply(
      `${getResponsePrelude()} deleted saved query: ${id}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting.`,
    );
  }
}
