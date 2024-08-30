import { ChatInputCommandInteraction } from "discord.js";
import SalvosQuery from "../../mongoSchema/salvosQuery.js";
import {
  getFailurePrelude,
  getResponsePrelude,
} from "../../functions/messagePreludes.js";

export default async function (interaction: ChatInputCommandInteraction) {
  try {
    const query = interaction.options.getString("query");
    const minPrice = interaction.options.getNumber("minPrice");
    const maxPrice = interaction.options.getNumber("maxprice");
    if (!query) {
      throw new Error("Invalid query paramaters.");
    }
    if (URL.canParse(query)) {
      throw new Error(
        "Do not enter a URL, set the content to what you would enter in the search box.",
      );
    }

    const salvosQuery = new SalvosQuery({
      name: query.toString(),
      minPrice: minPrice,
      maxPrice: maxPrice,
    });
    await salvosQuery.save();
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
