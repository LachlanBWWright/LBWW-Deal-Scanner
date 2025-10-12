import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function deleteSalvosQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  try {
    const existing = await db.salvos.findUnique({ where: { name: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved Salvos query found with that id.`,
      );
      return;
    }

    await db.salvos.delete({ where: { name: id } });
    await interaction.editReply(
      `${getResponsePrelude()} deleted Salvos query: ${id}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while deleting: ${err}`,
    );
  }
}
