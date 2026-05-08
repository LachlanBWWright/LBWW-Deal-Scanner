import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { resultAsync } from "../../functions/neverthrowUtils.js";

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

  const existingResult = await resultAsync(
    () => db.csTradeBot.findUnique({ where: { name: id } }),
    "Failed to load existing query",
  );
  if (existingResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while updating: ${existingResult.error.message}`,
    );
    return;
  }

  if (!existingResult.value) {
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

  const updateResult = await resultAsync(
    () => db.csTradeBot.update({ where: { name: id }, data }),
    "Failed to update query",
  );
  if (updateResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while updating: ${updateResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} updated multi-search query to ${query}`,
  );
}
