import { ChatInputCommandInteraction } from "discord.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";
import { err, ok } from "neverthrow";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";

export default async function (interaction: ChatInputCommandInteraction) {
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

  const createResult = await fromThrowableAsync(
    () =>
      db.query.create({
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
      }),
    "Failed to create Salvos query",
  );

  if (createResult.isErr()) {
    await interaction.editReply(
      `${getFailurePrelude()} your search was invalid! \n\n${createResult.error.message}`,
    );
    return;
  }

  await interaction.editReply(
    `${getResponsePrelude()} the search has been created${dmOnly ? " (DM only)" : ""}: https://www.salvosstores.com.au/search?search=${encodeURIComponent(
      validatedQuery,
    )}`,
  );
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
