import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function editSCMQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  const query = interaction.options.getString("query");
  const maxPrice = interaction.options.getNumber("maxprice");

  if (!id || !query || typeof maxPrice !== "number") {
    await interaction.editReply(
      `${getFailurePrelude()} missing id, query or maxprice option.`,
    );
    return;
  }

  try {
    const url = new URL(query).toString();
    const existing = await db.steamMarket.findUnique({ where: { name: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved SCM query found with that id.`,
      );
      return;
    }

    await db.steamMarket.update({
      where: { name: id },
      data: { name: url, displayUrl: query, maxPrice },
    });
    await interaction.editReply(
      `${getResponsePrelude()} updated SCM query to ${url}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} invalid URL or database error: ${err}`,
    );
  }
}
