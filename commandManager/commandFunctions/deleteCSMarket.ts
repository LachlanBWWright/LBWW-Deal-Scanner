import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function deleteCSMarket(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  try {
    const existing = await db.csMarket.findUnique({ where: { url: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved CS Market query found with that id.`,
      );
      return;
    }

    await db.csMarket.delete({ where: { url: id } });
    await interaction.editReply(
      `${getResponsePrelude()} deleted CS Market query: ${id}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while deleting.`,
    );
  }
}
