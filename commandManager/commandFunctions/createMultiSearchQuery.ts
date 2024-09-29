import { ChatInputCommandInteraction } from "discord.js";
import { db } from "../../globals/PrismaClient.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  const skinName = interaction.options.getString("skinname") ?? "placeholder";
  const maxPrice = interaction.options.getNumber("maxprice") ?? -1;
  const maxFloat = interaction.options.getNumber("maxfloat") ?? -1;
  const minFloat = interaction.options.getNumber("minfloat") ?? 0;

  if (minFloat > maxFloat)
    await interaction.editReply(
      `${getFailurePrelude()} the minimum float cannot be higher than the maximum float value.`,
    );
  else if (maxFloat <= 0 || maxFloat >= 1)
    await interaction.editReply(
      `${getFailurePrelude()} the maximum float must be between 0 and 1.`,
    );
  else if (minFloat < 0 || minFloat >= 1)
    await interaction.editReply(
      `${getFailurePrelude()} the minimum float must be between 0 and 1.`,
    );
  else if (maxPrice <= 0) {
    await interaction.editReply(
      `${getFailurePrelude} the price must be positive, and greater than $0.`,
    );
  } else {
    //Attempt creating new searches
    db.csTradeBot.create({
      data: {
        name: skinName,
        maxFloat,
        minFloat,
        maxPrice,
      },
    });
    await interaction.editReply(
      `${getResponsePrelude()} your search has been created.`,
    );
  }
}
