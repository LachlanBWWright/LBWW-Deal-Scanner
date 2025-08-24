import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function editCSMarket(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  const query = interaction.options.getString("query");
  const maxFloat = interaction.options.getNumber("maxfloat");
  const maxPrice = interaction.options.getNumber("maxprice");

  if (
    !id ||
    !query ||
    typeof maxFloat !== "number" ||
    typeof maxPrice !== "number"
  ) {
    await interaction.editReply(
      `${getFailurePrelude()} missing id, query, maxfloat or maxprice option.`,
    );
    return;
  }

  try {
    const url = new URL(query).toString();
    const existing = await db.csMarket.findUnique({ where: { url: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved CS Market query found with that id.`,
      );
      return;
    }

    await db.csMarket.update({
      where: { url: id },
      data: { url, displayUrl: url, maxFloat, maxPrice },
    });
    await interaction.editReply(
      `${getResponsePrelude()} updated CS Market query to ${url}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} invalid URL or database error: ${err}`,
    );
  }
}
