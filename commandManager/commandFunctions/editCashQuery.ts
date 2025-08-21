import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

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

  try {
    const url = new URL(query).toString();
    const existing = await db.cashConverters.findUnique({ where: { url: id } });

    // In the CashConverters table the unique key is `url`. We allow editing by URL id.
    if (!existing) {
      await interaction.editReply(
        `${getFailurePrelude()} no saved query found with that id.`,
      );
      return;
    }

    await db.cashConverters.update({ where: { url: id }, data: { url } });
    await interaction.editReply(
      `${getResponsePrelude()} updated query to: ${url}`,
    );
  } catch (err) {
    await interaction.editReply(
      `${getFailurePrelude()} invalid URL or database error.`,
    );
  }
}
