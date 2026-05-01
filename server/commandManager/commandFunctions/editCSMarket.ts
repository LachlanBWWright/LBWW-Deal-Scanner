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
    () => db.csMarket.findUnique({ where: { url: id } }),
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
      `${getFailurePrelude()} no saved CS Market query found with that id.`,
    );
    return;
  }

  const updateResult = await fromThrowableAsync(
    () =>
      db.csMarket.update({
        where: { url: id },
        data: { url, displayUrl: url, maxFloat, maxPrice },
      }),
    "Failed to update query",
  );
  if (updateResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error: ${updateResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} updated CS Market query to ${url}`,
  );
}
