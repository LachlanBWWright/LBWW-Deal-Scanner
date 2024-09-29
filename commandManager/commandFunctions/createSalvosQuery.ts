import { ChatInputCommandInteraction } from "discord.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";
import { db } from "../../globals/PrismaClient.js";

export default async function (interaction: ChatInputCommandInteraction) {
  try {
    const query = interaction.options.getString("query");
    const minPrice = interaction.options.getNumber("minPrice");
    const maxPrice = interaction.options.getNumber("maxprice");
    if (!query || minPrice == null || maxPrice == null) {
      throw new Error("Invalid query paramaters.");
    }
    if (URL.canParse(query)) {
      throw new Error(
        "Do not enter a URL, set the content to what you would enter in the search box.",
      );
    }

    await db.salvos.create({
      data: {
        name: query.toString(),
        minPrice,
        maxPrice,
      },
    });

    await interaction.editReply(
      `${getResponsePrelude()} the search has been created: https://www.salvosstores.com.au/search?search=${encodeURIComponent(
        query,
      )}`,
    );
  } catch (e) {
    if (e instanceof Error) {
      interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\n${e.message}`,
      );
    } else {
      interaction.editReply(
        `${getFailurePrelude()} your search was invalid! \n\n${e}`,
      );
    }
  }
}
