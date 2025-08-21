import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function editEbayQuery(
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
    const existing = await db.ebay.findUnique({ where: { url: id } });
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved eBay query found with that id.`,
      );
      return;
    }

    await db.ebay.update({ where: { url: id }, data: { url, maxPrice } });
    await interaction.editReply(
      `${getResponsePrelude()} updated eBay query to ${url} with maxPrice ${maxPrice}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error.`,
    );
  }
}
