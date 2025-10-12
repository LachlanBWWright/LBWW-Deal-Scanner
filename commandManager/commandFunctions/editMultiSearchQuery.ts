import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function editMultiSearchQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  const query = interaction.options.getString("query");
  const minFloat = interaction.options.getNumber("minfloat");
  const maxFloat = interaction.options.getNumber("maxfloat");
  const maxPrice = interaction.options.getNumber("maxprice");

  if (
    !id ||
    !query ||
    typeof minFloat !== "number" ||
    typeof maxFloat !== "number"
  ) {
    await interaction.editReply(
      `${getFailurePrelude()} missing id, query, minfloat or maxfloat option.`,
    );
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

    const data: Partial<{
      name: string;
      maxPrice: number;
      minFloat: number;
      maxFloat: number;
    }> = { name: query };
    if (typeof maxPrice === "number") data.maxPrice = maxPrice;
    if (typeof minFloat === "number") data.minFloat = minFloat;
    if (typeof maxFloat === "number") data.maxFloat = maxFloat;

    await db.csTradeBot.update({ where: { name: id }, data });
    await interaction.editReply(
      `${getResponsePrelude()} updated multi-search query to ${query}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} database error while updating: ${err}`,
    );
  }
}
