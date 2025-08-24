import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function deleteSCMQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  if (!id) {
    await interaction.editReply(`${getFailurePrelude()} missing id option.`);
    return;
  }

  try {
    const existing = await db.steamMarket.findUnique({ where: { name: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved SCM query found with that id.`,
      );
      return;
    }

    await db.steamMarket.delete({ where: { name: id } });
    await interaction.editReply(
      `${getResponsePrelude()} deleted SCM query: ${id}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while deleting: ${err}`,
    );
  }
}
