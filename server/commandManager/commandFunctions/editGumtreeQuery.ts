import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function editGumtreeQuery(
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
    const existing = await db.gumtree.findUnique({ where: { url: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved Gumtree query found with that id.`,
      );
      return;
    }

    await db.gumtree.update({ where: { url: id }, data: { url, maxPrice } });
    await interaction.editReply(
      `${getResponsePrelude()} updated Gumtree query to ${url} with maxPrice ${maxPrice}`,
    );
  } catch (err) {
    await interaction.editReply(
  `${getFailurePrelude()} invalid URL or database error: ${err}`,
    );
  }
}
