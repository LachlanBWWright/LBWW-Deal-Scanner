import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function editSalvosQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  const query = interaction.options.getString("query");
  const minPrice = interaction.options.getNumber("minprice");
  const maxPrice = interaction.options.getNumber("maxprice");

  if (!id || !query) {
    await interaction.editReply(
      `${getFailurePrelude()} missing id or query option.`,
    );
    return;
  }

  const existingResult = await fromThrowableAsync(
    () => db.salvos.findUnique({ where: { name: id } }),
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
      `${getFailurePrelude()} no saved Salvos query found with that id.`,
    );
    return;
  }

  const data: Partial<{ name: string; minPrice: number; maxPrice: number }> = {
    name: query,
  };
  if (typeof minPrice === "number") data.minPrice = minPrice;
  if (typeof maxPrice === "number") data.maxPrice = maxPrice;

  const updateResult = await fromThrowableAsync(
    () => db.salvos.update({ where: { name: id }, data }),
    "Failed to update query",
  );
  if (updateResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} database error while updating: ${updateResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} updated Salvos query to ${query}`,
  );
}
