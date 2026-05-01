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

export default async function editCashQuery(
  interaction: ChatInputCommandInteraction,
) {
  const id = interaction.options.getString("id");
  const query = interaction.options.getString("query");
  if (!id || !query) {
    await interaction.editReply(
      `${getFailurePrelude()} missing id or query option.`,
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
    () => db.cashConverters.findUnique({ where: { url: id } }),
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
      `${getFailurePrelude()} no saved query found with that id.`,
    );
    return;
  }

  const updateResult = await fromThrowableAsync(
    () => db.cashConverters.update({ where: { url: id }, data: { url } }),
    "Failed to update query",
  );
  if (updateResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error: ${updateResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} updated query to: ${url}`,
  );
}
