import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import {
  fromThrowableAsync,
  fromThrowableSync,
} from "../../functions/neverthrowUtils.js";

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

  const urlResult = fromThrowableSync(
    () => new URL(query).toString(),
    "Invalid URL",
  );
  if (urlResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error: ${urlResult.error.message}`,
    );
    return;
  }

  const url = urlResult.value;
  const existingResult = await fromThrowableAsync(
    () => db.ebay.findUnique({ where: { url: id } }),
    "Failed to load existing query",
  );
  if (existingResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error: ${existingResult.error.message}`,
    );
    return;
  }
  if (!existingResult.value) {
    await interaction.editReply(
      `${getFailurePrelude()} no saved eBay query found with that id.`,
    );
    return;
  }

  const updateResult = await fromThrowableAsync(
    () => db.ebay.update({ where: { url: id }, data: { url, maxPrice } }),
    "Failed to update query",
  );
  if (updateResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error: ${updateResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} updated eBay query to ${url} with maxPrice ${maxPrice}`,
  );
}
