import { ChatInputCommandInteraction } from "discord.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import { err, ok } from "neverthrow";

export default async function (interaction: ChatInputCommandInteraction) {
  try {
    const query = interaction.options.getString("query");
    const minPrice = interaction.options.getNumber("minprice") ?? 0;
    const maxPrice = interaction.options.getNumber("maxprice") ?? 99999;
    const dmOnly = interaction.options.getBoolean("dmonly") ?? false;
    const validationResult = validateSalvosQuery(query, minPrice, maxPrice);
    const validationErrorMessage = validationResult.match(
      () => null,
      (error) => error.message,
    );
    if (validationErrorMessage) {
      await interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\n${validationErrorMessage}`,
      );
      return;
    }
    const validatedQuery = validationResult.match(
      (value) => value,
      () => "",
    );

    if (URL.canParse(validatedQuery)) {
      await interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\nDo not enter a URL, set the content to what you would enter in the search box.`,
      );
      return;
    }

    // Create Query first, then link it to Salvos
    await db.query.create({
      data: {
        dmOnly,
        salvos: {
          create: {
            name: validatedQuery.toString(),
            minPrice,
            maxPrice,
          },
        },
      },
    });

    await interaction.editReply(
      `${getResponsePrelude()} the search has been created${dmOnly ? " (DM only)" : ""}: https://www.salvosstores.com.au/search?search=${encodeURIComponent(
        validatedQuery,
      )}`,
    );
  } catch (e) {
    if (e instanceof Error) {
      await interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\n${e.message}`,
      );
    } else {
      await interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\n${e}`,
      );
    }
  }
}

function validateSalvosQuery(
  query: string | null,
  minPrice: number | null,
  maxPrice: number | null,
) {
  if (!query || minPrice == null || maxPrice == null) {
    return err(new Error("Invalid query parameters."));
  }
  return ok(query);
}
